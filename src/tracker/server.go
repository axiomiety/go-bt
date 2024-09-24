package tracker

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"bytes"
	"fmt"
	"log"
	"net/http"
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
	log.Printf("serving torrents from %s on :%d", t.Directory, t.Port)
	if t.Cache == nil {
		t.Cache = &TrackerCache{
			Interval: 30,
			Store:    map[[20]byte]data.BETrackerResponse{},
		}
	}
	t.loadTorrents()

	http.HandleFunc("/announce", t.announce)
	http.ListenAndServe(fmt.Sprintf(":%d", t.Port), nil)
}

func (t *TrackerServer) announce(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	infoHash := [20]byte{}
	// this has already been decoded for us
	copy(infoHash[:], query.Get("info_hash"))

	sendFailure := func(reason string) {
		response := map[string]any{
			"failure reason": reason,
		}
		buffer := &bytes.Buffer{}
		bencode.Encode(buffer, response)
		w.Write(buffer.Bytes())
	}

	if _, found := t.Cache.Store[infoHash]; found {
		// quick sanity checks on the input
		peerId := query.Get("peer_id")
		peerPortStr := query.Get("port")
		// let's see if we need to add the caller to our list of peers
		fmt.Printf("%s:%s\n", peerId, peerPortStr)

	} else {
		sendFailure("unknown info hash")
	}
}
