package tracker

import (
	"axiomiety/go-bt/data"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

func EncodeInfoHash(infoHash [20]byte) string {
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
				pairs = append(pairs, fmt.Sprintf("%s=%s", tag, val))
			case reflect.Uint:
				pairs = append(pairs, fmt.Sprintf("%s=%d", tag, val))
			}
		}

	}
	return strings.Join(pairs, "&")
}
