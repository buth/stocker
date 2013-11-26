package redis

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

const (
	concurrentConnections int = 2
)

func prefixKey(key string) string {
	return "stocker:" + key
}

type redisBackend struct {
	connectionType, connectionString string
	pool                             *redis.Pool
	sem                              chan bool
}

func New(connectionType, connectionString string) *redisBackend {

	r := &redisBackend{connectionType: connectionType, connectionString: connectionString}

	// Build the underlying pool setting the maximum size to the number of
	// allowed concurrent connections.
	r.pool = redis.NewPool(r.dial, concurrentConnections)

	// Preload the semaphore.
	r.sem = make(chan bool, concurrentConnections)
	for i := 0; i < concurrentConnections; i += 1 {
		r.v()
	}

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

func (r *redisBackend) v() {
	r.sem <- true
}

func (r *redisBackend) p() {
	<-r.sem
}

func (r *redisBackend) Get(key string) (string, error) {
	prefixedKey := prefixKey(key)

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Return the results of the GET command.
	return redis.String(conn.Do("GET", prefixedKey))
}

func (r *redisBackend) Set(key, value string) error {
	prefixedKey := prefixKey(key)

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET command and return any error.
	_, err := conn.Do("SET", prefixedKey, value)
	return err
}

func (r *redisBackend) SetWithTTL(key, value string, ttl int) error {
	prefixedKey := prefixKey(key)

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET and EXPIRE commands as a single transaaction and return any
	// error.
	conn.Send("MULTI")
	conn.Send("SET", prefixedKey, value)
	conn.Send("EXPIRE", prefixedKey, ttl)
	_, err := conn.Do("EXEC")
	return err
}

func (r *redisBackend) Remove(key string) error {
	prefixedKey := prefixKey(key)

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("DEL", prefixedKey)
	return err
}

func (r *redisBackend) Publish(key, message string) error {
	prefixedKey := prefixKey(key)

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the PUBLISH command and return any error.
	_, err := conn.Do("PUBLISH", prefixedKey, message)
	return err
}

func (r *redisBackend) Subscribe(key string, process func(string, string) error) error {
	prefixedKey := prefixKey(key)

	// Don't pull these connections from the pool, as they will remain open.
	conn, err := r.dial()
	if err != nil {
		return err
	}

	// Use the redis Pub/Sub wrapper.
	pconn := redis.PubSubConn{conn}
	defer pconn.Close()

	// Subscribe using the prefixed key as a pattern.
	pconn.PSubscribe(prefixedKey)
	for {
		switch v := pconn.Receive().(type) {
		case redis.PMessage:

			err := process(v.Channel, string(v.Data))
			if err != nil {
				return err
			}
		case error:
			return v
		}
	}

	return nil
}
