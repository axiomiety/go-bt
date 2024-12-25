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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"sort"
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
	BitField        data.BitField
	PeerPoolSize    int
	BaseDirectory   string
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

func (p *PeerManager) ejectPeersInErrorState() {
	p.PeerHandlerLock.Lock()
	defer p.PeerHandlerLock.Unlock()
	peersToRemove := []string{}
	for peerId, peer := range p.PeerHandlers {
		if peer.State == ERROR {
			peersToRemove = append(peersToRemove, peerId)
		}
	}
	for _, peerId := range peersToRemove {
		log.Printf("dropping peer %s because it is in an ERROR state", hex.EncodeToString([]byte(peerId)))
		delete(p.PeerHandlers, peerId)
	}
}

func (p *PeerManager) ejectNotSoUsefulPeers() int {
	availability := p.GetPiecesAvailability()

	// we key by score
	ordered := map[uint32][]*PeerHandler{}
	keys := make([]uint32, len(p.PeerHandlers))

	for _, peer := range p.PeerHandlers {
		score := p.GetPeerScore(availability, peer)
		level := ordered[score]
		level = append(level, peer)
		ordered[score] = level
		keys = append(keys, score)
	}

	// lowest score first!
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	// eject up to 2 peers at a time - it's pretty arbitrary though...
	numToEject := 2
	// we'll cap this to peers with a score of 2 or lower
	numEjected := 0

	p.PeerHandlerLock.Lock()
	defer p.PeerHandlerLock.Unlock()

	for _, score := range keys {
		for _, peer := range ordered[score] {
			if numToEject == numEjected {
				break
			}
			peerId := string([]byte(peer.PeerId[:]))
			delete(p.PeerHandlers, peerId)
			log.Printf("dropping peer %s because of its low score: %d", hex.EncodeToString(peer.PeerId[:]), score)
			numEjected += 1
		}
	}
	return numEjected
}

func (p *PeerManager) UpdatePeers() {
	// TODO: expand
	// there's a ton of stuff we could do here - e.g. if our peers don't cover
	// the blocks we require, we could disconnect and find new ones
	// we can also discard peers we've had trouble connecting to in the past,
	// or ones that are chocked

	// start by ejecting peers
	p.ejectPeersInErrorState()
	p.ejectNotSoUsefulPeers()

	// if we have space in our peer pool, try to add a new one!
	if len(p.PeerHandlers) < p.PeerPoolSize {
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
				handler := MakePeerHandler(&myPeer, p.PeerId, p.InfoHash, uint32(len(p.Torrent.Info.Pieces)))
				p.PeerHandlers[peer.Id] = handler
				// now establish a connection!
				// TODO: mmm - each handler should have its own context
				go handler.Loop(p.Context)
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
	t := bencode.ParseFromReader[data.BETorrent](file)
	return &PeerManager{
		Torrent:         t,
		InfoHash:        digest,
		PeerHandlers:    make(map[string]*PeerHandler),
		PeerHandlerLock: &mu,
		PeerId:          [20]byte(peerId),
		TrackerURL:      *baseUrl,
		BitField: data.BitField{
			Field: make([]byte, len(t.Info.Pieces)/20),
		},
		// hard-coded for now
		PeerPoolSize:  5,
		BaseDirectory: "/tmp",
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

func (p *PeerManager) DownloadNextPiece() bool {
	didAnything := false

	// is there a race condition vs when that's updated?
	pendingPieces := map[uint32]bool{}
	for _, handler := range p.PeerHandlers {
		if handler.State == REQUESTING_PIECE {
			pendingPieces[handler.PendingPiece.Index] = true
		}
	}
	for pieceNum := range p.BitField.NumPieces() {
		_, pieceIsBeingDownloaded := pendingPieces[pieceNum]
		if !p.BitField.HasPiece(pieceNum) && !pieceIsBeingDownloaded {
			for peerId, handler := range p.PeerHandlers {
				if handler.State == UNCHOKED && handler.BitField.HasPiece(pieceNum) {
					log.Printf("peer %x is UNCHOKED and has piece %d", peerId, pieceNum)
					// usually we'd request PIECE_LENGTH, but if this is e.g. the last
					// piece, the size of the piece may be less than the piece size
					// specified in the info dict
					handler.RequestPiece(pieceNum, p.Torrent.Info.GetPieceSize(pieceNum))
					didAnything = true
				}
			}
		}
	}
	return didAnything
}

func (p *PeerManager) PeerHasPieceOfInterest(h *PeerHandler) bool {
	// a peer has a piece of interest if we don't already have it
	for idx := range p.BitField.NumPieces() {
		if !p.BitField.HasPiece(idx) && h.BitField.HasPiece(idx) {
			return true
		}
	}
	return false
}

func (p *PeerManager) GetPiecesAvailability() map[uint32]uint32 {
	availability := map[uint32]uint32{}
	for idx := range p.BitField.NumPieces() {
		if !p.BitField.HasPiece(idx) {
			availability[idx] = 0
			for _, peerHandler := range p.PeerHandlers {
				if peerHandler.BitField.HasPiece(idx) {
					availability[idx] += 1
				}
			}
		}
	}
	return availability
}

func GetPiecesScore(b data.BitField, availability map[uint32]uint32, numPeers uint32) uint32 {
	score := uint32(0)
	for pieceIdx, numPeersWithPiece := range availability {
		if b.HasPiece(pieceIdx) {
			score += 1 + numPeers - numPeersWithPiece
		}
	}
	return score
}

func (p *PeerManager) GetPeerScore(availability map[uint32]uint32, h *PeerHandler) uint32 {
	// not yet unchocked!
	if p.PeerHasPieceOfInterest(h) {
		score := GetPiecesScore(h.BitField, availability, uint32(len(p.PeerHandlers)))
		if h.State == READY {
			// we're chocked - halve the score
			return score / 2
		} else {
			return score
		}
	} else if h.State == UNCHOKED {
		// peer is unchocked but it doesn't currently have
		// any piece we're interested in
		return 1
	} else {
		// no interest here!
		return 0
	}
}

func (p *PeerManager) processCompletedPieces() {
	for peerId, peer := range p.PeerHandlers {
		if peer.State == PIECE_COMPLETE {
			h := sha1.New()
			h.Write(peer.PendingPiece.Data)
			digest := h.Sum(nil)
			pieceIdx := peer.PendingPiece.Index
			log.Printf("downloaded piece %d from peer %s with sha1: %s", pieceIdx, hex.EncodeToString([]byte(peerId)), hex.EncodeToString(digest))
			expectedDigest := []byte(p.Torrent.Info.Pieces[pieceIdx*20 : (pieceIdx+1)*20])
			if bytes.Equal(expectedDigest, digest) {
				segments := torrent.GetSegmentsForPiece(&p.Torrent.Info, pieceIdx)
				torrent.WriteSegments(segments, peer.PendingPiece.Data, p.BaseDirectory)
				p.BitField.SetPiece(pieceIdx)
				//TODO: let all other peers know we have a new piece
			} else {
				log.Printf("digest mismatch - expected %s, got %s", hex.EncodeToString(expectedDigest), hex.EncodeToString(digest))
			}
			// reset the state - it's ready to download pieces again
			peer.State = READY
		}
	}
}

func (p *PeerManager) Run() {
	log.Printf("peerManager ID (ours): %s", hex.EncodeToString(p.PeerId[:]))
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		cancelFunc()
	}()
	p.Context = ctx

	// periodic ping to the tracker to ensure we still show up
	// as a valid peer
	// TODO: this should ideally have the right event sent out
	// to the tracker depending on which state we're in
	go func(ctx context.Context) {
		p.queryTrackerAndUpdatePeersList(ctx)
		time.Sleep(30 * time.Second)
	}(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:

			p.processCompletedPieces()
			if p.DownloadNextPiece() {
				log.Print("found new piece(s) to download!")
			} else {
				log.Print("nothing to download - but are we complete?")
			}
		}
		time.Sleep(5 * time.Second)
	}
}
