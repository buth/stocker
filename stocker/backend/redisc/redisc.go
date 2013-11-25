package redisc

import (
	"github.com/buth/funnel"
	"github.com/garyburd/redigo/redis"
	"log"
)

// A redisc is a managed pool of connections to redis that restricts
// total redis connections and exposes a the backend interface.
type redisc struct {
	pool   *redis.Pool
	funnel *funnel.Funnel
}

func New(connectionType, connectionString string) *redisc {

	// Build the underlying pool.
	pool := redis.NewPool(
		func() (redis.Conn, error) {
			connection, err := redis.Dial(connectionType, connectionString)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			return connection, err
		},
		2,
	)

	// Use that pool to create a concurency-managed funnel.
	funnel := funnel.New(pool, 2)

	// Build the redisc object.
	return &redisc{pool: pool, funnel: funnel}
}

func (c *redisc) Get(key string) (value string, err error) {
	c.funnel.Get(func(conn redis.Conn) {
		value, err = redis.String(conn.Do("GET", key))
	})
	return
}

func (c *redisc) Set(key, value string) (err error) {
	c.funnel.Get(func(conn redis.Conn) {
		_, err = conn.Do("SET", key, value)
	})
	return
}

func (c *redisc) SetWithTTL(key, value string, ttl int) (err error) {
	c.funnel.Get(func(conn redis.Conn) {
		conn.Send("MULTI")
		conn.Send("SET", key, value)
		conn.Send("EXPIRE", key, ttl)
		_, err = conn.Do("EXEC")
	})
	return
}

func (c *redisc) Remove(key string) (err error) {
	c.funnel.Get(func(conn redis.Conn) {
		_, err = conn.Do("DEL", key)
	})
	return
}
