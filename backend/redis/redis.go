package redis

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

const (
	MaxIdle int = 2
)

type redisBackend struct {
	connectionType, connectionString string
	pool                             *redis.Pool
}

func New(connectionType, connectionString string) *redisBackend {

	r := &redisBackend{connectionType: connectionType, connectionString: connectionString}

	// Build the underlying pool setting the maximum size to the number of
	// allowed concurrent connections.
	r.pool = redis.NewPool(r.dial, MaxIdle)

	// Build the Backend object.
	return r
}

func (r *redisBackend) dial() (redis.Conn, error) {
	connection, err := redis.Dial(r.connectionType, r.connectionString)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return connection, err
}

func (r *redisBackend) Get(key string) (string, error) {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Return the results of the GET command.
	return redis.String(conn.Do("GET", key))
}

func (r *redisBackend) Set(key, value string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET command and return any error.
	_, err := conn.Do("SET", key, value)
	return err
}

func (r *redisBackend) Remove(key string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("DEL", key)
	return err
}
