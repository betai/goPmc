/*
Package pmc description
*/
package pmc

import (
	"fmt"
	"math/rand"
	"math"
	"time"
	"hash/fnv"
	"strings"
	"strconv"
	"errors"

	"github.com/willf/bitset"
)

var ( // change to const
	n_max float64 = 1e5
	phi_const float64 = 0.77351
)

type Sketch struct {
	m uint
	w uint
	l uint
	bitmap *bitset.BitSet
	n uint
}

func New(l uint, m uint, w uint) (*Sketch, error) {
 	if l == 0 || m == 0 || w == 0 {
		return nil, errors.New("All parameters must be > 0")
	} else if l > (1 << w){
		return nil, errors.New("l must be < 2**w")
	}
	return &Sketch{l: l, m: m, w: w, bitmap: bitset.New(l)}, nil
}

// Algorithm 1 B[H(f,i,j)] = 1
func (sketch *Sketch) PmcCount(f string) (error) {
	index, err := sketch.getIndexF(f)
	if err != nil {
		//fmt.Printf("%v\n", err.Error())		
		return err
	}
	sketch.bitmap.Set(index % sketch.l)
	return nil
}

// Algorithm 2
func (sketch *Sketch) getZSum(f string) (z uint) {
	z = 0
	for i := uint(0); i < sketch.m; i++ {
		for j := uint(0); j < sketch.w; j++ {
			if !sketch.bitmap.Test(getIndexFIJ(f, i, j)) {
				z += j - 1
				break
			}
		}
	}
	return z
}

// Algorithm 3
func (sketch *Sketch) getEmptyRows(f string) (k uint) {
	k = 0
	for i := uint(0); i < sketch.m; i ++ {
		if !sketch.bitmap.Test(getIndexFIJ(f, i, 1)) {
			k++
		}
	}
	return k
}

// Algorithm 4
func (sketch *Sketch) PmcEstimate(f string) (float64, error) {
	k := float64(sketch.getEmptyRows(f))
	p := sketch.p()
	m := float64(sketch.m)

	estimate := 0.0
	
	if kp := k/(1 - p); kp > 0.3*m {
		fmt.Println("small multiplicity")
		estimate = -2*m*math.Log(kp/m)
	} else {
		fmt.Println("large multiplicity")		
		if sketch.n <= 0 {
			return -1, errors.New("sketch.n should be positive")
		}
		z := float64(sketch.getZSum(f))
		estimate = m*math.Pow(2, z/m)/sketch.phi(p)
	}

	estimate = math.Ceil(math.Abs(estimate))
	sketch.n = uint(estimate)
	return estimate, nil
}


/**** Public Helpers ****/
func (sketch *Sketch) PrintBitmap() {
	fmt.Printf("Non-zero bits for sketch %v\n", sketch)
	count := 0
	for i := uint(0); i < sketch.l ; i++ {
		if sketch.bitmap.Test(i) {
			fmt.Printf("  %v\n", i)
			count++
		}
	}
	fmt.Printf("Total: %v\n", count)
}

/**** Private Helpers ****/
func (sketch *Sketch) p() (float64) {
	count := 0.0
	for i := uint(0); i < sketch.l ; i++ {
		if sketch.bitmap.Test(i) {
			count++
		}
	}
	return count/float64(sketch.l)
}

func (sketch *Sketch) phi(p float64) (float64) {
	n := float64(sketch.n)
	if n >= n_max {
		return phi_const
	}
	return math.Pow(2, sketch.expZ(n, p))/n
}

func (sketch *Sketch) expZ(n float64, p float64) (float64) {
	exp := 0.0
	w := float64(sketch.w)
	for k := float64(0); k < w; k++ {
		exp += k * (qk(k, n, p) - qk(k+1, n, p))
	}
	return exp
}

func qk(k float64, n float64, p float64) (float64) {
	product := 1.0
	for i := float64(0); i < k; i++ {
		product *= (1 - math.Pow((1 - math.Pow(2, -i)), n) * (1 - p))
	}
	return product
}

// Get index into sketch bitmap
func (sketch *Sketch) getIndexF(f string) (uint, error) { // TODO: Bubble up the error
	i := uint(rand.Intn(int(sketch.m)))
	j, err := geometric(sketch.w)
	if err != nil {
//		fmt.Printf("%v\n", err.Error())
		return 0, err
	}

	return getIndexFIJ(f, i, j), err
}

func getIndexFIJ(f string, i uint, j uint) (uint) {
	fij := concatenate(f, int(i), int(j))
	// fmt.Printf("fij = %v\n", fij)
	
	index := hash(fij)
	// fmt.Printf("index = %v\n", index)
	return uint(index)
}

// Join f i and j into a string
func concatenate(f string, i int, j int) (string) {
	return strings.Join([]string{f, strconv.Itoa(i), strconv.Itoa(j)}, "")
}

// fnv1a hash
func hash(fij string) (uint64) {
	h := fnv.New64a()
	h.Write([]byte(fij))
	return h.Sum64()
}

// random number generator based of geometric distribution
func geometric(w uint) (j uint, e error) {
	if w > 32 {
		return 0, errors.New("input parameter w to geometric function must be < 32") // uint is 32 bit
	}
	
	uniform := rand.Uint32()

	for i := uint32(0); i < uint32(w); i++ {
		if uniform & (1 << uint32(i)) != 0 {
			return uint(i), nil
		}
	}
	return w, nil
}
