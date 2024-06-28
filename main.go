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

type CgroupCollector struct {
	hierarchy cgroups.CGMode
	root      string
	patterns  []string
}

func main() {
	var slurm bool
	flag.BoolVar(&slurm, "slurm", false, "whether to collect slurm metrics")

	var collector CgroupCollector
	if slurm {
		collector = *NewSlurmCollector()
	} else {
		collector = *NewUserSliceCollector()
	}

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(collector)
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":2888", router))
}
