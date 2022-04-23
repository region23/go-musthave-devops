package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"
)

type gauge float64
type counter int64

type Metric struct {
	Alloc         gauge
	BuckHashSys   gauge
	Frees         gauge
	GCCPUFraction gauge
	GCSys         gauge
	HeapAlloc     gauge
	HeapIdle      gauge
	HeapInuse     gauge
	HeapObjects   gauge
	HeapReleased  gauge
	HeapSys       gauge
	LastGC        gauge
	Lookups       gauge
	MCacheInuse   gauge
	MCacheSys     gauge
	MSpanInuse    gauge
	MSpanSys      gauge
	Mallocs       gauge
	NextGC        gauge
	NumForcedGC   gauge
	NumGC         gauge
	OtherSys      gauge
	PauseTotalNs  gauge
	StackInuse    gauge
	StackSys      gauge
	Sys           gauge
	TotalAlloc    gauge
	PollCount     counter
	RandomValue   gauge
}

func getMetrics(curMetric Metric) Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	curMetric.Alloc = gauge(memStats.Alloc)
	curMetric.BuckHashSys = gauge(memStats.BuckHashSys)
	curMetric.Frees = gauge(memStats.Frees)
	curMetric.GCCPUFraction = gauge(memStats.GCCPUFraction)
	curMetric.GCSys = gauge(memStats.GCSys)
	curMetric.HeapAlloc = gauge(memStats.HeapAlloc)
	curMetric.HeapIdle = gauge(memStats.HeapIdle)
	curMetric.HeapInuse = gauge(memStats.HeapInuse)
	curMetric.HeapObjects = gauge(memStats.HeapObjects)
	curMetric.HeapReleased = gauge(memStats.HeapReleased)
	curMetric.HeapSys = gauge(memStats.HeapSys)
	curMetric.LastGC = gauge(memStats.LastGC)
	curMetric.Lookups = gauge(memStats.Lookups)
	curMetric.MCacheInuse = gauge(memStats.MCacheInuse)
	curMetric.MCacheSys = gauge(memStats.MCacheSys)
	curMetric.MSpanInuse = gauge(memStats.MSpanInuse)
	curMetric.MSpanSys = gauge(memStats.MSpanSys)
	curMetric.Mallocs = gauge(memStats.Mallocs)
	curMetric.NextGC = gauge(memStats.NextGC)
	curMetric.NumForcedGC = gauge(memStats.NumForcedGC)
	curMetric.NumGC = gauge(memStats.NumGC)
	curMetric.OtherSys = gauge(memStats.OtherSys)
	curMetric.PauseTotalNs = gauge(memStats.PauseTotalNs)
	curMetric.StackInuse = gauge(memStats.StackInuse)
	curMetric.StackSys = gauge(memStats.StackSys)
	curMetric.Sys = gauge(memStats.Sys)
	curMetric.TotalAlloc = gauge(memStats.TotalAlloc)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	curMetric.RandomValue = gauge(r1.Float64())
	curMetric.PollCount = curMetric.PollCount + 1

	return curMetric
}

func sendMetric(mType string, mName string, mValue string) (*http.Request, error) {
	//fmt.Printf("%v | %v | %v\n", mType, mName, mValue)
	url := fmt.Sprintf("http://127.0.0.1:8080/update/%v/%v/%v", mType, mName, mValue)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	request.Header.Set("Content-Type", "text/plain")
	if err != nil {
		return nil, err
	}

	return request, nil
}

func report(curMetric Metric) {
	var mType, mName, mValue string
	v := reflect.ValueOf(curMetric)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Interface().(type) {
		case gauge:
			mType = "gauge"
		case counter:
			mType = "counter"
		}

		mValue = fmt.Sprintf("%v", v.Field(i).Interface())
		mName = fmt.Sprintf("%v", typeOfS.Field(i).Name)

		_, err := sendMetric(mType, mName, mValue)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	var curMetric Metric
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
			report(curMetric)
		case <-osSigChan:
			os.Exit(0)
		}
	}
}
