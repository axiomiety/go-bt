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

type Segment struct {
	Filename string
	Offset   uint64
	Length   uint64
}

func GetSegmentsForPiece(i *data.BEInfo, index uint64) []Segment {
	segments := make([]Segment, 0)

	pieceStart := index * i.PieceLength
	pieceRemaining := i.PieceLength
	runningOffset := uint64(0)
	for _, file := range i.Files {
		if pieceRemaining == 0 || runningOffset > pieceStart+pieceRemaining {
			// we're done
			break
		} else if (runningOffset + uint64(file.Length)) < pieceStart {
			// this is beyond the current file's boundary
			runningOffset += uint64(file.Length)
		} else {
			// part of this piece belongs to this file
			fileBytesInPiece := min(runningOffset+uint64(file.Length)-pieceStart, pieceRemaining)
			segments = append(segments, Segment{
				Filename: file.Path[0],
				Offset:   pieceStart - runningOffset,
				Length:   fileBytesInPiece,
			})
			// this may well be 0 now
			pieceRemaining -= fileBytesInPiece
			pieceStart += fileBytesInPiece
			if pieceStart == (runningOffset + uint64(file.Length)) {
				runningOffset += uint64(file.Length)
			}
		}
	}
	return segments
}
