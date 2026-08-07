package main

/*
#include <stdint.h>
#include <stdbool.h>
#include "cue.h"
*/
import "C"

import (
	"cuelang.org/go/cue"
)

//export cue_fields_raw
func cue_fields_raw(v C.cue_value, opts *C.struct_cue_eopt, length C.size_t, count *C.size_t) *C.cue_value {
	iter, err := cueValue(v).Fields(options(opts, length)...)
	if err != nil {
		return nil
	}

	var fields []cue.Value
	for iter.Next() {
		fields = append(fields, iter.Value())
	}

	*count = C.size_t(len(fields))

	if len(fields) == 0 {
		return nil
	}

	s, ptr := calloc[C.cue_value](len(fields), C.sizeof_cue_value)
	for i, f := range fields {
		s[i] = cueValueHandle(f)
	}

	return ptr
}
