package backend

import (
	"github.com/buth/funnel"
	"github.com/garyburd/redigo/redis"
	"log"
)

// A Backend is a managed pool of connections to redis that restricts
// total redis connections and exposes a the backend interface.
type Backend struct {
	pool   *redis.Pool
	funnel *funnel.Funnel
}

func New(connectionType, connectionString string) *Backend {

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

	// Build the Backend object.
	return &Backend{pool: pool, funnel: funnel}
}

func (c *Backend) Get(key string) (value string, err error) {
	c.funnel.Get(func(conn redis.Conn) {
		value, err = redis.String(conn.Do("GET", key))
	})
	return
}

func (c *Backend) Set(key, value string) (err error) {
	c.funnel.Get(func(conn redis.Conn) {
		_, err = conn.Do("SET", key, value)
	})
	return
}

func (c *Backend) SetWithTTL(key, value string, ttl int) (err error) {
	c.funnel.Get(func(conn redis.Conn) {
		conn.Send("MULTI")
		conn.Send("SET", key, value)
		conn.Send("EXPIRE", key, ttl)
		_, err = conn.Do("EXEC")
	})
	return
}

func (c *Backend) Remove(key string) (err error) {
	c.funnel.Get(func(conn redis.Conn) {
		_, err = conn.Do("DEL", key)
	})
	return
}
