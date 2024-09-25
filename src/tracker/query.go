package tracker

import (
	"axiomiety/go-bt/bencode"
	"axiomiety/go-bt/common"
	"axiomiety/go-bt/data"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func EncodeBytes(infoHash [20]byte) string {
	var sb strings.Builder
	for _, val := range infoHash {
		sb.WriteString(encodeByte(val))
	}
	return sb.String()
}

func encodeByte(val byte) string {
	switch {
	case val == '.' || val == '-' || val == '_' || val == '~':
		fallthrough
	case '0' <= val && val <= '9':
		fallthrough
	case 'a' <= val && val <= 'z':
		fallthrough
	case 'A' <= val && val <= 'Z':
		return string(val)
	default:
		return fmt.Sprintf("%%%s", strings.ToUpper(hex.EncodeToString([]byte{val})))
	}
}

func ToQueryString(q *data.TrackerQuery) string {
	structure := reflect.TypeOf(q).Elem()
	pairs := []string{}
	for i := 0; i < structure.NumField(); i++ {
		f := structure.Field(i)
		tag := f.Tag.Get("url")
		if tag != "" {
			val := reflect.ValueOf(q).Elem().FieldByName(f.Name).Interface()
			switch f.Type.Kind() {
			case reflect.String:
				// empty strings like an empty event= can cause trackers to reject
				// our request
				if val != "" {
					pairs = append(pairs, fmt.Sprintf("%s=%s", tag, val))
				}
			case reflect.Uint:
				pairs = append(pairs, fmt.Sprintf("%s=%d", tag, val))
			case reflect.Bool:
				boolAsInt := 0
				if val.(bool) {
					boolAsInt = 1
				}
				pairs = append(pairs, fmt.Sprintf("%s=%d", tag, boolAsInt))
			default:
				panic(fmt.Sprintf("unknown value for tag=%s", tag))
			}
		}

	}
	return strings.Join(pairs, "&")
}

func QueryTrackerRaw(t *url.URL, q *data.TrackerQuery) []byte {
	t.RawQuery = ToQueryString(q)
	fmt.Printf("u:%s\n", t.String())
	resp, err := http.Get(t.String())
	common.Check(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	common.Check(err)
	return bodyBytes
}

func QueryTracker(t *url.URL, q *data.TrackerQuery) *data.BETrackerResponse {
	return bencode.ParseFromReader[data.BETrackerResponse](bytes.NewReader(QueryTrackerRaw(t, q)))
}
