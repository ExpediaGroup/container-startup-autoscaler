package integration

import (
	"fmt"
	"testing"
	"time"
)

func logMessage(t *testing.T, log any) {
	prefix := fmt.Sprintf("[%s]:", time.Now().Format(time.RFC3339Nano))

	if t != nil {
		t.Log(prefix, log)
		return
	}

	fmt.Println(prefix, log)
}
