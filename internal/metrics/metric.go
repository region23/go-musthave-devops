package metrics

type Gauge float64
type Counter int64

type Metric struct {
	Alloc           Gauge
	BuckHashSys     Gauge
	Frees           Gauge
	GCCPUFraction   Gauge
	GCSys           Gauge
	HeapAlloc       Gauge
	HeapIdle        Gauge
	HeapInuse       Gauge
	HeapObjects     Gauge
	HeapReleased    Gauge
	HeapSys         Gauge
	LastGC          Gauge
	Lookups         Gauge
	MCacheInuse     Gauge
	MCacheSys       Gauge
	MSpanInuse      Gauge
	MSpanSys        Gauge
	Mallocs         Gauge
	NextGC          Gauge
	NumForcedGC     Gauge
	NumGC           Gauge
	OtherSys        Gauge
	PauseTotalNs    Gauge
	StackInuse      Gauge
	StackSys        Gauge
	Sys             Gauge
	TotalAlloc      Gauge
	PollCount       Counter
	RandomValue     Gauge
	TotalMemory     Gauge
	FreeMemory      Gauge
	CPUutilization1 Gauge
}
