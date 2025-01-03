package data

type BETorrent struct {
	InfoHash     [20]byte
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	Info         BEInfo     `bencode:"info"`
}

type BEInfo struct {
	Name        string   `bencode:"name"`
	PieceLength uint32   `bencode:"piece length"` // bytes per piece
	Pieces      string   `bencode:"pieces"`       // byte string, 20-byte SHA1 for each piece
	Length      uint32   `bencode:"length"`       // of file(s), in bytes
	Files       []BEFile `bencode:"files"`
}

type BEFile struct {
	Path   []string `bencode:"path"`
	Length int      `bencode:"length"`
}

func (i *BEInfo) GetPieceSize(idx uint32) uint32 {
	numPieces := i.GetNumPieces()
	if idx == numPieces-1 {
		return min(i.PieceLength, i.GetTotalLength()%i.PieceLength)
	} else {
		return i.PieceLength
	}
}

func (i *BEInfo) GetTotalLength() uint32 {
	totalLength := i.Length
	if len(i.Files) > 0 {
		for _, file := range i.Files {
			totalLength += uint32(file.Length)
		}
	}
	return totalLength
}

func (i *BEInfo) GetNumPieces() uint32 {
	totalLength := i.GetTotalLength()
	numPieces := totalLength / i.PieceLength
	if totalLength%i.PieceLength > 0 {
		numPieces += 1
	}
	return numPieces
}
