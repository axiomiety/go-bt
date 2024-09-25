package tracker

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TrackerServer struct {
	Directory string
	Port      int32
	Cache     *TrackerCache
	Lock      *sync.Mutex
}

type TrackerCache struct {
	Interval int64
	Store    map[[20]byte]data.BETrackerResponse
	// keyed by info hash -> peer ID -> time
	PeersLastSeen map[[20]byte]map[string]time.Time
}

func (t *TrackerServer) removeStalePeers(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(t.Cache.Interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.Lock.Lock()
			defer t.Lock.Unlock()
			now := time.Now()
			// this is essentially n seconds ago - any peer that hasn't
			// given us a hearbeat since is considered stale
			timeThreshold := now.Add(time.Duration(-t.Cache.Interval) * time.Second)

			for infoHash, peers := range t.Cache.PeersLastSeen {
				peerIdsToRemove := make([]string, 0)
				for peerId, lastSeen := range peers {
					if lastSeen.Before(timeThreshold) {
						peerIdsToRemove = append(peerIdsToRemove, peerId)
					}
				}
				// now update the tracker resposne by removing the stale peers
				trackerResponse := t.Cache.Store[infoHash]
				existingPeers := trackerResponse.Peers
				peersToKeep := make([]data.BEPeer, 0)
				// this is quadratic but :shrug:
				for _, peer := range existingPeers {
					shouldRemove := false
					for _, peerId := range peerIdsToRemove {
						if peer.Id == peerId {
							shouldRemove = true
							break
						}
					}
					// this is a good peer, let's add it back
					if !shouldRemove {
						peersToKeep = append(peersToKeep, peer)
					} else {
						log.Printf("evicted peer ID %s from %s", peer.Id, infoHash)
					}
				}
				trackerResponse.Peers = peersToKeep
				t.Cache.Store[infoHash] = trackerResponse
			}
		}
	}

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
				Peers:      make([]data.BEPeer, 0),
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go t.removeStalePeers(ctx)

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

	if trackerResponse, found := t.Cache.Store[infoHash]; found {
		// this is our own special "key" - if it's provided we'll just
		// return what we already have
		if query.Get("quiet") == "" {
			// don't bother parsing anything, just return the response
			peerId := query.Get("peer_id")
			peerPortStr := query.Get("port")

			// technically it's 32bit, but ParseInt always returns an int64
			var parsedPeerPort int64
			var err error
			if parsedPeerPort, err = strconv.ParseInt(peerPortStr, 10, 32); err != nil {
				sendFailure(fmt.Sprintf("unable to parse port=%s", peerPortStr))
				return
			}
			if len([]byte(peerId)) != 20 {
				sendFailure("peer_id should be 20-bytes long")
				return
			}
			peerPort := uint32(parsedPeerPort)
			peerIp, _, err := net.SplitHostPort(req.RemoteAddr)
			common.Check(err)
			found := false
			// let's see if we already have a peer with that ID
			for _, peer := range trackerResponse.Peers {
				if peer.Id == peerId {
					// update the port, just in case
					peer.Port = peerPort
					peer.IP = peerIp
					found = true
					break
				}
			}
			if !found {
				newPeer := data.BEPeer{
					Id:   peerId,
					IP:   peerIp,
					Port: peerPort,
				}
				trackerResponse.Peers = append(trackerResponse.Peers, newPeer)
			}
			buffer := &bytes.Buffer{}
			bencode.Encode(buffer, bencode.ToDict(trackerResponse))
			w.Write(buffer.Bytes())
		}

	} else {
		sendFailure("unknown info hash")
	}
}
