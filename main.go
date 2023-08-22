package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/containerd/cgroups/v3"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	root     string
	patterns []string
}

type CgroupCollector struct {
	hierarchy cgroups.CGMode
	config    Config
}

func main() {
	var conf Config
	var s string
	flag.StringVar(&conf.root, "root", "/slurm", "path of the root cgroup to export, default is /slurm")
	flag.StringVar(&s, "sub-cgroup-patterns", "/uid_* /uid_*/job_*", "patterns of sub cgroups to export underneath the root cgroup, defaults are /uid_* and /uid_*/job_*")
	flag.Parse()
	conf.patterns = strings.Split(s, " ")

	collector := CgroupCollector{cgroups.Mode(), conf}
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(collector)
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":2112", router))
}
