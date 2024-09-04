package bencode_test

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/data"
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestBencodeDecode(t *testing.T) {

	testCases := []struct {
		data []byte
	}{
		{[]byte("i-42e")},
		{[]byte("3:foo")},
		{[]byte("12:foobarraboof")},
		{[]byte("li42ee")},
		{[]byte("li42ei43ee")},
		{[]byte("d3:fooi42ee")},
		{[]byte("d3:fooli42eee")},
		{[]byte("d3:fooi42e3:zari1ee")},
	}

	buf := &bytes.Buffer{}
	for _, testCase := range testCases {
		buf.Reset()
		bencode.Encode(buf, bencode.ParseBencoded2(bytes.NewReader(testCase.data)))
		if !bytes.Equal(buf.Bytes(), testCase.data) {
			t.Errorf("expected %s, got %s", testCase.data, buf.Bytes())
		}
	}
}
func TestBencodeRecursiveParser(t *testing.T) {

	// negative int!
	r := bytes.NewReader([]byte("i-42e"))
	ret := bencode.ParseBencoded2(r)
	if ret != -42 {
		t.Errorf("expected -42, got %v", ret)
	}

	// string, below 10 chars
	r = bytes.NewReader([]byte("3:foo"))
	ret = bencode.ParseBencoded2(r).(string)
	if ret != "foo" {
		t.Errorf("expected 'foo', got %v", ret)
	}

	// string, above 10 chars
	r = bytes.NewReader([]byte("12:foobarraboof"))
	ret = bencode.ParseBencoded2(r).(string)
	if ret != "foobarraboof" {
		t.Errorf("expected 'foo', got %v", ret)
	}

	// list with one int
	r = bytes.NewReader([]byte("li42ee"))
	retSlice, _ := bencode.ParseBencoded2(r).([]any)
	if len(retSlice) != 1 && retSlice[0] != 42 {
		t.Errorf("expected [42], got %v", ret)
	}

	// list with two items
	r = bytes.NewReader([]byte("li42ei43ee"))
	retSlice, _ = bencode.ParseBencoded2(r).([]any)
	if len(retSlice) != 2 && retSlice[0] != 42 && retSlice[1] != 43 {
		t.Errorf("expected [42, 43], got %v", ret)
	}

	// a simple map
	r = bytes.NewReader([]byte("d3:foo3:bare"))
	// r = bytes.NewReader([]byte("d3:fooi42ee"))
	retMap, _ := bencode.ParseBencoded2(r).(map[string]any)
	if retMap["foo"] != "bar" {
		t.Errorf("expected {'foo': 'bar'}, got %v", retMap)
	}

	// a map with a list
	r = bytes.NewReader([]byte("d3:fooli42eee"))
	retMap, _ = bencode.ParseBencoded2(r).(map[string]any)
	retSlice = retMap["foo"].([]interface{})
	if len(retSlice) != 1 && retSlice[0] != 42 {
		t.Errorf("expected {'foo': [42]}, got %v", ret)
	}
}

func TestBencodeParsing(t *testing.T) {

	testCases := []struct {
		f func(io.Reader) interface{}
	}{
		// {bencode.ParseBencoded},
		{bencode.ParseBencoded2},
	}
	for _, tc := range testCases {
		// single integer
		r := bytes.NewReader([]byte("i42e"))
		ret := tc.f(r)
		if ret != 42 {
			t.Errorf("expected 42, got %v", ret)
		}

		// string, below 10 chars
		r = bytes.NewReader([]byte("3:foo"))
		ret = tc.f(r).(string)
		if ret != "foo" {
			t.Errorf("expected 'foo', got %v", ret)
		}

		// string, above 10 chars
		r = bytes.NewReader([]byte("12:foobarraboof"))
		ret = tc.f(r).(string)
		if ret != "foobarraboof" {
			t.Errorf("expected 'foo', got %v", ret)
		}

		// list with one int
		r = bytes.NewReader([]byte("li42ee"))
		retSlice, _ := tc.f(r).([]interface{})
		if len(retSlice) != 1 && retSlice[0] != 42 {
			t.Errorf("expected [42], got %v", ret)
		}

		// list with two items
		r = bytes.NewReader([]byte("li42ei43ee"))
		retSlice, _ = tc.f(r).([]interface{})
		if len(retSlice) != 2 && retSlice[0] != 42 && retSlice[1] != 43 {
			t.Errorf("expected [42, 43], got %v", ret)
		}

		// a simple map
		r = bytes.NewReader([]byte("d3:fooi42ee"))
		retMap, _ := tc.f(r).(map[string]interface{})
		if retMap["foo"] != 42 {
			t.Errorf("expected [42], got %v", retMap)
		}

		// a map with a list
		r = bytes.NewReader([]byte("d3:fooli42eee"))
		retMap, _ = tc.f(r).(map[string]interface{})
		retSlice = retMap["foo"].([]interface{})
		if len(retSlice) != 1 && retSlice[0] != 42 {
			t.Errorf("expected {'foo': [42]}, got %v", ret)
		}
	}
}

func TestBencodeEncode(t *testing.T) {
	var b bytes.Buffer
	// int
	bencode.Encode(&b, 42)
	expected := []byte("i42e")
	if bb := b.Bytes(); !bytes.Equal(bb, expected) {
		t.Errorf("expected %v, got %v", expected, bb)
	}

	// string
	b.Reset()
	bencode.Encode(&b, "foobar")
	expected = []byte("6:foobar")
	if bb := b.Bytes(); !bytes.Equal(bb, expected) {
		t.Errorf("expected %s, got %s", string(expected), string(bb))

	}

	// list of ints
	b.Reset()
	bencode.Encode(&b, []int{1, 2, 3})
	expected = []byte("li1ei2ei3ee")
	if bb := b.Bytes(); !bytes.Equal(bb, expected) {
		t.Errorf("expected %v, got %v", expected, bb)
	}

	// ditto, but uint16
	b.Reset()
	bencode.Encode(&b, []uint16{1, 2, 3})
	expected = []byte("li1ei2ei3ee")
	if bb := b.Bytes(); !bytes.Equal(bb, expected) {
		t.Errorf("expected %v, got %v", expected, bb)
	}

	// list of strings
	b.Reset()
	bencode.Encode(&b, []string{"a", "bc", "def"})
	expected = []byte("l1:a2:bc3:defe")
	if bb := b.Bytes(); !bytes.Equal(bb, expected) {
		t.Errorf("expected %v, got %v", expected, bb)
	}

	// dictionary
	b.Reset()
	m := map[string]int{}
	m["def"] = 2
	m["abc"] = 1
	bencode.Encode(&b, m)
	// note the alphabetical order
	expected = []byte("d3:abci1e3:defi2ee")
	if bb := b.Bytes(); !bytes.Equal(bb, expected) {
		t.Errorf("expected %v, got %s", expected, bb)
	}

	// dictionary with nested list
	b.Reset()
	m2 := map[string]any{}
	m2["def"] = []int{1, 2, 3}
	m2["abc"] = "foo"
	bencode.Encode(&b, m2)
	expected = []byte("d3:abc3:foo3:defli1ei2ei3eee")
	if bb := b.Bytes(); !bytes.Equal(bb, expected) {
		t.Errorf("expected %v, got %s", expected, bb)
	}

	// floats are *not* supported!
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("expected a panic!")
		}
	}()

	b.Reset()
	bencode.Encode(&b, 3.44)
}

func TestBencodeStructTags(t *testing.T) {
	file, _ := os.Open("testdata/ubuntu.torrent")
	defer file.Close()
	btorrent := bencode.ParseFromReader[data.BETorrent](file)

	expectedName := "ubuntu-22.04.2-live-server-amd64.iso"
	if btorrent.Info.Name != expectedName {
		t.Errorf("expected %s, found %s", expectedName, btorrent.Info.Name)
	}
	if btorrent.Info.Length != 1975971840 {
		t.Errorf("expected %d, found %d", 1975971840, btorrent.Info.Length)
	}
	announceList := [][]string{{"https://torrent.ubuntu.com/announce"}, {"https://ipv6.torrent.ubuntu.com/announce"}}
	if !reflect.DeepEqual(btorrent.AnnounceList, announceList) {
		t.Errorf("expected %v, found %v", announceList, btorrent.AnnounceList)
	}

	// do the same for a tracker response - plenty of nested structs/slices!
	file2, _ := os.Open("testdata/tracker.response.beencoded")
	defer file2.Close()
	trackerResponse := bencode.ParseFromReader[data.BETrackerResponse](file2)
	if len(trackerResponse.Peers) != 33 {
		t.Errorf("expected 2 peers, got %d", len(trackerResponse.Peers))
	}
	if trackerResponse.Peers[0].IP != "2a02:1210:4831:9700:ba27:ebff:fe91:60cd" {
		t.Errorf("mismatch in first peer IP: %v", trackerResponse.Peers[0].IP)
	}

	// a torrent with multiple files
	file, _ = os.Open("testdata/files.torrent")
	defer file.Close()
	btorrent = bencode.ParseFromReader[data.BETorrent](file)
	if btorrent.Info.Length != 0 {
		t.Errorf("info.length should be nil (0), found %d", btorrent.Info.Length)
	}
	if len(btorrent.Info.Files) != 3 {
		t.Errorf("expecting 3 files, found %d", len(btorrent.Info.Files))
	}
	if !reflect.DeepEqual(btorrent.Info.Files[2].Path, []string{"/tmp/files/file3"}) {
		t.Errorf("expecting info.files[2].path to be file3, found %s instead", btorrent.Info.Files[2].Path)
	}
}

func TestBencodeStruct(t *testing.T) {
	// info dict with multiple files
	beinfo := data.BEInfo{
		Name:        "foo",
		PieceLength: 65536,
		Files: []data.BEFile{
			{
				Path:   []string{"path1"},
				Length: 123,
			},
			{
				Path:   []string{"path2"},
				Length: 456,
			},
		},
	}
	val := bencode.ToBencodedDict(beinfo)
	// due to how we encode, note how we need to specify unit64
	expected := map[string]any{
		"name":         "foo",
		"piece length": uint64(65536),
		// "length":       uint64(0),
		// "pieces":       "",
		"files": []map[string]any{
			{"path": []string{"path1"}, "length": 123},
			{"path": []string{"path2"}, "length": 456},
		},
	}
	if !reflect.DeepEqual(val, expected) {
		t.Errorf("exepcted %+v, got %+v", expected, val)
	}

	// info dict with a single file
	beinfo = data.BEInfo{
		Name:        "foo",
		PieceLength: 65536,
		Pieces:      "deadbeef",
		Length:      123456,
	}
	val = bencode.ToBencodedDict(beinfo)
	// due to how we encode, note how we need to specify unit64
	expected = map[string]any{
		"name":         "foo",
		"piece length": uint64(65536),
		"length":       uint64(123456),
		"pieces":       "deadbeef",
		// "files":        make([]map[string]any, 0),
	}
	if !reflect.DeepEqual(val, expected) {
		t.Errorf("exepcted %+v, got %+v", expected, val)
	}
}
