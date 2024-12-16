package peer

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"encoding/hex"
	"testing"
)

func TestGetHandshake(t *testing.T) {
	filename := "../bencode/testdata/ubuntu.torrent"
	obj := bencode.GetDictFromFile(&filename)
	infoDict := obj["info"].(map[string]any)
	digest := torrent.CalculateInfoHashFromInfoDict(infoDict)
	var peerId [20]byte
	copy(peerId[:], []byte("12345678901234567890"))
	handshake := data.GetHanshake(peerId, digest)
	expectedHexBytes := "13426974546f7272656e742070726f746f636f6c00000000000000009e638562ab1c1fced9def142864cdd5a7019e1aa3132333435363738393031323334353637383930"
	if hexBytes := hex.EncodeToString(handshake.ToBytes()); hexBytes != expectedHexBytes {
		t.Errorf("%v", hexBytes)
	}
}
