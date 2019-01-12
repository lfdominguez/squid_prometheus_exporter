package main

import (
	"flag"
)

const (
	defaultListenAddress = ":9399"
	defaultMetricsPath   = "/metrics"
	defaultSquidHost     = "localhost"
	defaultSquidPort     = 3128
)

var (
	versionFlag *bool
	configFile  *string
)

type config struct {
	ListenAddress string `yaml:"listen-address"`
	MetricPath    string `yaml:"metrics-path"`

	SquidHost string `yaml:"squid-host"`
	SquidPort int    `yaml:"squid-port"`

	EnableOnly  string `yaml:"enable-only"`
	DisableOnly string `yaml:"disable-only"`
}

func newConfig() *config {
	c := &config{}

	flag.StringVar(&c.ListenAddress, "listen-address", defaultListenAddress, "The address to listen on for HTTP requests.")
	flag.StringVar(&c.MetricPath, "metrics-path", defaultMetricsPath, "Metrics path to expose prometheus metrics.")

	flag.StringVar(&c.SquidHost, "squid-host", defaultSquidHost, "Squid address or hostname.")
	flag.IntVar(&c.SquidPort, "squid-port", defaultSquidPort, "Squid port")

	flag.StringVar(&c.EnableOnly, "enable-only", "", "Enable only the specific metrics. Can't be used with '-disable-only'")
	flag.StringVar(&c.DisableOnly, "disable-only", "", "Disable only the specific metrics. Can't be used with '-enable-only'")

	configFile = flag.String("config-file", "config.yml", "Configuration file")
	versionFlag = flag.Bool("version", false, "Print the version and exit.")

	flag.Parse()

	return c
}
