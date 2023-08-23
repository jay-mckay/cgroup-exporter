package main

import (
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/prometheus/client_golang/prometheus"
)

func (c V2Collector) GetRelativeJobPaths() []string {
	return getRelativeJobPaths("/sys/fs/cgroup")
}

func (c V2Collector) Describe(ch chan<- *prometheus.Desc) {
	describe(ch)
}

func (c V2Collector) Collect(ch chan<- prometheus.Metric) {
	collect(ch, c, c.host)
}

func (c V2Collector) GetStats(cgroup string) []Stat {
	manager, err := cgroup2.Load(cgroup, nil)
	check(err)
	s, err := manager.Stat()
	check(err)
	return []Stat{{"kernel", s.CPU.SystemUsec}, {"user", s.CPU.UsageUsec}, {"total", s.CPU.UsageUsec}}
}
