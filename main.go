package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/jinzhu/configor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	regexes   = make(map[string]*regexp.Regexp)
	indexMaps = make(map[string]map[string]int)

	mode bool

	enables  = make(map[string]struct{})
	disables = make(map[string]struct{})

	exporterConfig = newConfig()
)

func init() {
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())
}

func main() {
	if *versionFlag {
		fmt.Println(version.Print("squid_prometheus_exporter"))
		os.Exit(0)
	}

	os.Setenv("CONFIGOR_ENV_PREFIX", "-")

	if *configFile != "" {
		configor.Load(&exporterConfig, *configFile)
	}

	if exporterConfig.EnableOnly != "" && exporterConfig.DisableOnly != "" {
		log.Fatal("You can't use enable-only and disable-only at same time.")
	}

	if exporterConfig.EnableOnly != "" {
		mode = true
		for _, metric := range strings.Split(exporterConfig.EnableOnly, ",") {
			enables[metric] = struct{}{}
		}
	}

	if exporterConfig.DisableOnly != "" {
		mode = false
		for _, metric := range strings.Split(exporterConfig.DisableOnly, ",") {
			disables[metric] = struct{}{}
		}
	}

	initActiveRequests()
	// initInfoRequests()

	log.Infof("Starting Server: %s", exporterConfig.ListenAddress)

	http.Handle(exporterConfig.MetricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
		<html lang="es">
		  <head>
		    <title>Squid Exporter</title>
		  </head>
		  <body>
		    <h1>Squid Exporter</h1>
		    <p><a href="` + exporterConfig.MetricPath + `">Metrics</a></p>
			</body>
		  </html>`))
	})
	http.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	log.Fatal(http.ListenAndServe(exporterConfig.ListenAddress, nil))
}
