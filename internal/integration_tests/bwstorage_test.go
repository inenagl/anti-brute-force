//go:build integration

package integration_tests

import (
	"database/sql"
	"errors"
	"net"
	"testing"

	"github.com/inenagl/anti-brute-force/internal/storage/bwlist"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type BWStorageSuite struct {
	suite.Suite
	storage *bwliststorage.Storage
	db      *sqlx.DB
}

func (s *BWStorageSuite) SetupSuite() {
	settings, err := newDbSettings()
	s.Require().NoError(err)

	db, err := newDB(settings)
	s.Require().NoError(err)
	s.db = db

	s.storage = bwliststorage.New(
		settings.host,
		settings.port,
		settings.dbname,
		settings.user,
		settings.password,
		settings.sslmode,
		settings.timeout,
	)
	err = s.storage.Connect()
	s.Require().NoError(err)
}

func (s *BWStorageSuite) TearDownSuite() {
	err := s.db.Close()
	s.Require().NoError(err)
	err = s.storage.Close()
	s.Require().NoError(err)
}

func (s *BWStorageSuite) SetupTest() {
}

func (s *BWStorageSuite) TearDownTest() {
	_ = s.db.MustExec("TRUNCATE TABLE bw_lists")
}

func (s *BWStorageSuite) isRecordExists(network net.IPNet, t bwliststorage.ListType) bool {
	var dest bool
	if err := s.db.Get(
		&dest,
		`SELECT true FROM bw_lists WHERE network = $1 AND type = $2`,
		network.String(),
		t,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		s.Fail(err.Error())
	}
	return true
}

func (s *BWStorageSuite) addRecord(network net.IPNet, t bwliststorage.ListType) {
	s.Require().False(s.isRecordExists(network, t))

	_, err := s.db.Exec(
		`INSERT INTO bw_lists (network, type) VALUES ($1, $2)`,
		network.String(),
		t,
	)
	s.Require().NoError(err)

	s.Require().True(s.isRecordExists(network, t))
}

func (s *BWStorageSuite) TestAdd() {
	n := net.IPNet{IP: net.IPv4(192, 168, 5, 0), Mask: net.IPv4Mask(255, 255, 255, 0)}
	t := bwliststorage.TypeWhite
	rec := bwliststorage.ListRecord{
		Network: n,
		Type:    t,
	}

	s.Require().False(s.isRecordExists(n, t))
	s.Require().NoError(s.storage.Add(rec))
	s.Require().True(s.isRecordExists(n, t))

	s.Require().Error(s.storage.Add(rec))
}

func (s *BWStorageSuite) TestRemove() {
	n1 := net.IPNet{IP: net.IPv4(192, 168, 5, 0), Mask: net.IPv4Mask(255, 255, 255, 0)}
	t1 := bwliststorage.TypeWhite
	rec1 := bwliststorage.ListRecord{
		Network: n1,
		Type:    t1,
	}
	n2 := net.IPNet{IP: net.IPv4(110, 255, 78, 0), Mask: net.IPv4Mask(255, 255, 255, 128)}
	t2 := bwliststorage.TypeWhite

	s.addRecord(n1, t1)
	s.addRecord(n2, t2)

	s.Require().NoError(s.storage.Remove(rec1))
	s.Require().False(s.isRecordExists(n1, t1))
	s.Require().True(s.isRecordExists(n2, t2))
}

func (s *BWStorageSuite) TestGetByIP() {
	n := net.IPNet{IP: net.IPv4(192, 192, 1, 0), Mask: net.IPv4Mask(255, 255, 255, 0)}
	t := bwliststorage.TypeBlack

	s.addRecord(n, t)

	res, err := s.storage.GetByIP(net.IPv4(192, 191, 1, 10))
	s.Require().NoError(err)
	s.Require().Nil(res)

	for _, v := range []net.IP{net.IPv4(192, 192, 1, 0), net.IPv4(192, 192, 1, 100), net.IPv4(192, 192, 1, 255)} {
		res, err = s.storage.GetByIP(v)
		s.Require().NoError(err)
		s.Require().Equal(n.String(), res.Network.String())
		s.Require().Equal(t, res.Type)
	}
}

func (s *BWStorageSuite) TestRemoveAll() {
	init := []bwliststorage.ListRecord{
		{
			Network: net.IPNet{IP: net.IPv4(1, 1, 1, 0), Mask: net.IPv4Mask(255, 255, 255, 0)},
			Type:    bwliststorage.TypeBlack,
		},
		{
			Network: net.IPNet{IP: net.IPv4(125, 1, 11, 0), Mask: net.IPv4Mask(255, 255, 255, 128)},
			Type:    bwliststorage.TypeBlack,
		},
		{
			Network: net.IPNet{IP: net.IPv4(254, 43, 0, 0), Mask: net.IPv4Mask(255, 255, 128, 0)},
			Type:    bwliststorage.TypeBlack,
		},
	}
	for _, rec := range init {
		s.addRecord(rec.Network, rec.Type)
	}

	s.Require().NoError(s.storage.RemoveAll())

	for _, rec := range init {
		s.Require().False(s.isRecordExists(rec.Network, rec.Type))
	}
}

func TestBWStorageSuite(t *testing.T) {
	suite.Run(t, new(BWStorageSuite))
}
