package handlers

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	ErrNoKeyPrefix = errors.New("cannot get session keys without a key prefix")
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
	var err error
	rs := &RedisStore{opts, nil}
	rs.conn, err = redis.DialTimeout(opts.Network, opts.Address, opts.ConnectTimeout,
		opts.ReadTimeout, opts.WriteTimeout)
	if err != nil {
		panic(err)
	}
	return rs
}

func (this *RedisStore) Get(id string) (*Session, error) {
	key := id
	if this.opts.KeyPrefix != "" {
		key = this.opts.KeyPrefix + ":" + id
	}
	strSess, err := redis.String(this.conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	var sess Session
	err = json.Unmarshal([]byte(strSess), &sess)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (this *RedisStore) Set(sess *Session) error {
	bufSess, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	key := sess.ID()
	if this.opts.KeyPrefix != "" {
		key = this.opts.KeyPrefix + ":" + sess.ID()
	}
	_, err = this.conn.Do("SETEX", key, int(sess.maxAge.Seconds()), string(bufSess))
	if err != nil {
		return err
	}
	return nil
}

func (this *RedisStore) Delete(id string) error {
	key := id
	if this.opts.KeyPrefix != "" {
		key = this.opts.KeyPrefix + ":" + id
	}
	_, err := this.conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

func (this *RedisStore) Clear() error {
	vals, err := this.getSessionKeys()
	if err != nil {
		return err
	}
	if len(vals) > 0 {
		this.conn.Send("MULTI")
		for _, v := range vals {
			this.conn.Send("DEL", v)
		}
		_, err = this.conn.Do("EXEC")
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *RedisStore) Len() int {
	vals, err := this.getSessionKeys()
	if err != nil {
		return -1
	}
	return len(vals)
}

func (this *RedisStore) getSessionKeys() ([]interface{}, error) {
	if this.opts.KeyPrefix != "" {
		return redis.Values(this.conn.Do("KEYS", this.opts.KeyPrefix+":*"))
	}
	return nil, ErrNoKeyPrefix
}
