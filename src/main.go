package main

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	bencodeCmd := flag.NewFlagSet("bencode", flag.ExitOnError)
	bencodeDecode := bencodeCmd.String("decode", "-", "decode file/stdin")

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createAnnounce := createCmd.String("announce", "", "tracker URL")
	createName := createCmd.String("name", "", "info.name")

	switch os.Args[1] {
	case "bencode":
		bencodeCmd.Parse(os.Args[2:])
		var contents []byte
		var err error
		if *bencodeDecode == "-" {
			contents, err = io.ReadAll(os.Stdin)
			common.Check(err)
		} else {
			contents, err = os.ReadFile(*bencodeDecode)
			common.Check(err)
		}
		obj := bencode.ParseBencoded2(bytes.NewReader(contents)).(map[string]any)
		b, err := json.MarshalIndent(obj, "", "  ")
		common.Check(err)
		fmt.Printf("%s", string(b))
	case "create":
		createCmd.Parse(os.Args[2:])
		common.CreateTorrent(*createAnnounce, *createName, createCmd.Args()...)
	default:
		panic("Unknown option!")
	}
}
