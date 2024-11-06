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
	"sync"
	"time"
)

type PeerManager struct {
	Torrent         *data.BETorrent
	TrackerResponse *data.BETrackerResponse
	PeerHandlers    map[string]*PeerHandler
	PeerHandlerLock *sync.Mutex
	InfoHash        [20]byte
	Context         context.Context
	PeerId          [20]byte
	TrackerURL      url.URL
	BitFied         data.BitField
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
		p.PeerHandlerLock.Lock()
		defer p.PeerHandlerLock.Unlock()
		for _, peer := range p.TrackerResponse.Peers {
			// do we know the peer?
			if _, ok := p.PeerHandlers[peer.Id]; ok {
				log.Printf("peer %s is already known, skipping", hex.EncodeToString([]byte(peer.Id)))
				continue
			}
			// let's not try to connect to ourselves
			// TODO: fix that hack
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

	var mu sync.Mutex
	return &PeerManager{
		Torrent:         bencode.ParseFromReader[data.BETorrent](file),
		InfoHash:        digest,
		PeerHandlers:    make(map[string]*PeerHandler),
		PeerHandlerLock: &mu,
		PeerId:          [20]byte(peerId),
		TrackerURL:      *baseUrl,
	}
}

func (p *PeerManager) queryTrackerAndUpdatePeersList(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		p.QueryTracker()
		p.UpdatePeers()
	}
}

func (p *PeerManager) refreshPeerPool(ctx context.Context) {
	peersToRemove := []string{}
	for peerId, peer := range p.PeerHandlers {
		switch peer.State {
		case UNSET:
			// likely a new peer - let's connect!
			go peer.Loop(ctx)
		case ERROR:
			peersToRemove = append(peersToRemove, peerId)
		}
	}

	p.PeerHandlerLock.Lock()
	defer p.PeerHandlerLock.Unlock()
	for _, peerId := range peersToRemove {
		log.Printf("dropping peer %s because it is in an ERROR state", hex.EncodeToString([]byte(peerId)))
		delete(p.PeerHandlers, peerId)
	}
}

func (p *PeerManager) selectPieceToDownload() uint64 {
	return 0
}

func (p *PeerManager) assignPieceToPeer(idx uint64) string {
	// iterate through each peer to see if they have the piece in question
	for _, peer := range p.PeerHandlers {
		if peer.State == READY && peer.BitField.HasBlock(idx) {
			peer.DownloadPiece(idx)
			return peer.Peer.Id
		}
	}
}

func (p *PeerManager) Run() {
	log.Printf("peerManager ID (ours): %s", hex.EncodeToString(p.PeerId[:]))
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		cancelFunc()
	}()
	for _, peer := range p.PeerHandlers {
		if peer.State == UNSET {
			// eventually this will need to go into a goroutine
			go peer.Loop(ctx)
		}
	}

	// periodic ping to the tracker to ensure we still show up
	// as a valid peer
	// TODO: this should ideally have the right event sent out
	// to the tracker
	go func(ctx context.Context) {
		p.queryTrackerAndUpdatePeersList(ctx)
		time.Sleep(30 * time.Second)
	}(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			p.refreshPeerPool(ctx)
			pieceIdx := p.selectPieceToDownload()
			p.assignPieceToPeer(pieceIdx)
		}
		time.Sleep(5 * time.Second)
	}
}
