package data

type BETorrent struct {
	InfoHash     [20]byte
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	Info         BEInfo     `bencode:"info"`
}

type BEInfo struct {
	Name        string   `bencode:"name"`
	PieceLength uint64   `bencode:"piece length"` // bytes per piece
	Pieces      string   `bencode:"pieces"`       // byte string, 20-byte SHA1 for each piece
	Length      uint64   `bencode:"length"`       // of file(s), in bytes
	Files       []BEFile `bencode:"files"`
}

type BEFile struct {
	Path   []string `bencode:"path"`
	Length int      `bencode:"length"`
}
