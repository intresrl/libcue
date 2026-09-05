// Copyright 2024 The CUE Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

/*
#include <stdint.h>
#include <stdbool.h>
#include "cue.h"
*/
import "C"

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"unsafe"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
)

//export cue_compile_string_raw
func cue_compile_string_raw(ctx C.cue_ctx, str *C.char, opts *C.struct_cue_bopt, len C.size_t, v *C.cue_value) C.cue_error {
	bopts := buildOptions(opts, len)
	val := cueContext(ctx).CompileString(C.GoString(str), bopts...)

	if err := val.Err(); err != nil {
		return cueErrorHandle(err)
	}
	*v = cueValueHandle(val)
	return 0
}

//export cue_compile_bytes_raw
func cue_compile_bytes_raw(ctx C.cue_ctx, buf unsafe.Pointer, bufLen C.size_t, opts *C.struct_cue_bopt, optLen C.size_t, v *C.cue_value) C.cue_error {
	bopts := buildOptions(opts, optLen)
	val := cueContext(ctx).CompileBytes(C.GoBytes(buf, C.int(bufLen)), bopts...)

	if err := val.Err(); err != nil {
		return cueErrorHandle(err)
	}
	*v = cueValueHandle(val)
	return 0
}

//export cue_top
func cue_top(ctx C.cue_ctx) C.cue_value {
	val := cueContext(ctx).CompileString("_")
	return cueValueHandle(val)
}

//export cue_bottom
func cue_bottom(ctx C.cue_ctx) C.cue_value {
	val := cueContext(ctx).CompileString("_|_")
	return cueValueHandle(val)
}

//export cue_unify
func cue_unify(x C.cue_value, y C.cue_value) C.cue_value {
	u := cueValue(x).Unify(cueValue(y))
	return cueValueHandle(u)
}

//export cue_instance_of_raw
func cue_instance_of_raw(v C.cue_value, typ C.cue_value, opts *C.struct_cue_eopt, len C.size_t) C.cue_error {
	err := cueValue(typ).Subsume(cueValue(v), options(opts, len)...)
	if err != nil {
		return cueErrorHandle(err)
	}
	return 0
}

//export cue_lookup_string
func cue_lookup_string(v C.cue_value, str *C.char, res *C.cue_value) C.cue_error {
	path := cue.ParsePath(C.GoString(str))
	target := cueValue(v).LookupPath(path)
	if err := target.Err(); err != nil {
		return cueErrorHandle(err)
	}
	*res = cueValueHandle(target)
	return 0
}

//export cue_lookup_any_index
func cue_lookup_any_index(v C.cue_value, res *C.cue_value) C.cue_error {
	target := cueValue(v).LookupPath(cue.MakePath(cue.AnyIndex))
	if err := target.Err(); err != nil {
		return cueErrorHandle(err)
	}
	*res = cueValueHandle(target)
	return 0
}

//export cue_lookup_any_string
func cue_lookup_any_string(v C.cue_value, res *C.cue_value) C.cue_error {
	target := cueValue(v).LookupPath(cue.MakePath(cue.AnyString))
	if err := target.Err(); err != nil {
		return cueErrorHandle(err)
	}
	*res = cueValueHandle(target)
	return 0
}

//export cue_from_int64
func cue_from_int64(ctx C.cue_ctx, v C.int64_t) C.cue_value {
	val := cueContext(ctx).Encode(int64(v))
	return cueValueHandle(val)
}

//export cue_from_uint64
func cue_from_uint64(ctx C.cue_ctx, v C.uint64_t) C.cue_value {
	val := cueContext(ctx).Encode(uint64(v))
	return cueValueHandle(val)
}

//export cue_from_bool
func cue_from_bool(ctx C.cue_ctx, v C.bool) C.cue_value {
	val := cueContext(ctx).Encode(bool(v))
	return cueValueHandle(val)
}

//export cue_from_double
func cue_from_double(ctx C.cue_ctx, v C.double) C.cue_value {
	val := cueContext(ctx).Encode(float64(v))
	return cueValueHandle(val)
}

//export cue_from_string
func cue_from_string(ctx C.cue_ctx, str *C.char) C.cue_value {
	val := cueContext(ctx).Encode(C.GoString(str))
	return cueValueHandle(val)
}

//export cue_from_bytes
func cue_from_bytes(ctx C.cue_ctx, buf unsafe.Pointer, len C.size_t) C.cue_value {
	val := cueContext(ctx).Encode(C.GoBytes(buf, C.int(len)))
	return cueValueHandle(val)
}

//export cue_from_list
func cue_from_list(ctx C.cue_ctx, values *C.cue_value, count C.size_t) C.cue_value {
	var v []cue.Value
	for _, val := range unsafe.Slice(values, count) {
		v = append(v, cueValue(val))
	}
	nl := cueContext(ctx).NewList(v...)
	return cueValueHandle(nl)
}

//export cue_dec_int64
func cue_dec_int64(v C.cue_value, res *C.int64_t) C.cue_error {
	n, err := cueValue(v).Int64()
	if err != nil {
		return cueErrorHandle(err)
	}
	*res = C.int64_t(n)
	return 0
}

//export cue_dec_uint64
func cue_dec_uint64(v C.cue_value, res *C.uint64_t) C.cue_error {
	n, err := cueValue(v).Uint64()
	if err != nil {
		return cueErrorHandle(err)
	}
	*res = C.uint64_t(n)
	return 0
}

//export cue_dec_bool
func cue_dec_bool(v C.cue_value, res *C.bool) C.cue_error {
	b, err := cueValue(v).Bool()
	if err != nil {
		return cueErrorHandle(err)
	}
	*res = C.bool(b)
	return 0
}

//export cue_dec_double
func cue_dec_double(v C.cue_value, res *C.double) C.cue_error {
	x, err := cueValue(v).Float64()
	if err != nil {
		return cueErrorHandle(err)
	}
	*res = C.double(x)
	return 0
}

//export cue_dec_float
func cue_dec_float(v C.cue_value, res *C.cue_float) C.cue_error {
	cv := cueValue(v)

	src, err := format.Node(cv.Syntax())
	if err != nil {
		return cueErrorHandle(err)
	}

	// multiply by 4 to get an overappoximation of base-2 precision from base-10
	precision := decimalDigits(string(src)) * uint64(4)

	var floatVal big.Float
	floatVal.SetPrec(uint(min(precision, uint64(math.MaxUint))))

	x, err := cv.Float(&floatVal)
	if err != nil {
		return cueErrorHandle(err)
	}

	ToCueFloat(x, res)
	return 0
}

var decimalDigitsRE = regexp.MustCompile(`\s*^[+-]?0*(\d*)\.?(\d+)(?:[eE][+-]?\d+)?\s*$`)

func decimalDigits(s string) uint64 {
	m := decimalDigitsRE.FindStringSubmatch(s)

	if len(m) == 0 {
		return 1
	}

	return uint64(len(m[1])) + uint64(len(m[2]))
}

//export cue_dec_string
func cue_dec_string(v C.cue_value, res **C.char) C.cue_error {
	s, err := cueValue(v).String()
	if err != nil {
		return cueErrorHandle(err)
	}
	*res = C.CString(s)
	return 0
}

//export cue_dec_bytes
func cue_dec_bytes(v C.cue_value, res *unsafe.Pointer, size *C.size_t) C.cue_error {
	b, err := cueValue(v).Bytes()
	if err != nil {
		return cueErrorHandle(err)
	}
	*res = C.CBytes(b)
	*size = C.size_t(len(b))
	return 0
}

//export cue_dec_json
func cue_dec_json(v C.cue_value, res *unsafe.Pointer, size *C.size_t) C.cue_error {
	b, err := cueValue(v).MarshalJSON()
	if err != nil {
		return cueErrorHandle(err)
	}
	*res = C.CBytes(b)
	*size = C.size_t(len(b))
	return 0
}

//export cue_validate_raw
func cue_validate_raw(v C.cue_value, opts *C.struct_cue_eopt, len C.size_t) C.cue_error {
	err := cueValue(v).Validate(options(opts, len)...)
	if err != nil {
		return cueErrorHandle(err)
	}
	return 0
}

//export cue_default
func cue_default(v C.cue_value, res *C.bool) C.cue_value {
	def, ok := cueValue(v).Default()
	if res == nil {
		return cueValueHandle(def)
	}
	*res = C.bool(ok)
	return cueValueHandle(def)
}

//export cue_is_equal
func cue_is_equal(x C.cue_value, y C.cue_value) C.bool {
	return C.bool(cueValue(x).Equals(cueValue(y)))
}

const allKindMask = cue.BoolKind |
	cue.IntKind |
	cue.FloatKind |
	cue.StringKind |
	cue.BytesKind |
	cue.NumberKind |
	cue.StructKind |
	cue.ListKind

//export cue_concrete_kind
func cue_concrete_kind(v C.cue_value) C.cue_kind {
	kind := cueValue(v).Kind()
	switch kind {
	case cue.BottomKind:
		return C.CUE_KIND_BOTTOM
	case cue.NullKind:
		return C.CUE_KIND_NULL
	case cue.BoolKind:
		return C.CUE_KIND_BOOL
	case cue.IntKind:
		return C.CUE_KIND_INT
	case cue.FloatKind:
		return C.CUE_KIND_FLOAT
	case cue.StringKind:
		return C.CUE_KIND_STRING
	case cue.BytesKind:
		return C.CUE_KIND_BYTES
	case cue.StructKind:
		return C.CUE_KIND_STRUCT
	case cue.ListKind:
		return C.CUE_KIND_LIST
	case cue.NumberKind:
		return C.CUE_KIND_NUMBER
	case cue.TopKind:
		return C.CUE_KIND_TOP
	}
	if kind&allKindMask != 0 { // disjunctions where kind is different are represented as a bitwise OR of all kinds of the branches
		return C.CUE_KIND_TOP
	}
	panic(fmt.Sprintf("unknown value kind %d\n", kind))
}

//export cue_incomplete_kind
func cue_incomplete_kind(v C.cue_value) C.cue_kind {
	kind := cueValue(v).IncompleteKind()
	switch kind {
	case cue.BottomKind:
		return C.CUE_KIND_BOTTOM
	case cue.NullKind:
		return C.CUE_KIND_NULL
	case cue.BoolKind:
		return C.CUE_KIND_BOOL
	case cue.IntKind:
		return C.CUE_KIND_INT
	case cue.FloatKind:
		return C.CUE_KIND_FLOAT
	case cue.StringKind:
		return C.CUE_KIND_STRING
	case cue.BytesKind:
		return C.CUE_KIND_BYTES
	case cue.StructKind:
		return C.CUE_KIND_STRUCT
	case cue.ListKind:
		return C.CUE_KIND_LIST
	case cue.NumberKind:
		return C.CUE_KIND_NUMBER
	case cue.TopKind:
		return C.CUE_KIND_TOP
	}
	if kind&allKindMask != 0 {
		return C.CUE_KIND_TOP
	}
	panic(fmt.Sprintf("unknown value kind %d\n", kind))
}

//export cue_value_error
func cue_value_error(v C.cue_value) C.cue_error {
	if err := cueValue(v).Err(); err != nil {
		return cueErrorHandle(err)
	}
	return 0
}

//export cue_path
func cue_path(v C.cue_value) *C.char {
	label := cueValue(v).Path().String()
	return C.CString(label)
}

//export cue_list
func cue_list(v C.cue_value, count *C.size_t) *C.cue_value {
	iter, err := cueValue(v).List()
	if err != nil {
		return nil
	}

	var elements []cue.Value
	for iter.Next() {
		elements = append(elements, iter.Value())
	}

	*count = C.size_t(len(elements))

	if len(elements) == 0 {
		return nil
	}

	s, ptr := calloc[C.cue_value](len(elements), C.sizeof_cue_value)
	for i, elem := range elements {
		s[i] = cueValueHandle(elem)
	}

	return ptr
}

//export cue_is_concrete
func cue_is_concrete(v C.cue_value) C.bool {
	return C.bool(cueValue(v).IsConcrete())
}

func opConversion(op cue.Op) C.cue_op {
	switch op {
	case cue.NoOp:
		return C.CUE_OP_NO
	case cue.AndOp:
		return C.CUE_OP_AND
	case cue.OrOp:
		return C.CUE_OP_OR
	case cue.SelectorOp:
		return C.CUE_OP_SELECTOR
	case cue.IndexOp:
		return C.CUE_OP_INDEX
	case cue.SliceOp:
		return C.CUE_OP_SLICE
	case cue.CallOp:
		return C.CUE_OP_CALL
	case cue.BooleanAndOp:
		return C.CUE_OP_BOOLEAN_AND
	case cue.BooleanOrOp:
		return C.CUE_OP_BOOLEAN_OR
	case cue.EqualOp:
		return C.CUE_OP_EQUAL
	case cue.NotOp:
		return C.CUE_OP_NOT
	case cue.NotEqualOp:
		return C.CUE_OP_NOT_EQUAL
	case cue.LessThanOp:
		return C.CUE_OP_LESS_THAN
	case cue.LessThanEqualOp:
		return C.CUE_OP_LESS_THAN_EQUAL
	case cue.GreaterThanOp:
		return C.CUE_OP_GREATER_THAN
	case cue.GreaterThanEqualOp:
		return C.CUE_OP_GREATER_THAN_EQUAL
	case cue.RegexMatchOp:
		return C.CUE_OP_REGEX_MATCH
	case cue.NotRegexMatchOp:
		return C.CUE_OP_NOT_REGEX_MATCH
	case cue.AddOp:
		return C.CUE_OP_ADD
	case cue.SubtractOp:
		return C.CUE_OP_SUBTRACT
	case cue.MultiplyOp:
		return C.CUE_OP_MULTIPLY
	case cue.FloatQuotientOp:
		return C.CUE_OP_FLOAT_QUOTIENT
	case cue.InterpolationOp:
		return C.CUE_OP_INTERPOLATION
	case cue.SpreadOp:
		return C.CUE_OP_SPREAD
	default:
		panic("unknown op")
	}
}

//export cue_expr
func cue_expr(v C.cue_value) C.cue_expr_result {
	op, values := cueValue(v).Expr()

	var result C.cue_expr_result
	result.op = opConversion(op)

	if op == cue.CallOp {
		b, err := format.Node(values[0].Syntax())
		if err != nil {
			panic(fmt.Sprintf("format.Node cannot print func name node: %s", err.Error()))
		}

		result.call_name = C.CString(string(b))
		values = values[1:]
	} else {
		result.call_name = nil
	}

	if len(values) == 0 {
		result.count = 0
		result.values = nil
		return result
	}

	result.count = C.size_t(len(values))

	s, ptr := calloc[C.cue_value](len(values), C.sizeof_cue_value)
	for i, d := range values {
		s[i] = cueValueHandle(d)
	}

	result.values = ptr

	return result
}

//export cue_len
func cue_len(v C.cue_value) C.cue_value {
	return cueValueHandle(cueValue(v).Len())
}
