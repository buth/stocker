package redis

import (
	"bytes"
	"github.com/garyburd/redigo/redis"
	"log"
)

const (
	MaxIdle int = 2
	KeySep      = ':'
)

type redisBackend struct {
	namespace, protocol, address string
	pool                         *redis.Pool
}

func New(namespace, protocol, address string) *redisBackend {

	r := &redisBackend{namespace: namespace, protocol: protocol, address: address}

	// Build the underlying pool setting the maximum size to the number of
	// allowed concurrent connections.
	r.pool = redis.NewPool(r.dial, MaxIdle)

	// Build the Backend object.
	return r
}

func (r *redisBackend) dial() (redis.Conn, error) {
	connection, err := redis.Dial(r.protocol, r.address)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return connection, err
}

func (r *redisBackend) Key(group string) []byte {
	buf := bytes.NewBufferString(r.namespace)
	buf.WriteRune(KeySep)
	buf.WriteString(group)
	return buf.Bytes()
}

func (r *redisBackend) GetVariable(group, variable string) (string, error) {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Return the results of the GET command.
	return redis.String(conn.Do("HGET", r.Key(group), variable))
}

func (r *redisBackend) SetVariable(group, variable, value string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET command and return any error.
	_, err := conn.Do("HMSET", r.Key(group), variable, value)
	return err
}

func (r *redisBackend) RemoveVariable(group, variable string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("HDEL", r.Key(group), variable)
	return err
}

func (r *redisBackend) RemoveGroup(group string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("DEL", r.Key(group))
	return err
}
