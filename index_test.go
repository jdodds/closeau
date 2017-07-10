package closeau

import (
	"os"
	"path"
	"testing"
)

func TestIndex(t *testing.T) {
	in := `{"a":"sometext"}`
	in2 := `{"a": "hey baby yo"}`
	ss := NewIndex(NewDirStore(path.Join(os.TempDir(), "closeau")))
	ss.Add(1, in)
	ss.Add(2, in2)
	ss.Add(3, `{"a": "bob"}`)
	ss.Add(4, `{"a": "sailboat doorknob"}`)
	r := ss.Search("bob")
	if len(r) != 2 {
		t.Errorf("len %v.Search('bob') = %d, want %d)", ss, len(r), 2)
	}
}
