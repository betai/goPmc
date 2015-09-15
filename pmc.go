/*
Package pmc description
*/
package pmc

import (
	"fmt"
	"math/rand"
	"math"
	"errors"

	"github.com/willf/bitset"
	"github.com/dgryski/go-farm"
)

var ( // change to const
	n_max float64 = 1e5
	phi_const float64 = 0.77351
)

type Sketch struct {
	m uint64
	w uint64
	l uint64
	bitmap *bitset.BitSet
	n uint64
	cachedP float64
	changed bool
}

func New(l uint64, m uint64, w uint64) (*Sketch, error) {
 	if l == 0 || m == 0 || w == 0 {
		return nil, errors.New("All parameters must be > 0")
	} else if l > (1 << w) || l > (1 << 32){
		return nil, errors.New("l must be < 2**w and <= 2**32")
	}
	return &Sketch{l: l, m: m, w: w, bitmap: bitset.New(uint(l)), changed: true}, nil
}

// Algorithm 1 B[H(f,i,j)] = 1
func (sketch *Sketch) PmcCount(f []byte) error {
	index, err := sketch.getIndexF(f)
	if err != nil {
		//fmt.Printf("%v\n", err.Error())
		return err
	}
	if sketch.bitmap.Test(uint(index) % uint(sketch.l)) {

		// fmt.Printf("Bit at index %v is already set. Sketch changed %v\n", index, sketch.changed)
		return nil
	}

	sketch.changed = true
	sketch.bitmap.Set(uint(index) % uint(sketch.l))
	return nil
}

// Algorithm 2
func (sketch *Sketch) getZSum(f []byte) uint64 {
	z := uint64(0)
	for i := uint64(0); i < sketch.m; i++ {
		for j := uint64(0); j < sketch.w; j++ {
			if !sketch.bitmap.Test(uint(sketch.getIndexFIJ(f, i, j))) {
				z += j - 1
				break
			}
		}
	}
	return z
}

// Algorithm 3
func (sketch *Sketch) getEmptyRows(f []byte) uint64 {
	k := uint64(0)
	for i := uint64(0); i < sketch.m; i ++ {
		if !sketch.bitmap.Test(uint(sketch.getIndexFIJ(f, i, 0))) {
			k++
		}
	}
	return k
}

// Algorithm 4
func (sketch *Sketch) PmcEstimate(f []byte) (uint64, error) {
	k := float64(sketch.getEmptyRows(f))
	p := sketch.p()
	m := float64(sketch.m)
	phi := sketch.phi(p)

	estimate := 0.0
	
	if kp := k/(1 - p); kp > 0.3*m {
		//fmt.Println("Small multiplicity")
		estimate = -2*m*math.Log(kp/m)
	} else {
		//fmt.Println("large multiplicity")		
		if sketch.n <= 0 {
			return 0, errors.New("sketch.n should be positive")
		}
		z := float64(sketch.getZSum(f))
		estimate = m*math.Pow(2, z/m)/phi
	}

	estimate = math.Ceil(math.Abs(estimate))
	sketch.n = uint64(estimate)
	return sketch.n, nil
}


/**** Public Helpers ****/
func (sketch *Sketch) PrintBitmap() {
	fmt.Printf("Non-zero bits for sketch %v\n", sketch)
	count := 0
	for i := uint(0); i < uint(sketch.l) ; i++ {
		if sketch.bitmap.Test(i) {
			fmt.Printf("  %v\n", i)
			count++
		}
	}
	fmt.Printf("Total: %v\n", count)
}

/**** Private Helpers ****/
func (sketch *Sketch) p() float64 {
	if !sketch.changed {
		return sketch.cachedP
	}
	count := 0.0
	for i := uint(0); i < uint(sketch.l) ; i++ {
		if sketch.bitmap.Test(i) {
			count++
		}
	}
	sketch.cachedP = count/float64(sketch.l)
	sketch.changed = false
	return sketch.cachedP
}

func (sketch *Sketch) phi(p float64) float64 {
	n := float64(sketch.n)
	if n >= n_max {
		return phi_const
	}
	return math.Pow(2, sketch.expZ(n, p))/n
}

func (sketch *Sketch) expZ(n float64, p float64) float64 {
	exp := 0.0
	w := float64(sketch.w)
	for k := float64(0); k < w; k++ {
		exp += k * (qk(k, n, p) - qk(k+1, n, p))
	}
	return exp
}

func qk(k float64, n float64, p float64) float64 {
	product := 1.0
	for i := float64(0); i < k; i++ {
		product *= (1 - math.Pow((1 - math.Pow(2, -i)), n) * (1 - p))
	}
	return product
}

// Get index into sketch bitmap
func (sketch *Sketch) getIndexF(f []byte) (uint64, error) {
	i := uniform(sketch.m)
	j, err := geometric(sketch.w)
	if err != nil {
//		fmt.Printf("%v\n", err.Error())
		return 0, err
	}

	return sketch.getIndexFIJ(f, i, j), err
}

func (sketch *Sketch) getIndexFIJ(f []byte, i uint64, j uint64) uint64 {
	return farm.Hash64WithSeeds(f, i, j) % sketch.l
}

func uniform (m uint64) uint64 {
	return uint64(rand.Int63n(int64(m))) % m
}

// random number generator based of geometric distribution
func geometric(w uint64) (j uint64, e error) {
	if w > 64 {
		return 0, errors.New("input parameter w to geometric function must be <= 64") // uint64 is 64 bit
	}
	uniform := rand.Uint32()

	for i := uint32(0); i < uint32(w); i++ {
		if uniform & (1 << uint32(i)) != 0 {
			return uint64(i), nil
		}
	}
	return w, nil
}
