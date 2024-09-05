package torrent

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/data"
	"bytes"
	"crypto/sha1"
)

func CalculateInfoHash(info *data.BEInfo) [20]byte {
	return CalculateInfoHashFromInfoDict(bencode.ToDict(*info))
}

func CalculateInfoHashFromInfoDict(info map[string]any) [20]byte {
	var buf bytes.Buffer
	bencode.Encode(&buf, info)
	return sha1.Sum(buf.Bytes())
}
