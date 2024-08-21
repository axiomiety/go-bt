package data

type BETorrent struct {
	InfoHash     [20]byte
	Announce     string   `bencode:"announce"`
	AnnounceList []string `bencode:"annouce-list"`
	Info         BEInfo   `bencode:"info"`
}

type BEInfo struct {
	Name        string `bencode:"name"`
	PieceLength uint64 // bytes per piece
	Pieces      string // byte string, 20-byte SHA1 for each piece
	Length      uint64 `bencode:"length"` // of file(s), in bytes
}
