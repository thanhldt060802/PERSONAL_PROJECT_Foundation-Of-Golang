### Document for go-redis library usage

- func redis.NewClient(opt *redis.Options) *redis.Client
	This function will initialize a *redis.Client with options (it is Redis client to Redis server).

- func (c redis.cmdable) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
	This function will publish message to channel.

- func (c *redis.Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	This function will subscribe to channel.

- func (c *redis.PubSub) Channel(opts ...redis.ChannelOption) <-chan *redis.Message
	This function will start a listener to channel for receiving messages.