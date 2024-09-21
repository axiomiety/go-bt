package tracker

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"fmt"
	"log"
	"os"
	"strings"
)

type TrackerServer struct {
	Directory string
	Port      int32
	Cache     *TrackerCache
}

type TrackerCache struct {
	Interval int64
	Store    map[[20]byte]data.BETrackerResponse
}

func (t *TrackerServer) loadTorrents() {
	files, err := os.ReadDir(t.Directory)
	common.Check(err)
	for _, filename := range files {
		if strings.HasSuffix(filename.Name(), ".torrent") {
			log.Printf("torrent file found: %s\n", filename.Name())
			fullPath := fmt.Sprintf("%s/%s", t.Directory, filename.Name())
			file, err := os.Open(fullPath)
			common.Check(err)
			defer file.Close()
			btorrent := bencode.ParseFromReader[data.BETorrent](file)
			t.Cache.Store[torrent.CalculateInfoHash(&btorrent.Info)] = data.BETrackerResponse{
				Complete:   0,
				Incomplete: 0,
				Peers:      make([]data.BEPeer, 1),
				Interval:   t.Cache.Interval,
			}
		}
	}
}

func (t *TrackerServer) Serve() {
	log.Printf("serving torrents from %s", t.Directory)
	if t.Cache == nil {
		t.Cache = &TrackerCache{
			Interval: 30,
			Store:    map[[20]byte]data.BETrackerResponse{},
		}
	}
	t.loadTorrents()
}
