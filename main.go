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

	exporter_config = newConfig()
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
		configor.Load(&exporter_config, *configFile)
	}

	if exporter_config.EnableOnly != "" && exporter_config.DisableOnly != "" {
		log.Fatal("You can't use enable-only and disable-only at same time.")
	}

	if exporter_config.EnableOnly != "" {
		mode = true
		for _, metric := range strings.Split(exporter_config.EnableOnly, ",") {
			enables[metric] = struct{}{}
		}
	}

	if exporter_config.DisableOnly != "" {
		mode = false
		for _, metric := range strings.Split(exporter_config.DisableOnly, ",") {
			disables[metric] = struct{}{}
		}
	}

	initActiveRequests()

	log.Infof("Starting Server: %s", exporter_config.ListenAddress)

	http.Handle(exporter_config.MetricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
		<html lang="es">
		  <head>
		    <title>Squid Exporter</title>
		  </head>
		  <body>
		    <h1>Squid Exporter</h1>
		    <p><a href="` + exporter_config.MetricPath + `">Metrics</a></p>
			</body>
		  </html>`))
	})
	http.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	log.Fatal(http.ListenAndServe(exporter_config.ListenAddress, nil))
}
