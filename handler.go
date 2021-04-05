package gonsque

import (
	"github.com/nsqio/go-nsq"
)

type Handler func(msg *Message) error

// HandleFunc возвращает обработчик NSQ
func (h Handler) HandleFunc(model interface{}) nsq.HandlerFunc {
	return func(msg *nsq.Message) (err error) {
		m := NewMessage(msg, model)
		if err = m.Unmarshal(); err != nil {
			return
		}

		return h(m)
	}
}

// Middleware - обёртка для создания цепочек промежуточной обработки
func (h Handler) Middleware(fn Middleware, args ...interface{}) Handler {
	return fn(h, args...)
}
