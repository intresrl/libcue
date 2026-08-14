#include "cue.h"
#include "option.h"
#include "cue_export.h"
#include "_cgo_export.h"

CUE_EXPORT
cue_value*
cue_fields(cue_value v, cue_eopt *opts, size_t *count) {
	return cue_fields_raw(v, opts, cue_eopt_len(opts), count);
}
