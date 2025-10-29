### Document for ttlcache library usage

- func New[K comparable, V any](opts ...Option[K, V]) *Cache[K, V]
	This function will initialize a *Cache[K, V] with options.

- func (c *Cache[K, V]) Start()
	This function will start an automatic cleanup process that periodically deletes expired items.

- func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) *Item[K, V]
	This function will set an element with key: value format and come with TTL.

- func (c *Cache[K, V]) Get(key K, opts ...Option[K, V]) *Item[K, V]
	This function will get an element by key, it returns *Item[K, V] which contains key and value, and refresh the TTL of element.

- func (c *Cache[K, V]) Delete(key K)
	This function will delete an element by key.