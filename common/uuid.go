package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
)

func uuid4ForBytes(bytes []byte) string {
	bytes[6] = byte(0x40 | (int(bytes[6]) & 0xf))
	bytes[8] = byte(0x80 | (int(bytes[8]) & 0x3f))
	result := hex.EncodeToString(bytes)

	return fmt.Sprintf("%s-%s-%s-%s-%s", result[:8], result[8:12],
		result[12:16], result[16:20], result[20:])
}

// UUID4For converts interface to json bytes, hashes it and uses result to
// generate an UUID4 string.
func UUID4For(i interface{}) string {
	hash := fnv.New128()
	if data, err := json.Marshal(i); err == nil {
		hash.Write([]byte(data))
	} else {
		hash.Write(make([]byte, 16))
	}
	return uuid4ForBytes(hash.Sum([]byte{}))
}

func UUID4() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return uuid4ForBytes(bytes)
}
