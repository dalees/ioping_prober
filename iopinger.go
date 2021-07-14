package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/common/log"
)

func NewIopinger(target string) *Iopinger {
	return &Iopinger{
		Target:       target,
		Interval:     time.Second,
		Measurements: 0,
		done:         make(chan interface{}),
	}
}

type Iopinger struct {
	Target string

	// Interval is the wait time between each packet send. Default is 1s.
	Interval time.Duration

	// Number of measurements performed
	Measurements uint64

	// Channel and mutex used to communicate when the Pinger should stop between goroutines.
	done chan interface{}
	//lock sync.Mutex

	// Handler, called after each measurement
	OnMeasure func(*Statistics)
}

// Run runs the pinger. This is a blocking function that will run
// continuously until interrupted.
func (p *Iopinger) Run() {
	// Start pinging the host.
	//timeout := time.NewTicker(p.Timeout)
	interval := time.NewTicker(p.Interval)
	defer func() {
		p.Stop()
		interval.Stop()
		//timeout.Stop()
	}()

	// TODO: validate that ioping is a version we're okay with.

	// Require ioping with:
	// '-a 0', so we can use '-c 1'
	// otherwise need to use '-c 2 -i 0ms'
	// TODO: Consider changing to '-c 10 -B -p 1'. Less exec overhead? Requires ignoring final summary line.
	target := "/tmp"

	for range interval.C {
		// "-sync"
		cmd := exec.Command("/usr/bin/ioping", "-warmup=0", "-count=1", "-interval=0ms", string(target), "-batch")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with '%s'\n", err)
		}

		stats := Statistics{
			Target: target,
		}
		stats.parseRawStatistics(out.String())

		handler := p.OnMeasure
		if handler != nil {
			handler(&stats)
		}

		//var ns_to_ms float64 = 0.000001
		//fmt.Printf("Max time: %f ms\n", float64(stats.Max)*ns_to_ms)
		//fmt.Printf("iops: %f\n", stats.Iops)
		//fmt.Printf("Requests: %d\n", stats.Count)
	}

}

func (p *Iopinger) Stop() {
	//return
}

//func (s *Iopinger) Statistics() {
//	return s
//}

// Statistics represent the batch mode stats of a completed operation.
type Statistics struct {
	Target string

	// dump_statistics: https://github.com/koct9i/ioping/blob/f549dffc224b3fcab10ad718dc243e1b0ba845f7/ioping.c#L1418
	// struct def: https://github.com/koct9i/ioping/blob/f549dffc224b3fcab10ad718dc243e1b0ba845f7/ioping.c#L1341

	//(0) count of requests in statistics
	Valid uint64

	//(1) running time         (nsec)
	Sum float64

	//(2) requests per second  (iops)
	Iops float64

	//(3) transfer speed       (bytes/sec)
	Speed float64

	//(4) minimal request time (nsec)
	Min uint64

	//(5) average request time (nsec)
	Avg float64

	//(6) maximum request time (nsec)
	Max uint64

	//(7) request time standard deviation (nsec)
	Mdev float64

	//(8) total requests       (including too slow and too fast)
	Count uint64

	//(9) total running time  (nsec)
	Load_time uint64
}

func (stat *Statistics) parseRawStatistics(raw string) {

	stats_raw := strings.Split(raw, " ")

	valid_requests, err := strconv.ParseUint(stats_raw[0], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	stat.Valid = valid_requests

	iops_value, err := strconv.ParseFloat(stats_raw[2], 64)
	if err != nil {
		log.Fatal(err)
	}
	stat.Iops = iops_value

	bytespersec_value, err := strconv.ParseFloat(stats_raw[3], 64)
	if err != nil {
		log.Fatal(err)
	}
	stat.Speed = bytespersec_value

	max_request_ns, err := strconv.ParseUint(stats_raw[6], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	stat.Max = max_request_ns

	count_requests, err := strconv.ParseUint(stats_raw[8], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	stat.Count = count_requests
}
