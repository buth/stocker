package redis

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

const (
	concurrentConnections int = 2
)

type redisBackend struct {
	connectionType, connectionString string
	pool                             *redis.Pool
	sem                              chan bool
	subscriptions                    map[string]*redis.PubSubConn
}

func New(connectionType, connectionString string) *redisBackend {

	r := &redisBackend{connectionType: connectionType, connectionString: connectionString}

	// Build the underlying pool setting the maximum size to the number of
	// allowed concurrent connections.
	r.pool = redis.NewPool(r.dial, concurrentConnections)

	// Subscriptions.
	r.subscriptions = make(map[string]*redis.PubSubConn)

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

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Return the results of the GET command.
	return redis.String(conn.Do("GET", key))
}

func (r *redisBackend) Set(key, value string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET command and return any error.
	_, err := conn.Do("SET", key, value)
	return err
}

func (r *redisBackend) SetWithTTL(key, value string, ttl int) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET and EXPIRE commands as a single transaaction and return any
	// error.
	conn.Send("MULTI")
	conn.Send("SET", key, value)
	conn.Send("EXPIRE", key, ttl)
	_, err := conn.Do("EXEC")
	return err
}

func (r *redisBackend) Remove(key string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the DEL command and return any error.
	_, err := conn.Do("DEL", key)
	return err
}

func (r *redisBackend) Publish(key, message string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the PUBLISH command and return any error.
	_, err := conn.Do("PUBLISH", key, message)
	return err
}

func (r *redisBackend) Subscribe(key string, process func(string, string) error) error {

	// Don't pull these connections from the pool, as they will remain open.
	conn, err := r.dial()
	if err != nil {
		return err
	}

	// Use the redis Pub/Sub wrapper.
	r.subscriptions[key] = &redis.PubSubConn{conn}
	if err = r.subscriptions[key].PSubscribe(key); err != nil {
		return err
	}

	defer r.subscriptions[key].Close()

	for {
		switch v := r.subscriptions[key].Receive().(type) {
		case redis.PMessage:

			err := process(v.Channel, string(v.Data))
			if err != nil {
				return err
			}
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

func (r *redisBackend) Unsubscribe(key string) error {

	// Unsubscribe and return any error.
	return r.subscriptions[key].PUnsubscribe(key)
}

func (r *redisBackend) Add(key, value string) error {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET command and return any error.
	_, err := conn.Do("SADD", key, value)
	return err
}

func (r *redisBackend) List(key string) ([]string, error) {

	// Wait for a signal from the semaphore and then pull a new connection from
	// the pool. Defer signalling the semaphore and closing the connection.
	r.p()
	defer r.v()
	conn := r.pool.Get()
	defer conn.Close()

	// Run the SET command and return any error.
	return redis.Strings(conn.Do("SMEMBERS", key))
}
