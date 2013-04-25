package handlers

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedisStoreOptions struct {
	Network        string
	Address        string
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	Database       int
	KeyPrefix      string
}

type RedisStore struct {
	opts *RedisStoreOptions
	conn redis.Conn
}

func NewRedisStore(opts *RedisStoreOptions) *RedisStore {
	rs := &RedisStore{opts, nil}
	rs.conn, err := redis.DialTimeout(opts.Network, opts.Address, opts.ConnectTimeout,
		opts.ReadTimeout, opts.WriteTimeout)
	if err != nil {
		panic(err)
	}
	return rs
}

func (this *RedisStore) Get(id string) (*Session, error) {
	strSess, err := redis.String(this.conn.Do("GET", id))
	if err != nil {
		return nil, err
	}
	return nil, nil
}
