package main

import (
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	resty "gopkg.in/resty.v0"
)

func getFloat(value string) float64 {
	float, _ := strconv.ParseFloat(value, 64)
	return float
}

func canInit(metric string) bool {
	_, okEnabled := enables[metric]
	_, okDisabled := disables[metric]

	if (mode == true && !okEnabled) || (mode == false && okDisabled) {
		return false
	}

	return true
}

func whereAmI() string {
	_, file, _, _ := runtime.Caller(1)
	return chopPath(file)
}

// return the source filename after the last slash
func chopPath(original string) string {
	i := strings.LastIndex(original, "/")
	j := strings.LastIndex(original, ".")

	if i == -1 {
		return original
	}

	return original[i+1 : j]
}

func registerMetric(name string, regex string, collector prometheus.Collector) {
	if canInit(name) {
		regexes[name] = regexp.MustCompile(regex)
		indexMaps[name] = make(map[string]int)

		for i, groupName := range regexes[name].SubexpNames() {
			if i != 0 && groupName != "" {
				indexMaps[name][groupName] = i
			}
		}

		prometheus.MustRegister(collector)
	}
}

func getSquidResponse(metric string) (matches [][]string, err bool) {
	resp, errResty := resty.R().Get("http://" + exporter_config.SquidHost + ":" + strconv.Itoa(exporter_config.SquidPort) + "/squid-internal-mgr/" + metric)

	if errResty != nil {
		log.Errorf("Error scraping squid '%s': %v", metric, err)
		return nil, true
	}

	return regexes[metric].FindAllStringSubmatch(resp.String(), -1), false
}
