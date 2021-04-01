package gonsque

type Subscriber struct {
	topic       string
	channel     string
	concurrency int
	handler     Handler
}

func NewSubscriber(topic, channel string, concurrency int, handler Handler) *Subscriber {
	return &Subscriber{
		topic,
		channel,
		concurrency,
		handler,
	}
}
