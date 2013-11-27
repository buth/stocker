package backend

type Backend interface {
	Get(key string) (string, error)
	Set(key, value string) error
	SetWithTTL(key, value string, ttl int) error
	Remove(key string) error
	Publish(key, message string) error
	Subscribe(key string, process func(channel, message string) error) error
	Unsubscribe(key string) error
}
