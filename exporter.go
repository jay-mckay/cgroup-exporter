package main

import (
	"log"

	"github.com/containerd/cgroups"
	"github.com/prometheus/client_golang/prometheus"
)

var cgroupMetricLabels = []string{"cgroup"}
var cgroupMetricPrefix = "cgroup_"

type Metric struct {
	promDesc *prometheus.Desc
	promType prometheus.ValueType
}

type Collector struct {
	subsystem cgroups.Name
}

func (c *Collector) Name() cgroups.Name {
	return c.subsystem
}

var CPUMetrics = map[string]Metric{
	"kernel": {prometheus.NewDesc(cgroupMetricPrefix+"cpu_kernel_ns", "kernel cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
	"user":   {prometheus.NewDesc(cgroupMetricPrefix+"cpu_user_ns", "user cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
	"total":  {prometheus.NewDesc(cgroupMetricPrefix+"cpu_total_ns", "total cpu time for a cgroup in ns", cgroupMetricLabels, nil), prometheus.CounterValue},
}

func (c *cgroupCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range CPUMetrics {
		ch <- m.promDesc
	}
}

func (c *cgroupCollector) subsystem() ([]cgroups.Subsystem, error) {
	return []cgroups.Subsystem{&Collector{cgroups.Cpu}}, nil
}

func collectCPU(controller cgroups.Cgroup, path string, ch chan<- prometheus.Metric) {
	stats, err := controller.Stat()
	if err != nil {
		log.Println("unable to get stats of cgroup: ", err)
	}
	if stats.CPU != nil {
		if stats.CPU.Usage != nil {
			for name, m := range CPUMetrics {
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

func (collector *cgroupCollector) Collect(ch chan<- prometheus.Metric) {
	controller, err := cgroups.Load(collector.subsystem, cgroups.StaticPath(collector.config.path))
	if err != nil {
		log.Println("unable to load cgroup controllers: ", err)
	}
	collectCPU(controller, collector.config.path, ch)
}
