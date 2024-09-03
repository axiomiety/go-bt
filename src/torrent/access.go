package torrent

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/data"
	"bytes"
	"crypto/sha1"
)

func InfoHash(i *data.BEInfo) string {
	var encodedInfo bytes.Buffer
	bencode.Encode(&encodedInfo, i)
	return string(sha1.New().Sum(encodedInfo.Bytes()))
}
