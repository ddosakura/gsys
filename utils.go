package gsys

import (
	"os/exec"
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
