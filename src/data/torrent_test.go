package data

import (
	"testing"
)

func TestGetPieceLength(t *testing.T) {
	// sample info struct with 2 files
	beinfo := BEInfo{
		Name:        "foo",
		PieceLength: 100,
		Files: []BEFile{
			{
				Path:   []string{"path1"},
				Length: 90,
			},
			{
				Path:   []string{"path2"},
				Length: 50,
			},
		},
	}

	numPieces := beinfo.GetNumPieces()
	if numPieces != 2 {
		t.Errorf("expected 2 pieces, got %d", numPieces)
	}
	pieceSize := beinfo.GetPieceSize(0)
	if pieceSize != beinfo.PieceLength {
		t.Errorf("expected piece 0 to be of size %d, got %d instead", beinfo.PieceLength, pieceSize)
	}
	pieceSize = beinfo.GetPieceSize(1)
	if beinfo.GetPieceSize(1) != 40 {
		t.Errorf("expected piece 1 to be of size %d, got %d instead", 40, pieceSize)
	}
}
