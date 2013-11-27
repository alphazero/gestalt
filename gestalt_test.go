package gestalt

import (
	"testing"
	"fmt"
)

func TestLoadFile(t *testing.T) {
	fname := "test/test.conf"
	_, e := Load(fname)
	if e != nil {
		t.Errorf("TestLoadFile - gestalt.Load - %s", e)
	}
}

func TestMultilineArray(t *testing.T) {
	spec :=`
%s = a,  \
	b,       \
		c,    \
			d
`
	expected := []string{"a", "b", "c", "d"}
	key := "foo[]"
	spec = fmt.Sprintf(spec, key)
	prop, e := DefineStr(spec)
	if e != nil {
		t.Errorf("TestMultilineArray - gestalt.DefineStr - %s", e)
	}

	got := prop.GetArray(key)
	if got == nil {
		t.Errorf("TestMultilineArray - GetArray(%s) - expected: %s, got: %s", key, expected, got)
	}
	for i, av := range got {
		if av != expected[i] {
			t.Errorf("TestMultilineArray - GetArray(%s) - at index %d expected: %s, got: %s", key, i, expected[i], av)
		}
	}
	return

}

func TestMultilineString(t *testing.T) {
	spec := `
%s = this is supposed to be\
 a very\
 long sentence.
`
	expected := "this is supposed to be a very long sentence."
	key := "a.long.sentence"
	spec = fmt.Sprintf(spec, key)

	prop, e := DefineStr(spec)
	if e != nil {
		t.Errorf("TestMultilineString - New - %s", e)
	}

	got := prop.GetString(key)
	if got != expected {
		t.Errorf("TestMultilineString - GetString(%s) - expected: %s, got: %s", key, expected, got)
	}
}

func TestNew(t *testing.T) {
	spec := `
%s=bar
 woof = meow
 quoted = "INFO "
some.array[] = a, b , 	c, d
some.array.with.quoted.values[] = a, b, "   c", "d "
`
	key := "foo"
	spec = fmt.Sprintf(spec, key)

	prop, e := DefineStr(spec)
	if e != nil {
		t.Errorf("TestNew - New - %s", e)
	}

	got := prop.GetString(key)
	expected := "bar"
	if got != expected {
		t.Errorf("TestNew - GetString(%s) - expected: %s, got: %s", key, expected, got)
	}

	v := prop.GetString("quoted")
	expected = "INFO "
	if v != expected {
		t.Errorf("TestNew - gestalt.GetString(%s) - expected: <%s>, got: <%s>", key, expected, v)
	}
	return
}
