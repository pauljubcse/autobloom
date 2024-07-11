package autobloom

import (
	"fmt"
	"math"
)

type ScalableBloomFilter struct {
	filters  []*BloomFilter
	n        uint64 //Number of elements in current filter
	totaln   uint64
	currentFilterSize uint64
	maxPerFilter uint64
	fpRate   float64
	fpGrowth float64
	maxFilter  uint64
	sizeGrowthRate float64
	seed int
}

func NewScalableBloomFilter(initialSize uint64, maxFilterSize uint64, maxFilter uint64, sizeGrowthRate float64, fpRate float64, fpGrowth float64, h Hasher, seed int) (*ScalableBloomFilter, error) {
	if initialSize <= 0 {
		return nil, fmt.Errorf("invalid initial size")
	}
	if fpRate <= 0 || fpRate >= 1 {
		return nil, fmt.Errorf("invalid false positive rate")
	}
	if fpGrowth <= 0 {
		return nil, fmt.Errorf("invalid false positive growth rate")
	}

	bf, err := NewBloomFilter(initialSize, fpRate, h, seed)
	if err != nil {
		return nil, err
	}

	return &ScalableBloomFilter{
		filters:  []*BloomFilter{bf}, // Start with one filter slice
		fpRate:   fpRate,             // Set the initial false positive rate
		fpGrowth: fpGrowth,           // By what factor does fpRate for successive filters increase
		n:        0,                  // Initialize with zero elements added
		currentFilterSize: initialSize,
		maxPerFilter: maxFilterSize,
		maxFilter: uint64(maxFilter),
		sizeGrowthRate: sizeGrowthRate,
		seed: seed,
	}, nil
}
func estimateCapacity(m uint64, k uint64, p float64) float64 {
	// Calculate ln(1 - e^(ln(p) / k))
	innerExp := math.Exp(math.Log(p) / float64(k))
	denominator := -float64(k) / math.Log(1 - innerExp)
	
	// Calculate the capacity n
	n := float64(m) / denominator
	
	// Return the ceiling of n
	return float64(math.Ceil(n))
}
func (sbf *ScalableBloomFilter) Add(data []byte) {
	for _, filter := range sbf.filters {
	    for i := uint64(0); i < filter.k; i++ {
	        filter.Add(data)
	    }
	}
   
	currentFilter := sbf.filters[len(sbf.filters)-1] // Get the most recent filter
   
	currentFPRate:=currentFilter.falsePositiveProb()
	//fmt.Printf("m: %d \t fpRate: %f \t n: %d\n", currentFilter.m, sbf.fpRate, sbf.n)
	//fmt.Printf("Current FPRate: %f\n", currentFPRate)

	
	// If current estimate of false positive probability > desired, we add a new filter
	if currentFPRate > float64(sbf.fpRate) && len(sbf.filters)<int(sbf.maxFilter) {
		//fmt.Println("Filter Added")
	    newFpRate := sbf.fpRate * sbf.fpGrowth
		newHasher := NewFNVHasher()
	    // Create and append the new filter slice.
		fmt.Printf("n: %d\n", sbf.n)
		fmt.Printf("Set Bit Count: %d\n", currentFilter.setBitCount)
		fmt.Printf("New fp rate: %f\n", newFpRate)
		sbf.fpRate=newFpRate
	    nbf, _ := NewBloomFilter(
			max(
				uint64(float64(sbf.currentFilterSize)*sbf.sizeGrowthRate), 
				sbf.maxPerFilter),
			newFpRate, newHasher, sbf.seed)
	    sbf.filters = append(sbf.filters, nbf)
		sbf.n=0
	}else{
		// Increment the total number of items added across all filter slices.
	    sbf.n++
	}
	
	sbf.totaln++
}
   
func (sbf *ScalableBloomFilter) Test(data []byte) bool {
	// Check the item against all filter slices from the oldest to the newest.
	for _, filter := range sbf.filters {
	    // Assume the item is in the filter until proven otherwise.
	    isPresent := true
   
	    for i := uint64(0); i < filter.k; i++ {
	        // If any of the bits corresponding to the item's hash values are not set, it's definitely not present in this filter.
	        if !filter.Test(data) {
	            isPresent = false
	            // We break out of the hash function loop as soon as we find a bit that is not set.
	            break
	        }
	    }
   
	    // If all the bits for this filter are set, then the item is potentially present (with some false positive rate).
	    if isPresent {
	        return true
	    }
	    // Otherwise, continue checking the next filter to see if the item may be present there.
	}
   
	// If none of the filters had all bits set, the item is definitely not in the set.
	return false
}
