package peer

import (
	"axiomiety/go-bt/data"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
	Incoming   chan *data.Message
	Outgoing   chan *data.Message
	BitField   data.BitField
}

func MakePeerHandler(peer *data.BEPeer, peerId [20]byte, infoHash [20]byte) *PeerHandler {
	return &PeerHandler{
		Peer:       peer,
		PeerId:     peerId,
		InfoHash:   infoHash,
		Connection: nil,
		State:      UNSET,
		Incoming:   make(chan *data.Message),
		Outgoing:   make(chan *data.Message),
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
	log.Printf("connected to %s", address)
}

func (p *PeerHandler) Handshake() {
	// a handshake consists of both sending and receiving one!
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		handshakeMsg := data.GetHanshake(p.PeerId, p.InfoHash)
		// fmt.Printf("%+v", handshakeMsg)
		numBytesWritten, err := p.Connection.Write(handshakeMsg.ToBytes())
		if err != nil || numBytesWritten == 0 {
			p.State = ERROR
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 1)
		// it really shouldn't take the peer that long to get back with
		// a handshake - if it does, we're probably not getting anything from them
		p.Connection.SetReadDeadline(time.Now().Add(5 * time.Second))
		numBytesRead, err := io.ReadFull(p.Connection, buf)
		if err != nil && err != io.EOF || numBytesRead == 0 {
			log.Printf("handshake error (pstrlen): %s", err)
			p.State = ERROR
			return
		}
		pstrLength := buf[0]
		buf = make([]byte, 49+pstrLength-1)
		numBytesRead, err = p.Connection.Read(buf)
		if err != nil && err != io.EOF || numBytesRead == 0 {
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
		if peerHandShake.InfoHash != p.InfoHash {
			log.Printf("info_hash doesn't match!")
			p.State = ERROR
		}
		// peer spoofing?
		// if string(peerHandShake.PeerId[:]) != p.Peer.Id {
		// 	log.Printf("peer_id doesn't match!")
		// 	p.State = ERROR
		// }
	}()
	wg.Wait()
	// if we reach here, we're ready!
	if p.State != ERROR {
		p.State = READY
	}
}

func getMessage(conn net.Conn) (*data.Message, error) {

	timeoutWaitDuration := 2 * time.Minute
	conn.SetReadDeadline(time.Now().Add(timeoutWaitDuration))
	header := make([]byte, 4)
	numBytesRead, err := io.ReadFull(conn, header)

	processBadResponse := func(err error, numBytesRead int) (*data.Message, error) {
		if numBytesRead == 0 {
			log.Printf("no data!")
			return &data.Message{}, errors.New("no data")
		} else if os.IsTimeout(err) {
			log.Println("timed out reading length header from client")
			return &data.Message{}, err
		} else {
			return &data.Message{}, err
		}
	}

	if (err != nil && err != io.EOF) || numBytesRead == 0 {
		return processBadResponse(err, numBytesRead)
	}

	length := binary.BigEndian.Uint32(header[:])

	// keep-alive
	if length == 0 {
		return &data.Message{}, nil
	}

	buffer := make([]byte, length)
	numBytesRead, err = io.ReadFull(conn, buffer)
	if (err != nil && err != io.EOF) || numBytesRead == 0 {
		return processBadResponse(err, numBytesRead)
	}

	msg := &data.Message{
		Length:    [4]byte(header),
		MessageId: buffer[0],
	}

	// some messages don't have a payload
	if len(buffer) > 1 {
		msg.Payload = buffer[1:]
	}

	return msg, nil
}

func (p *PeerHandler) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("shutting down listener")
		default:
			msg, err := getMessage(p.Connection)
			if err != nil {
				log.Printf("error: %s", err)
				break
			}
			p.Incoming <- msg
		}
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
	log.Printf("lock 'n load!")
	go p.Listen(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context is done, closing connection to %s", hex.EncodeToString([]byte(p.Peer.Id)))
			p.Connection.Close()
			return
		case msg := <-p.Incoming:
			log.Printf("msg received: %x", msg.MessageId)
			log.Printf("payload: %v", msg.Payload)
		case msg := <-p.Outgoing:
			log.Printf("msg to send: %x", msg.MessageId)
		}
	}
}
