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

	// Get a connection from the pool and defer its closing.
	conn := r.pool.Get()
	defer conn.Close()

	// Return the results of the GET command.
	return redis.String(conn.Do("HGET", r.Key(group), variable))
}

func (r *redisBackend) SetVariable(group, variable, value string) error {

	// Get a connection from the pool and defer its closing.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET command and return any error.
	_, err := conn.Do("HMSET", r.Key(group), variable, value)
	return err
}

func (r *redisBackend) RemoveVariable(group, variable string) error {

	// Get a connection from the pool and defer its closing.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("HDEL", r.Key(group), variable)
	return err
}

func (r *redisBackend) GetGroup(group string) (map[string]string, error) {

	// Create an empty map.
	variables := make(map[string]string)

	// Get a connection from the pool and defer its closing.
	conn := r.pool.Get()
	defer conn.Close()

	// Get the values as a flat string.
	values, err := redis.Strings(conn.Do("HGETALL", r.Key(group)))
	if err != nil {
		return variables, err
	}

	// Write the values into the variables map.
	for i := 0; i < len(values)-1; i += 2 {
		variables[values[i]] = values[i+1]
	}

	// Return the map with no error.
	return variables, nil
}

func (r *redisBackend) RemoveGroup(group string) error {

	// Get a connection from the pool and defer its closing.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("DEL", r.Key(group))
	return err
}
