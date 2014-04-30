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

func (r *redisBackend) GetVariable(group, variable string) (string, error) {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Return the results of the GET command.
	return redis.String(conn.Do("HGET", group, variable))
}

func (r *redisBackend) SetVariable(group, variable, value string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET command and return any error.
	_, err := conn.Do("HMSET", group, variable, value)
	return err
}

func (r *redisBackend) RemoveVariable(group, variable string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("HDEL", group, variable)
	return err
}

func (r *redisBackend) RemoveGroup(group string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("DEL", group)
	return err
}
