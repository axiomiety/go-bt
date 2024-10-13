package peer

import (
	"axiomiety/go-bt/data"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type StateType int

const (
	UNSET = iota
	ERROR
	READY
)

type PeerHandler struct {
	Peer       *data.BEPeer
	PeerId     [20]byte
	InfoHash   [20]byte
	Connection net.Conn
	State      StateType
	Context    context.Context
	Incoming   chan data.Message
	Outgoing   chan data.Message
}

func MakePeerHandler(peer *data.BEPeer, peerId [20]byte) *PeerHandler {
	return &PeerHandler{
		Peer:       peer,
		PeerId:     peerId,
		Connection: nil,
		State:      UNSET,
		Context:    nil,
		Incoming:   make(chan data.Message),
		Outgoing:   make(chan data.Message),
	}
}

func (p *PeerHandler) connect() {
	address := net.JoinHostPort(p.Peer.IP, fmt.Sprintf("%d", p.Peer.Port))
	conn, err := net.DialTimeout("tcp", address, time.Second*5)
	if err != nil {
		log.Printf("error connecting to peer %s: %s", hex.EncodeToString([]byte(p.Peer.Id)), err)
		p.State = ERROR
		return
	}
	p.Connection = conn
}

func (p *PeerHandler) handshake() {
	// a handshake consists of both sending and receiving one!
	// let's add a timer so we don't wait for the peer indefinitely
	var wg sync.WaitGroup
	wg.Add(2)
	// read the handshake from the peer
	go func() {
		defer wg.Done()
		buf := make([]byte, 1)
		_, err := p.Connection.Read(buf)
		if err != nil {
			p.State = ERROR
			return
		}
		pstrLength := buf[0]
		buf = make([]byte, 49+pstrLength-1)
		_, err = p.Connection.Read(buf)
		if err != nil {
			p.State = ERROR
			return
		}
		peerHandShake := data.Handshake{
			PstrLen:  pstrLength,
			Pstr:     buf[1:pstrLength],
			Reserved: [8]byte(buf[pstrLength : pstrLength+8]),
			InfoHash: [20]byte(buf[pstrLength+8 : pstrLength+8+20]),
			PeerId:   [20]byte(buf[pstrLength+8+20:]),
		}
		// validate it all matches
		log.Printf("hs: %v", peerHandShake)
	}()
	// and send ours
	go func() {
		defer wg.Done()
		handshakeMsg := data.GetHanshake(string(p.PeerId[:]), p.InfoHash)
		_, err := p.Connection.Write(handshakeMsg.ToBytes())
		if err != nil {
			p.State = ERROR
		}
	}()
	wg.Wait()
	// if we reach here, we're ready!
	p.State = READY
}

func (p *PeerHandler) Loop() {
	p.connect()
	defer p.Connection.Close()
	p.handshake()
	for {
		select {
		case <-p.Context.Done():
			log.Printf("Context is done, closing connection to %s", hex.EncodeToString([]byte(p.Peer.Id)))
			p.Connection.Close()
			return
		case msg := <-p.Incoming:
			log.Printf("msg received: %x", msg.MessageId)
		case msg := <-p.Outgoing:
			log.Printf("msg to send: %x", msg.MessageId)
		}
	}
}
