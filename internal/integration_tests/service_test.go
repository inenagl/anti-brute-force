//go:build integration

package integration_tests

import (
	"context"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/inenagl/anti-brute-force/internal/api"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	maxLogins    int
	maxPasswords int
	maxIPs       int
)

type ABFSuite struct {
	suite.Suite
	client api.AntiBruteForceClient
	conn   *grpc.ClientConn
	db     *sqlx.DB
}

func (s *ABFSuite) SetupSuite() {
	grpcAddr := os.Getenv("GOABF_GRPCADDR")
	if grpcAddr == "" {
		grpcAddr = ":8889"
	}
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.client = api.NewAntiBruteForceClient(conn)
	s.conn = conn

	settings, err := newDbSettings()
	s.Require().NoError(err)

	db, err := newDB(settings)
	s.Require().NoError(err)
	s.db = db

	maxLogins, err = strconv.Atoi(os.Getenv("GOABF_MAIN_MAXLOGINS"))
	s.Require().NoError(err)
	maxPasswords, err = strconv.Atoi(os.Getenv("GOABF_MAIN_MAXPASSWORDS"))
	s.Require().NoError(err)
	maxIPs, err = strconv.Atoi(os.Getenv("GOABF_MAIN_MAXIPS"))
	s.Require().NoError(err)
}

func (s *ABFSuite) TearDownSuite() {
	err := s.conn.Close()
	s.Require().NoError(err)
	err = s.db.Close()
	s.Require().NoError(err)
}

func (s *ABFSuite) SetupTest() {
}

func (s *ABFSuite) TearDownTest() {
	_ = s.db.MustExec("TRUNCATE TABLE bw_lists")
}

// Формирует запрос для заданных логина, пароля, IP.
// Если передано нуль-значение, то вместо него генерирует случайное.
func authReq(login, passwd, ip string) *api.AuthRequest {
	l := faker.Username()
	p := faker.Password()
	i := faker.IPv4()
	if login != "" {
		l = login
	}
	if passwd != "" {
		p = passwd
	}
	if ip != "" {
		i = ip
	}
	return &api.AuthRequest{Login: l, Password: p, Ip: i}
}

// Проверка авторизации с учетом бакетов.
// Здесь же проверка сброса бакетов.
func (s *ABFSuite) TestAuthWithLimits() {
	// Функция позволяет тестировать вызовы для одного неизменного поля, а остальных меняющихся.
	// Чтобы проверить ограничения отдельно по логинам, паролям и IP.
	test := func(limit int, login, passwd, ip string) {
		ctx := context.Background()
		var i int
		var err error
		var resp *api.AuthResponse
		// Первые попытки будут успешные, пока не упрёмся в ограничение по логинам.
		for i = 0; i < limit; i++ {
			resp, err = s.client.Auth(ctx, authReq(login, passwd, ip))
			s.Require().NoError(err)
			s.Require().True(resp.Ok)
		}

		// Следующие попытки с этим логином всегда будут безуспешные, т.к. лимит исчерпан и не уменьшается.
		for i = 0; i < 100; i++ {
			resp, err = s.client.Auth(ctx, authReq(login, passwd, ip))
			s.Require().NoError(err)
			s.Require().False(resp.GetOk())
		}

		// Подождём, пока один слот в бакете освободится, и сделаем успешную попытку.
		time.Sleep(time.Second * 60 / time.Duration(limit))
		resp, err = s.client.Auth(ctx, authReq(login, passwd, ip))
		s.Require().NoError(err)
		s.Require().True(resp.GetOk())

		// Следующие попытки снова будут отклонены.
		resp, err = s.client.Auth(ctx, authReq(login, passwd, ip))
		s.Require().NoError(err)
		s.Require().False(resp.GetOk())

		// Сбросим бакеты и попробуем снова.
		_, err = s.client.Reset(ctx, &api.ResetRequest{Login: login, Password: passwd, Ip: ip})
		s.Require().NoError(err)
		// Новые попытки должны пройти успешно.
		for i = 0; i < limit; i++ {
			resp, err = s.client.Auth(ctx, authReq(login, passwd, ip))
			s.Require().NoError(err)
			s.Require().True(resp.GetOk())
		}
	}

	login := faker.Username()
	password := faker.Password()
	ip := faker.IPv4()

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		test(maxLogins, login, "", "")
	}()

	go func() {
		defer wg.Done()
		test(maxPasswords, "", password, "")
	}()

	go func() {
		defer wg.Done()
		test(maxIPs, "", "", ip)
	}()

	wg.Wait()
}

// Проверка авторизации через черно-белые списки без учета бакетов.
func (s *ABFSuite) TestAuthByBWLists() {
	ip := faker.IPv4()
	netw := net.IPNet{IP: net.ParseIP(ip), Mask: net.IPv4Mask(255, 255, 255, 0)}
	ctx := context.Background()

	var err error
	var resp *api.AuthResponse
	var i int

	// Добавляем сеть в белый список.
	_, err = s.client.AddToWhiteList(ctx, &api.IpNetRequest{Inet: netw.String()})
	s.Require().NoError(err)

	// Теперь все запросы с этого IP будут разрешены.
	for i = 0; i <= maxIPs*2; i++ {
		resp, err = s.client.Auth(ctx, authReq("", "", ip))
		s.Require().NoError(err)
		s.Require().True(resp.Ok)
	}

	// Добавляем сеть в черный список.
	_, err = s.client.RemoveFromWhiteList(ctx, &api.IpNetRequest{Inet: netw.String()})
	s.Require().NoError(err)
	_, err = s.client.AddToBlackList(ctx, &api.IpNetRequest{Inet: netw.String()})
	s.Require().NoError(err)
	// И сбрасываем бакет для верности.
	_, err = s.client.Reset(ctx, &api.ResetRequest{Login: "", Password: "", Ip: ip})
	s.Require().NoError(err)

	// Но все запросы все-равно будут отклоняться.
	for i = 0; i <= 10; i++ {
		resp, err = s.client.Auth(ctx, authReq("", "", ip))
		s.Require().NoError(err)
		s.Require().False(resp.Ok)
	}
}

// Проверка методов работы с черно-белыми списками.
func (s *ABFSuite) TestBWLists() {
	netw1 := net.IPNet{
		IP:   net.IPv4(125, 125, 1, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	netw2 := net.IPNet{
		IP:   net.IPv4(125, 125, 0, 0),
		Mask: net.IPv4Mask(255, 255, 0, 0),
	}
	netw3 := net.IPNet{
		IP:   net.IPv4(192, 168, 1, 0),
		Mask: net.IPv4Mask(255, 255, 255, 128),
	}
	ctx := context.Background()
	var err error

	// Добавим первую сеть в белый список.
	_, err = s.client.AddToWhiteList(ctx, &api.IpNetRequest{Inet: netw1.String()})
	s.Require().NoError(err)
	// Попробуем добавить ещё раз.
	_, err = s.client.AddToWhiteList(ctx, &api.IpNetRequest{Inet: netw1.String()})
	s.Require().Error(err)
	// А если её в черный список?
	_, err = s.client.AddToBlackList(ctx, &api.IpNetRequest{Inet: netw1.String()})
	s.Require().Error(err)
	// А если добавить попробовать вторую, которая пересекается?
	_, err = s.client.AddToWhiteList(ctx, &api.IpNetRequest{Inet: netw2.String()})
	s.Require().Error(err)
	// А если вторую в черный список?
	_, err = s.client.AddToBlackList(ctx, &api.IpNetRequest{Inet: netw2.String()})
	s.Require().Error(err)
	// С третьей должно получиться, т.к. нет пересечений.
	_, err = s.client.AddToBlackList(ctx, &api.IpNetRequest{Inet: netw3.String()})
	s.Require().NoError(err)

	// Удалим первую из белого списка.
	_, err = s.client.RemoveFromWhiteList(ctx, &api.IpNetRequest{Inet: netw1.String()})
	s.Require().NoError(err)
	// Теперь можно добавить вторую, например, в чёрный.
	_, err = s.client.AddToBlackList(ctx, &api.IpNetRequest{Inet: netw2.String()})
	s.Require().NoError(err)
	// Но нельзя первую.
	_, err = s.client.AddToWhiteList(ctx, &api.IpNetRequest{Inet: netw1.String()})
	s.Require().Error(err)
	_, err = s.client.AddToBlackList(ctx, &api.IpNetRequest{Inet: netw1.String()})
	s.Require().Error(err)

	// Удалим вторую и третью и проверим, что в базе ничего не осталось.
	_, err = s.client.RemoveFromBlackList(ctx, &api.IpNetRequest{Inet: netw2.String()})
	s.Require().NoError(err)
	_, err = s.client.RemoveFromBlackList(ctx, &api.IpNetRequest{Inet: netw3.String()})
	s.Require().NoError(err)

	var dest int
	s.db.Get(&dest, `SELECT COUNT(*) FROM bw_lists`)
	s.Require().Equal(0, dest)
}

func TestABFSuite(t *testing.T) {
	suite.Run(t, new(ABFSuite))
}
