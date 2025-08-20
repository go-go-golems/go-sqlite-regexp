package main

// #cgo pkg-config: sqlite3
// #include <stdlib.h>
// #include "regexp_extension.h"
import "C"

import (
	"regexp"
	"unsafe"
)

//export go_register_regexp
func go_register_regexp(db *C.sqlite3) C.int {
	rc := C.create_regexp(db)
	if rc != C.SQLITE_OK {
		return rc
	}
	return C.SQLITE_OK
}

//export go_regexp
func go_regexp(ctx *C.sqlite3_context, argc C.int, argv **C.sqlite3_value) {
	if argc != 2 {
		msg := C.CString("regexp(): requires exactly 2 arguments: pattern, text")
		C.result_error(ctx, msg)
		C.free(unsafe.Pointer(msg))
		return
	}

	vPattern := C.value_at(argv, 0)
	vText := C.value_at(argv, 1)

	cText := (*C.uchar)(C.value_text(vText))
	cPattern := (*C.uchar)(C.value_text(vPattern))

	if cText == nil || cPattern == nil {
		C.result_null(ctx)
		return
	}

	pattern := C.GoString((*C.char)(unsafe.Pointer(cPattern)))
	text := C.GoString((*C.char)(unsafe.Pointer(cText)))

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		msg := C.CString(err.Error())
		C.result_error(ctx, msg)
		C.free(unsafe.Pointer(msg))
		return
	}

	if compiled.MatchString(text) {
		C.result_int(ctx, 1)
	} else {
		C.result_int(ctx, 0)
	}
}
