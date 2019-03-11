package gsys

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"time"
)

// Execute Command
func Execute(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	b, e := cmd.CombinedOutput()
	if e != nil {
		panic(e)
	}
	return string(b)
}

// CommandWatch Callback
type CommandWatch func(times int, frame []byte, size int, e error) (stop bool)

// WatchConfig for ExecuteWatch
type WatchConfig struct {
	Callback  CommandWatch
	Errorback CommandWatch
	FrameSize int
	KillError func(error, *exec.Cmd)
	Sleep     time.Duration
}

// ExecuteWatch Command
func ExecuteWatch(cfg *WatchConfig, name string, args ...string) (*sync.WaitGroup, *exec.Cmd) {
	var wg sync.WaitGroup
	if cfg == nil {
		panic("WatchConfig chouldn't be nil")
	}
	fn := cfg.Callback
	if fn == nil {
		return &wg, nil
	}
	framwSize := cfg.FrameSize
	if framwSize < 1024 {
		framwSize = 1024
	}

	cmd := exec.Command(name, args...)
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		panic(e)
	}
	stderr, e := cmd.StderrPipe()
	if e != nil {
		panic(e)
	}

	cmd.Start()
	r := bufio.NewReader(stdout)
	r2 := bufio.NewReader(stderr)

	stop := false
	wg.Add(1)
	go func() {
		times := 0
		for !stop {
			if cfg.Sleep > 0 {
				time.Sleep(cfg.Sleep)
			}
			buf := make([]byte, framwSize)
			n, e := r.Read(buf)
			times++
			if e == io.EOF {
				wg.Done()
				return
			}
			if e != nil {
				stop = fn(times, nil, 0, e)
			} else {
				stop = fn(times, buf[:n], n, nil)
			}
			if stop {
				e = cmd.Process.Kill()
				if e != nil && cfg.KillError != nil {
					cfg.KillError(e, cmd)
				}
				wg.Done()
			}
		}
	}()

	fn2 := cfg.Errorback
	if fn2 != nil {
		wg.Add(1)
		go func() {
			times := 0
			for !stop {
				if cfg.Sleep > 0 {
					time.Sleep(cfg.Sleep)
				}
				buf := make([]byte, framwSize)
				n, e := r2.Read(buf)
				times++
				if e == io.EOF {
					wg.Done()
					return
				}
				if e != nil {
					stop = fn2(times, nil, 0, e)
				} else {
					stop = fn2(times, buf[:n], n, nil)
				}
				if stop {
					e = cmd.Process.Kill()
					if e != nil && cfg.KillError != nil {
						cfg.KillError(e, cmd)
					}
					wg.Done()
				}
			}
		}()
	}

	return &wg, cmd
}
