// Copyright 2012 Joubin Houshyar. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Properties provides basic property file utility.
//
// Property keys are typed.
//
// The key suffixes `[]` and `[:]` specify []string and map[string]string, respectively, but
// otherwise can be used as prefix or embedded in key or value without reservation.
//
// The `#` char is reserved for comments and can not be used in keys or values.
// The `\` char is reserved for line continuation and can not be used in comments, keys, or values.
//
// Syntax supports:
//
// • Embedded white space {' ', '\t'} in keys and values.  Leading and trailing whitespace is ignored.
//
// • Typed properties: (Go) string, []string, and map[string]string properties
//
// • Value definitions can span multiple lines
//
// • Single line & trailing comments
//
// Example demonstrating format:
//
//  # a comment line
//  # note that blank lines are ignored
//
//  # ------------------------------------------
//  # examples of string properties - single line
//
//  the property key = property value                  # key"the property key", value:"property value"
//  the property key=property value                    # same as above
//  a.property@the.key.called!foo = joe@schmoe.com     # only embedded hashsign and/or forward-slashes are disallowed
//
//  # example of string properties - multiline
//  # => "value that "
//  this is a multiline property = value that spans multiple lines. \
//  Note that value line continuations \
//          include whitespace leading each new line.  # e.g. this line appends "        include whitespace ..."
//
//  # ------------------------------------------
//  # examples of []string properties - single line
//  # NOTE that the key includes the trailing `[]`
//
//  this.is.a.string.array.key[] = alpha  , omega      # => []string{"alpha", "omega"}
//  so.is.this.[] = alpha, omega                       # only the suffix [] is significant of []string property type
//
//  # array values can have embedded white space as well
//  # basically, any leading/trailing whitespace around `,` is trimmed
//  # for example
//  another.array[] =  hi there  , bon voyage          # => []string{"hi there", "bon voyage"}
//
//  # array values can also be quoted if trailing and/or leading whitespace is required
//  # for example
//  yet.another[] = " lead", or, "trail "              # => []string{" lead", "or", "trail  "}
//
//  # example of []string property - multiline
//  # note that layout is insignificant
//  web.resource.type.extensions[] = js,    \
//                                   css  , \
//                                   gif      \
//                              ,     jpeg,  \
//                                   png               # => []string{"js", "css", "gif", "jpeg", "png"}
//
//  # ------------------------------------------
//  # examples of map[string]string properties - single line
//  # map key must end in `[:]`.
//  # value must be of form <map-key>:<map-value>
//  # map values must be seperated by `,`
//
//  this.is.a.map[:] = a:b, b:c
//
//  # key set is {"*", "list", "login"}
//  dispatch.table[:] = *:/ , list : /do/list, login: /do/user/login
//
//  # same thing spanning multiple lines:
//
//  dispatch.tablex[:] = *:/ , \
//                         list:/do/list, \           # note the `,`
//                         login:/do/user/login
//
// The associated Properties (type) defines the properties API, but is itself simply a
// a map[string]interface{} and can be used as such (without any type safety).
//
//
//
package gestalt

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"unicode/utf8"
)

// ----------------------------------------------------------------------
// Property file Constants
// ----------------------------------------------------------------------

// REVU - too many flavors of whitespace

const (
	NIL           = ""
	VAL_DELIM     = ","
	KV_DELIM      = ":"
	QUOTE         = "\""
	PKV_SEP       = "="
	TRIMSET       = "\n\r \t"
	WS            = " \t"
	ARRAY         = "[]"
	ARRAY_LEN     = len(ARRAY)
	MAP           = "[:]"
	MAP_LEN       = len(MAP)
	MIN_ENTRY_LEN = len("a=b")

//	INHERIT       = "*"
//	INHERIT_MAP   = INHERIT + KV_DELIM + INHERIT
)
const (
	R_CONTINUATION = '\\'
	R_COMMENT      = '#'
)

// Properties is based on map and can be accessed as such
// but best to use the API
type Properties map[string]interface{}

// ----------------------------------------------------------------------
// API
// ----------------------------------------------------------------------

// Instantiates a new Properties object initialized from the
// content of the specified file.
func Load(filename string) (p Properties, e error) {

	if filename == "" {
		e = fmt.Errorf("filename is nil")
		return
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		e = fmt.Errorf("Error reading gestalt file <%s> : %s", filename, e)
		return
	}

	return loadBuffer(bytes.NewBuffer(b).String())
}

// Support embedded properties (e.g. without files)
func LoadStr(spec string) (p Properties, e error) {
	return loadBuffer(spec)
}

// Return a clone of the argument Properties object
func (p Properties) Clone() (clone Properties) {

	for k, v := range p {
		clone[k] = v
	}
	return
}

// Copy all entries from specified Properties to the receiver
// Note this will overwrite existing matching values if overwrite is true,
// otherwise if overwrite is false it will only append keys that do not exist
// in receiver
func (p Properties) Copy(from Properties, overwrite bool) {
	// TODO - REVU - either silently Debug log or return error on nil 'from'
	for k, v := range from {
		if p[k] == nil || overwrite {
			p[k] = v
		}
	}
}

// Inherits from the parent key/value pairs if receiver[key] is nil.
// If key is array receiver's value array will be PRE-prended with parent's.
// If key is map receiver's value map will be augmented with parent's.
// nil input is silently ignored.
//  REVU - issue regarding preserving order in parent array key values
func (p Properties) Inherit(from Properties) {
	if from == nil {
		return
	}
	for k, v := range from {
		pv := p[k]
		if pv == nil {
			p[k] = v
		} else {
			switch {
			case IsArrayKey(k):
				// REVU - somewhat funky semantics here
				// attempting to preserve order of array values (in child)
				// but parent's order is chomped
				temp := make(map[string]string)
				for _, av := range pv.([]string) {
					temp[av] = av
				}
				narrv := []string{}
				for _, av := range v.([]string) {
					if temp[av] == "" {
						narrv = append(narrv, av)
					}
				}
				p[k] = append(narrv, pv.([]string)...)
			case IsMapKey(k):
				mapv := v.(map[string]string)
				pmapv := pv.(map[string]string)
				for mk, mv := range mapv {
					if pmapv[mk] == "" {
						pmapv[mk] = mv
					}
				}
			}
		}
	}
}

// Verifies that the receiver has values for the set of keys.
// Returns true, 0 len array if verified, or, if keys arg is nil.
// Returns false, and array of missing values otherwise.
func (p Properties) VerifyMust(keys ...string) (bool, []string) {
	missing := []string{}
	if keys != nil {
		for _, rk := range keys {
			if p[rk] == nil {
				missing = append(missing, rk)
			}
		}
	}
	return len(missing) == 0, missing
}

//// returns generic (interface{} prop value or default values if nil
//func (p Properties) GetOrDefault(key string, defval interface{}) (v interface{}) {
//	if v = p[key]; v == nil {
//		v = defval
//	}
//	return
//}

// returns nil/zero-value if no such key or not an array
//  REVU - this silently returns nil if key type is mismatched ..
// TODO: return error
func (p Properties) GetArray(key string) []string {
	if IsArrayKey(key) {
		if v := p[key]; v == nil {
			return nil
		}
		return p[key].([]string)
	}
	return nil
}

// returns prop value or default values if nil
func (p Properties) GetArrayOrDefault(key string, defval []string) (v []string) {
	if v = p.GetArray(key); v == nil {
		v = defval
	}
	return
}

// returns nil/zero-value if no such key or not a map
//  REVU - this silently returns nil if key type is mismatched ..
func (p Properties) GetMap(key string) map[string]string {
	if IsMapKey(key) {
		if v := p[key]; v == nil {
			return nil
		}
		return p[key].(map[string]string)
	}
	return nil
}

// returns prop value or default values if nil
func (p Properties) GetMapOrDefault(key string, defval map[string]string) (v map[string]string) {
	if v = p.GetMap(key); v == nil {
		v = defval
	}
	return
}

// String value property - returns nil/zero-value if no such key or not a map
func (p Properties) GetString(key string) string {
	if !(IsMapKey(key) || IsArrayKey(key)) {
		if v := p[key]; v == nil {
			return ""
		}
		return p[key].(string)
	}
	return ""
}

func (p Properties) GetStringOrDefault(key string, defval string) (v string) {
	if v = p.GetString(key); v == "" {
		v = defval
	}
	return
}
func (p Properties) MustGetString(key string) (v string) {
	return p.GetString(key)
}

// Returns true if provided key is a valid array property value key,
// suitable for use with GetMap(mapkey)
func IsMapKey(key string) bool {
	if strings.HasSuffix(key, MAP) {
		//		if idx := strings.LastIndex(key, MAP); idx == len(key)-MAP_LEN {
		return true
	}
	return false
}

// Returns true if provided key is a valid map property value key
// suitable for use with GetArray(arrkey)
func IsArrayKey(key string) bool {
	if !IsMapKey(key) && strings.HasSuffix(key, ARRAY) {
		return true
	}
	return false
}

// Returns a pretty print string for Properties.
// See also Properties#Print
func (p Properties) String() string {
	srep := "-- properties --\n"
	for k, v := range p {
		srep += fmt.Sprintf("'%s' => '%s'", k, v)
		srep += "\n"
	}
	srep += "----------------\n"
	return srep
}

// Pretty print dumps the Properties content to stdout
func (p Properties) Print() {
	fmt.Print(p)
}

// ----------------------------------------------------------------------
// internal ops
// ----------------------------------------------------------------------

func loadBuffer(s string) (p Properties, e error) {

	if s == NIL {
		e = errors.New("s is nil")
		return
	}

	specs := splitCleanPropSpecs(s)

	p = make(Properties)
	for _, spec := range specs {
		k, v, err := parseProperty(spec)
		if err != nil {
			e = fmt.Errorf("error parsing properties- %s", err)
			return
		}
		if k != NIL {
			p[k] = v
		}
	}
	return
}

// converts to []string of lines.  this is mainly addressing
// comments (both flavors) & continuations (multi-line values)
// beyond a general split on crlf
func splitCleanPropSpecs(s string) (pspecs []string) {

	// trim overall buffer
	s = strings.Trim(s, TRIMSET)

	erase := false
	cont := false
	reset := false
	b := make([]byte, len(s))
	off := 0
	s = strings.Trim(s, TRIMSET)
	for _, c := range s {
		if c == rune(R_CONTINUATION) {
			erase = true
			cont = true
		} else if c == R_COMMENT {
			erase = true
		} else if c == '\n' {
			if cont {
				cont = false
				reset = true
			} else {
				erase = false
			}
		} else if reset {
			erase = false
			reset = false
		}
		if !erase {
			off += utf8.EncodeRune(b[off:], c)
		}
	}
	s = string(b[0:off])

	// split to get distinct specs.
	pspecs = strings.Split(s, "\n")

	return
}

// attempts to parse a single <key> = <value> property def spec.
// Returns ("", "") if comment or malformed.
// Otherwise (key, value) pair are returned.
func parseProperty(spec string) (key string, value interface{}, e error) {
	if len(spec) < MIN_ENTRY_LEN {
		return NIL, value, e
	}

	propTuple := strings.Split(strings.Trim(spec, TRIMSET), PKV_SEP)

	// Verify well-formedness
	if len(propTuple) != 2 || propTuple[1] == NIL {
		e = errors.New(fmt.Sprintf("property spec '%s' is malformed", spec))
		return
	}

	key = strings.Trim(propTuple[0], WS)
	vrep := strings.Trim(propTuple[1], WS)

	// do NOT change order of parse - maps first
	if IsMapKey(key) {
		kvmap := make(map[string]string)
		kvpairs := strings.Split(vrep, VAL_DELIM)
		for _, _kv := range kvpairs {
			_kv = strings.Trim(_kv, WS)
			_kvarr := strings.Split(_kv, KV_DELIM)
			ek := strings.Trim(_kvarr[0], WS)
			ev := strings.Trim(_kvarr[1], WS)
			kvmap[strings.Trim(ek, QUOTE)] = strings.Trim(ev, QUOTE)
		}
		value = kvmap
	} else if IsArrayKey(key) {
		arrv := strings.Split(vrep, VAL_DELIM)
		for i, v := range arrv {
			v = strings.Trim(v, WS)
			arrv[i] = strings.Trim(v, QUOTE)
		}
		value = arrv
	} else {
		value = strings.Trim(propTuple[1], WS)
		value = strings.Trim(vrep, QUOTE)
	}

	return
}
