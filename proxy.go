package gonsque

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"net/http"
)

var (
	counterMsgs prometheus.Counter
)

// ProxyPublish - конструктор http.Handler, который отправляет содержимое POST запроса в соответствующий топик
func (q *Queue) ProxyPublish(topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = q.Publish(topic, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ProxyMetrics - конструктор http.Handler для сбора статистики в Prometheus по работе каналов в указанном топике
func (q *Queue) ProxyMetrics(topic string) http.Handler {
	gaugeMsgs := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "golang_nsq",
			Name: fmt.Sprintf("nsq_%s_gauge_msg", topic),
		},
	)

	counterMsgs = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "golang_nsq",
			Name: fmt.Sprintf("nsq_%s_counter_msg", topic),
		},
	)

	prometheus.MustRegister(gaugeMsgs)
	prometheus.MustRegister(counterMsgs)

	return promhttp.Handler()
}
