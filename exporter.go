package main

import (
	"flag"
	"net/http"
	"regexp"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	resty "gopkg.in/resty.v0"
)

const (
	namespace = "squid" // For Prometheus metrics.
)

var (
	addr     = flag.String("listen-address", ":9399", "The address to listen on for HTTP requests.")
	squidURL = flag.String("squid-url", "http://localhost:3128/squid-internal-mgr/active_requests", "Squid cache manager active requests URL.")

	regex    = regexp.MustCompile(`Connection: 0x(?P<ConnID>[a-f0-9]+)\n\s+.+\s+.+\s+.+\s+remote: (?P<IP>[0-9\.]+).+\s+.+\s+.+\nuri (?P<URI>.+)\n.+\n.+out.size (?P<DataDown>\d+)\n.+\n.+\n.+\((?P<Duration>[\d\.]+) seconds ago\)\nusername (?P<Username>.+)\ndelay_pool (?P<DelayPool>\d+)`)
	indexMap = make(map[string]int)
)

type squidCollector struct {
	URL                   string
	mutex                 sync.Mutex
	up                    prometheus.Gauge
	activeRequestDataDown *prometheus.Desc
	activeRequestDuration *prometheus.Desc
}

func newSquidCollector(url string) *squidCollector {
	return &squidCollector{
		URL: url,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of squid successful.",
		}),
		activeRequestDataDown: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "active_requests_data_down"),
			"How much data is downloaded.",
			[]string{"connection", "ip", "uri", "username", "delay_pool"},
			nil,
		),
		activeRequestDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "active_requests_duration"),
			"Time elapsed on connection.",
			[]string{"connection", "ip", "uri", "username", "delay_pool"},
			nil,
		),
	}
}

// Describe the prometheus metrics
func (collector *squidCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.activeRequestDataDown
	ch <- collector.activeRequestDuration
	collector.up.Describe(ch)
}

// Collect all metrics
func (collector *squidCollector) Collect(ch chan<- prometheus.Metric) {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	resp, err := resty.R().Get(collector.URL)

	if err != nil {
		log.Errorf("Error scraping squid active requests: %v", err)
		collector.up.Set(0)
		collector.up.Collect(ch)
		return
	}

	matches := regex.FindAllStringSubmatch(resp.String(), -1)

	for _, match := range matches {
		ch <- prometheus.MustNewConstMetric(
			collector.activeRequestDataDown,
			prometheus.GaugeValue,
			getFloat(match[indexMap["DataDown"]]),
			match[indexMap["ConnID"]],
			match[indexMap["IP"]],
			match[indexMap["URI"]],
			match[indexMap["Username"]],
			match[indexMap["DelayPool"]],
		)

		ch <- prometheus.MustNewConstMetric(
			collector.activeRequestDuration,
			prometheus.GaugeValue,
			getFloat(match[indexMap["Duration"]]),
			match[indexMap["ConnID"]],
			match[indexMap["IP"]],
			match[indexMap["URI"]],
			match[indexMap["Username"]],
			match[indexMap["DelayPool"]],
		)
	}

	collector.up.Set(1)
	collector.up.Collect(ch)
}

func init() {
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())
}

func main() {
	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" {
			indexMap[name] = i
		}
	}

	flag.Parse()
	exporter := newSquidCollector(*squidURL)
	prometheus.MustRegister(exporter)

	log.Infof("Starting Server: %s", *addr)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Squid Exporter</title></head>
			<body>
			<h1>Squid Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func getFloat(value string) float64 {
	float, _ := strconv.ParseFloat(value, 64)
	return float
}
