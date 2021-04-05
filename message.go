package gonsque

import (
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"reflect"
	"time"
)

// Message - обёртка поверх nsq сообщения с поддержкой десериализации в указанную модель
type Message struct {
	msg   *nsq.Message
	model interface{}

	Body interface{}
}

type Messenger interface {
	Requeue(delay time.Duration)
	RequeueWithoutBackoff(delay time.Duration)

	ID() string
	Attempts() uint16

	Unmarshal() error
}

func NewMessage(msg *nsq.Message, model interface{}) *Message {
	return &Message{
		msg,
		model,
		nil,
	}
}

func (m *Message) Requeue(delay time.Duration) {
	m.msg.Requeue(delay)
}

func (m *Message) RequeueWithoutBackoff(delay time.Duration) {
	m.msg.RequeueWithoutBackoff(delay)
}

func (m *Message) ID() string {
	return string(m.msg.ID[:])
}

func (m *Message) Attempts() uint16 {
	return m.msg.Attempts
}

func (m *Message) Unmarshal() (err error) {
	modelType := reflect.TypeOf(m.model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	m.Body = reflect.New(modelType).Interface()
	if err = json.Unmarshal(m.msg.Body, m.Body); err != nil {
		return
	}

	return
}
