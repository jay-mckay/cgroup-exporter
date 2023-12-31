package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
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

func (c CgroupCollector) getRelativeSubCgroups(cgroup string, pattern string) []string {
	var root string
	if c.hierarchy == cgroups.Unified {
		root = "sys/fs/cgroup"
	} else {
		root = "/sys/fs/cgroup/cpu" // use cpu here, cpu always(?) enabled
	}
	patterns, err := filepath.Glob(root + cgroup + pattern)
	check(err)
	var cgroups []string
	for _, p := range patterns {
		tokens := strings.Split(p, cgroup)
		cgroups = append(cgroups, cgroup+tokens[1])
	}
	return cgroups
}

func (c CgroupCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range CPUMetrics {
		ch <- m.promDesc
	}
}

func (c CgroupCollector) Collect(ch chan<- prometheus.Metric) {
	log.Println("collecting", c.config.root)
	c.collectCPU(ch, c.config.root)
	for _, pattern := range c.config.patterns {
		cgroups := c.getRelativeSubCgroups(c.config.root, pattern)
		for _, cgroup := range cgroups {
			log.Println("collecting", cgroup)
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
	manager, err := cgroup2.Load(cgroup, nil)
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{{"kernel", s.CPU.SystemUsec}, {"user", s.CPU.UsageUsec}, {"total", s.CPU.UsageUsec}}
}
