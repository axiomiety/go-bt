package data

type BEPeer struct {
	Id   string `bencode:"peer_id"`
	IP   string `bencode:"ip"`
	Port uint32 `bencode:"port"`
}

type BETrackerResponse struct {
	Complete   int64    `bencode:"complete"`   // seeds
	Incomplete int64    `bencode:"incomplete"` // leechers
	Interval   int64    `bencode:"interval"`   // in seconds
	Peers      []BEPeer `bencode:"peers"`
}

type TrackerQuery struct {
	InfoHash   string `url:"info_hash"`
	PeerId     string `url:"peer_id"`
	Port       uint   `url:"port"`
	Uploaded   uint   `url:"uploaded"`
	Downloaded uint   `url:"downloaded"`
	Left       uint   `url:"left"`
	Event      string `url:"event"`
	// Compact    bool   `url:"compact"`
	Numwant uint `url:"numwant"`
}
