package bwliststorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type ListType string

const (
	TypeBlack ListType = "Black"
	TypeWhite ListType = "White"
)

type ListRecord struct {
	Network net.IPNet
	Type    ListType
}

type dbRecord struct {
	Network string
	Type    ListType
}

type Storage struct {
	dsn     string
	db      *sqlx.DB
	timeout time.Duration
}

func New(host string, port int, dbname, user, password, sslmode string, timeout time.Duration) *Storage {
	portStr := ""
	if port != 0 {
		portStr = ":" + strconv.Itoa(port)
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s%s/%s?sslmode=%s", user, password, host, portStr, dbname, sslmode)

	return &Storage{
		dsn:     dsn,
		timeout: timeout,
	}
}

func (s Storage) createTimeoutCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), s.timeout) //nolint: govet
	return ctx
}

func (s *Storage) Connect() error {
	db, err := sqlx.ConnectContext(s.createTimeoutCtx(), "pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	s.db = db

	return nil
}

func (s *Storage) Ping() error {
	if s.db == nil {
		return s.Connect()
	}

	if err := s.db.PingContext(s.createTimeoutCtx()); err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}

	return nil
}

func (s *Storage) Add(record ListRecord) error {
	if err := s.Ping(); err != nil {
		return err
	}

	// Вставляем запись с проверкой того, что вставляемая сеть не пересекается с уже имеющимися.
	res, err := s.db.ExecContext(
		s.createTimeoutCtx(),
		`INSERT INTO bw_lists (network, type)
		SELECT INET(network), LIST_TYPE(type) FROM (
			 SELECT
			 $1 AS network,
			 $2 AS type
		) t
		WHERE NOT EXISTS (
			 SELECT 1 FROM bw_lists bw
			 WHERE bw.network && INET(t.network)
		)`,
		record.Network.String(),
		record.Type,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	// Если rows отличается от нуля, значит вставка прошла успешно.
	if rows != 0 {
		return nil
	}

	// Если rows = 0, то ничего не вставилось. Это значит добавляемая сеть пересекается с уже имеющимися в базе.
	// Найдём такие сети и добавим в текст ошибки.
	var dest []dbRecord
	if err = s.db.SelectContext(
		s.createTimeoutCtx(),
		&dest,
		`SELECT network, type FROM bw_lists WHERE network && $1`,
		record.Network.String(),
	); err != nil {
		return err
	}

	// Формируем текст ошибки.
	nets := make([]string, len(dest))
	for i, v := range dest {
		nets[i] = string(v.Type) + " - " + v.Network
	}
	msg := fmt.Sprintf(
		"Can't insert '%s' into black/white list. Intersection with: %s",
		record.Network.String(),
		strings.Join(nets, ", "),
	)

	return errors.New(msg)
}

func (s *Storage) Remove(record ListRecord) error {
	if err := s.Ping(); err != nil {
		return err
	}

	_, err := s.db.ExecContext(
		s.createTimeoutCtx(),
		"DELETE FROM bw_lists WHERE network=$1 AND type=$2",
		record.Network.String(),
		record.Type,
	)

	return err
}

func (s *Storage) GetByIP(ip net.IP) (*ListRecord, error) {
	var err error
	if err = s.Ping(); err != nil {
		return nil, err
	}

	var dest dbRecord
	if err = s.db.GetContext(
		s.createTimeoutCtx(),
		&dest,
		`SELECT network, type FROM bw_lists WHERE $1 <<= network`,
		ip.String(),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return nil, err
	}

	network, err := ParseIPNet(dest.Network)
	if err != nil {
		return nil, err
	}

	return &ListRecord{Network: *network, Type: dest.Type}, err
}

func (s *Storage) RemoveAll() error {
	if err := s.Ping(); err != nil {
		return err
	}

	_, err := s.db.ExecContext(
		s.createTimeoutCtx(),
		"TRUNCATE TABLE bw_lists",
	)

	return err
}

func ParseIPNet(s string) (*net.IPNet, error) {
	_, IPNet, err := net.ParseCIDR(s)
	// Допускаются единичные IP, поэтому пробуем распарсить как IP
	if err != nil {
		ip := net.ParseIP(s)
		if ip == nil {
			return nil, fmt.Errorf(`%w: can't parse "%s" to IP Network`, err, s)
		}
		IPNet = &net.IPNet{
			IP:   ip,
			Mask: net.IPv4Mask(255, 255, 255, 255),
		}
	}
	return IPNet, nil
}
