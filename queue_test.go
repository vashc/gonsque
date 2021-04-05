package gonsque

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

type model struct {
	Title string `json:"title"`
}

func HandleMessage(msg *Message) error {
	log.Print("handling message")
	log.Printf("model: %+v", msg.Body)

	time.Sleep(2 * time.Second)

	return nil
}

func TestQueue(t *testing.T) {
	q := &Queue{
		NsqD:       "localhost:4150",
		NsqLookupD: "localhost:4161",
	}

	if err := q.Init(); err != nil {
		log.Fatal(err)
	}

	var handler Handler = HandleMessage
	handler = handler.
		Middleware(WithTimer)

	if err := q.Subscribe(NewSubscriber("topic", "channel", 2, handler, model{})); err != nil {
		log.Fatal(err)
	}

	if err := q.Connect(); err != nil {
		log.Fatal(err)
	}

	tc := time.NewTicker(time.Second * 1)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-tc.C:
			obj := &model{Title: fmt.Sprintf("wow, it's %s o'clock!", time.Now())}
			if err := q.Publish("topic", obj); err != nil {
				log.Fatal(err)
			}
		case <-ch:
			q.Stop()
			log.Fatal("terminated")
		}
	}
}

func TestStart(t *testing.T) {
	q := &Queue{
		NsqD:       "localhost:4150",
		NsqLookupD: "localhost:4161",
	}

	var handler Handler = HandleMessage
	handler = handler.
		Middleware(WithTimer)

	if err := q.Start(NewSubscriber("topic", "channel", 2, handler, model{})); err != nil {
		log.Fatal(err)
	}

	tc := time.NewTicker(time.Second * 1)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-tc.C:
			obj := &model{Title: fmt.Sprintf("wow, it's %s o'clock!", time.Now())}
			if err := q.Publish("topic", obj); err != nil {
				log.Fatal(err)
			}
		case <-ch:
			q.Stop()
			log.Fatal("terminated")
		}
	}
}
