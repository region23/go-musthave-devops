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
	"runtime"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/region23/go-musthave-devops/internal/serializers"
	"github.com/rs/zerolog/log"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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

func getMainMetrics(metrics *serializers.Metrics) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics.Add("Alloc", "gauge", memStats.Alloc)
	metrics.Add("BuckHashSys", "gauge", memStats.BuckHashSys)
	metrics.Add("Frees", "gauge", memStats.Frees)
	metrics.Add("GCCPUFraction", "gauge", memStats.GCCPUFraction)
	metrics.Add("GCSys", "gauge", memStats.GCSys)
	metrics.Add("HeapAlloc", "gauge", memStats.HeapAlloc)
	metrics.Add("HeapIdle", "gauge", memStats.HeapIdle)
	metrics.Add("HeapInuse", "gauge", memStats.HeapInuse)
	metrics.Add("HeapObjects", "gauge", memStats.HeapObjects)
	metrics.Add("HeapReleased", "gauge", memStats.HeapReleased)
	metrics.Add("HeapSys", "gauge", memStats.HeapSys)
	metrics.Add("LastGC", "gauge", memStats.LastGC)
	metrics.Add("Lookups", "gauge", memStats.Lookups)
	metrics.Add("MCacheInuse", "gauge", memStats.MCacheInuse)
	metrics.Add("MCacheSys", "gauge", memStats.MCacheSys)
	metrics.Add("MSpanInuse", "gauge", memStats.MSpanInuse)
	metrics.Add("MSpanSys", "gauge", memStats.MSpanSys)
	metrics.Add("Mallocs", "gauge", memStats.Mallocs)
	metrics.Add("NextGC", "gauge", memStats.NextGC)
	metrics.Add("NumForcedGC", "gauge", memStats.NumForcedGC)
	metrics.Add("NumGC", "gauge", memStats.NumGC)
	metrics.Add("OtherSys", "gauge", memStats.OtherSys)
	metrics.Add("PauseTotalNs", "gauge", memStats.PauseTotalNs)
	metrics.Add("StackInuse", "gauge", memStats.StackInuse)
	metrics.Add("StackSys", "gauge", memStats.StackSys)
	metrics.Add("Sys", "gauge", memStats.Sys)
	metrics.Add("TotalAlloc", "gauge", memStats.TotalAlloc)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	metrics.Add("RandomValue", "gauge", r1.Float64())

	pollCount, exist := metrics.Get("PollCount")
	var val float64 = 1
	if exist {
		val = *pollCount.Value + 1
	}
	metrics.Add("PollCount", "gauge", val)
}

func getGopsUitilMetrics(metrics *serializers.Metrics) {
	cpuUtilization, err := cpu.Percent(0, true)
	if err != nil {
		log.Error().Err(err).Msg("При получении процента загрузки процессоров возникла ошибка")
	}

	for i := 0; i < len(cpuUtilization); i++ {
		metrics.Add(fmt.Sprintf("%s%d", "CPUutilization", i+1), "gauge", cpuUtilization[i])
	}

	v, err := mem.VirtualMemory()

	if err != nil {
		log.Error().Err(err).Msg("При получении данных о виртуальной памяти возникла ошибка")
	}

	metrics.Add("TotalMemory", "gauge", v.Total)
	metrics.Add("FreeMemory", "gauge", v.Free)
}

// Отправляем метрику на сервер
func sendMetric(metrics *serializers.Metrics) error {
	u := url.URL{
		Scheme: "http",
		Host:   cfg.Address,
		Path:   "updates",
	}

	postBody, err := json.Marshal(metrics.GetAll())

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

	// После отправки сбрасываем счётчик
	metrics.Add("PollCount", "gauge", 1)

	return nil
}

func main() {
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Error().Err(err).Msgf("%+v\n", err)
	}

	metrics := serializers.InitMetrics(cfg.Key)

	osSigChan := make(chan os.Signal, 1)
	signal.Notify(osSigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	pollTick := time.NewTicker(cfg.PollInterval)
	reportTick := time.NewTicker(cfg.ReportInterval)
	for {
		select {
		case <-pollTick.C:
			go getMainMetrics(metrics)
			go getGopsUitilMetrics(metrics)
		case <-reportTick.C:
			go sendMetric(metrics)
		case <-osSigChan:
			os.Exit(0)
		}
	}

}
