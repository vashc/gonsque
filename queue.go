package gonsque

import (
	"encoding/json"
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Queue struct {
	// Адрес поставщика
	NsqD string
	// Адрес для поиска поставщиков (используется потребителями)
	NsqLookupD string
	// Опции настройки конфига для поставщика/потребителей
	NsqOpts map[string]interface{}
	//TODO: logger

	producer  *nsq.Producer
	consumers []*nsq.Consumer
	nsqConfig *nsq.Config

	initialized bool
}

// Init - инициализация очереди, создание конфига, установка дополнительных опций, создание поставщика.
// Должна быть вызвана до добавления потребителей и работы с сообщениями (публикации)
func (q *Queue) Init() (err error) {
	q.nsqConfig = nsq.NewConfig()

	for opt, val := range q.NsqOpts {
		if err = q.nsqConfig.Set(opt, val); err != nil {
			return errors.Wrap(err, "setting NSQ config option")
		}
	}

	if q.producer, err = nsq.NewProducer(q.NsqD, q.nsqConfig); err != nil {
		return errors.Wrap(err, "creating producer")
	}

	q.initialized = true

	return
}

// AddConsumer добавляет потребителя для работы с очередью, возможно с указанием обработчиков
func (q *Queue) AddConsumer(topic, channel string, handlers ...nsq.HandlerFunc) (consumer *nsq.Consumer, err error) {
	q.assertInitialized()

	if consumer, err = nsq.NewConsumer(topic, channel, q.nsqConfig); err != nil {
		err = errors.Wrap(err, "creating consumer")
		return
	}

	q.consumers = append(q.consumers, consumer)
	for _, h := range handlers {
		consumer.AddHandler(h)
	}

	return
}

// AddConsumerWithConcurrency добавляет потребителя с указанной степенью конкурентности.
// Степень конкурентности показывает количество горутин, используемых при обработке сообщений
func (q *Queue) AddConsumerWithConcurrency(topic, channel string, concurrency int, handler nsq.HandlerFunc) (consumer *nsq.Consumer, err error) {
	q.assertInitialized()

	if consumer, err = nsq.NewConsumer(topic, channel, q.nsqConfig); err != nil {
		err = errors.Wrap(err, "creating concurrent consumer")
		return
	}

	q.consumers = append(q.consumers, consumer)
	consumer.AddConcurrentHandlers(handler, concurrency)
	// Увеличение максимального числа сообщений в обработке для этого потребителя.
	// 10 - эмпирическое число
	consumer.ChangeMaxInFlight(concurrency * 10)

	return
}

// Connect подключает всех потребителей к очереди
func (q *Queue) Connect() (err error) {
	var (
		connect func(consumer *nsq.Consumer, addr string) error
		addr    string
	)

	if q.NsqD == "" {
		connect = (*nsq.Consumer).ConnectToNSQLookupd
		addr = q.NsqLookupD
	} else {
		connect = (*nsq.Consumer).ConnectToNSQD
		addr = q.NsqD
	}

	for i := range q.consumers {
		err = connect(q.consumers[i], addr)
		if errors.Is(err, nsq.ErrAlreadyConnected) {
			continue
		}
		if err != nil {
			return errors.Wrap(err, "connecting consumer to NSQ")
		}
	}

	return
}

// Subscribe добавляет к очереди обработчик с указанной степенью конкурентности
func (q *Queue) Subscribe(sub *Subscriber) (err error) {
	_, err = q.AddConsumerWithConcurrency(sub.topic, sub.channel, sub.concurrency, sub.handler.HandleFunc(sub.model))
	return
}

// Publish - сериализация в JSON и публикация объекта
func (q *Queue) Publish(topic string, obj interface{}) error {
	q.assertInitialized()

	body, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "marshaling JSON")
	}

	return errors.Wrap(q.producer.Publish(topic, body), "publishing to NSQ")
}

// DeferredPublish - сериализация в JSON и отложенная публикация объекта
func (q *Queue) DeferredPublish(topic string, obj interface{}, delay time.Duration) error {
	q.assertInitialized()

	body, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "marshaling JSON")
	}

	return errors.Wrap(
		q.producer.DeferredPublish(topic, delay, body),
		"deferred publishing to NSQ",
	)
}

// BulkPublish - массовая публикация объектов в заданный топик.
// Уменьшает количество раундтрипов и увеличивает пропускную способность очереди
func (q *Queue) BulkPublish(topic string, objs []interface{}) error {
	q.assertInitialized()

	bodies := make([][]byte, len(objs))
	for i := range objs {
		body, err := json.Marshal(objs[i])
		if err != nil {
			return errors.Wrap(err, "marshaling JSON")
		}

		bodies[i] = body
	}

	return errors.Wrap(
		q.producer.MultiPublish(topic, bodies),
		"multi publishing to NSQ",
	)
}

// Start запускает работу очереди с указанными потребителями
func (q *Queue) Start(subs ...*Subscriber) error {
	if err := q.Init(); err != nil {
		return errors.Wrap(err, "initializing queue")
	}

	for i := range subs {
		if err := q.Subscribe(subs[i]); err != nil {
			return errors.Wrap(err, fmt.Sprintf("subscribing to %s/%s", subs[i].topic, subs[i].channel))
		}
	}

	return errors.Wrap(q.Connect(), "connecting consumers")
}

// Stop останавливает работу очереди: отправляет сигнал потребителям/поставщику и ждёт их завершения
func (q *Queue) Stop() {
	var wg sync.WaitGroup

	for i := range q.consumers {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			q.consumers[i].Stop()
			<-q.consumers[i].StopChan
		}(i)
	}

	wg.Wait()

	// Stop() поставщика блокируется до завершения остановки
	q.producer.Stop()

	// После закрытия нельзя пользоваться публикацией, потребители/поставщик завершили работу
	q.initialized = false
}

// Unregister удаляет указанный топик/канал в текущем подключенном nsqd.
// Может быть полезным на teardown этапе при тестировании
func (q *Queue) Unregister(topic, channel string) {
	nsq.UnRegister(topic, channel)
}

func (q *Queue) assertInitialized() {
	// TODO: recover
	if !q.initialized {
		panic("queue is not initialized")
	}
}
