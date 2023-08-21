package main

import (
	cgroups "github.com/containerd/cgroups/v3"
	cgroup1 "github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/prometheus/client_golang/prometheus"
)

var cgroupMetricLabels = []string{"cgroup"}
var cgroupMetricPrefix = "cgroup_"

type Metric struct {
	promDesc *prometheus.Desc
	promType prometheus.ValueType
}

var CPUMetrics = map[string]Metric{
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

func (c CgroupCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range CPUMetrics {
		ch <- m.promDesc
	}
}

func (c CgroupCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectCPU(ch)
}

func (c CgroupCollector) collectCPU(ch chan<- prometheus.Metric) {
	var cpustats []Stat
	if c.hierarchy == cgroups.Unified {
		cpustats = c.collectCPUUnified()
	} else {
		cpustats = c.collectCPULegacy()
	}
	for _, s := range cpustats {
		m := CPUMetrics[s.name]
		ch <- prometheus.MustNewConstMetric(m.promDesc, m.promType, float64(s.value), c.path)
	}
}

func (c CgroupCollector) collectCPULegacy() []Stat {
	manager, err := cgroup1.Load(cgroup1.StaticPath(c.path))
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{{"kernel", s.CPU.Usage.Kernel}, {"user", s.CPU.Usage.User}, {"total", s.CPU.Usage.Total}}
}

func (c CgroupCollector) collectCPUUnified() []Stat {
	manager, err := cgroup2.Load(c.path, nil)
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{{"kernel", s.CPU.SystemUsec}, {"user", s.CPU.UsageUsec}, {"total", s.CPU.UsageUsec}}
}
