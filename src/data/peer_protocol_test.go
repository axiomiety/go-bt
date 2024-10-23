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

func TestChoke(t *testing.T) {
	choke := Choke()
	if !bytes.Equal(choke.ToBytes(), []byte{0, 0, 0, 1, 0}) {
		t.Errorf("was expecting [0,0,0,0], got %v", choke.ToBytes())
	}
}
