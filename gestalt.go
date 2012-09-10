// Gestalt is a basic properties files utility.

package gestalt

import (
	"fmt"
	"strings"
)

// Gestalt is just a map[string]interface{} and can be used as such.
type Gestalt map[string]interface{}

// ----------------------------------------------------------------------
// API
// ----------------------------------------------------------------------

// Create a new Gestalt object from spec
func New(spec string) (Gestalt, error) {
	if spec == "" {
		return nil, fmt.Errorf("spec is zerovalue")
	}

	gestalt, e := parse(spec)
	if e != nil {
		return nil, e
	}

	return gestalt, nil
}

func (g Gestalt) GetOrDefault(key string, defvalue interface{}) (value interface{}) {
	value = g[key]
	if value == nil {
		value = defvalue
	}
	return
}

func (g Gestalt) GetArray(key string) (value []string, e error) {
	if !strings.HasSuffix(key, "map[]") && strings.HasSuffix(key, "[]") {
		if g[key] != nil {
			value = g[key].([]string)
		}
	} else {
		e = fmt.Errorf("%s is not an array key", key)
	}
	return
}

func (g Gestalt) GetArrayOrDefault(key string, defvalue []string) (value []string, e error) {
	value, e = g.GetArray(key)
	if e != nil {
		return
	}
	if value == nil {
		value = defvalue
	}
	return
}

func (g Gestalt) GetMap(key string) (value map[string]string, e error) {
	if strings.HasSuffix(key, "map[]") {
		if g[key] != nil {
			value = g[key].(map[string]string)
		}
	} else {
		e = fmt.Errorf("%s is not a map key", key)
	}
	return
}

func (g Gestalt) GetMapOrDefault(key string, defvalue map[string]string) (value map[string]string, e error) {
	value, e = g.GetMap(key)
	if e != nil {
		return
	}
	if value == nil {
		value = defvalue
	}
	return
}

func (g Gestalt) GetString(key string) (value string, e error) {
	if !(strings.HasSuffix(key, "map[]") && strings.HasSuffix(key, "[]")) {
		value = ""
		if g[key] != nil {
			value = g[key].(string)
		}
	} else {
		e = fmt.Errorf("%s is not a string key", key)
	}
	return
}

func (g Gestalt) GetStringOrDefault(key string, defvalue string) (value string, e error) {
	value, e = g.GetString(key)
	if e != nil {
		return
	}
	if value == "" {
		value = defvalue
	}
	return
}

// ----------------------------------------------------------------------
// internal ops
// ----------------------------------------------------------------------

// parse the given s (string) into a generic map[string]interface{}
func parse(s string) (Gestalt, error) {
	s = strings.Trim(s, " \t\n\r")
	s = strings.Replace(s, "\r\n", "\n", -1)
	s = strings.Replace(s, "\\", "", -1)
	sarr := strings.Split(s, "\n")
	if len(sarr) == 0 {
		return nil, fmt.Errorf("BUG - zero len array from spec parse step 1")
	}

	gestalt := make(Gestalt)

	for _, pspec := range sarr {
		k, v, e := parsePSpec(pspec)
		if e != nil {
			return nil, fmt.Errorf("Parse property - %s", e)
		}
		gestalt[k] = v
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

	stuple := strings.Split(s, "=")
	if len(stuple) != 2 {
		e = fmt.Errorf("format error for `%s`", s)
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
			mapv[ek] = ev
		}
		v = mapv
	case strings.HasSuffix(k, "[]"):
		arrayv := strings.Split(vspec, ",")
		for i, arrvi := range arrayv {
			arrayv[i] = strings.Trim(arrvi, " \t")
		}
		v = arrayv
	default:
		v = vspec
	}

	return
}

func cleanTuple(tuple []string) (t0, t1 string) {
	t0 = strings.Trim(tuple[0], " \t")
	t1 = strings.Trim(tuple[1], " \t")
	return
}
