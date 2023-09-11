package main

import (
	"log"
	"net/http"
	"os"

	"github.com/containerd/cgroups/v3"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type V2Collector struct {
	host string
}

type V1Collector struct {
	host string
}

func main() {
	host, err := os.Hostname()
	check(err)
	mode := cgroups.Mode()
	var collector prometheus.Collector
	switch mode {
	case cgroups.Hybrid:
		collector = V2Collector{host}
	case cgroups.Unified:
		collector = V2Collector{host}
	case cgroups.Legacy:
		collector = V1Collector{host}
	default:
		log.Fatal("cgroups are not enabled")
	}
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(collector)
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":2112", router))
}
