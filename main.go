package main

import (
	"flag"
	"log"

	"github.com/containerd/cgroups"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	bind string
	path string
	proc bool
	mode string
}

type cgroupCollector struct {
	config Config
}

func main() {
	var config Config
	flag.StringVar(&config.bind, "bind", ":2112", "port to bind exporter to")
	flag.StringVar(&config.path, "path", "/sys/fs/cgroup", "path of cgroup to export metrics from")
	flag.BoolVar(&config.proc, "proc", false, "whether to export process metrics")
	flag.Parse()

	switch cgroups.Mode() {
	case cgroups.Unified:
		log.Println("cgroups using unified mount")
	case cgroups.Hybrid:
		log.Println("cgroups using hybrid mount")
	case cgroups.Legacy:
		log.Println("cgroups using legacy mount")
	case cgroups.Unavailable:
		log.Fatal("cgroups not mounted")
	}

	cgroupCollector := cgroupCollector{config}
	registry := prometheus.NewRegistry()
	registry.MustRegister(cgroupCollector)
	router := mux.NewRouter
	router.Handle("/metrics", promhttp.HandleFor(registry, promhttp.HandlerOpts{}))
}
