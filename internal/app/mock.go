package app

import (
	"net"

	lrucache "github.com/inenagl/anti-brute-force/internal/cache"
	"github.com/inenagl/anti-brute-force/internal/ratelimit"
	bwliststorage "github.com/inenagl/anti-brute-force/internal/storage/bwlist"
	"github.com/stretchr/testify/mock"
)

type mockBWListStorage struct {
	mock.Mock
}

func newMockBWListStorage() *mockBWListStorage {
	return &mockBWListStorage{}
}

func (s *mockBWListStorage) Add(record bwliststorage.ListRecord) error {
	args := s.Called(record)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (s *mockBWListStorage) Remove(record bwliststorage.ListRecord) error {
	args := s.Called(record)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (s *mockBWListStorage) GetByIP(ip net.IP) (record *bwliststorage.ListRecord, err error) {
	args := s.Called(ip)
	var res *bwliststorage.ListRecord
	if args.Get(0) != nil {
		res = args.Get(0).(*bwliststorage.ListRecord)
	}
	return res, args.Error(1)
}

func (s *mockBWListStorage) RemoveAll() error {
	// TODO implement me
	panic("implement me")
}

func (s *mockBWListStorage) Connect() error {
	// TODO implement me
	panic("implement me")
}

func (s *mockBWListStorage) Close() error {
	// TODO implement me
	panic("implement me")
}

type mockBucketStorage struct {
	mock.Mock
}

func newMockBucketStorage() *mockBucketStorage {
	return &mockBucketStorage{}
}

func (s *mockBucketStorage) Set(key string, bucket ratelimit.Bucket) {
	s.Called(key, bucket)
}

func (s *mockBucketStorage) Get(key string) (bucket ratelimit.Bucket, ok bool) {
	args := s.Called(key)
	var res ratelimit.Bucket
	if args.Get(0) != nil {
		res = args.Get(0).(ratelimit.Bucket)
	}
	return res, args.Bool(1)
}

func (s *mockBucketStorage) Remove(_ string) {
	// TODO implement me
	panic("implement me")
}

func (s *mockBucketStorage) ClearByTTL() {
	// TODO implement me
	panic("implement me")
}

func (s *mockBucketStorage) ClearAll() {
	// TODO implement me
	panic("implement me")
}

type mockCache struct {
	mock.Mock
}

func newMockCache() *mockCache {
	return &mockCache{}
}

func (c *mockCache) Set(_ lrucache.Key, _ interface{}) bool {
	// TODO implement me
	panic("implement me")
}

func (c *mockCache) Get(_ lrucache.Key) (interface{}, bool) {
	// TODO implement me
	panic("implement me")
}

func (c *mockCache) Clear() {
	c.Called()
}
