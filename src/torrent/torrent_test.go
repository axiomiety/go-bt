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
	file, _ := os.Open("../bencode/testdata/files.torrent")
	defer file.Close()
	btorrent := bencode.ParseFromReader[data.BETorrent](file)
	binfo := btorrent.Info

	segments := torrent.GetSegmentsForPiece(&binfo, 0)
	expected := []torrent.Segment{
		{
			Filename: "/tmp/files/file1",
			Offset:   uint64(0),
			Length:   binfo.PieceLength,
		},
	}
	if !reflect.DeepEqual(segments, expected) {
		t.Errorf("expected %+v, got %+v ", expected, segments)
	}

}
