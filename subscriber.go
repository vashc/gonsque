package gonsque

type Subscriber struct {
	topic       string
	channel     string
	concurrency int
	handler     Handler
	model       interface{}
}

func NewSubscriber(topic, channel string, concurrency int, handler Handler, model interface{}) *Subscriber {
	return &Subscriber{
		topic,
		channel,
		concurrency,
		handler,
		model,
	}
}
