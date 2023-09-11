package main

import (
	"path/filepath"
	"regexp"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/prometheus/client_golang/prometheus"
)

var CPU_KERNEL = "cpu_kernel"
var CPU_USER = "cpu_user"
var CPU_TOTAL = "cpu_total"
var MEM_USAGE = "mem_usage"

var labels = []string{"uid", "jobid", "host"}
var metrics = map[string]Metric{
	CPU_KERNEL: {prometheus.NewDesc(CPU_KERNEL, "kernel cpu time for a cgroup in ns", labels, nil), prometheus.CounterValue},
	CPU_USER:   {prometheus.NewDesc(CPU_USER, "user cpu time for a cgroup in ns", labels, nil), prometheus.CounterValue},
	CPU_TOTAL:  {prometheus.NewDesc(CPU_TOTAL, "total cpu time for a cgroup in ns", labels, nil), prometheus.CounterValue},
	MEM_USAGE:  {prometheus.NewDesc(MEM_USAGE, "total memory usage for a cgroup in bytes", labels, nil), prometheus.GaugeValue},
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

func (c V1Collector) GetRelativeJobPaths() []string {
	return getRelativeJobPaths("/sys/fs/cgroup/cpu")
}

func (c V2Collector) GetRelativeJobPaths() []string {
	return getRelativeJobPaths("/sys/fs/cgroup")
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

func (c V1Collector) Describe(ch chan<- *prometheus.Desc) {
	describe(ch)
}

func (c V2Collector) Describe(ch chan<- *prometheus.Desc) {
	describe(ch)
}

func describe(ch chan<- *prometheus.Desc) {
	for _, m := range metrics {
		ch <- m.promDesc
	}
}

func (c V1Collector) Collect(ch chan<- prometheus.Metric) {
	collect(ch, c, c.host)
}

func (c V2Collector) Collect(ch chan<- prometheus.Metric) {
	collect(ch, c, c.host)
}

func collect(ch chan<- prometheus.Metric, c Exporter, host string) {
	r := regexp.MustCompile("/slurm(?:/uid_([0-9]+)/job_([0-9]+))")
	cgroups := c.GetRelativeJobPaths()
	for _, cgroup := range cgroups {
		match := r.FindAllStringSubmatch(cgroup, -1)
		job, uid := match[0][1], match[0][2]
		stats := c.GetStats(cgroup)
		for name, m := range metrics {
			ch <- prometheus.MustNewConstMetric(m.promDesc, m.promType, float64(stats[name]), job, uid, host)

		}
	}
}

func (c V1Collector) GetStats(cgroup string) map[string]uint64 {
	manager, err := cgroup1.Load(cgroup1.StaticPath(cgroup))
	check(err)
	s, err := manager.Stat()
	check(err)
	return map[string]uint64{
		CPU_KERNEL: s.CPU.Usage.Kernel,
		CPU_USER:   s.CPU.Usage.User,
		CPU_TOTAL:  s.CPU.Usage.Total,
		MEM_USAGE:  s.Memory.Usage.Usage,
	}
}

func (c V2Collector) GetStats(cgroup string) map[string]uint64 {
	manager, err := cgroup2.Load(cgroup, nil)
	check(err)
	s, err := manager.Stat()
	check(err)
	return map[string]uint64{
		CPU_KERNEL: s.CPU.SystemUsec,
		CPU_USER:   s.CPU.UserUsec,
		CPU_TOTAL:  s.CPU.UsageUsec,
		MEM_USAGE:  s.Memory.Usage,
	}
}
