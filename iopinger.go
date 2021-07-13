package main

import (
	"net"
	"sync"
	"time"
)

func NewIopinger(target string) *Iopinger {
	return &Iopinger{
		target:      target,
		PacketsSent: 0,
	}
}

type Iopinger struct {
	target string

	// Number of packets sent
	PacketsSent int
}

func (s *Iopinger) Run() {
	// Start pinging the host. This is a go run function
	return
}

func (s *Iopinger) Stop() {
	return
}

//func (s *Iopinger) Statistics() {
//	return s
//}

type IopingerSample struct {
	// Interval is the wait time between each packet send. Default is 1s.
	Interval time.Duration

	// Timeout specifies a timeout before ping exits, regardless of how many
	// packets have been received.
	Timeout time.Duration

	// Count tells pinger to stop after sending (and receiving) Count echo
	// packets. If this option is not specified, pinger will operate until
	// interrupted.
	Count int

	// Debug runs in debug mode
	Debug bool

	// Number of packets sent
	PacketsSent int

	// Number of packets received
	PacketsRecv int

	// Number of duplicate packets received
	PacketsRecvDuplicates int

	// Round trip time statistics
	minRtt    time.Duration
	maxRtt    time.Duration
	avgRtt    time.Duration
	stdDevRtt time.Duration
	stddevm2  time.Duration
	statsMu   sync.RWMutex

	// If true, keep a record of rtts of all received packets.
	// Set to false to avoid memory bloat for long running pings.
	RecordRtts bool

	// rtts is all of the Rtts
	rtts []time.Duration

	// OnSetup is called when Pinger has finished setting up the listening socket
	OnSetup func()

	// OnSend is called when Pinger sends a packet
	//OnSend func(*Packet)

	// OnRecv is called when Pinger receives and processes a packet
	//OnRecv func(*Packet)

	// OnFinish is called when Pinger exits
	//OnFinish func(*Statistics)

	// OnDuplicateRecv is called when a packet is received that has already been received.
	//OnDuplicateRecv func(*Packet)

	// Size of packet being sent
	Size int

	// Tracker: Used to uniquely identify packets - Deprecated
	Tracker uint64

	// Source is the source IP address
	Source string

	// Channel and mutex used to communicate when the Pinger should stop between goroutines.
	done chan interface{}
	lock sync.Mutex

	ipaddr *net.IPAddr
	addr   string

	// trackerUUIDs is the list of UUIDs being used for sending packets.
	//trackerUUIDs []uuid.UUID

	ipv4     bool
	id       int
	sequence int
	// awaitingSequences are in-flight sequence numbers we keep track of to help remove duplicate receipts
	//awaitingSequences map[uuid.UUID]map[int]struct{}
	// network is one of "ip", "ip4", or "ip6".
	network string
	// protocol is "icmp" or "udp".
	protocol string

	//logger Logger
}
