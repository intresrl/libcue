package main

/*
#include <stdint.h>
#include <stdbool.h>
#include "cue.h"
*/
import "C"

import "math/big"

func ToCueFloat(x *big.Float, out *C.cue_float) {
	if x == nil {
		panic("big.Float value is nil")
	}

	if x.Sign() == 0 {
		out.exponent = 0
		out.sign = false
		out.mantissa_len = 1
		out.mantissa = C.CBytes([]byte{0})
	}

	// precision in bits of mantissa
	prec := x.Prec()

	// frac = normalized fraction (0.5 < |frac| < 1), so x = frac * 2^exp
	var frac big.Float
	bigFloatExp := x.MantExp(&frac)

	// frac = frac * 2^prec, so that frac is now an integer
	frac.SetMantExp(&frac, int(prec))

	// since it is an integer, fetch it
	mantissa, _ := frac.Int(nil)

	// Cue does not support infinity values (failed arithmetic: division by zero)
	if mantissa == nil {
		panic("infinity value in Cue")
	}

	// separate sign from mantissa
	isNegative := mantissa.Sign() < 0
	mantissa.Abs(mantissa)

	// Absolute value, big-endian.
	b := mantissa.Bytes()

	out.exponent = C.int64_t(int64(bigFloatExp) - int64(prec))
	out.sign = C.bool(isNegative)
	out.mantissa_len = C.size_t(len(b))
	out.mantissa = C.CBytes(b)
}
