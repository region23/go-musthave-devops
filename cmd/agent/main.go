package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"

	"github.com/region23/go-musthave-devops/internal/metrics"
)

func getMetrics(curMetric metrics.Metric) metrics.Metric {
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
	curMetric.PollCount = curMetric.PollCount + 1

	return curMetric
}

func sendMetric(mType string, mName string, mValue string) error {
	//fmt.Printf("%v | %v | %v\n", mType, mName, mValue)
	url := fmt.Sprintf("http://127.0.0.1:8080/update/%v/%v/%v", mType, mName, mValue)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	request.Header.Set("Content-Type", "text/plain")
	if err != nil {
		return err
	}

	client := &http.Client{}
	// отправляем запрос
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	// печатаем код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	// и печатаем его
	fmt.Println(string(body))

	return nil
}

func report(curMetric metrics.Metric) metrics.Metric {
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

		err := sendMetric(mType, mName, mValue)
		if err != nil {
			fmt.Println(err)
		}
	}

	curMetric.PollCount = 1
	return curMetric
}

func main() {
	var curMetric metrics.Metric
	curMetric = getMetrics(curMetric)

	osSigChan := make(chan os.Signal, 1)
	signal.Notify(osSigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	pollInterval := time.NewTicker(2 * time.Second)
	reportInterval := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-pollInterval.C:
			curMetric = getMetrics(curMetric)
		case <-reportInterval.C:
			curMetric = report(curMetric)
		case <-osSigChan:
			os.Exit(0)
		}
	}
}
