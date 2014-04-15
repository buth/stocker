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
	subscriptions                    map[string]*redis.PubSubConn
}

func New(connectionType, connectionString string) *redisBackend {

	r := &redisBackend{connectionType: connectionType, connectionString: connectionString}

	// Build the underlying pool setting the maximum size to the number of
	// allowed concurrent connections.
	r.pool = redis.NewPool(r.dial, MaxIdle)

	// // Subscriptions.
	// r.subscriptions = make(map[string]*redis.PubSubConn)

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

func (r *redisBackend) publish(key, message string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	conn := r.pool.Get()
	defer conn.Close()

	// Run the PUBLISH command and return any error.
	_, err := conn.Do("PUBLISH", key, message)
	return err
}

func (r *redisBackend) Subscribe(key string, process func(string)) error {

	// Don't pull these connections from the pool, as they will remain open.
	conn, err := r.dial()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Use the redis Pub/Sub wrapper.
	r.subscriptions[key] = &redis.PubSubConn{conn}
	if err = r.subscriptions[key].PSubscribe(key); err != nil {
		return err
	}

	defer r.subscriptions[key].Close()

	for {
		switch v := r.subscriptions[key].Receive().(type) {
		case redis.PMessage:
			// v.Channel,
			process(string(v.Data))

		case redis.Subscription:
			if v.Kind == "punsubscribe" {
				return nil
			}
		case error:
			return v
		}
	}

	// The only returnable error was with the subscription.
	return nil
}
