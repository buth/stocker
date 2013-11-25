package backend

type Backend interface {
	Get(key string) (value string, err error)
	Set(key, value string) (err error)
	SetWithTTL(key, value string, ttl int) (err error)
	Remove(key string) (err error)
}
