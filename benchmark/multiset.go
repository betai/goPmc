package main

import (
	"fmt"
	"sort"
)

type Multiset struct {
	set map[string]int
}

func NewMultiset () *Multiset {
	return &Multiset{set: make(map[string]int)}
}

func (multiset *Multiset) Len() int {
	return len(multiset.set)
}

func (multiset *Multiset) Add(key string) {
	_, ok := multiset.set[key]
	if !ok {
		multiset.set[key] = 0
	}
	multiset.set[key]++
}

func (multiset *Multiset) Get(key string) (int, bool) {
	count, alreadyIn := multiset.set[key]
	return count, alreadyIn
}


func (multiset *Multiset) Keys() []string {
	keys := make([]string, 0, len(multiset.set))
	for key, _ := range multiset.set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (multiset *Multiset) PrintKeys() {
	for key, value := range multiset.set {
		fmt.Printf("m[%v] = %v\n", key, value)
	}
}

func (multiset *Multiset) PrintCount() {
	sum := 0
	for _, value := range multiset.set {
		sum += value
	}
	fmt.Printf("the sum is %v\n", sum)
}
