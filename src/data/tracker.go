package data

type BEPeer struct {
	Id   string `bencode:"peer id"`
	IP   string `bencode:"ip"`
	Port int64  `bencode:"port"`
}

type BETrackerResponse struct {
	Complete   int64    `bencode:"complete"`   // seeds
	Incomplete int64    `bencode:"incomplete"` // leechers
	Interval   int64    `bencode:"interval"`   // in seconds
	Peers      []BEPeer `bencode:"peers"`
}
