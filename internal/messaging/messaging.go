package messaging

type Producer interface {
	Publish(name, topic string, message interface{}) error
	Close() error
}

type Consumer interface {
	Subscribe(topic string, handler func([]byte) error) error
	Start() error
	Stop() error
	Close() error
}

type RabbitMQOpts struct {
	Host          string
	Port          string
	Username      string
	Password      string
	VHost         string
	ExchangeName  string
	ExchangeType  string
	ExchangeTopic string
}
