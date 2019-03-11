package gsys

import (
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
