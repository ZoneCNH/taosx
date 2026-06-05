package taosx

import "time"

type Option func(*options)

type options struct {
	driver  Driver
	metrics Metrics
	clock   func() time.Time
}

func defaultOptions() options {
	return options{
		driver:  unavailableDriver{},
		metrics: NoopMetrics{},
		clock:   time.Now,
	}
}

func WithDriver(driver Driver) Option {
	return func(o *options) {
		if driver != nil {
			o.driver = driver
		}
	}
}

func WithMetrics(metrics Metrics) Option {
	return func(o *options) {
		if metrics != nil {
			o.metrics = metrics
		}
	}
}

func WithClock(clock func() time.Time) Option {
	return func(o *options) {
		if clock != nil {
			o.clock = clock
		}
	}
}
