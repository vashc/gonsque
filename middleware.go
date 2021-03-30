package gonsque

import (
	"log"
	"time"
)

type Middleware func(Handler) Handler

func WithTimer(h Handler) Handler {
	return func(msg interface{}) (err error) {
		startTime := time.Now()
		log.Printf("start: %s", startTime)
		err = h(msg)
		log.Printf("finish: %s", time.Now())
		log.Printf("elapsed time: %d", time.Since(startTime).Milliseconds())
		return
	}
}

func WithFoo(h Handler) Handler {
	return func(msg interface{}) (err error) {
		log.Print("foo")
		return h(msg)
	}
}
