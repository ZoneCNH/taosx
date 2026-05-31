package templatex

type Metrics interface {
	IncCounter(name string, labels map[string]string)
	ObserveHistogram(name string, value float64, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
}

type NoopMetrics struct{}

func (NoopMetrics) IncCounter(name string, labels map[string]string) {}

func (NoopMetrics) ObserveHistogram(name string, value float64, labels map[string]string) {}

func (NoopMetrics) SetGauge(name string, value float64, labels map[string]string) {}
