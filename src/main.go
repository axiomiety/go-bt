package main

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/torrent"
	"axiomiety/go-bt/tracker"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
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
		fmt.Printf("hex: %x\nurl: %s\n", digest, tracker.EncodeInfoHash(digest))
	case "tracker":
		trackerCmd.Parse(os.Args[2:])
		obj := getDictFromFile(trackerTorrentFile)
		infoDict := obj["info"].(map[string]any)
		digest := torrent.CalculateInfoHashFromInfoDict(infoDict)
		baseUrl := obj["announce"].(string)
		q := data.TrackerQuery{
			InfoHash: tracker.EncodeInfoHash(digest),
			PeerId:   "-TD18327382910944852",
		}
		fmt.Printf("%s?%s", baseUrl, tracker.ToQueryString(&q))
	default:
		panic("Unknown option!")
	}
}
