package tracker

import (
	"encoding/hex"
	"fmt"
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
