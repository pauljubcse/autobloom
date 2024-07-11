package autobloom

import (
	"fmt"
	"hash"
	"hash/fnv"
	"math/rand"
	"testing"
	"time"
)

// A mock hasher for testing purposes
type MockHasher struct{}

func NewMockHasher() *MockHasher {
	return &MockHasher{}
}

func (h *MockHasher) GetHashes(n uint64) []hash.Hash64 {
	hashers := make([]hash.Hash64, n)
	for i := uint64(0); i < n; i++ {
		hashers[i] = fnv.New64a()
	}
	return hashers
}

func TestOptimalHashCount(t *testing.T) {
	n := uint64(1000)
	p := 0.01
	m, k := OptimalHashCount(n, p)

	if m == 0 {
		t.Errorf("Expected m to be greater than 0, got %d", m)
	}

	if k == 0 {
		t.Errorf("Expected k to be greater than 0, got %d", k)
	}
}

func TestNewBloomFilter(t *testing.T) {
	n := uint64(1000)
	p := 0.01
	h := NewMockHasher()

	bf, err := NewBloomFilter(n, p, h, 1234)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bf.m == 0 {
		t.Errorf("Expected bf.m to be greater than 0, got %d", bf.m)
	}

	if bf.k == 0 {
		t.Errorf("Expected bf.k to be greater than 0, got %d", bf.k)
	}

	if len(bf.bitset) != int(bf.m) {
		t.Errorf("Expected bitset length to be %d, got %d", bf.m, len(bf.bitset))
	}

	if len(bf.hashes) != int(bf.k) {
		t.Errorf("Expected number of hashes to be %d, got %d", bf.k, len(bf.hashes))
	}
}

func TestAddAndTest(t *testing.T) {
	n := uint64(1000)
	p := 0.01
	h := NewFNVHasher()
	bf, err := NewBloomFilter(n, p, h, 124) //seed
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data := []byte("testdata")
	bf.Add(data)

	if !bf.Test(data) {
		t.Errorf("Expected data to be found in BloomFilter, but it was not")
	}

	nonexistentData := []byte("nonexistentdata")
	if bf.Test(nonexistentData) {
		t.Errorf("Expected nonexistent data to not be found in BloomFilter, but it was")
	}
}

// Generate random bytes for testing
func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func BenchmarkFalsePositiveRate(b *testing.B) {
	n := uint64(100000)  // Number of elements
	p := 0.01           // Desired false positive rate
	h := NewFNVHasher()
	bf, err := NewBloomFilter(n, p, h, 1234)
	if err != nil {
		b.Fatalf("Failed to create BloomFilter: %v", err)
	}
	fmt.Printf("Calculated m: %d \t\tCalculated k: %d", bf.m, bf.k)
	fmt.Println()
	falsePositives := 0
	totalTests := 1000

	// Add elements to the BloomFilter
	for i := 0; i < int(n); i++ {
		data := randomBytes(8) // Add random data
		bf.Add(data)
	}

	// Test for random elements that were not added to the BloomFilter
	for i := 0; i < totalTests; i++ {
		data := randomBytes(8) // Generate random data
		if bf.Test(data) {
			falsePositives++
		}
	}

	falsePositiveRate := float64(falsePositives) / float64(totalTests)
	fmt.Printf("False positive rate: %f", falsePositiveRate)
	fmt.Println()
}
func TestFalsePositiveRate(t *testing.T) {
	n := uint64(100000)  // Number of elements
	p := 0.01           // Desired false positive rate
	h := NewFNVHasher()
	bf, err := NewBloomFilter(n, p, h, 1234)
	if err != nil {
		t.Fatalf("Failed to create BloomFilter: %v", err)
	}
	fmt.Printf("Calculated m: %d \t\tCalculated k: %d", bf.m, bf.k)
	fmt.Println()
	// falsePositives := 0
	// totalTests := 1000

	// Add elements to the BloomFilter
	for i := 0; i < int(n); i++ {
		data := randomBytes(8) // Add random data
		bf.Add(data)
		if i%100 == 0{
			fmt.Printf("Current Fill Ratio: %f\n", bf.fillRatio())
			fmt.Printf("Current False Positive Probability: %f\n", bf.falsePositiveProb())
		}
	}

	
}
// }
func TestScalableBloomFilter_AddAndTest(t *testing.T) {
	hasher := NewFNVHasher()
	sbf, err := NewScalableBloomFilter(1000, 1000, 10, 1, 0.01, 1.01, hasher, 1234)
	if err != nil {
		t.Fatalf("failed to create ScalableBloomFilter: %v", err)
	}

	// Test adding and testing elements
	data := [][]byte{
		[]byte("test1"),
		[]byte("test2"),
		[]byte("test3"),
	}

	// Add elements to the filter
	for _, d := range data {
		sbf.Add(d)
	}

	// Test if elements are present
	for _, d := range data {
		if !sbf.Test(d) {
			t.Errorf("expected element %s to be present", d)
		}
	}

	// Test for an element not added to the filter
	if sbf.Test([]byte("not_present")) {
		t.Error("expected element 'not_present' to be absent")
	}
}

func TestScalableBloomFilter_Expand(t *testing.T) {
	hasher := NewFNVHasher()
	n:=10000
	initialSize := uint64(n)
	fpRate := 0.001
	fpGrowth := 1.01
	rand.Seed(time.Now().UnixNano())
	sbf, err := NewScalableBloomFilter(initialSize, initialSize, 10, 1, fpRate, fpGrowth, hasher, 1234)
	if err != nil {
		t.Fatalf("failed to create ScalableBloomFilter: %v", err)
	}
	fmt.Printf("m: %d\t k:%d\n", sbf.filters[len(sbf.filters)-1].m, sbf.filters[len(sbf.filters)-1].k)

	// Add elements to exceed the initial filter capacity
	for i := 0; i < n*110; i++ {
		data := []byte(fmt.Sprintf("LmaoooXDXD-%d", rand.Intn(1000000)))
		sbf.Add(data)
	}

	// Check that we have more than one filter (since we exceed the initial filter capacity)
	// if len(sbf.filters) <= 1 {
	// 	t.Errorf("expected more than one filter, got %d", len(sbf.filters))
	// }

	// Check that the added elements are still present
	// for i := 0; i < n*100; i++ {
	// 	data := []byte{byte(i)}
	// 	if !sbf.Test(data) {
	// 		t.Errorf("expected element %d to be present", i)
	// 	}
	// }
	fmt.Printf("False Positive Prob: %f\n", sbf.filters[len(sbf.filters)-1].falsePositiveProb())
	fmt.Printf("Set Bit Count: %d\n", sbf.filters[len(sbf.filters)-1].setBitCount)
	fmt.Printf("Collisions: : %d\n", sbf.filters[len(sbf.filters)-1].collisions)
	fmt.Printf("m: %d\t k:%d\n", sbf.filters[len(sbf.filters)-1].m, sbf.filters[len(sbf.filters)-1].k)
	fmt.Printf("Filters: %d", len(sbf.filters))
}
// BenchmarkScalableBloomFilter benchmarks the false positive rate of the ScalableBloomFilter over 100000 iterations.
func BenchmarkScalableBloomFilter(b *testing.B) {
	hasher := NewFNVHasher()
	initialSize := uint64(100000)
	fpRate := 0.001
	fpGrowth := 1.1

	sbf, err := NewScalableBloomFilter(initialSize, initialSize, 10, 0.9, fpRate, fpGrowth, hasher, 1234)
	if err != nil {
		b.Fatalf("failed to create ScalableBloomFilter: %v", err)
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Add elements to the filter
	for i := 0; i < 50000; i++ {
		data := []byte(fmt.Sprintf("Lmao data-%d", i))
		sbf.Add(data)
	}

	// Benchmark the false positive rate over 100000 iterations
	falsePositives := 0
	iterations := 1000000

	for i := 0; i < iterations; i++ {
		// Generate a random non-existent element
		data := []byte(fmt.Sprintf("non-existent-%d", rand.Intn(1000000)+50000))
		if sbf.Test(data) {
			falsePositives++
		}
	}

	falsePositiveRate := float64(falsePositives) / float64(iterations)
	b.Logf("False positive rate: %f", falsePositiveRate)
}

func TestSingleKeyDeletion(t *testing.T) {
	hasher := NewFNVHasher()
	bf, err := NewDeletableBloomFilter(100, 0.01, hasher, 10, 12344)
	if err != nil {
		t.Fatalf("Error creating Deletable Bloom Filter: %v\n", err)
	}
	fmt.Printf("m: %d\n", bf.m)
	key := []byte("test-key")
	
	// Add the key
	bf.Add(key)
	
	// Check if the key is in the Bloom Filter
	if !bf.Test(key) {
		t.Fatalf("Expected key to be present in the Bloom Filter after addition")
	}

	// Remove the key
	if !bf.Remove(key){
		t.Fatalf("Deletion failed")
	}
	
	// Check if the key is deleted
	if bf.Test(key) {
		t.Fatalf("Expected key to be absent in the Bloom Filter after deletion")
	} else {
		t.Logf("Key successfully deleted from the Bloom Filter")
	}
}
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
func TestDeletableBloomFilterDeletability(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	numElements := 10000
	falsePositiveRate := 0.01
	numRegions := uint64(1000)
	randomStrings := make([]string, numElements)
	for i := range randomStrings {
		randomStrings[i] = generateRandomString(20)
	}

	hasher := NewFNVHasher()
	bf, err := NewDeletableBloomFilter(uint64(numElements), falsePositiveRate, hasher, numRegions, 12121)
	if err != nil {
		t.Fatalf("Error creating Deletable Bloom Filter: %v\n", err)
	}

	stages := 4
	step := numElements / stages
	deletability := make([]float64, stages)

	for stage := 1; stage <= stages; stage++ {
		for i := (stage - 1) * step; i < stage*step; i++ {
			bf.Add([]byte(randomStrings[i]))
		}

		successfulDeletions := 0
		for i := 0; i < stage*step; i++ {
			if bf.Test([]byte(randomStrings[i])) {
				if bf.Remove([]byte(randomStrings[i])){
					successfulDeletions++
				}
			}
		}
		deletability[stage-1] = float64(successfulDeletions) / float64(stage*step)
		t.Logf("Stage %d deletability: %.2f%%\n", stage, deletability[stage-1]*100)
	}
}