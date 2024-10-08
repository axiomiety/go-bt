package main

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"axiomiety/go-bt/tracker"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"
)

func getDictFromFile(file *string) map[string]any {
	var contents []byte
	var err error
	if *file == "-" {
		contents, err = io.ReadAll(os.Stdin)
		common.Check(err)
	} else {
		contents, err = os.ReadFile(*file)
		common.Check(err)
	}
	return bencode.ParseBencoded2(bytes.NewReader(contents)).(map[string]any)
}

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

	switch os.Args[1] {
	case "bencode":
		bencodeCmd.Parse(os.Args[2:])
		obj := getDictFromFile(bencodeDecode)
		b, err := json.MarshalIndent(obj, "", "  ")
		common.Check(err)
		fmt.Printf("%s", string(b))
	case "create":
		createCmd.Parse(os.Args[2:])
		torrent.CreateTorrent(*createOutputFile, *createAnnounce, *createName, *createPieceLength, createCmd.Args()...)
	case "infohash":
		infoHashCmd.Parse(os.Args[2:])
		obj := getDictFromFile(infoHashFile)
		digest := torrent.CalculateInfoHashFromInfoDict(obj["info"].(map[string]any))
		fmt.Printf("hex: %x\nurl: %s\n", digest, tracker.EncodeBytes(digest))
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
			obj := getDictFromFile(trackerTorrentFile)
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
				Compact:  false,
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
