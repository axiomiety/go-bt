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
		// 0b 1110 1111, 0111 1111, 0000 0100
		Field: []byte{0xef, 0x7f, 0x04},
	}
	blocksPresent := []uint32{0, 1, 2, 4, 5, 6, 7, 9, 10, 11, 12, 13, 14, 15, 21}
	for _, idx := range blocksPresent {
		if !b.HasPiece(idx) {
			t.Errorf("We should have block %d", idx)
		}

	}
	blocksMissing := []uint32{3, 8, 16, 17, 18, 19, 20}
	for _, idx := range blocksMissing {
		if b.HasPiece(idx) {
			t.Errorf("We should *not* have block %d", idx)
		}

	}

	// now for updates
	for _, idx := range blocksMissing {
		b.SetPiece(idx)
		if !b.HasPiece(3) {
			t.Errorf("Tried to set block %d but it is still reported as missing", idx)
		}
	}
}

func TestReques(t *testing.T) {
	msg := Request(1, 2, 3)
	expected := []byte{
		0, 0, 0, 13,
		6,
		0, 0, 0, 1,
		0, 0, 0, 2,
		0, 0, 0, 3,
	}
	if !bytes.Equal(msg.ToBytes(), expected) {
		t.Errorf("exepected %v but got %v", expected, msg.ToBytes())
	}
}
