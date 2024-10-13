package peer

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"axiomiety/go-bt/tracker"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/url"
	"os"
)

type PeerManager struct {
	Torrent         *data.BETorrent
	TrackerResponse *data.BETrackerResponse
	PeerHandlers    []*PeerHandler
	InfoHash        [20]byte
	PeerId          [20]byte
	TrackerURL      url.URL
}

func (p *PeerManager) QueryTracker() {

	q := data.TrackerQuery{
		InfoHash: tracker.EncodeBytes(p.InfoHash),
		PeerId:   tracker.EncodeBytes(p.PeerId),
		// eventually that'll be an option
		Port:    6688,
		Compact: false,
	}
	resp := tracker.QueryTrackerRaw(&p.TrackerURL, &q)
	p.TrackerResponse = bencode.ParseFromReader[data.BETrackerResponse](bytes.NewReader(resp))
}

func (p *PeerManager) UpdatePeers() {
	// TODO: expand
	// there's a ton of stuff we could do here - e.g. if our peers don't cover
	// the blocks we require, we could disconnect and find new ones
	// we can also discard peers we've had trouble connecting to in the past,
	// or ones that are chocked
	if len(p.PeerHandlers) < 5 {
		for _, peer := range p.TrackerResponse.Peers {
			// let's not try to connect to ourselves
			if peer.Id != string(p.PeerId[:]) {
				log.Printf("enquing peer %s", hex.EncodeToString([]byte(peer.Id)))
				handler := MakePeerHandler(&peer, p.PeerId)
				log.Printf("handler.Peer: %s", handler.Peer.Id)
				p.PeerHandlers = append(p.PeerHandlers, handler)
			}
		}
	}
	// we should probably drop ones that are in a bad state
	for _, handler := range p.PeerHandlers {
		log.Printf("peerHandler: remote peer %s, state=%d", hex.EncodeToString([]byte(handler.Peer.Id)), handler.State)
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
		PeerHandlers: make([]*PeerHandler, 0),
		PeerId:       [20]byte(peerId),
		TrackerURL:   *baseUrl,
	}
}

func (p *PeerManager) Run() {
	log.Printf("peerManager ID: %s", hex.EncodeToString(p.PeerId[:]))
	p.QueryTracker()
	p.UpdatePeers()
	p.PeerHandlers[0].connect()
}
