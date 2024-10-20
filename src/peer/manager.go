package peer

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"axiomiety/go-bt/tracker"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"
)

type PeerManager struct {
	Torrent         *data.BETorrent
	TrackerResponse *data.BETrackerResponse
	PeerHandlers    map[string]*PeerHandler
	InfoHash        [20]byte
	Context         context.Context
	PeerId          [20]byte
	TrackerURL      url.URL
}

func (p *PeerManager) QueryTracker() {

	q := data.TrackerQuery{
		InfoHash: tracker.EncodeBytes(p.InfoHash),
		PeerId:   tracker.EncodeBytes(p.PeerId),
		// eventually that'll be an option
		Port: 6688,
		// Compact: false,
	}
	resp := tracker.QueryTrackerRaw(&p.TrackerURL, &q)
	p.TrackerResponse = bencode.ParseFromReader[data.BETrackerResponse](bytes.NewReader(resp))
	log.Print("tracker responded")
}

func (p *PeerManager) UpdatePeers() {
	// TODO: expand
	// there's a ton of stuff we could do here - e.g. if our peers don't cover
	// the blocks we require, we could disconnect and find new ones
	// we can also discard peers we've had trouble connecting to in the past,
	// or ones that are chocked
	if len(p.PeerHandlers) < 5 {
		for _, peer := range p.TrackerResponse.Peers {
			// do we know the peer?
			if _, ok := p.PeerHandlers[peer.Id]; ok {
				log.Printf("peer %s is already known, skipping", hex.EncodeToString([]byte(peer.Id)))
				continue
			}
			// let's not try to connect to ourselves
			if peer.Id != string(p.PeerId[:]) && peer.Port != 6688 {
				log.Printf("enquing peer %s - %s", hex.EncodeToString([]byte(peer.Id)), net.JoinHostPort(peer.IP, fmt.Sprintf("%d", peer.Port)))
				// we're using a range - peer gets reassigned
				// at every iteration! c.f. the below for a more in-depth explanation
				// https://medium.com/swlh/use-pointer-of-for-range-loop-variable-in-go-3d3481f7ffc9
				myPeer := peer
				handler := MakePeerHandler(&myPeer, p.PeerId, p.InfoHash)
				p.PeerHandlers[peer.Id] = handler
			}
		}
	}
	// we should probably drop ones that are in a bad state
	for _, handler2 := range p.PeerHandlers {
		log.Printf("peerHandler: remote peer %s, state=%d", hex.EncodeToString([]byte(handler2.Peer.Id)), handler2.State)
	}
}

func FromTorrentFile(filename string) *PeerManager {
	obj := bencode.GetDictFromFile(&filename)
	infoDict := obj["info"].(map[string]any)
	digest := torrent.CalculateInfoHashFromInfoDict(infoDict)

	baseUrl, err := url.Parse(obj["announce"].(string))
	common.Check(err)

	// generate a random peer ID
	peerId := make([]byte, 20)
	rand.Read(peerId)
	// maybe we should read this once only o_O
	file, _ := os.Open(filename)
	defer file.Close()

	return &PeerManager{
		Torrent:      bencode.ParseFromReader[data.BETorrent](file),
		InfoHash:     digest,
		PeerHandlers: make(map[string]*PeerHandler),
		PeerId:       [20]byte(peerId),
		TrackerURL:   *baseUrl,
	}
}

func (p *PeerManager) Run() {
	log.Printf("peerManager ID: %s", hex.EncodeToString(p.PeerId[:]))
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		cancelFunc()
	}()
	p.QueryTracker()
	p.UpdatePeers()
	for _, peer := range p.PeerHandlers {
		if peer.State == UNSET {
			// eventually this will need to go into a goroutine
			go peer.Loop(ctx)
		}
	}
	time.Sleep(20 * time.Second)
}
