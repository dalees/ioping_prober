// Copyright 2018 Ben Kochie <superq@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/log"
)

const (
	namespace = "ioping"
)

var (
	labelNames = []string{"target", "target_size"}

	pingResponseTtl = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "response_ttl",
			Help:      "The last response Time To Live (TTL).",
		},
		labelNames,
	)
	pingResponseDuplicates = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "response_duplicates_total",
			Help:      "The number of duplicated response packets.",
		},
		labelNames,
	)
)

func newPingResponseHistogram(buckets []float64) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "response_duration_seconds",
			Help:      "A histogram of latencies for ping responses.",
			Buckets:   buckets,
		},
		labelNames,
	)
}

// IopingCollector collects metrics from the pinger.
type IopingCollector struct {
	pingers *[]*Iopinger

	requestsSent *prometheus.Desc
}

func NewIopingCollector(pingers *[]*Iopinger, pingResponseSeconds prometheus.HistogramVec) *IopingCollector {
	for _, pinger := range *pingers {
		// Init all metrics to 0s.
		target := pinger.Target
		pingResponseSeconds.WithLabelValues(target, "1")

		// Setup handler functions.
		pinger.OnMeasure = func(stats *Statistics) {
			measurement_value := float64(stats.Max)
			pingResponseSeconds.WithLabelValues(stats.Target).Observe(measurement_value)
			var nsec_to_sec float64 = 0.000000001
			log.Debugf("Measurement time: %f sec (%f nanosec)\n", measurement_value*nsec_to_sec, measurement_value)
		}
		//pinger.OnFinish = func(stats *ping.Statistics) {
		//	log.Debugf("\n--- %s ping statistics ---\n", stats.Addr)
		//	log.Debugf("%d packets transmitted, %d packets received, %v%% packet loss\n",
		//		stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		//	log.Debugf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
		//		stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
		//}
	}

	return &IopingCollector{
		pingers: pingers,
		requestsSent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "requests_total"),
			"Number of measurements performed",
			labelNames,
			nil,
		),
	}
}

func (s *IopingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.requestsSent
}

func (s *IopingCollector) Collect(ch chan<- prometheus.Metric) {
	for _, pinger := range *s.pingers {
		//stats := pinger.Statistics()

		ch <- prometheus.MustNewConstMetric(
			s.requestsSent,
			prometheus.CounterValue,
			float64(pinger.Measurements),
			pinger.Target,
		)
	}
}
