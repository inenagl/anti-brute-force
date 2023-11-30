package app

import (
	"errors"
	"net"
	"testing"
	"time"

	lrucache "github.com/inenagl/anti-brute-force/internal/cache"
	"github.com/inenagl/anti-brute-force/internal/logger"
	"github.com/inenagl/anti-brute-force/internal/ratelimit"
	bucketstorage "github.com/inenagl/anti-brute-force/internal/storage/bucket"
	bwliststorage "github.com/inenagl/anti-brute-force/internal/storage/bwlist"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testLogger(t *testing.T) *zap.Logger {
	t.Helper()
	l, err := logger.New("dev", "fatal", "json", []string{}, []string{})
	require.NoError(t, err)

	return l
}

// Проверка работы предфильтра с черно-белыми списками, а также кэширования этих списков.
func TestAuthByBWList(t *testing.T) {
	bwStorage := newMockBWListStorage()
	lStorage := newMockBucketStorage()
	pStorage := newMockBucketStorage()
	ipStorage := newMockBucketStorage()
	cache := lrucache.New(10, time.Millisecond)

	ip := net.IPv4(1, 1, 1, 1)
	record := bwliststorage.ListRecord{
		Network: net.IPNet{IP: net.IPv4(1, 1, 1, 1), Mask: net.IPv4Mask(255, 255, 255, 0)},
		Type:    bwliststorage.TypeWhite,
	}
	bwStorage.On("GetByIP", ip).Return(&record, nil)

	a := New(*testLogger(t), bwStorage, lStorage, pStorage, ipStorage, cache, 60, 60, 60)

	res, err := a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.True(t, res)

	lStorage.AssertNotCalled(t, "Get")
	pStorage.AssertNotCalled(t, "Get")
	ipStorage.AssertNotCalled(t, "Get")
	bwStorage.AssertNumberOfCalls(t, "GetByIP", 1)
	cached, ok := cache.Get(lrucache.Key(ip.String()))
	require.True(t, ok)
	require.Equal(t, record, *cached.(*bwliststorage.ListRecord))

	// Вызываем ещё раз. Теперь подсеть вернётся из кэша.
	res, err = a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.True(t, res)

	lStorage.AssertNotCalled(t, "Get")
	pStorage.AssertNotCalled(t, "Get")
	ipStorage.AssertNotCalled(t, "Get")
	bwStorage.AssertNumberOfCalls(t, "GetByIP", 1)

	// Ждём инвалидации кэша.
	time.Sleep(time.Millisecond + 1)

	// Изменим тип возвращаемого списка на "чёрный". И проверим, что ответ метода изменится на false.
	record.Type = bwliststorage.TypeBlack
	res, err = a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.False(t, res)

	lStorage.AssertNotCalled(t, "Get")
	pStorage.AssertNotCalled(t, "Get")
	ipStorage.AssertNotCalled(t, "Get")
	bwStorage.AssertNumberOfCalls(t, "GetByIP", 2)
}

// Проверка работы core части алгоритма ограничения частоты.
func TestAuthByBuckets(t *testing.T) {
	bwStorage := newMockBWListStorage()
	lStorage := newMockBucketStorage()
	pStorage := newMockBucketStorage()
	ipStorage := newMockBucketStorage()
	cache := lrucache.New(10, time.Millisecond)

	ip := net.IPv4(2, 2, 2, 2)
	bwStorage.On("GetByIP", ip).Return(nil, nil)

	lBucket := ratelimit.NewBucket(1, 10)  // бакет ёмкостью 1 и скоростью освобождения 10 в секунду.
	pBucket := ratelimit.NewBucket(1, 10)  // бакет ёмкостью 1 и скоростью освобождения 10 в секунду.
	ipBucket := ratelimit.NewBucket(1, 10) // бакет ёмкостью 1 и скоростью освобождения 10 в секунду.
	lStorage.On("Get", "a").Return(lBucket, true)
	pStorage.On("Get", "b").Return(pBucket, true)
	ipStorage.On("Get", ip.String()).Return(ipBucket, true)
	lStorage.On("Set", "a", lBucket).Return()
	pStorage.On("Set", "b", pBucket).Return()
	ipStorage.On("Set", ip.String(), ipBucket).Return()

	a := New(*testLogger(t), bwStorage, lStorage, pStorage, ipStorage, cache, 2, 2, 2)

	res, err := a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.True(t, res)

	lStorage.AssertNumberOfCalls(t, "Get", 1)
	pStorage.AssertNumberOfCalls(t, "Get", 1)
	ipStorage.AssertNumberOfCalls(t, "Get", 1)
	lStorage.AssertNumberOfCalls(t, "Set", 1)
	pStorage.AssertNumberOfCalls(t, "Set", 1)
	ipStorage.AssertNumberOfCalls(t, "Set", 1)

	// Повторная попытка авторизации отклоняется, т.к. все бакеты переполнены.
	res, err = a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.False(t, res)

	wait := 100 * time.Millisecond
	// Через 100 миллисекунд бакеты освобождаются и авторизация разрешена снова.
	time.Sleep(wait)
	res, err = a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.True(t, res)

	// Заполним полностью каждый бакет по очереди
	time.Sleep(wait)
	lBucket.Increment()
	res, err = a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.False(t, res)
	// Проверяем, что данные в каждом бакете инкрементятся независимо от того, в каком из них переполнение.
	lStorage.AssertNumberOfCalls(t, "Set", 4)
	pStorage.AssertNumberOfCalls(t, "Set", 4)
	ipStorage.AssertNumberOfCalls(t, "Set", 4)

	time.Sleep(wait)
	pBucket.Increment()
	res, err = a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.False(t, res)
	lStorage.AssertNumberOfCalls(t, "Set", 5)
	pStorage.AssertNumberOfCalls(t, "Set", 5)
	ipStorage.AssertNumberOfCalls(t, "Set", 5)

	time.Sleep(wait)
	ipBucket.Increment()
	res, err = a.Auth("a", "b", ip)
	require.NoError(t, err)
	require.False(t, res)
	lStorage.AssertNumberOfCalls(t, "Set", 6)
	pStorage.AssertNumberOfCalls(t, "Set", 6)
	ipStorage.AssertNumberOfCalls(t, "Set", 6)
}

func TestResetBuckets(t *testing.T) {
	bwStorage := newMockBWListStorage()
	lStorage := bucketstorage.New(time.Minute)
	pStorage := bucketstorage.New(time.Minute)
	ipStorage := bucketstorage.New(time.Minute)
	cache := lrucache.New(10, time.Millisecond)

	a := New(*testLogger(t), bwStorage, lStorage, pStorage, ipStorage, cache, 2, 4, 6)

	login := "login"
	passwd := "password"
	ip := net.IPv4(1, 1, 1, 1)
	ipKey := ip.String()

	checkEmptyBuckets := func() {
		_, ok := lStorage.Get(login)
		require.False(t, ok)
		_, ok = pStorage.Get(passwd)
		require.False(t, ok)
		_, ok = ipStorage.Get(ipKey)
		require.False(t, ok)
	}
	checkBucketsAfterReset := func() {
		b, ok := lStorage.Get(login)
		require.True(t, ok)
		require.Equal(t, 0, b.GetCount())
		require.Equal(t, float64(2)/60, b.GetRate())
		b, ok = pStorage.Get(passwd)
		require.True(t, ok)
		require.Equal(t, 0, b.GetCount())
		require.Equal(t, float64(4)/60, b.GetRate())
		b, ok = ipStorage.Get(ipKey)
		require.True(t, ok)
		require.Equal(t, 0, b.GetCount())
		require.Equal(t, float64(6)/60, b.GetRate())
	}
	setBucket := func(storage *bucketstorage.Storage, key string, bucket ratelimit.Bucket) {
		bucket.Increment()
		bucket.Increment()
		storage.Set(key, bucket)
		res, ok := storage.Get(key)
		require.True(t, ok)
		require.Equal(t, bucket, res)
	}

	checkEmptyBuckets()
	// Ресет несуществующих бакетов ни к чему не приводит.
	a.ResetBuckets(login, passwd, ip)
	checkEmptyBuckets()

	setBucket(lStorage, login, ratelimit.NewBucket(2, float64(2)/60))
	setBucket(pStorage, passwd, ratelimit.NewBucket(4, float64(4)/60))
	setBucket(ipStorage, ipKey, ratelimit.NewBucket(6, float64(6)/60))

	a.ResetBuckets(login, passwd, ip)
	checkBucketsAfterReset()
}

func TestAddToBWList(t *testing.T) { //nolint:dupl
	bwStorage := newMockBWListStorage()
	bucketStorage := bucketstorage.New(time.Minute)
	cache := newMockCache()
	cache.On("Clear").Return()

	a := New(*testLogger(t), bwStorage, bucketStorage, bucketStorage, bucketStorage, cache, 1, 1, 1)

	networkBlack := net.IPNet{
		IP:   net.IPv4(1, 2, 3, 4),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	networkWhite := net.IPNet{
		IP:   net.IPv4(11, 22, 33, 44),
		Mask: net.IPv4Mask(255, 255, 0, 0),
	}
	recordBlack := bwliststorage.ListRecord{
		Network: networkBlack,
		Type:    bwliststorage.TypeBlack,
	}
	recordWhite := bwliststorage.ListRecord{
		Network: networkWhite,
		Type:    bwliststorage.TypeWhite,
	}
	err := errors.New("add error")

	// Успешное добавление. Кэш чистится.
	var e error
	bwStorage.On("Add", recordBlack).Return(nil).Once()
	bwStorage.On("Add", recordWhite).Return(nil).Once()

	e = a.AddToBlackList(networkBlack)
	require.NoError(t, e)
	cache.AssertNumberOfCalls(t, "Clear", 1)

	e = a.AddToWhiteList(networkWhite)
	require.NoError(t, e)
	cache.AssertNumberOfCalls(t, "Clear", 2)

	// Добавление с ошибкой. Кэш не чистится.
	bwStorage.On("Add", recordBlack).Return(err).Once()
	bwStorage.On("Add", recordWhite).Return(err).Once()

	e = a.AddToBlackList(networkBlack)
	require.Equal(t, err, e)
	cache.AssertNumberOfCalls(t, "Clear", 2)

	e = a.AddToWhiteList(networkWhite)
	require.Equal(t, err, e)
	cache.AssertNumberOfCalls(t, "Clear", 2)
}

func TestRemoveFromBWList(t *testing.T) { //nolint:dupl
	bwStorage := newMockBWListStorage()
	bucketStorage := bucketstorage.New(time.Minute)
	cache := newMockCache()
	cache.On("Clear").Return()

	a := New(*testLogger(t), bwStorage, bucketStorage, bucketStorage, bucketStorage, cache, 1, 1, 1)

	networkBlack := net.IPNet{
		IP:   net.IPv4(1, 2, 3, 4),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	networkWhite := net.IPNet{
		IP:   net.IPv4(11, 22, 33, 44),
		Mask: net.IPv4Mask(255, 255, 0, 0),
	}
	recordBlack := bwliststorage.ListRecord{
		Network: networkBlack,
		Type:    bwliststorage.TypeBlack,
	}
	recordWhite := bwliststorage.ListRecord{
		Network: networkWhite,
		Type:    bwliststorage.TypeWhite,
	}
	err := errors.New("remove error")

	// Успешное удаление. Кэш чистится.
	var e error
	bwStorage.On("Remove", recordBlack).Return(nil).Once()
	bwStorage.On("Remove", recordWhite).Return(nil).Once()

	e = a.RemoveFromBlackList(networkBlack)
	require.NoError(t, e)
	cache.AssertNumberOfCalls(t, "Clear", 1)

	e = a.RemoveFromWhiteList(networkWhite)
	require.NoError(t, e)
	cache.AssertNumberOfCalls(t, "Clear", 2)

	// Удаление с ошибкой. Кэш не чистится.
	bwStorage.On("Remove", recordBlack).Return(err).Once()
	bwStorage.On("Remove", recordWhite).Return(err).Once()

	e = a.RemoveFromBlackList(networkBlack)
	require.Equal(t, err, e)
	cache.AssertNumberOfCalls(t, "Clear", 2)

	e = a.RemoveFromWhiteList(networkWhite)
	require.Equal(t, err, e)
	cache.AssertNumberOfCalls(t, "Clear", 2)
}
