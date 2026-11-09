package fio

import (
	"bufio"
	_ "embed"
	"strings"
	"sync"
)

//go:embed data/names.txt
var namesRaw string

//go:embed data/last_names.txt
var lastNamesRaw string

var (
	loadOnce   sync.Once
	firstNames map[string]struct{}
	lastNames  map[string]struct{}
)

func dictionaries() (first, last map[string]struct{}) {
	loadOnce.Do(func() {
		first = loadSet(namesRaw)
		last = loadSet(lastNamesRaw)
		for k := range first {
			delete(last, k)
		}
		for k := range last {
			delete(first, k)
		}
		firstNames = first
		lastNames = last
	})
	return firstNames, lastNames
}

func loadSet(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := prepareSearch(strings.TrimSpace(scanner.Text()))
		if line != "" {
			out[line] = struct{}{}
		}
	}
	return out
}

func isFirstName(name string) bool {
	first, _ := dictionaries()
	_, ok := first[prepareSearch(name)]
	return ok
}

func isLastName(name string) bool {
	_, last := dictionaries()
	_, ok := last[prepareSearch(name)]
	return ok
}
