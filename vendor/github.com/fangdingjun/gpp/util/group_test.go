package util

import (
	"testing"
)

func TestLookupGroup(t *testing.T) {
	testData := map[string]int{
		"dingjun": 1000,
		"root":    0,
	}

	for name, gid := range testData {
		g, err := LookupGroup(name)
		if err != nil {
			t.Error(err)
			continue
		}
		if g.Gid != gid {
			t.Errorf("expected gid: %d, got: %d\n", gid, g.Gid)
		}
	}
}

func TestLookupGroupId(t *testing.T) {
	testData := map[int]string{
		0:    "root",
		1000: "dingjun",
	}
	for gid, name := range testData {
		g, err := LookupGroupID(gid)
		if err != nil {
			t.Error(err)
			continue
		}
		if g.Name != name {
			t.Errorf("expected name: %s, got: %s\n", name, g.Name)
		}
	}
}
