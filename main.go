package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/containerd/cgroups/v3"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	bind string
	path string
}

type CgroupCollector struct {
	hierarchy cgroups.CGMode
	path      string
}

func main() {
	var conf Config
	flag.StringVar(&conf.bind, "bind", ":2112", "port to bind exporter to")
	flag.StringVar(&conf.path, "path", "user.slice", "path of cgroup to export metrics from")
	flag.Parse()

	collector := CgroupCollector{cgroups.Mode(), conf.path}
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(collector)
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(conf.bind, router))
}
