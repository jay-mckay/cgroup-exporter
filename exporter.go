package main

import (
	"path/filepath"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
)

var slurmMetricLabels = []string{"uid", "jobid", "host"}
var slurmMetricPrefix = "slurm_job_"
var Metrics = map[string]Metric{
	"kernel_cpu": {prometheus.NewDesc(slurmMetricPrefix+"cpu_kernel_ns", "kernel cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
	"user_cpu":   {prometheus.NewDesc(slurmMetricPrefix+"cpu_user_ns", "user cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
	"total_cpu":  {prometheus.NewDesc(slurmMetricPrefix+"cpu_total_ns", "total cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
}

type Metric struct {
	promDesc *prometheus.Desc
	promType prometheus.ValueType
}

type Exporter interface {
	GetRelativeJobPaths() []string
	GetStats(string) map[string]uint64
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getRelativeJobPaths(root string) []string {
	r := regexp.MustCompile("/slurm/uid_([0-9]+)/job_([0-9]+)")
	paths, err := filepath.Glob(root + "/slurm/uid_*/job_*")
	check(err)
	var cgroups []string
	for _, p := range paths {
		s := r.FindString(p)
		if s != "" {
			cgroups = append(cgroups, s)
		}
	}
	return cgroups
}

func describe(ch chan<- *prometheus.Desc) {
	for _, m := range Metrics {
		ch <- m.promDesc
	}
}

func collect(ch chan<- prometheus.Metric, c Exporter, host string) {
	r := regexp.MustCompile("/slurm(?:/uid_([0-9]+)/job_([0-9]+))")
	cgroups := c.GetRelativeJobPaths()
	for _, cgroup := range cgroups {
		match := r.FindAllStringSubmatch(cgroup, -1)
		job, uid := match[0][1], match[0][2]
		stats := c.GetStats(cgroup)
		for name, m := range Metrics {
			ch <- prometheus.MustNewConstMetric(m.promDesc, m.promType, float64(stats[name]), job, uid, host)

		}
	}
}
