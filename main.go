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
	root string
	uids bool
	jobs bool
}

type CgroupCollector struct {
	hierarchy cgroups.CGMode
	config    Config
}

func main() {
	var conf Config
	flag.StringVar(&conf.root, "path", "/slurm", "path of cgroup to export")
	flag.BoolVar(&conf.uids, "export-uid-data", true, "whether or not to export nested user metrics")
	flag.BoolVar(&conf.jobs, "export-job-data", true, "whether or not to export nested job metrics")
	flag.Parse()

	collector := CgroupCollector{cgroups.Mode(), conf}
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(collector)
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":2112", router))
}
