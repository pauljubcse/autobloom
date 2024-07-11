package autobloom

import "hash"

type Interface interface{
	Add([] byte)
	Test([]byte) bool
}

type Hasher interface{
	GetHashes(n uint64) []hash.Hash64
}

