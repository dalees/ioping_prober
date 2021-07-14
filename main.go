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
	"fmt"
	"math"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// Generated with: prometheus.ExponentialBuckets(0.00005, 2, 20)
	defaultBuckets = "5e-05,0.0001,0.0002,0.0004,0.0008,0.0016,0.0032,0.0064,0.0128,0.0256,0.0512,0.1024,0.2048,0.4096,0.8192,1.6384,3.2768,6.5536,13.1072,26.2144"
)

type targetList []string

func (h *targetList) Set(value string) error {
	if value == "" {
		return fmt.Errorf("'%s' is not valid directory|file|device", value)
	} else {
		*h = append(*h, value)
		return nil
	}
}

func (h *targetList) String() string {
	return ""
}

func (h *targetList) IsCumulative() bool {
	return true
}

func TargetList(s kingpin.Settings) (target *[]string) {
	target = new([]string)
	s.SetValue((*targetList)(target))
	return
}

func init() {
	prometheus.MustRegister(version.NewCollector("ioping_prober"))
}

func parseBuckets(buckets string) ([]float64, error) {
	bucketstrings := strings.Split(buckets, ",")
	bucketlist := make([]float64, len(bucketstrings))
	for i := range bucketstrings {
		value, err := strconv.ParseFloat(bucketstrings[i], 64)
		if err != nil {
			return nil, err
		}
		bucketlist[i] = value
	}
	return bucketlist, nil
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9374").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()

		buckets         = kingpin.Flag("buckets", "A comma delimited list of buckets to use").Default(defaultBuckets).String()
		interval        = kingpin.Flag("ping.interval", "Ping interval duration").Short('i').Default("1s").Duration()
		writemode       = kingpin.Flag("write", "Write to target. Uses ioping -W and is safe for directory target.").Default("false").Bool()
		unsafewritemode = kingpin.Flag("unsafewrite", "Unsafely write to target. Uses ioping -WWW and is destructive to file|device- read ioping manpage.").Default("false").Bool()
		targets         = TargetList(kingpin.Arg("target", "List of target directory/file/device to ioping").Required())
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("ioping_prober"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting ioping_prober", version.Info())
	log.Infoln("Build context", version.BuildContext())
	bucketlist, err := parseBuckets(*buckets)
	if err != nil {
		log.Errorf("failed to parse buckets: %s\n", err.Error())
		return
	}
	pingResponseSeconds := newPingResponseHistogram(bucketlist)
	prometheus.MustRegister(pingResponseSeconds)

	num_targets := len(*targets)
	num_pingers := num_targets
	if *writemode || *unsafewritemode {
		// Create an extra pinger per target for writes
		num_pingers *= 2
	}

	pingers := make([]*Iopinger, num_pingers)
	for i, target := range *targets {
		// Create read mode pingers
		pinger := NewIopinger(target)
		pinger.Interval = *interval
		pinger.Timeout = time.Duration(math.MaxInt64)
		pingers[i] = pinger

		// create write mode pingers
		if *writemode || *unsafewritemode {
			pinger := NewIopinger(target)
			pinger.Interval = *interval
			pinger.Timeout = time.Duration(math.MaxInt64)
			pinger.WriteMode = *writemode
			pinger.UnsafeWriteMode = *unsafewritemode
			pingers[num_targets+i] = pinger
		}
	}

	splay := time.Duration(interval.Nanoseconds() / int64(len(pingers)))
	log.Infof("Waiting %s between starting pingers", splay)
	for i, pinger := range pingers {
		log.Infof("Starting prober for %s", pinger.Target)
		go pinger.Run()
		if i < len(pingers)-1 {
			time.Sleep(splay)
		}
	}

	prometheus.MustRegister(NewIopingCollector(&pingers, *pingResponseSeconds))

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>ioping Exporter</title></head>
			<body>
			<h1>ioping Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	log.Infof("Listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
	//for _, pinger := range pingers {
	//		pinger.Stop()
	//	}
}
