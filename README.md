# Prometheus exporter for Squid Proxy
[![Build Status](https://travis-ci.com/lfdominguez/squid_prometheus_exporter.svg?branch=master)](https://travis-ci.com/lfdominguez/squid_prometheus_exporter)
[![](https://img.shields.io/github/release/lfdominguez/squid_prometheus_exporter.svg)](https://github.com/lfdominguez/squid_prometheus_exporter/releases)
![](https://img.shields.io/github/license/lfdominguez/squid_prometheus_exporter.svg)
![](https://img.shields.io/github/downloads/lfdominguez/squid_prometheus_exporter/total.svg)
![](https://img.shields.io/github/release-date/lfdominguez/squid_prometheus_exporter.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/lfdominguez/squid_prometheus_exporter)](https://goreportcard.com/report/github.com/lfdominguez/squid_prometheus_exporter)
[![Maintainability](https://api.codeclimate.com/v1/badges/4e10de65bb2e9fce7d2e/maintainability)](https://codeclimate.com/github/lfdominguez/squid_prometheus_exporter/maintainability)

This project try to extract all the stats from manager page of Squid > 3.5 (in this version the squid manager can be acceded from HTTP endpoint directly using `/squid-internal-mgr/`). The metrics got from Squid are:

 1. Active Requests

## Active Requests

Print all the information about the current connections of Squid proxy, the metrics are:

 * `squid_active_requests_data_down` Show data transfered of a connection
 * `squid_active_requests_duration` Show duration of a connection

The labels are:

 * `connection` ID of the connection
 * `ip` IP of source connection
 * `uri` URL of request
 * `username` Username of the source connection
 * `delay_pool` Delay pool matched for the source connection

## How to use

First you need configure squid to allow the requests to cache manager and later execute the exporter.

### Configure Squid

You need configure in `squid.conf` the access to cache manager with this:

```
http_access allow localhost manager
http_access deny manager
```

You can change `localhost` with any other acl to allow the access. More info at [Squid Manager Doc](https://wiki.squid-cache.org/Features/CacheManager)

### Execute exporter

Compile the Go source code

```bash
go build -o squid_exporter .
```

Then execute with these parameters:

 * `-listen-address` The address to listen on for HTTP requests. (default ':9399')
 * `-squid-url` Squid URL. (default 'http://localhost:3128/')
 * `-enable-only` Enable only specific metrics. Can't be used with `-disable-only`
 * `-disable-only` Disable only specific metrics. Can't be used with `-enable-only`