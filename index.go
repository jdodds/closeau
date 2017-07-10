package closeau

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type Id uint64
type charPair [2]rune

func (c charPair) String() string {
	return string(c[0]) + string(c[1])
}

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

type Index struct {
	charPairs map[charPair]IdSet
	store     IndexStore
}

func NewIndex(store IndexStore) Index {
	i := Index{
		charPairs: make(map[charPair]IdSet),
		store:     store,
	}
	i.init()
	return i
}

func (ss *Index) init() {
	m := ss.store.Read()
	for cp, ids := range m {
		if ss.charPairs[cp] == nil {
			ss.charPairs[cp] = make(IdSet)
		}
		for _, id := range ids {
			ss.charPairs[cp][id] = true
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
	ss.store.Start()
	ss.add(id, m)
	ss.store.Finish()
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
		_, exists := ss.charPairs[l][id]
		if !exists {
			ss.charPairs[l][id] = true
			ss.store.Add(l, id)
		}
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

type IndexStore interface {
	Start()
	Add(cp charPair, id Id)
	Finish()
	Read() map[charPair][]Id
}

type DirStore struct {
	charPairs map[charPair][]Id
	dir       string
	writers   map[charPair]io.Writer
}

func NewDirStore(d string) IndexStore {
	os.MkdirAll(d, os.ModeDir|0755)
	return &DirStore{
		charPairs: make(map[charPair][]Id),
		dir:       d,
		writers:   make(map[charPair]io.Writer),
	}
}

func (d *DirStore) Start() {
}

func (d *DirStore) Add(cp charPair, id Id) {
	ids, exists := d.charPairs[cp]
	if !exists {
		ids = []Id{id}
		d.charPairs[cp] = ids
		p := path.Join(d.dir, cp.String())
		d.writers[cp], _ = os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	} else {
		d.charPairs[cp] = append(ids, id)
	}
}

func (d *DirStore) Finish() {
	for cp, ids := range d.charPairs {
		buf := new(bytes.Buffer)
		for _, id := range ids {
			binary.Write(buf, binary.BigEndian, id)
		}
		_, err := d.writers[cp].Write(buf.Bytes())
		if err != nil {
			fmt.Println(err)
		}
		d.charPairs[cp] = []Id{}
	}
}

func (d *DirStore) Read() map[charPair][]Id {
	m := make(map[charPair][]Id)
	fs, _ := ioutil.ReadDir(d.dir)
	for _, f := range fs {
		n := f.Name()
		p := path.Join(d.dir, n)
		b, _ := ioutil.ReadFile(p)
		buf := bytes.NewReader(b)
		l := len(b) / 8
		ids := make([]Id, l)
		binary.Read(buf, binary.BigEndian, &ids)
		cp := charPair{rune(n[0]), rune(n[1])}
		m[cp] = ids
	}
	return m
}
