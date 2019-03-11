package testing

import (
	"fmt"
	"testing"

	"github.com/ddosakura/gsys/icmp"
)

func TestIcmpPing(t *testing.T) {
	result, e := icmp.Ping(&icmp.PingCfg{
		Target: "localhost",
		Count:  5,
	})
	if e != nil {
		t.Error(e.Error())
	}
	fmt.Printf("%#v\n%#v\n", result, result.RTT)
}

func TestIcmpPingIPv6(t *testing.T) {
	result, e := icmp.Ping(&icmp.PingCfg{
		Target:  "localhost",
		Count:   5,
		UseIPv6: true,
	})
	if e != nil {
		t.Error(e.Error())
		return
	}
	fmt.Printf("%#v\n%#v\n", result, result.RTT)
}
