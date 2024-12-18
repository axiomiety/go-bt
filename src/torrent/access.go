package torrent

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"bytes"
	"crypto/sha1"
	"io"
	"os"
	"path"
)

func CalculateInfoHash(info *data.BEInfo) [20]byte {
	return CalculateInfoHashFromInfoDict(bencode.ToDict(*info))
}

func CalculateInfoHashFromInfoDict(info map[string]any) [20]byte {
	var buf bytes.Buffer
	bencode.Encode(&buf, info)
	return sha1.Sum(buf.Bytes())
}

type Segment struct {
	Filename string
	Offset   uint32
	Length   uint32
}

func GetSegmentsForPiece(i *data.BEInfo, index uint32) []Segment {
	segments := make([]Segment, 0)

	pieceStart := index * i.PieceLength
	bytesRemainingInPiece := i.PieceLength
	runningOffset := uint32(0)
	for _, file := range i.Files {
		if bytesRemainingInPiece == 0 || runningOffset > pieceStart+bytesRemainingInPiece {
			// we're done
			break
		} else if (runningOffset + uint32(file.Length)) < pieceStart {
			// this is beyond the current file's boundary
			runningOffset += uint32(file.Length)
		} else {
			// part of this piece belongs to this file
			fileBytesInPiece := min(runningOffset+uint32(file.Length)-pieceStart, bytesRemainingInPiece)
			segments = append(segments, Segment{
				Filename: file.Path[0],
				Offset:   pieceStart - runningOffset,
				Length:   fileBytesInPiece,
			})
			// this may well be 0 now
			bytesRemainingInPiece -= fileBytesInPiece
			pieceStart += fileBytesInPiece
			// if we're at a file boundary we should move on to the next one
			if pieceStart == (runningOffset + uint32(file.Length)) {
				runningOffset += uint32(file.Length)
			}
		}
	}
	return segments
}

func WriteSegments(segments []Segment, data []byte, baseDir string) {
	dataOffset := 0
	for _, segment := range segments {
		filePath := path.Join(baseDir, segment.Filename)
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
		common.Check(err)
		defer file.Close()
		writer := io.NewOffsetWriter(file, int64(segment.Offset))
		writer.Write(data[dataOffset : dataOffset+int(segment.Length)])
		dataOffset += int(segment.Length)
	}
}
