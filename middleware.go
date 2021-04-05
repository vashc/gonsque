package gonsque

import (
	"log"
	"time"
)

type Middleware func(Handler, ...interface{}) Handler

func WithTimer(h Handler, args ...interface{}) Handler {
	return func(msg *Message) (err error) {
		startTime := time.Now()
		log.Printf("start: %s", startTime)
		err = h(msg)
		log.Printf("finish: %s", time.Now())
		log.Printf("elapsed time: %d ms", time.Since(startTime).Milliseconds())
		return
	}
}

func WithNotifier(h Handler, args ...interface{}) Handler {
	ch := args[0].(chan struct{})

	return func(msg *Message) (err error) {
		err = h(msg)
		ch <- struct{}{}
		return
	}
}
