package shell

import (
	"context"
	"testing"
	"time"
)

func TestUnix(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	v, err := ExecShell(ctx, "sleep 3; echo \"test1\"")
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("### output: %s", v)
	}
}
