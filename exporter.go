package main

import (
	"log"

	"github.com/containerd/cgroups"
	"github.com/prometheus/client_golang/prometheus"
)

var cgroupMetricLabels = []string{"cgroup"}
var cgroupMetricPrefix = "cgroup_"

type Metric struct {
	promDesc *prometheus.promDesc
	promType prometheus.ValueType
}

var cgroupCPUMetrics = map[string]Metric{
	"kernel": {prometheus.NewDesc(cgroupMetricPrefix+"cpu_kernel_ns", "kernel cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
	"user":   {prometheus.NewDesc(cgroupMetricPrefix+"cpu_user_ns", "user cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
	"total":  {prometheus.NewDesc(cgroupMetricPrefix+"cpu_total_ns", "total cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
}

func (collector *cgroupCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range cgroupCPUMetrics {
		ch <- m.promDesc
	}
}

func subsystem() ([]cgroups.Subsystem, error) {
	s := cgroups.Subsystem{
		cgroups.Cpu,
		cgroups.Memory,
	}
	return s
}

func (cg cgroups.Cgroup) collectCPU(path string, ch chan<- prometheus.Metric) {
	stats, err := cg.Stats()
	if err != nil {
		log.Println("unable to get stats of cgroup: ", err)
	}
	if stats.CPU != nil {
		if stats.CPU.Usage != nil {
			for name, m := range cgroupCPUMetrics {
				var value float64
				switch name {
				case "kernel":
					value = float64(stats.CPU.Usage.Kernel)
				case "user":
					value = float64(stats.CPU.Usage.User)
				case "total":
					value = float64(stats.CPU.Usage.Total)
				}
				ch <- prometheus.MustNewConstMetric(m.promDesc, m.promType, value, path)
			}

		}
	}
}

func (c *cgroupCollector) Collect(ch chan<- prometheus.Metric) {
	cg, err := cgroups.Load(subsystem, cgroups.StaticPath(c.config.path))
	if err != nil {
		log.Println("unable to load cgroup controllers: ", err)
	}
	cg.collectCPU(path, ch)
}
