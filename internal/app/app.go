package app

import (
	"net"

	lrucache "github.com/inenagl/anti-brute-force/internal/cache"
	"github.com/inenagl/anti-brute-force/internal/ratelimit"
	bwliststorage "github.com/inenagl/anti-brute-force/internal/storage/bwlist"
	"go.uber.org/zap"
)

type Application interface {
	Auth(login, password string, ip net.IP) (bool, error)
	ResetBuckets(login, password string, ip net.IP)
	AddToBlackList(network net.IPNet) error
	AddToWhiteList(network net.IPNet) error
	RemoveFromBlackList(network net.IPNet) error
	RemoveFromWhiteList(network net.IPNet) error
}

type BWListStorage interface {
	Add(record bwliststorage.ListRecord) error
	Remove(record bwliststorage.ListRecord) error
	GetByIP(ip net.IP) (record *bwliststorage.ListRecord, err error)
	RemoveAll() error
	Connect() error
	Close() error
}

type BucketStorage interface {
	Set(key string, bucket ratelimit.Bucket)
	Get(key string) (bucket ratelimit.Bucket, ok bool)
	Remove(key string)
	ClearByTTL()
	ClearAll()
}

type App struct {
	logger          zap.Logger
	bwLists         BWListStorage
	loginBuckets    BucketStorage
	passwordBuckets BucketStorage
	ipBuckets       BucketStorage
	cache           lrucache.Cache
	maxLogins       int
	maxPasswords    int
	maxIPs          int
}

func New(lg zap.Logger, bwStorage BWListStorage, loginBucketStorage BucketStorage, passwordBucketStorage BucketStorage,
	ipBucketStorage BucketStorage, cache lrucache.Cache, maxLogins int, maxPasswords int, maxIPs int,
) *App {
	return &App{
		logger:          lg,
		bwLists:         bwStorage,
		loginBuckets:    loginBucketStorage,
		passwordBuckets: passwordBucketStorage,
		ipBuckets:       ipBucketStorage,
		cache:           cache,
		maxLogins:       maxLogins,
		maxPasswords:    maxPasswords,
		maxIPs:          maxIPs,
	}
}

func (a *App) Auth(login, password string, ip net.IP) (bool, error) {
	// Проверяем черно-белые списки
	item, err := a.getNetByIP(ip)
	if err != nil {
		return false, err
	}
	if item.Type != "" {
		return item.Type == bwliststorage.TypeWhite, nil
	}

	res1 := a.processBucket(a.loginBuckets, login, a.maxLogins)
	res2 := a.processBucket(a.passwordBuckets, password, a.maxPasswords)
	res3 := a.processBucket(a.ipBuckets, ip.String(), a.maxIPs)

	return res1 && res2 && res3, nil
}

// Достаём запись из кэша или из БД.
func (a *App) getNetByIP(ip net.IP) (*bwliststorage.ListRecord, error) {
	key := lrucache.Key(ip.String())
	res, ok := a.cache.Get(key)

	if ok {
		r := res.(*bwliststorage.ListRecord)
		return r, nil
	}

	rec, err := a.bwLists.GetByIP(ip)
	if err != nil {
		return &bwliststorage.ListRecord{}, err
	}
	if rec == nil {
		rec = &bwliststorage.ListRecord{}
	}
	a.cache.Set(key, rec)

	return rec, nil
}

// Вычисляем скорость опустошения бакета в единицах в секунду.
// На основании того, что ёмкость задаётся в единицах в минуту.
func calcLeakRate(capacity int) float64 {
	return float64(capacity) / 60
}

// Получаем бакет из хранилища (или создаём новый) и инкрементим. Возвращаем успех или неуспех операции инкремента.
func (a *App) processBucket(storage BucketStorage, key string, capacity int) bool {
	b, ok := storage.Get(key)
	if !ok {
		b = ratelimit.NewBucket(capacity, calcLeakRate(capacity))
	}
	res := b.Increment()
	storage.Set(key, b)

	return res
}

func (a *App) ResetBuckets(login, password string, ip net.IP) {
	if login != "" {
		resetBucket(a.loginBuckets, login, a.maxLogins)
	}
	if password != "" {
		resetBucket(a.passwordBuckets, password, a.maxPasswords)
	}
	if ip != nil {
		resetBucket(a.ipBuckets, ip.String(), a.maxIPs)
	}
}

func resetBucket(storage BucketStorage, key string, capacity int) {
	_, ok := storage.Get(key)
	if ok {
		storage.Set(key, ratelimit.NewBucket(capacity, calcLeakRate(capacity)))
	}
}

func (a *App) AddToBlackList(network net.IPNet) error {
	if err := a.bwLists.Add(
		bwliststorage.ListRecord{
			Network: network,
			Type:    bwliststorage.TypeBlack,
		},
	); err != nil {
		return err
	}

	a.cache.Clear()
	return nil
}

func (a *App) AddToWhiteList(network net.IPNet) error {
	if err := a.bwLists.Add(
		bwliststorage.ListRecord{
			Network: network,
			Type:    bwliststorage.TypeWhite,
		},
	); err != nil {
		return err
	}

	a.cache.Clear()
	return nil
}

func (a *App) RemoveFromBlackList(network net.IPNet) error {
	if err := a.bwLists.Remove(
		bwliststorage.ListRecord{
			Network: network,
			Type:    bwliststorage.TypeBlack,
		},
	); err != nil {
		return err
	}

	a.cache.Clear()
	return nil
}

func (a *App) RemoveFromWhiteList(network net.IPNet) error {
	if err := a.bwLists.Remove(
		bwliststorage.ListRecord{
			Network: network,
			Type:    bwliststorage.TypeWhite,
		},
	); err != nil {
		return err
	}

	a.cache.Clear()
	return nil
}
