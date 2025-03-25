package cache

type Manager interface {
	Get(key string) (string, bool)
	Set(key, value string)
	Exists(key string) bool
	Delete(key string)
}
