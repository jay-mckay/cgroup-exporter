package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var slurmMetricLabels = []string{"uid", "jobid", "host"}
var slurmMetricPrefix = "slurm_job_"
var Metrics = map[string]Metric{
	"kernel_cpu": {prometheus.NewDesc(slurmMetricPrefix+"cpu_kernel_ns", "kernel cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
	"user_cpu":   {prometheus.NewDesc(slurmMetricPrefix+"cpu_user_ns", "user cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
	"total_cpu":  {prometheus.NewDesc(slurmMetricPrefix+"cpu_total_ns", "total cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
	"mem_rss":    {prometheus.NewDesc(slurmMetricPrefix+"cpu_total_ns", "total cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
}

type Metric struct {
	promDesc *prometheus.Desc
	promType prometheus.ValueType
}

type Stat struct {
	name  string
	value uint64
}

type Exporter interface {
	GetRelativeJobPaths() []string
	GetStats(string) []Stat
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getRelativeJobPaths(root string) []string {
	paths, err := filepath.Glob(root + "/slurm/uid_*/job_*")
	check(err)
	var cgroups []string
	for _, p := range paths {
		// TODO: REGEX please
		tokens := strings.Split(p, "/slurm") // return path relative to slurm
		cgroups = append(cgroups, "/slurm"+tokens[1])
	}
	return cgroups
}

func describe(ch chan<- *prometheus.Desc) {
	for _, m := range Metrics {
		ch <- m.promDesc
	}
}

func collect(ch chan<- prometheus.Metric, c Exporter, host string) {
	cgroups := c.GetRelativeJobPaths()
	if len(cgroups) < 1 {
		log.Println("no jobs running on host")
		return
	}
	for _, cgroup := range cgroups {
		// TODO: REGEX please, I mean c'mon
		_, t1, cut := strings.Cut(cgroup, "/slurm/uid_")
		if !cut {
			continue
		}
		uid, t2, cut := strings.Cut(t1, "/")
		if !cut {
			continue
		}
		_, job, cut := strings.Cut(t2, "/job_")
		if !cut {
			continue
		}
		stats := c.GetStats(cgroup)
		for _, s := range stats {
			m := Metrics[s.name]
			ch <- prometheus.MustNewConstMetric(m.promDesc, m.promType, float64(s.value), uid, job, host)
		}
	}
}
