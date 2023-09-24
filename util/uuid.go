package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func UUID4() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	bytes[6] = byte(0x40 | (int(bytes[6]) & 0xf))
	bytes[8] = byte(0x80 | (int(bytes[8]) & 0x3f))
	result := hex.EncodeToString(bytes)

	return fmt.Sprintf("%s-%s-%s-%s-%s", result[:8], result[8:12],
		result[12:16], result[16:20], result[20:])
}
