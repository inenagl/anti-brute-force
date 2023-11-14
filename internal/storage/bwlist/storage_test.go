package bwliststorage

import (
	"net"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/stretchr/testify/require"
)

// TODO Временные тесты для себя. Нужно переделать на функциональные и запускать в контейнере.
func SkipCI(t *testing.T) {
	t.Helper()
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

func TestAdd(t *testing.T) {
	SkipCI(t)
	s := New(
		"localhost",
		5432,
		"abf",
		"abfuser",
		"abfpassword",
		"disable",
		time.Second*5,
	)

	rec := ListRecord{
		Network: net.IPNet{IP: net.IPv4(192, 168, 5, 0), Mask: net.IPv4Mask(255, 255, 255, 0)},
		Type:    TypeWhite,
	}

	require.NoError(t, s.Add(rec))
	require.Error(t, s.Add(rec))

	require.NoError(t, s.Remove(rec))
}

func TestRemove(t *testing.T) {
	SkipCI(t)
	s := New(
		"localhost",
		5432,
		"abf",
		"abfuser",
		"abfpassword",
		"disable",
		time.Second*5,
	)

	rec := ListRecord{
		Network: net.IPNet{IP: net.IPv4(192, 168, 5, 0), Mask: net.IPv4Mask(255, 255, 255, 0)},
		Type:    TypeWhite,
	}

	require.NoError(t, s.Remove(rec))
	require.NoError(t, s.Add(rec))
	res, err := s.GetByIP(net.IPv4(192, 168, 5, 1))
	require.NoError(t, err)
	require.NotNil(t, res)

	require.NoError(t, s.Remove(rec))

	res, err = s.GetByIP(net.IPv4(192, 168, 5, 1))
	require.NoError(t, err)
	require.Nil(t, res)
}

func TestGetByIP(t *testing.T) {
	SkipCI(t)
	s := New(
		"localhost",
		5432,
		"abf",
		"abfuser",
		"abfpassword",
		"disable",
		time.Second*5,
	)

	rec := ListRecord{
		Network: net.IPNet{IP: net.IPv4(192, 192, 1, 0), Mask: net.IPv4Mask(255, 255, 255, 0)},
		Type:    TypeBlack,
	}

	require.NoError(t, s.Remove(rec))
	require.NoError(t, s.Add(rec))

	res, err := s.GetByIP(net.IPv4(192, 191, 1, 10))
	require.NoError(t, err)
	require.Nil(t, res)

	res, err = s.GetByIP(net.IPv4(192, 192, 1, 0))
	require.NoError(t, err)
	require.Equal(t, rec.Network.String(), res.Network.String())
	require.Equal(t, rec.Type, res.Type)

	res, err = s.GetByIP(net.IPv4(192, 192, 1, 100))
	require.NoError(t, err)
	require.Equal(t, rec.Network.String(), res.Network.String())
	require.Equal(t, rec.Type, res.Type)

	res, err = s.GetByIP(net.IPv4(192, 192, 1, 255))
	require.NoError(t, err)
	require.Equal(t, rec.Network.String(), res.Network.String())
	require.Equal(t, rec.Type, res.Type)

	require.NoError(t, s.Remove(rec))
}

func TestRemoveAll(t *testing.T) {
	SkipCI(t)
	s := New(
		"localhost",
		5432,
		"abf",
		"abfuser",
		"abfpassword",
		"disable",
		time.Second*5,
	)

	rec := ListRecord{
		Network: net.IPNet{IP: net.IPv4(192, 168, 5, 0), Mask: net.IPv4Mask(255, 255, 255, 0)},
		Type:    TypeWhite,
	}

	require.NoError(t, s.Remove(rec))
	require.NoError(t, s.Add(rec))
	res, err := s.GetByIP(net.IPv4(192, 168, 5, 1))
	require.NoError(t, err)
	require.NotNil(t, res)

	require.NoError(t, s.RemoveAll())

	res, err = s.GetByIP(net.IPv4(192, 168, 5, 1))
	require.NoError(t, err)
	require.Nil(t, res)
}
