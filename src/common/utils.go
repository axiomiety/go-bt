package common

import "fmt"

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func CreateTorrent(announce string, name string, filenames ...string) {
	// build the info dict
	infoDict := map[string]any{}

	torrentMap := map[string]any{
		"announce": announce,
		"info":     infoDict,
	}
	fmt.Printf("torrentMap: %+v", torrentMap)
}
