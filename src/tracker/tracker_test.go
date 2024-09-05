package tracker_test

import (
	"axiomiety/go-bt/tracker"
	"testing"
)

func TestByteEncode(t *testing.T) {
	infoHashStr := "\x12\x34\x56\x78\x9a\xbc\xde\xf1\x23\x45\x67\x89\xab\xcd\xef\x12\x34\x56\x78\x9a"
	var infoHash [20]byte
	copy(infoHash[:], string(infoHashStr))
	encodedHash := tracker.EncodeInfoHash(infoHash)
	expected := "%124Vx%9A%BC%DE%F1%23Eg%89%AB%CD%EF%124Vx%9A"
	if encodedHash != expected {
		t.Errorf("expected %s but got %s", expected, encodedHash)
	}
}
