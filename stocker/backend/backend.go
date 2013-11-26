package backend

type Backend interface {
	Get(key string) (string, error)
	Set(key, value string) error
	SetWithTTL(key, value string, ttl int) error
	Remove(key string) error
	Publish(key, message string) error
	Subscribe(key string, process func(string, string) error) error
}
