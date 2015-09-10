/*
Package pmc description
*/
package pmc

import (
	"fmt"
	"math/rand"
	"time"
	"hash/fnv"
	"strings"
	"strconv"
	"errors"

	"github.com/willf/bitset"
)

type Sketch struct {
	m uint
	w uint
	l uint
	bitmap *bitset.BitSet
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
func (sketch *Sketch) PmcCount(f string) {
	// i := rand.Intn(m)
	// j := geometric(w)

	i := 234
	j := 345

	fij := concatenate(f,i,j)
	fmt.Printf("fij = %v\n", fij)
	
	index := hash(fij)
	fmt.Printf("index = %v\n", index)

	sketch.bitmap.Set(uint(index) % sketch.l)
}

// Algorithm 2


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
func geometric(w uint) (j int, e error) {
	if w > 32 || w < 0 {
		return 0, errors.New("input parameter w to geometric function must be > 0")
	}

	rand.Seed(time.Now().UTC().UnixNano())
	uniform := uint(rand.Int())

	fmt.Printf("uniform = %v \t", uniform)
	
	for i := uint(0); i < w; i++ {
		if uniform & (1 << uint(i)) != 0 {
			return j, nil
		}
		j++
	}
	return j, nil
}
