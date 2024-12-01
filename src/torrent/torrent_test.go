package torrent_test

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"encoding/hex"
	"os"
	"reflect"
	"testing"
)

func TestInfoHash(t *testing.T) {
	file, _ := os.Open("../bencode/testdata/files.torrent")
	defer file.Close()
	btorrent := bencode.ParseFromReader[data.BETorrent](file)
	rawDigest := torrent.CalculateInfoHash(&btorrent.Info)
	infoDigest := hex.EncodeToString(rawDigest[:])
	expectedDigest := "b6e355aa9e2a9b510cf67f0b4be76d9da36ddbbf"
	if infoDigest != expectedDigest {
		t.Errorf("expected %s, got %s", expectedDigest, infoDigest)
	}
}

func TestGetSegmentsForPiece(t *testing.T) {
	// total size is 23 bytes for a total of 3 pieces
	binfo := &data.BEInfo{
		Files: []data.BEFile{
			{
				Path:   []string{"file1"},
				Length: 12,
			},
			{
				Path:   []string{"file2"},
				Length: 4,
			},
			{
				Path:   []string{"file3"},
				Length: 7,
			},
		},
		PieceLength: 10,
	}

	expected := map[int][]torrent.Segment{
		0: {
			{
				Filename: "file1",
				Offset:   uint64(0),
				Length:   10,
			}},
		// this is the most interesting piece - it spans 3 files!
		1: {
			{
				Filename: "file1",
				Offset:   uint64(10),
				Length:   2,
			},
			{
				Filename: "file2",
				Offset:   uint64(0),
				Length:   4,
			},
			{
				Filename: "file3",
				Offset:   uint64(0),
				Length:   4,
			}},
		2: {
			{
				Filename: "file3",
				Offset:   uint64(4),
				Length:   3,
			}},
	}

	for pieceIdx, expectedSegments := range expected {
		segments := torrent.GetSegmentsForPiece(binfo, uint64(pieceIdx))
		if !reflect.DeepEqual(segments, expectedSegments) {
			t.Errorf("expected %+v, got %+v ", expectedSegments, segments)
		}
	}
}
