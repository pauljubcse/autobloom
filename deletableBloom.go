package autobloom

import (
	"fmt"
	"hash"
	"math"
	"strconv"
	"sync"
)

type DeletableBloomFilter struct{
	bitset []bool
	m uint64
	hashes []hash.Hash64
	k uint64
	mutex sync.Mutex
	setBitCount uint64
	collisions uint64
	r uint64
	collisionBitmap []bool
	seed int
}


func NewDeletableBloomFilter(n uint64, p float64, h Hasher, r uint64, seed int) (*DeletableBloomFilter, error){
	if n == 0 {
		return nil, fmt.Errorf("number of elements can't be 0")
	}
	if p <= 0 || p >=1 {
		return nil, fmt.Errorf("invalid false positive rate")
	}
	if h == nil{
		return nil, fmt.Errorf("hasher dependency not injected")
	}
	m, k := OptimalHashCount(n, p)
	if r == 0 {
		return nil, fmt.Errorf("number of regions can't be 0")
	}
	return &DeletableBloomFilter{
		m: m,
		k: k,
		bitset: make([]bool , m),
		hashes: h.GetHashes(k),
		setBitCount: 0,
		collisions: 0,
		r: r,
		collisionBitmap: make([]bool, r),
		seed: seed,
	}, nil
}

func (bf *DeletableBloomFilter) Add(data []byte){
	bf.mutex.Lock()
	defer bf.mutex.Unlock()
	for i, hash := range bf.hashes {
		hash.Reset()
		hash.Write([]byte(strconv.Itoa(bf.seed+i)))
		hash.Write(data)
		hashValue := hash.Sum64() % bf.m
		region := hashValue / uint64(math.Ceil(float64(bf.m)/ float64(bf.r)))
		//fmt.Printf("hashValue: %d\t region: %d\n", hashValue, region)
		if !bf.bitset[hashValue] {
			bf.setBitCount += 1
		} else {
			bf.collisionBitmap[region] = true
		}
		bf.bitset[hashValue] = true
	}
}

func (bf *DeletableBloomFilter) Remove(data []byte) bool{
	bf.mutex.Lock()
	defer bf.mutex.Unlock()
	flag:=false
	for i, hash := range bf.hashes {
		hash.Reset()
		hash.Write([]byte(strconv.Itoa(bf.seed+i)))
		hash.Write(data)
		hashValue := hash.Sum64() % bf.m
		region := hashValue / uint64(math.Ceil(float64(bf.m)/ float64(bf.r)))
		if !bf.collisionBitmap[region] {
			bf.bitset[hashValue] = false
			bf.setBitCount -= 1
			flag=true
		}
	}
	return flag
}

func (bf *DeletableBloomFilter) fillRatio() float64{
	fr := float64(bf.setBitCount)/float64(bf.m)
	return fr
}

func (bf *DeletableBloomFilter) falsePositiveProb() float64{
	p := math.Pow(bf.fillRatio(), float64(bf.k))
	return p
}

func (bf *DeletableBloomFilter) Test(data []byte) bool {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()
	for i, hash := range bf.hashes {
		hash.Reset()
		hash.Write([]byte(strconv.Itoa(bf.seed+i)))
		hash.Write(data)
		hashValue := hash.Sum64() % bf.m
		if !bf.bitset[hashValue] {
			return false
		}
	}
	return true
}
