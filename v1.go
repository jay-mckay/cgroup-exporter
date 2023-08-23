package main

import (
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/prometheus/client_golang/prometheus"
)

func (c V1Collector) GetRelativeJobPaths() []string {
	return getRelativeJobPaths("/sys/fs/cgroup/cpu")
}

func (c V1Collector) Describe(ch chan<- *prometheus.Desc) {
	describe(ch)
}

func (c V1Collector) Collect(ch chan<- prometheus.Metric) {
	collect(ch, c, c.host)
}

func (c V1Collector) GetStats(cgroup string) []Stat {
	manager, err := cgroup1.Load(cgroup1.StaticPath(cgroup))
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{
		{"kernel_cpu", s.CPU.Usage.Kernel}, {"user_cpu", s.CPU.Usage.User}, {"total_cpu", s.CPU.Usage.Total},
		//{"mem_rss", s.Memory.RSS},
	}
}
