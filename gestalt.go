// Copyright 2012 Joubin Houshyar. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Gestalt defines a basic configuration utility object supporting
// string, []string, and map[string]string configuration keys. Instances are typically
// created by the user provided configuration specs
// (or files), but it is also possible to directly add entries view the
// via underlying map[string]interface{}.
//
// Both keys and values may have embedded whitespace (spaces or tabs) and
// quoted values support leading and/or trailing whitespace in values.
// Beyond that, mainly note that the `=` character is reserved and can not be
// used in keys.
//
// Single line and trailing comments are supported.
//
// Gestalt provides a semantic interface for access to configuration elements,
// but can be access in a raw format via its underlying map[string]interface{}.
// It follows that you can also add entries to a gestalt instance via the map
// semantics.  In that case, no type enforcement is provided, e.g. you can add
// a []string bound to a key not ending in "[]", but you will not be able to
// retrieve via GetArray as the semantic methods do check against key.  It is
// possible to reflectively address this, but the added complexity for a corner
// use-case, and the reflection overhead for general usecase argues against doing so.
//
// Example:
//
//  # A canonical Gestalt spec:
//
//  # -------------------------------------------
//  # demonstrate string type properties
//  # -------------------------------------------
//
//  # basic - note leading and trailing whitespace (including tabs)
//  # are removed from both keys and values.
//  # also note the trailing comment form.
//  prop one=prop one value                 # "prop one" => "prop one value"
//  another property   =  value             # "another property" => "value"
//
//  # use quotes for leading or trailing spaces/tabs in values
//  log.info.level.id = "INFO "             # "log.info.level.id" => "INFO "
//  leading.whitespace = " test"            # "leading.whitespace" => " test"
//
//  # multiline values use the '\' line continuation char.
//  # note that ALL leading characters on a continued line are appended to value.
//  # and any trailing white space chars immediately before `\` are also appended.
//  long one = This sentence ends \
//  in 4 spaces\
//      .                                   # "long one" => "This sentence ends in 4 spaces    ."
//
//  zerovalue =                             # "zerovalue" => ""
//
//  # -------------------------------------------
//  # demonstrate []string type properties
//  # any key ending in "[]" (and not "map[]")
//  # is treated as a []string type
//  # -------------------------------------------
//
//  # note that whitespace between `,` is removed.
//  an array [] = 1 , 2 , 3                 # "an array []" => ["1" "2" "3"]
//  another.array[] = "  1" , " 20", 300    # "an array []" => ["  1" " 20" "300"]
//
//  # note that for arrays (and maps) the leading and trailing
//  # white space for individual values on continued lines
//  # are trimmed (that is if a `,' precedes the `\`.
//  # Also note tha
//  multi-line[] = a, b, c, \
//                 12\
//   4567  ,\
//                 d, e         # "multi-line[]" => ["a" "b" "c" "12 4567" "d" "e"]
//
//  another.one[] = \
//      a, \
//      b, \
//      c                       # "another.one[]" => ["a" "b" "c"]
//
//  # you can define empty []string properties
//  empty[] =                   # "empty[]" => []
//
//
//  # -------------------------------------------
//  # demonstrate map[string]string type properties
//  # e.g. any key ending in "map[]"
//  # -------------------------------------------
//  a map[] = a:1 , b:2, c : 3 , d:4         # "a map[]" => map[a:1 c:3 b:2 d:4]
//
//  # maps can be defined in multi-lines per usual patterns and WS considerations
//  multline.map[] = \
//    a:1 , \
//    b:2, c:3, \
//    d:4                                    # "multline.map[]" => map[a:1 c:3 b:2 d:4]
//
//  # maps can have zerovalue entries
//  zv.entry.map[] =  foo:bar, zerovalue:    # "zv.entry.map[]" => map[foo:bar zerovalue:]
//
//  # and finally, you can define empty map[string]string, as well.
//  empty.map[] =                            # "empty.map[]" => map[]
//
package gestalt

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// Gestalt is just a map[string]interface{} and can be used as such.
type Gestalt map[string]interface{}

// ----------------------------------------------------------------------
// API
// ----------------------------------------------------------------------

// Create a new Gestalt object from spec.
// If spec is zero value string, it simply returns an empty intances.
func New(spec string) (Gestalt, error) {

	gestalt, e := parse(spec)
	if e != nil {
		return nil, e
	}

	return gestalt, nil
}

// Creates a new Gestalt instance from the specified file.
// Returns error if fname is zerovalue or on io errors.
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

// Just for completeness sake -- you can always use gestalk[key].
func (g Gestalt) Get(key string) interface{} {
	return g[key]
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
	value = g.GetMap(key)
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
		// ignore blank lines
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
			if len(etuple) == 2 {
				ek, ev := cleanTuple(etuple)
				mapv[ek] = unquoteValue(ev)
			}
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
