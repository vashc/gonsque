package gonsque

import (
	"encoding/json"
	"log"
	"reflect"
	"time"
)

type Middleware func(Handler, ...interface{}) Handler

func WithTimer(h Handler, args ...interface{}) Handler {
	return func(msg interface{}) (err error) {
		startTime := time.Now()
		log.Printf("start: %s", startTime)
		err = h(msg)
		log.Printf("finish: %s", time.Now())
		log.Printf("elapsed time: %d", time.Since(startTime).Milliseconds())
		return
	}
}

func WithModel(h Handler, args ...interface{}) Handler {
	if len(args) == 0 {
		return h
	}

	modelType := reflect.TypeOf(args[0])
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
		return h(msg)
	}
}
