// Gestalt is a basic properties files utility.

package gestalt

import (
	"fmt"
	"strings"
	"io/ioutil"
)

// Gestalt is just a map[string]interface{} and can be used as such.
type Gestalt map[string]interface{}

// ----------------------------------------------------------------------
// API
// ----------------------------------------------------------------------

// Create a new Gestalt object from spec.
// If spec is zero value string, it simply returns an empty intances.
func New(spec string) (Gestalt, error) {
//	if spec == "" {
//		return nil, fmt.Errorf("spec is zerovalue")
//	}

	gestalt, e := parse(spec)
	if e != nil {
		return nil, e
	}

	for k, v := range gestalt {
		fmt.Printf("DEBUG <%s> => <%s>\n", k, v)
	}
	return gestalt, nil
}

func ReadFile(fname string) (g Gestalt, e error) {
	if fname == "" {
		e = fmt.Errorf("fname is zero-value")
		return
	}
	b, ioe := ioutil.ReadFile(fname)
	if ioe != nil {
		e = fmt.Errorf("Read error - %s", ioe)
		return
	}

	return New(string(b))
}

func (g Gestalt) GetOrDefault(key string, defvalue interface{}) (value interface{}) {
	value = g[key]
	if value == nil {
		value = defvalue
	}
	return
}

func (g Gestalt) GetArray(key string) (value []string) {
	if !strings.HasSuffix(key, "map[]") && strings.HasSuffix(key, "[]") {
		if g[key] != nil {
			value = g[key].([]string)
		}
	}
	return
}

// Returns default value if key not present.
// Returns value if it is.
func (g Gestalt) GetArrayOrDefault(key string, defvalue []string) (value []string) {
	value = g.GetArray(key)
	if value == nil {
		value = defvalue
	}
	return
}

// Returns map value if present.  Returns nil if key is not a valid
// array key or if no such key is present.
func (g Gestalt) GetMap(key string) (value map[string]string) {
	if strings.HasSuffix(key, "map[]") {
		if g[key] != nil {
			value = g[key].(map[string]string)
		}
	}
	return
}

func (g Gestalt) GetMapOrDefault(key string, defvalue map[string]string) (value map[string]string) {
	value= g.GetMap(key)
	if value == nil {
		value = defvalue
	}
	return
}

func (g Gestalt) GetString(key string) (value string) {
	if !(strings.HasSuffix(key, "map[]") && strings.HasSuffix(key, "[]")) {
		value = ""
		if g[key] != nil {
			value = g[key].(string)
		}
	}
	return
}

func (g Gestalt) GetStringOrDefault(key string, defvalue string) (value string) {
	value = g.GetString(key)
	if value == "" {
		value = defvalue
	}
	return
}

// Checks if the Gestalt instance has the specified key
// values.  Returns nil if all are present, or the missing
// subset of specified keys.
func (g Gestalt) MustHave(keys ...string) []string {
	mset := make([]string, 0)
	for _, k := range keys {
		if g[k] == nil {
			mset = append(mset, k)
		}
	}
	if len(mset) == 0 {
		return nil
	}
	return mset
}
// ----------------------------------------------------------------------
// internal ops
// ----------------------------------------------------------------------

// parse the given s (string) into a generic map[string]interface{}
func parse(s string) (Gestalt, error) {
	s = strings.Trim(s, " \t\n\r")
	s = strings.Replace(s, "\r\n", "\n", -1)
	s = strings.Replace(s, "\\\n", "", -1)
	sarr := strings.Split(s, "\n")
	if len(sarr) == 0 {
		return nil, fmt.Errorf("BUG - zero len array from spec parse step 1")
	}
//	for _, l := range sarr {
//		fmt.Printf("DEBUG - after split: '%s'\n", l)
//	}
	gestalt := make(Gestalt)

	for _, pspec := range sarr {
		if len(pspec) == 0 {
			continue
		}

		k, v, e := parsePSpec(pspec)
		if e != nil {
			return nil, fmt.Errorf("Parse property - %s", e)
		}
		if k != "" {
			gestalt[k] = v
		}
	}
	return gestalt, nil
}

func parsePSpec(s string) (k string, v interface{}, e error) {
	// comments
	if strings.HasPrefix(s, "#") {
		return
	}
	if i := strings.Index(s, "#"); i > 0 {
		s = s[:i]
	}

	s = strings.Trim(s, " \t\r\n")
	if len(s) == 0 {
		return
	}
	stuple := strings.Split(s, "=")
	if len(stuple) != 2 {
		e = fmt.Errorf("format error for `%s`", s)
		return
	}

	k, vspec := cleanTuple(stuple)

	switch {
	case strings.HasSuffix(k, "map[]"):
		mapv := make(map[string]string)
		entries := strings.Split(vspec, ",")
		for _, entry := range entries {
			entry = strings.Trim(entry, " \t")
			etuple := strings.Split(entry, ":")
			ek, ev := cleanTuple(etuple)
			mapv[ek] = unquoteValue(ev)
		}
		v = mapv
	case strings.HasSuffix(k, "[]"):
		arrayv := strings.Split(vspec, ",")
		for i, arrvi := range arrayv {
			arrayv[i] = unquoteValue(strings.Trim(arrvi, " \t"))
		}
		v = arrayv
	default:
		v = unquoteValue(vspec)
	}

	return
}

// trims leading and trailing whitespace from values
func cleanTuple(tuple []string) (t0, t1 string) {
	t0 = strings.Trim(tuple[0], " \t")
	t1 = strings.Trim(tuple[1], " \t")
	return
}

// removes the enclosing " chars from the value spec
func unquoteValue(v string) string {
	return strings.Trim(v, "\"")
}
