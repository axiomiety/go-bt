package data

import (
	"bytes"
	"testing"
)

func TestKeepAlive(t *testing.T) {
	keepalive := KeepAlive()
	if !bytes.Equal(keepalive.ToBytes(), []byte{0, 0, 0, 0}) {
		t.Errorf("was expecting [0,0,0,0], got %v", keepalive.ToBytes())
	}
}
