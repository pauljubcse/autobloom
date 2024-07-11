package autobloom

import (
	"fmt"
	"hash"
	"math"
	"strconv"
	"sync"
)

type BloomFilter struct{
	bitset []bool //bit array
	m uint64 //length of bit array
	hashes []hash.Hash64
	k uint64 //Number of hashes
	mutex sync.Mutex
	setBitCount uint64
	collisions uint64
	seed int
}

//Expected Number of Elements: n
//Desired False Positive Rate: p
//Returns optimal size of bitset, and number of hash functions
func OptimalHashCount(n uint64, p float64) (uint64, uint64){
	//fmt.Printf("p: %f\n", p)
	m := uint64(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
	if m == 0 {
		m = 1
	}
	k := uint64(math.Ceil((float64(m) / float64(n)) * math.Log(2)))
	if k == 0 {
		k=1
	}
	//fmt.Printf("m: %d \t k: %d\n", m, k)
	return m, k
}

func NewBloomFilter(n uint64, p float64, h Hasher, seed int) (*BloomFilter, error){
	if n == 0 {
		return nil, fmt.Errorf("number of elements can't be 0")
	}
	if p <= 0 || p >=1 {
		return nil, fmt.Errorf("invalid false positive rate")
	}
	if h == nil{
		return nil, fmt.Errorf("hasher dependancy not injected")
	}
	m, k := OptimalHashCount(n, p)
	return &BloomFilter{
		m: m,
		k: k,
		bitset: make([]bool , m),
		hashes: h.GetHashes(k),
		setBitCount: 0,
		collisions: 0,
		seed: seed,
	}, nil
}

func (bf *BloomFilter) Add(data []byte){
	bf.mutex.Lock()
	defer bf.mutex.Unlock()
	for i, hash := range bf.hashes {
		hash.Reset()
		hash.Write([]byte(strconv.Itoa(bf.seed+i)))
		hash.Write(data)
		hashValue := hash.Sum64() % bf.m
		fmt.Printf("hashValue: %d\n", hashValue)
		if(!bf.bitset[hashValue]){
			bf.setBitCount+=1
			//fmt.Printf("Setting: %d\n", bf.setBitCount)
		}else{
			bf.collisions++;
		}
		bf.bitset[hashValue] = true
	}
}
func (bf *BloomFilter) fillRatio() (float64){
	// bf.mutex.Lock()
	// defer bf.mutex.Unlock()
	//fmt.Printf("Set Bit Count: %d \t m: %d\n", bf.setBitCount, bf.m)
	fr := float64(bf.setBitCount)/float64(bf.m)
	return fr
}
func (bf *BloomFilter) falsePositiveProb() (float64){
	// bf.mutex.Lock()
	// defer bf.mutex.Unlock()
	p := math.Pow(bf.fillRatio(), float64(bf.k))
	return p
}
func (bf *BloomFilter) Test(data []byte) bool {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()
	for i, hash := range bf.hashes{
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