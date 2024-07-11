package autobloom

import (
	"hash"
	"hash/fnv"
)

type FNVHasher struct{
}

func NewFNVHasher() *FNVHasher {
	return &FNVHasher{
	}
}

// func (h *FNVHasher) GetHashes(n uint64) []hash.Hash64 {
// 	hashers := make([]hash.Hash64, n)
// 	for i := 0; uint64(i) < n; i++ {
// 		hasher := fnv.New64a()
// 		hasher.Write([]byte{byte(h.seed >> 56), byte(h.seed >> 48), byte(h.seed >> 40), byte(h.seed >> 32), byte(h.seed >> 24), byte(h.seed >> 16), byte(h.seed >> 8), byte(h.seed)})
// 		hasher.Write([]byte{byte(i >> 56), byte(i >> 48), byte(i >> 40), byte(i >> 32), byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
// 		hashers[i] = hasher
// 	}
// 	return hashers
// }
func (h *FNVHasher) GetHashes(n uint64) []hash.Hash64 {
	hashers := make([]hash.Hash64, n)
	for i := uint64(0); i < n; i++ {
		hasher := fnv.New64a()
		hashers[i] = hasher
	}
	return hashers
}