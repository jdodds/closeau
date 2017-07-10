package closeau

import (
	"encoding/json"
	"strconv"
)

type Index struct {
	charPairs map[charPair]IdSet
}

type charPair [2]rune
type Id uint64

type IdSet map[Id]bool

func (i *IdSet) Copy() IdSet {
	r := make(IdSet)
	for k, _ := range *i {
		r[k] = true
	}
	return r
}

func (i *IdSet) Intersect(o IdSet) {
	for k, _ := range *i {
		if !o[k] {
			delete(*i, k)
		}
	}
}

func (ss Index) String() string {
	s := make([]byte, 0)
	s = append(s, '{')
	for k, v := range ss.charPairs {
		s = append(s, '"')
		s = strconv.AppendQuoteRune(s, k[0])
		s = append(s, ',')
		s = strconv.AppendQuoteRune(s, k[1])
		s = append(s, '"', ':', '[')
		for kk, _ := range v {
			s = strconv.AppendUint(s, uint64(kk), 10)
			s = append(s, ',', ' ')
		}
		s = append(s, ']', ',')
	}
	s = append(s, '}')
	return string(s)
}

func (ss *Index) Add(id Id, item string) {
	if ss.charPairs == nil {
		ss.charPairs = make(map[charPair]IdSet)
	}
	var j interface{}
	i := []byte(item)
	json.Unmarshal(i, &j)
	m := j.(map[string]interface{})
	ss.add(id, m)
}

func (ss *Index) add(id Id, m map[string]interface{}) {
	for _, v := range m {
		switch vv := v.(type) {
		case string:
			ss.index(id, vv)
		case map[string]interface{}:
			ss.add(id, vv)
		}
	}
}

func (ss *Index) index(id Id, st string) {
	for i, j := 0, 1; j < len(st); i, j = i+1, j+1 {
		l := charPair{rune(st[i]), rune(st[j])}
		if ss.charPairs[l] == nil {
			ss.charPairs[l] = make(IdSet)
		}
		ss.charPairs[l][id] = true
	}
}

func (ss *Index) Search(st string) []Id {
	var r []Id
	l := charPair{rune(st[0]), rune(st[1])}
	initial := ss.charPairs[l]
	if initial == nil {
		return r
	}

	results := initial.Copy()
	for i, j := 1, 2; j < len(st); i, j = i+1, j+1 {
		o := ss.charPairs[charPair{rune(st[i]), rune(st[j])}]
		results.Intersect(o)
		if len(results) == 0 {
			return r
		}
	}

	for k, _ := range results {
		r = append(r, k)
	}
	return r
}
