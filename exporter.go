package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/prometheus/client_golang/prometheus"
)

var slurmMetricLabels = []string{"uid", "jobid", "host"}
var slurmMetricPrefix = "slurm_job_"
var CPUMetrics = map[string]Metric{
	"kernel": {prometheus.NewDesc(slurmMetricPrefix+"cpu_kernel_ns", "kernel cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
	"user":   {prometheus.NewDesc(slurmMetricPrefix+"cpu_user_ns", "user cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
	"total":  {prometheus.NewDesc(slurmMetricPrefix+"cpu_total_ns", "total cpu time for a cgroup in ns", slurmMetricLabels, nil), prometheus.CounterValue},
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
	getRelativeJobPaths() []string
	stat(string) []Stat
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (c V1Collector) getRelativeJobPaths() []string {
	return getRelativeJobPaths("/sys/fs/cgroup/cpu")
}

func (c V2Collector) getRelativeJobPaths() []string {
	return getRelativeJobPaths("/sys/fs/cgroup")
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

func (c V1Collector) Describe(ch chan<- *prometheus.Desc) {
	describe(ch)
}

func (c V2Collector) Describe(ch chan<- *prometheus.Desc) {
	describe(ch)
}

func describe(ch chan<- *prometheus.Desc) {
	for _, m := range CPUMetrics {
		ch <- m.promDesc
	}
}

func (c V2Collector) Collect(ch chan<- prometheus.Metric) {
	collect(ch, c, c.host)
}

func (c V1Collector) Collect(ch chan<- prometheus.Metric) {
	collect(ch, c, c.host)
}

func collect(ch chan<- prometheus.Metric, c Exporter, host string) {
	cgroups := c.getRelativeJobPaths()
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
		stats := c.stat(cgroup)
		for _, s := range stats {
			m := CPUMetrics[s.name]
			ch <- prometheus.MustNewConstMetric(m.promDesc, m.promType, float64(s.value), uid, job, host)
		}
	}
}

func (c V2Collector) stat(cgroup string) []Stat {
	manager, err := cgroup1.Load(cgroup1.StaticPath(cgroup))
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{{"kernel", s.CPU.Usage.Kernel}, {"user", s.CPU.Usage.User}, {"total", s.CPU.Usage.Total}}
}

func (c V1Collector) stat(cgroup string) []Stat {
	manager, err := cgroup2.Load(cgroup, nil)
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{{"kernel", s.CPU.SystemUsec}, {"user", s.CPU.UsageUsec}, {"total", s.CPU.UsageUsec}}
}
