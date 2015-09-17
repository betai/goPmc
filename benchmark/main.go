package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"log"
	"os"
	"strconv"
	".." //goPmc
)

func main() {
	file, _ := os.Open("zipf.csv")

	r := csv.NewReader(file)
	r.Comma = ';'

	l := uint64(8e7)
	m := uint64(64)
	w := uint64(32)
	
	fmt.Printf("l = %v\tm = %v\tw= %v\n", l, m, w)
	pmc, err := pmc.New(l, m, w)

	if err != nil {
		fmt.Println(err.Error())
	}

	multiset := NewMultiset()
	
	x := 0
	for {
		record, err := r.Read()
		x++
		if x == 1 {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		id := fmt.Sprintf("flow-%s", record[0])
		counts, _ := strconv.ParseFloat(record[1], 64)
		
		for i := 0.0; i < counts; i++ {
			if x <= 11 { // just take the first 10
				multiset.Add(id)
			}
			pmc.PmcCount([]byte(id))
		}
	}

	keys := multiset.Keys()
	min := math.MaxFloat64
	max := 0.0
	avg := 0.0
	std := 0.0
	
	for _, key := range keys {
		est, err := pmc.PmcEstimate([]byte(key))
		if err != nil {
			fmt.Println(err.Error())
		}

		count, _ := multiset.Get(key)
		actual := float64(count)
		diff := 100*math.Abs(float64(est) - actual)/actual
		
		if diff > max {
			max = diff
		}
		if diff < min {
			min = diff
		}

		avg += diff
		std += diff*diff
		
		fmt.Printf("id: %v\texpected: %v\test: %v\t(1 - est/expected[i])*100: %v%%\n", key, actual, est, diff)
	}

	n := float64(len(keys))
	fmt.Printf("error: min/max/avg/std  %v / %v / %v / %v\n", min, max, avg/n, math.Sqrt(std/(n - 1.0)))
	fmt.Println("fill rate:", pmc.GetFillRate())
}
