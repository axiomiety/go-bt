package main

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/peer"
	"axiomiety/go-bt/torrent"
	"axiomiety/go-bt/tracker"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
)

func main() {
	bencodeCmd := flag.NewFlagSet("bencode", flag.ExitOnError)
	bencodeDecode := bencodeCmd.String("decode", "-", "decode file/stdin")

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createOutputFile := createCmd.String("out", "", "tracker URL")
	createAnnounce := createCmd.String("announce", "", "tracker URL")
	createName := createCmd.String("name", "", "info.name")
	createPieceLength := createCmd.Int("pieceLength", 262144, "length of each piece")

	infoHashCmd := flag.NewFlagSet("infohash", flag.ExitOnError)
	infoHashFile := infoHashCmd.String("file", "-", "file/stdin")

	trackerCmd := flag.NewFlagSet("tracker", flag.ExitOnError)
	trackerTorrentFile := trackerCmd.String("torrent", "-", "file/stdin")
	trackerServe := trackerCmd.Bool("serve", false, "serve")
	trackerDir := trackerCmd.String("dir", "", "directory with torrents, defaults to current directory")
	trackerPort := trackerCmd.Int("port", 8080, "tracker listening port")

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	downloadTorrentFile := downloadCmd.String("torrent", "", "file/stdin")

	handshakeCmd := flag.NewFlagSet("handshake", flag.ExitOnError)
	handshakeTorrentFile := handshakeCmd.String("torrent", "", "file/stdin")
	handshakePeerIp := handshakeCmd.String("ip", "", "IP of peer")
	handshakePeerPort := handshakeCmd.Uint("port", 0, "IP of peer")
	handhsakePeerId := handshakeCmd.String("id", "", "peer ID, in hex")

	switch os.Args[1] {
	case "bencode":
		bencodeCmd.Parse(os.Args[2:])
		obj := bencode.GetDictFromFile(bencodeDecode)
		b, err := json.MarshalIndent(obj, "", "  ")
		common.Check(err)
		fmt.Printf("%s", string(b))
	case "create":
		createCmd.Parse(os.Args[2:])
		torrent.CreateTorrent(*createOutputFile, *createAnnounce, *createName, *createPieceLength, createCmd.Args()...)
	case "infohash":
		infoHashCmd.Parse(os.Args[2:])
		obj := bencode.GetDictFromFile(infoHashFile)
		digest := torrent.CalculateInfoHashFromInfoDict(obj["info"].(map[string]any))
		fmt.Printf("hex: %x\nurl: %s\n", digest, tracker.EncodeBytes(digest))
	case "download":
		downloadCmd.Parse(os.Args[2:])
		manager := peer.FromTorrentFile(*downloadTorrentFile)
		obj := bencode.GetDictFromFile(downloadTorrentFile)
		infoDict := obj["info"].(map[string]any)
		log.Printf("hash of idx 0: %s", hex.EncodeToString([]byte(infoDict["pieces"].(string)[0:20*1])))
		manager.Run()
		log.Printf("manager has shut down")
	case "handshake":
		handshakeCmd.Parse(os.Args[2:])
		obj := bencode.GetDictFromFile(handshakeTorrentFile)
		infoDict := obj["info"].(map[string]any)
		digest := torrent.CalculateInfoHashFromInfoDict(infoDict)
		peerId, err := hex.DecodeString(*handhsakePeerId)
		common.Check(err)
		peerIdB := make([]byte, 20)
		rand.Read(peerId)
		bepeer := data.BEPeer{
			IP:   *handshakePeerIp,
			Port: uint32(*handshakePeerPort),
			Id:   string(peerId),
		}
		ph := peer.MakePeerHandler(&bepeer, [20]byte(peerIdB), digest, infoDict["piece"].(uint32))
		ph.Connect()
		ph.Handshake()
		log.Printf("peer state: %d", ph.State)
	case "tracker":
		trackerCmd.Parse(os.Args[2:])
		if *trackerServe {
			// serve 'em trackers!
			var directory string
			if *trackerDir == "" {
				cwd, err := os.Getwd()
				common.Check(err)
				directory = cwd
			} else {
				directory = *trackerDir
			}
			var mu sync.Mutex
			tracker := &tracker.TrackerServer{
				Directory: directory,
				Port:      int32(*trackerPort),
				Lock:      &mu,
			}
			tracker.Serve()
		} else {
			obj := bencode.GetDictFromFile(trackerTorrentFile)
			infoDict := obj["info"].(map[string]any)
			digest := torrent.CalculateInfoHashFromInfoDict(infoDict)
			baseUrl, err := url.Parse(obj["announce"].(string))
			common.Check(err)
			// generate a random peer ID
			peerId := make([]byte, 20)
			rand.Read(peerId)
			q := data.TrackerQuery{
				InfoHash: tracker.EncodeBytes(digest),
				PeerId:   tracker.EncodeBytes([20]byte(peerId)),
				Port:     6688,
				// Compact:  false,
				// if it's too small, some trackers won't send us peers!
				Left:    45536,
				Numwant: 100,
			}
			resp := tracker.QueryTrackerRaw(baseUrl, &q)
			raw := bencode.ParseBencoded2(bytes.NewReader(resp))
			b, err := json.MarshalIndent(raw, "", "  ")
			common.Check(err)
			fmt.Printf("%s", string(b))
		}
	default:
		panic("Unknown option!")
	}
}
