package statsd

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/segmentio/stats"
	"github.com/segmentio/stats/netstats"
)

type Config struct {
	Network       string
	Address       string
	BufferSize    int
	QueueSize     int
	RetryAfterMin time.Duration
	RetryAfterMax time.Duration
	FlushTimeout  time.Duration
	WriteTimeout  time.Duration
	Dial          func(string, string) (net.Conn, error)
	Fail          func(error)
	Rand          func() float64
}

func NewBackend(addr string) stats.Backend {
	network, address := netstats.SplitNetworkAddress(addr)
	return NewBackendWith(Config{
		Network: network,
		Address: address,
	})
}

func NewBackendWith(config Config) stats.Backend {
	config = setConfigDefaults(config)
	return netstats.NewBackendWith(netstats.Config{
		Protocol:      protocol{},
		Network:       config.Network,
		Address:       config.Address,
		BufferSize:    config.BufferSize,
		QueueSize:     config.QueueSize,
		RetryAfterMin: config.RetryAfterMin,
		RetryAfterMax: config.RetryAfterMax,
		FlushTimeout:  config.FlushTimeout,
		WriteTimeout:  config.WriteTimeout,
		Dial:          config.Dial,
		Fail:          config.Fail,
		Rand:          config.Rand,
	})
}

func setConfigDefaults(config Config) Config {
	if len(config.Network) == 0 {
		config.Network = "udp"
	}

	if len(config.Address) == 0 {
		config.Address = "localhost"
	}

	if _, port, _ := net.SplitHostPort(config.Address); len(port) == 0 {
		config.Address = net.JoinHostPort(config.Address, "8125")
	}

	return config
}

type protocol struct{}

func (p protocol) WriteSet(w io.Writer, m stats.Metric, v float64, t time.Time) error {
	return p.write(w, Gauge, m, v)
}

func (p protocol) WriteAdd(w io.Writer, m stats.Metric, v float64, t time.Time) error {
	return p.write(w, Counter, m, v)
}

func (p protocol) WriteObserve(w io.Writer, m stats.Metric, v float64, t time.Time) error {
	return p.write(w, Histogram, m, v)
}

func (p protocol) write(w io.Writer, t MetricType, m stats.Metric, v float64) (err error) {
	_, err = fmt.Fprint(w, Metric{
		Name:       sanitize(m.Name()),
		Value:      int64(v),
		Type:       t,
		SampleRate: SampleRate(m.Sample()),
	})
	return
}

func sanitize(s string) string {
	s = replace(s, ",")
	s = replace(s, ":")
	s = replace(s, "|")
	s = replace(s, "@")
	s = replace(s, "#")
	return s
}

func replace(s string, b string) string {
	if strings.IndexByte(s, b[0]) >= 0 {
		s = strings.Replace(s, b, "_", -1)
	}
	return s
}
