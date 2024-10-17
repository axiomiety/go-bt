package peer

import (
	"axiomiety/go-bt/data"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
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
	Incoming   chan data.Message
	Outgoing   chan data.Message
}

func MakePeerHandler(peer *data.BEPeer, peerId [20]byte) *PeerHandler {
	return &PeerHandler{
		Peer:       peer,
		PeerId:     peerId,
		Connection: nil,
		State:      UNSET,
		Incoming:   make(chan data.Message),
		Outgoing:   make(chan data.Message),
	}
}

func (p *PeerHandler) Connect() {
	address := net.JoinHostPort(p.Peer.IP, fmt.Sprintf("%d", p.Peer.Port))
	conn, err := net.DialTimeout("tcp", address, time.Second*5)
	if err != nil {
		log.Printf("error connecting to peer %s: %s", hex.EncodeToString([]byte(p.Peer.Id)), err)
		p.State = ERROR
		return
	}
	p.Connection = conn
	log.Printf("connected! %s %s", address, err)
}

func (p *PeerHandler) Handshake() {
	// a handshake consists of both sending and receiving one!
	// TODO: let's add a timer so we don't wait for the peer indefinitely
	// var wg sync.WaitGroup
	func() {
		// defer wg.Done()
		handshakeMsg := data.GetHanshake(string(p.PeerId[:]), p.InfoHash)
		numBytesWritten, err := p.Connection.Write(handshakeMsg.ToBytes())
		if err != nil {
			p.State = ERROR
		}
		log.Printf("sent hs: %d, err %s", numBytesWritten, err)
	}()
	func() {
		// defer wg.Done()
		buf := make([]byte, 1)
		// numBytesRead1, err := p.Connection.Read(buf)
		numBytesRead1, err := io.ReadFull(p.Connection, buf)
		log.Printf("read data1: %d	, %v", numBytesRead1, buf)
		if err != nil {
			log.Printf("handshake error (pstrlen): %s", err)
			p.State = ERROR
			return
		}
		pstrLength := buf[0]
		buf = make([]byte, 49+pstrLength-1)
		log.Printf("read data2")
		numBytesRead2, err := p.Connection.Read(buf)
		log.Printf("read data3: %d", numBytesRead2)
		if err != nil {
			log.Printf("handshake error: %s", err)
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

	// if we reach here, we're ready!
	if p.State != ERROR {
		p.State = READY
	}
}

func (p *PeerHandler) Loop(ctx context.Context) {
	p.Connect()
	if p.State == ERROR {
		return
	}
	defer p.Connection.Close()
	p.Handshake()
	if p.State == ERROR {
		return
	}
	log.Printf("peer read? %d", p.State)
	for {
		select {
		case <-ctx.Done():
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
