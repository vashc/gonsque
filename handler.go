package gonsque

import (
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"reflect"
)

type Handler func(msg interface{}) error

// HandleFunc возвращает обработчик NSQ
func (h Handler) HandleFunc() nsq.HandlerFunc {
	return func(msg *nsq.Message) error {
		return h(msg.Body)
	}
}

// Unmarshal создаёт обработчик с автоматической десериализацией в заданную модель
func (h Handler) Unmarshal(model interface{}, next Handler) Handler {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	return func(msg interface{}) (err error) {
		if data, ok := msg.([]byte); ok {
			msg = reflect.New(modelType).Interface()
			if err = json.Unmarshal(data, msg); err != nil {
				return
			}
		}
		return next(msg)
	}
}

// Middleware - обёртка для создания цепочек промежуточной обработки
func (h Handler) Middleware(fn Middleware, args ...interface{}) Handler {
	return fn(h, args...)
}
