package main

import (
	"flag"
	"net/http"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var (
	addr        = flag.String("listen-address", ":9399", "The address to listen on for HTTP requests.")
	squidURL    = flag.String("squid-url", "http://localhost:3128/", "Squid cache manager URL.")
	enableOnly  = flag.String("enable-only", "", "Enable only the specific metrics. Can't be used with '-disable-only'")
	disableOnly = flag.String("disable-only", "", "Disable only the specific metrics. Can't be used with '-enable-only'")

	regexes   = make(map[string]*regexp.Regexp)
	indexMaps = make(map[string]map[string]int)

	mode bool

	enables  = make(map[string]struct{})
	disables = make(map[string]struct{})
)

func init() {
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())
}

func main() {
	flag.Parse()

	if *enableOnly != "" && *disableOnly != "" {
		log.Fatal("You can't use enable-only and disable-only at same time.")
	}

	if *enableOnly != "" {
		mode = true
		for _, metric := range strings.Split(*enableOnly, ",") {
			enables[metric] = struct{}{}
		}
	}

	if *disableOnly != "" {
		mode = false
		for _, metric := range strings.Split(*disableOnly, ",") {
			disables[metric] = struct{}{}
		}
	}

	initActiveRequests()

	log.Infof("Starting Server: %s", *addr)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
		<html lang="es">
		  <head>
		    <title>Squid Exporter</title>
		  </head>
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
