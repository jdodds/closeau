package closeau

import (
	"encoding/json"
	"strconv"
)

type SearchSpace struct {
	uniqueChars map[rune]pairLookup
}

type pairLookup map[rune]map[int]bool

func (ss SearchSpace) String() string {
	s := make([]byte, 0)
	s = append(s, '{')
	for k, v := range ss.uniqueChars {
		s = append(s, '"')
		s = strconv.AppendQuoteRune(s, k)
		s = append(s, '"', ':', '{', '"')
		for kk, vv := range v {
			s = strconv.AppendQuoteRune(s, kk)
			s = append(s, '"', ':', '[')
			for kkk, _ := range vv {
				s = strconv.AppendInt(s, int64(kkk), 10)
				s = append(s, ',', ' ')
			}
			s = append(s, ']', ',')
		}
		s = append(s, '}', ',')
	}
	s = append(s, '}')
	return string(s)
}

func (ss *SearchSpace) Add(index int, item string) {
	if ss.uniqueChars == nil {
		ss.uniqueChars = make(map[rune]pairLookup)
	}
	var j interface{}
	i := []byte(item)
	json.Unmarshal(i, &j)
	m := j.(map[string]interface{})
	ss.add(index, m)
}

func (ss *SearchSpace) add(i int, m map[string]interface{}) {
	for _, v := range m {
		switch vv := v.(type) {
		case string:
			ss.index(i, vv)
		case map[string]interface{}:
			ss.add(i, vv)
		}

	}
}

func (ss *SearchSpace) index(i int, st string) {
	s := []rune(st)
	for k := range s {
		if ss.uniqueChars[s[k]] == nil {
			ss.uniqueChars[s[k]] = make(pairLookup)
		}
		l := k + 1
		if l < len(s) {
			if ss.uniqueChars[s[k]][s[l]] == nil {
				ss.uniqueChars[s[k]][s[l]] = make(map[int]bool)
			}
			ss.uniqueChars[s[k]][s[l]][i] = true
		}
	}
}

func (ss *SearchSpace) Search(s string) []int {
	r := make([]int, 0)
	a := s[0]
	if ss.uniqueChars[rune(a)] == nil {
		return r
	}
	if ss.uniqueChars[rune(a)][rune(s[1])] == nil {
		return r
	}
	results := make(map[int]bool)
	for ook, _ := range ss.uniqueChars[rune(a)][rune(s[1])] {
		results[ook] = true
	}
	for i := 2; i < len(s); i++ {
		bros, ok := ss.uniqueChars[rune(a)][rune(s[i])]
		if !ok {
			continue
		}
		for k, _ := range results {
			_, ok := bros[k]
			if !ok {
				delete(results, k)
			}
		}
	}
	for k, _ := range results {
		r = append(r, k)
	}
	return r
}
