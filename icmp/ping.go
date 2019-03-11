package icmp

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/ddosakura/gsys"
)

// PingCfg for `ping`
type PingCfg struct {
	Target             string
	Count              int
	PacketSize         int
	packetSizeWithHead int
	backPacketSize     int

	UseIPv6 bool
}

// PingResult for `ping`
type PingResult struct {
	cfg *PingCfg

	IP     string
	Frames []PingFrame

	TransmittedPackets int
	ReceivedPackets    int
	PacketLoss         float64
	TotalTime          string
	RTT                *PingRTT
}

// PingFrame for `ping`
type PingFrame struct {
	Seq  int
	TTL  int
	Time string
}

// PingRTT for `ping`
type PingRTT struct {
	Min  float64
	Avg  float64
	Max  float64
	MDev float64

	Unit string
}

var (
	regPingIP   = regexp.MustCompile(" [(][0-9.]+[)]")
	regPingIPv6 = regexp.MustCompile(" [(][0-9A-F:]+[)]")

	// eg. icmp_seq=1 ttl=64 time=0.035 ms
	regPingSeq  = regexp.MustCompile("icmp_seq=[0-9]+")
	regPingTTL  = regexp.MustCompile("ttl=[0-9]+")
	regPingTime = regexp.MustCompile("time=[0-9. ms]+")

	// 5 packets transmitted, 5 received, 0% packet loss, time 70ms
	regPingTransmittedPackets = regexp.MustCompile("[0-9]+ packets")
	regPingReceivedPackets    = regexp.MustCompile("[0-9]+ received")
	regPingPacketLoss         = regexp.MustCompile("[0-9.]+%")
	regPingTotalTime          = regexp.MustCompile("time [0-9.ms]+")

	// rtt min/avg/max/mdev = 0.040/0.060/0.076/0.012 ms
	regPingRTT = regexp.MustCompile("[0-9.]+")
)

// Ping cmd
func Ping(cfg *PingCfg, args ...string) (pr *PingResult, err error) {
	if cfg.Target == "" {
		cfg.Target = "localhost"
	}
	if cfg.Count < 1 {
		cfg.Count = 1
	}
	if cfg.PacketSize < 1 {
		cfg.PacketSize = 56
	}
	cfg.packetSizeWithHead = cfg.PacketSize + 28
	cfg.backPacketSize = cfg.PacketSize + 8

	t := make([]string, 0, 6+len(args))
	t = append(t, cfg.Target,
		"-c", strconv.Itoa(cfg.Count),
		"-s", strconv.Itoa(cfg.PacketSize),
	)
	if cfg.UseIPv6 {
		t = append(t, "-6")
	} else {
		t = append(t, "-4")
	}
	t = append(t, args...)

	pr = &PingResult{
		cfg:    cfg,
		Frames: make([]PingFrame, cfg.Count),
	}
	wg, _ := gsys.ExecuteWatch(&gsys.WatchConfig{
		Callback: func(times int, f []byte, n int, e error) bool {
			var p []int
			var pf PingFrame
			defer func() {
				_ = recover()
				if pf.Seq > 0 && pf.Seq <= cfg.Count {
					pr.Frames[pf.Seq-1] = pf
				}
			}()
			lines := strings.Split(string(f), "\n")

			line := lines[0]
			if times == 1 {
				if cfg.UseIPv6 {
					p = regPingIPv6.FindStringIndex(lines[0])
				} else {
					p = regPingIP.FindStringIndex(lines[0])
				}
				pr.IP = lines[0][p[0]+2 : p[1]-1]
				line = lines[1]
			}

			p = regPingSeq.FindStringIndex(line)
			pf.Seq, _ = strconv.Atoi(line[p[0]+9 : p[1]])
			p = regPingTTL.FindStringIndex(line)
			pf.TTL, _ = strconv.Atoi(line[p[0]+4 : p[1]])
			p = regPingTime.FindStringIndex(line)
			pf.Time = line[p[0]+5 : p[1]]

			if len(lines) > 3 {
				line = lines[len(lines)-3]
				p = regPingTransmittedPackets.FindStringIndex(line)
				pr.TransmittedPackets, _ = strconv.Atoi(line[p[0] : p[1]-8])
				p = regPingReceivedPackets.FindStringIndex(line)
				pr.ReceivedPackets, _ = strconv.Atoi(line[p[0] : p[1]-9])
				p = regPingPacketLoss.FindStringIndex(line)
				pr.PacketLoss, _ = strconv.ParseFloat(line[p[0]:p[1]-1], 64)
				p = regPingTotalTime.FindStringIndex(line)
				pr.TotalTime = line[p[0]+5 : p[1]]

				line = lines[len(lines)-2]
				ps := regPingRTT.FindAllStringIndex(line, 4)
				Min, _ := strconv.ParseFloat(line[ps[0][0]:ps[0][1]], 64)
				Avg, _ := strconv.ParseFloat(line[ps[1][0]:ps[1][1]], 64)
				Max, _ := strconv.ParseFloat(line[ps[2][0]:ps[2][1]], 64)
				MDev, _ := strconv.ParseFloat(line[ps[3][0]:ps[3][1]], 64)
				lines := strings.Split(line, " ")
				Unit := lines[len(lines)-1]
				pr.RTT = &PingRTT{
					Min,
					Avg,
					Max,
					MDev,
					Unit,
				}
			}

			return false
		},
		Errorback: func(times int, f []byte, n int, e error) bool {
			err = errors.New(string(f))
			return true
		},
		FrameSize: 1024,
	}, "ping", t...)
	wg.Wait()

	return
}
