package gsys

import (
	"fmt"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	ret := Execute("uname")
	if ret != "Linux\n" {
		t.Errorf("`uname` should return `Linux`, but `%s`", ret)
	}
}

func TestExecuteWithArgs(t *testing.T) {
	ret := Execute("uname", "-mo")
	if ret != "x86_64 GNU/Linux\n" {
		t.Errorf("`uname` should return `x86_64 GNU/Linux`, but `%s`", ret)
	}
}

func TestExecuteWatch(t *testing.T) {
	wg, cmd := ExecuteWatch(&WatchConfig{
		Callback: func(times int, f []byte, n int, e error) bool {
			lines := strings.Split(string(f), "\n")
			switch times {
			case 1:
				println(lines[1])
			default:
				println(lines[0])
			}
			return false
		},
		FrameSize: 1024 * 1024,
		// Sleep:     time.Duration(1) * time.Second,
	}, "ping", "baidu.com", "-c 5")
	wg.Wait()
	fmt.Println(*cmd)
}
