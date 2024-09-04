package torrent

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"bytes"
	"crypto/sha1"
	"io"
	"os"
)

func CreateTorrent(outputFile string, announce string, name string, pieceLength int, filenames ...string) {
	// build the info dict
	infoDict := mkInfoDict(name, filenames, pieceLength)

	torrentMap := map[string]any{
		"announce":   announce,
		"created by": "go-bt",
		"info":       infoDict,
	}
	var buf bytes.Buffer
	bencode.Encode(&buf, torrentMap)
	err := os.WriteFile(outputFile, buf.Bytes(), 0644)
	common.Check(err)
}

func mkInfoDict(name string, filenames []string, pieceLength int) map[string]any {
	infoDict := map[string]any{"name": name}

	getNumBytes := func(filename string) int64 {
		stat, err := os.Stat(filename)
		common.Check(err)
		return stat.Size()
	}

	if len(filenames) == 1 {
		infoDict["length"] = getNumBytes(filenames[0])
	} else {
		files := make([]map[string]any, len(filenames))
		for idx, filename := range filenames {
			files[idx] = map[string]any{
				"path":   []string{filename},
				"length": getNumBytes(filename),
			}
		}
		infoDict["files"] = files
	}
	infoDict["pieces"] = calculatePieces(pieceLength, filenames)
	infoDict["piece length"] = pieceLength
	return infoDict
}

func calculatePieces(pieceLength int, filenames []string) string {
	/*
		pieces are calculated based on the continuous byte stream
		of the provided files
	*/
	var pieces bytes.Buffer
	pieceBuffer := bytes.NewBuffer(make([]byte, 0, pieceLength))
	h := sha1.New()
	for _, filename := range filenames {
		f, err := os.Open(filename)
		common.Check(err)
		defer f.Close()
		for {
			bytesToRead := pieceBuffer.Available()
			readBuffer := make([]byte, bytesToRead)
			bytesRead, err := f.Read(readBuffer)
			pieceBuffer.Write(readBuffer[:bytesRead])
			// time to process the next file, if any
			if err == io.EOF {
				break
			}
			// we filled a full piece - let's hash it
			// and append the digest
			if pieceBuffer.Available() == 0 {
				h.Write(pieceBuffer.Bytes())
				pieces.Write(h.Sum(nil))
				pieceBuffer.Reset()
				h.Reset()
			}
		}
	}
	if pieceBuffer.Available() != pieceBuffer.Cap() {
		// we have bytes left in the buffer!
		h.Write(pieceBuffer.Bytes())
		pieces.Write(h.Sum(nil))
	}
	return pieces.String()
}

func CalculateInfoHash(info *data.BEInfo) {
	// infoMap := bencode.ToBencodedDict(info)
}
