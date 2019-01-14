package main

import (
	"net/url"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/publicsuffix"
)

type activeRequestsCollector struct {
	mutex                 sync.Mutex
	up                    prometheus.Gauge
	activeRequestDataDown *prometheus.Desc
	activeRequestDuration *prometheus.Desc
}

func initActiveRequests() {
	namespace := "squid_active_requests"

	registerMetric(
		whereAmI(),
		`Connection: 0x(?P<ConnID>[a-f0-9]+)\n\s+.+\s+.+\s+.+\s+remote: (?P<IP>[0-9\.]+).+\s+.+\s+.+\nuri (?P<URI>.+)\n.+\n.+out.size (?P<DataDown>\d+)\n.+\n.+\n.+\((?P<Duration>[\d\.]+) seconds ago\)\nusername (?P<Username>.+)\ndelay_pool (?P<DelayPool>\d+)`,
		&activeRequestsCollector{
			up: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "up",
				Help:      "Was the last scrape of squid successful?",
			}),
			activeRequestDataDown: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "data_down"),
				"How much data is downloaded.",
				[]string{"connection", "ip", "uri", "tld", "tld_plus", "username", "delay_pool"},
				nil,
			),
			activeRequestDuration: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "duration"),
				"Time elapsed on connection.",
				[]string{"connection", "ip", "uri", "tld", "tld_plus", "username", "delay_pool"},
				nil,
			),
		},
	)

}

// Describe the prometheus metrics
func (collector *activeRequestsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.activeRequestDataDown
	ch <- collector.activeRequestDuration
	collector.up.Describe(ch)
}

// Collect all metrics
func (collector *activeRequestsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	matches, err := getSquidResponse(whereAmI())

	if err {
		collector.up.Set(0)
		collector.up.Collect(ch)
		return
	}

	indexMap := indexMaps[whereAmI()]

	for _, match := range matches {
		connID := match[indexMap["ConnID"]]
		IP := match[indexMap["IP"]]
		uri := match[indexMap["URI"]]
		user := match[indexMap["Username"]]
		delay := match[indexMap["DelayPool"]]

		tld := "unknown"
		tldPlusOne := "unknown"

		if parser, error := url.Parse(uri); error == nil {
			tld, _ = publicsuffix.PublicSuffix(parser.Hostname())
			tldPlusOne, _ = publicsuffix.EffectiveTLDPlusOne(parser.Hostname())
		}

		ch <- prometheus.MustNewConstMetric(
			collector.activeRequestDataDown,
			prometheus.GaugeValue,
			getFloat(match[indexMap["DataDown"]]),
			connID,
			IP,
			uri,
			tld,
			tldPlusOne,
			user,
			delay,
		)

		ch <- prometheus.MustNewConstMetric(
			collector.activeRequestDuration,
			prometheus.GaugeValue,
			getFloat(match[indexMap["Duration"]]),
			connID,
			IP,
			uri,
			tld,
			tldPlusOne,
			user,
			delay,
		)
	}

	collector.up.Set(1)
	collector.up.Collect(ch)
}
