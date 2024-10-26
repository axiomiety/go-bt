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

func TestBitField(t *testing.T) {
	b := BitField{
		// 3 bytes to hold 24 bits
		NumBlocks: 22,
		// 0b 1110 1111, 0111 1111, 0000 1100
		Field: []byte{0xef, 0x7f, 0x0c},
	}
	blocksPresent := []uint64{0, 1, 2, 4, 5, 6, 7}
	for _, idx := range blocksPresent {
		if !b.HasBlock(idx) {
			t.Errorf("We should have block %d", idx)
		}

	}
}
