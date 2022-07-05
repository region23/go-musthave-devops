package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/region23/go-musthave-devops/internal/metrics"
	"github.com/region23/go-musthave-devops/internal/serializers"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	Key            string        `env:"KEY"`
}

var cfg Config = Config{}

func init() {
	flag.StringVar(&cfg.Address, "a", "127.0.0.1:8080", "server address")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "report interval")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "poll interval")
	flag.StringVar(&cfg.Key, "k", "", "key for hashing")
}

func getMetrics(curMetric *metrics.Metric) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	curMetric.Alloc = metrics.Gauge(memStats.Alloc)
	curMetric.BuckHashSys = metrics.Gauge(memStats.BuckHashSys)
	curMetric.Frees = metrics.Gauge(memStats.Frees)
	curMetric.GCCPUFraction = metrics.Gauge(memStats.GCCPUFraction)
	curMetric.GCSys = metrics.Gauge(memStats.GCSys)
	curMetric.HeapAlloc = metrics.Gauge(memStats.HeapAlloc)
	curMetric.HeapIdle = metrics.Gauge(memStats.HeapIdle)
	curMetric.HeapInuse = metrics.Gauge(memStats.HeapInuse)
	curMetric.HeapObjects = metrics.Gauge(memStats.HeapObjects)
	curMetric.HeapReleased = metrics.Gauge(memStats.HeapReleased)
	curMetric.HeapSys = metrics.Gauge(memStats.HeapSys)
	curMetric.LastGC = metrics.Gauge(memStats.LastGC)
	curMetric.Lookups = metrics.Gauge(memStats.Lookups)
	curMetric.MCacheInuse = metrics.Gauge(memStats.MCacheInuse)
	curMetric.MCacheSys = metrics.Gauge(memStats.MCacheSys)
	curMetric.MSpanInuse = metrics.Gauge(memStats.MSpanInuse)
	curMetric.MSpanSys = metrics.Gauge(memStats.MSpanSys)
	curMetric.Mallocs = metrics.Gauge(memStats.Mallocs)
	curMetric.NextGC = metrics.Gauge(memStats.NextGC)
	curMetric.NumForcedGC = metrics.Gauge(memStats.NumForcedGC)
	curMetric.NumGC = metrics.Gauge(memStats.NumGC)
	curMetric.OtherSys = metrics.Gauge(memStats.OtherSys)
	curMetric.PauseTotalNs = metrics.Gauge(memStats.PauseTotalNs)
	curMetric.StackInuse = metrics.Gauge(memStats.StackInuse)
	curMetric.StackSys = metrics.Gauge(memStats.StackSys)
	curMetric.Sys = metrics.Gauge(memStats.Sys)
	curMetric.TotalAlloc = metrics.Gauge(memStats.TotalAlloc)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	curMetric.RandomValue = metrics.Gauge(r1.Float64())
	curMetric.PollCount += 1
}

// Отправляем метрику на сервер
func sendMetric(metricsToSend []serializers.Metrics) error {
	u := url.URL{
		Scheme: "http",
		Host:   cfg.Address,
		Path:   "updates",
	}

	postBody, err := json.Marshal(metricsToSend)

	if err != nil {
		return err
	}

	responseBody := bytes.NewBuffer(postBody)
	request, err := http.NewRequest(http.MethodPost, u.String(), responseBody)
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	client := &http.Client{}
	// отправляем запрос
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	// печатаем код ответа
	log.Debug().Msgf("Статус-код %v", response.Status)
	defer response.Body.Close()

	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка при чтении ответа")
	}
	// и печатаем его
	log.Debug().Msg(string(body))

	return nil
}

// Отправка метрик на сервер
func report(curMetric *metrics.Metric, key string) {
	metricsBatch := []serializers.Metrics{}

	var mType, mName, mValue string
	v := reflect.ValueOf(curMetric)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Interface().(type) {
		case metrics.Gauge:
			mType = "gauge"
		case metrics.Counter:
			mType = "counter"
		}

		mValue = fmt.Sprintf("%v", v.Field(i).Interface())
		mName = fmt.Sprintf("%v", typeOfS.Field(i).Name)

		metricToSend := serializers.NewMetrics(mName, mType)

		if mType == "gauge" {
			if s, err := strconv.ParseFloat(mValue, 64); err == nil {
				metricToSend.Value = &s
				if key != "" {
					metricToSend.Hash = serializers.Hash(mType, mName, fmt.Sprintf("%f", s), key)
				}

			} else {
				log.Panic().Err(err).Msg("Ошибка при парсинге числа метрики")
			}

		} else if mType == "counter" {
			if s, err := strconv.ParseInt(mValue, 10, 64); err == nil {
				metricToSend.Delta = &s
				if key != "" {
					metricToSend.Hash = serializers.Hash(mType, mName, fmt.Sprintf("%d", s), key)
				}
			} else {
				log.Panic().Err(err).Msg("Ошибка при парсинге числа счетчика метрики")
			}
		}

		metricsBatch = append(metricsBatch, metricToSend)

		err := sendMetric(metricsBatch)
		if err != nil {
			log.Error().Err(err).Msg("Ошибка при отправке метрики на сервер")
		}
	}

	curMetric.PollCount = 1
}

func main() {
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Error().Err(err).Msgf("%+v\n", err)
	}

	curMetric := new(metrics.Metric)

	go getMetrics(curMetric)

	osSigChan := make(chan os.Signal, 1)
	signal.Notify(osSigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	pollTick := time.NewTicker(cfg.PollInterval)
	reportTick := time.NewTicker(cfg.ReportInterval)
	for {
		select {
		case <-pollTick.C:
			go getMetrics(curMetric)
		case <-reportTick.C:
			go report(curMetric, cfg.Key)
		case <-osSigChan:
			os.Exit(0)
		}
	}
}
