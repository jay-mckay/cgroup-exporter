package main

import (
	"path/filepath"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/prometheus/client_golang/prometheus"
)

var cgroupMetricLabels = []string{"cgroup"}
var cgroupMetricPrefix = "cgroup_"

type cgroupMetric struct {
	promDesc *prometheus.Desc
	promType prometheus.ValueType
}

var CPUMetrics = map[string]cgroupMetric{
	"kernel": {prometheus.NewDesc(cgroupMetricPrefix+"cpu_kernel_ns", "kernel cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
	"user":   {prometheus.NewDesc(cgroupMetricPrefix+"cpu_user_ns", "user cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
	"total":  {prometheus.NewDesc(cgroupMetricPrefix+"cpu_total_ns", "total cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
}

type Stat struct {
	name  string
	value uint64
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func NewSlurmCollector() *CgroupCollector {
	var slice = "slurm"
	var groupGlobs = []string{"uid_*", "uid_*/job_*"}
	return NewCgroupCollector(slice, groupGlobs)
}

func NewUserSliceCollector() *CgroupCollector {
	var slice = "user.slice"
	var groupGlobs = []string{"user-*.slice"}
	return NewCgroupCollector(slice, groupGlobs)
}

func NewCgroupCollector(root string, patterns []string) *CgroupCollector {
	collector := CgroupCollector{cgroups.Mode(), root, patterns}
	return &collector
}

func (c CgroupCollector) getRelativeSubCgroups(slice string, groupPattern string) []string {
	glob := filepath.Join(slice, groupPattern)
	paths, err := filepath.Glob(glob)
	check(err)
	return paths
}

func (c CgroupCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range CPUMetrics {
		ch <- m.promDesc
	}
}

func (c CgroupCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectCPU(ch, "/"+c.root) // leading slash see cgroup2.Manager.VerifyGroupPath
	for _, pattern := range c.patterns {
		cgroups := c.getRelativeSubCgroups(c.root, pattern)
		for _, cgroup := range cgroups {
			c.collectCPU(ch, cgroup)
		}
	}
}

func (c CgroupCollector) collectCPU(ch chan<- prometheus.Metric, cgroup string) {
	var cpustats []Stat
	if c.hierarchy == cgroups.Unified {
		cpustats = c.collectCPUUnified(cgroup)
	} else {
		cpustats = c.collectCPULegacy(cgroup)
	}
	for _, s := range cpustats {
		m := CPUMetrics[s.name]
		ch <- prometheus.MustNewConstMetric(m.promDesc, m.promType, float64(s.value), cgroup)
	}
}

func (c CgroupCollector) collectCPULegacy(cgroup string) []Stat {
	manager, err := cgroup1.Load(cgroup1.StaticPath(cgroup))
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{{"kernel", s.CPU.Usage.Kernel}, {"user", s.CPU.Usage.User}, {"total", s.CPU.Usage.Total}}
}

func (c CgroupCollector) collectCPUUnified(cgroup string) []Stat {
	manager, err := cgroup2.Load(cgroup)
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{{"kernel", s.CPU.SystemUsec}, {"user", s.CPU.UsageUsec}, {"total", s.CPU.UsageUsec}}
}
