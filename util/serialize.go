package util

import "encoding/binary"

func SerializeInt64(v int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	return buf
}

func DeserializeInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
