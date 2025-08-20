## Building a SQLite Loadable Extension (.so/.dylib) from Go code

This guide documents how we added a SQLite loadable extension for `REGEXP` to `go-sqlite-regexp`, what pitfalls we hit, and a clean recipe to repeat next time.

The end result is a shared library you can load in the SQLite CLI or any embedding that supports extension loading:

- Linux: `./regexp.so`
- macOS: `./regexp.dylib`

Load and use in the `sqlite3` shell:

```sql
.load ./regexp
SELECT 'hello' REGEXP 'h.llo';   -- 1
SELECT regexp('h.llo','hello');  -- 1
```

### Overview

- We keep our Go logic (RE2 matching) in Go, but expose it as a SQLite function through cgo.
- We implement the required SQLite extension entrypoint (`sqlite3_<name>_init`) in C, not Go, because of macro usage.
- We compile with `-buildmode=c-shared` to produce a shared object.

### Files Added

- `extension/regexp_extension.c`: C entrypoint, wrappers for SQLite macros, trampoline into Go.
- `extension/regexp_extension.h`: C declarations for helpers exposed to Go.
- `extension/regexp_extension.go`: Go implementation of `regexp(pattern, text)` wired to SQLite via cgo.
- `extension/main_dummy.go`: Tiny `main()` required by `-buildmode=c-shared`.

### Key C code

```1:30:go-sqlite-regexp/extension/regexp_extension.c
#include <sqlite3ext.h>
#include <stdlib.h>
#include "regexp_extension.h"
SQLITE_EXTENSION_INIT1

// Forward declarations for Go
extern void go_regexp(sqlite3_context *ctx, int argc, sqlite3_value **argv);
extern int go_register_regexp(sqlite3* db);

// Helper to index argv
sqlite3_value* value_at(sqlite3_value **argv, int idx) { return argv[idx]; }
const unsigned char* value_text(sqlite3_value* v) { return sqlite3_value_text(v); }
void result_null(sqlite3_context* ctx) { sqlite3_result_null(ctx); }
void result_error(sqlite3_context* ctx, const char* msg) { sqlite3_result_error(ctx, msg, -1); }
void result_int(sqlite3_context* ctx, int v) { sqlite3_result_int(ctx, v); }

// C-visible trampoline that calls into Go implementation
static void call_go_regexp(sqlite3_context *ctx, int argc, sqlite3_value **argv) {
    go_regexp(ctx, argc, argv);
}

// Helper to register the function with SQLite
int create_regexp(sqlite3* db) {
    return sqlite3_create_function(db, "regexp", 2, SQLITE_UTF8, NULL, call_go_regexp, NULL, NULL);
}

int sqlite3_regexp_init(sqlite3 *db, char **pzErrMsg, const sqlite3_api_routines *pApi) {
    SQLITE_EXTENSION_INIT2(pApi);
    return go_register_regexp(db);
}
```

Notes:
- `SQLITE_EXTENSION_INIT1/2` macros must be used in C, not Go.
- We expose small helper functions so Go can call into SQLite C API cleanly (macros are not callable from cgo).

### Key Go code

```1:58:go-sqlite-regexp/extension/regexp_extension.go
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
```

Notes:
- The SQLite REGEXP operator uses the order `(text REGEXP pattern)`, but the function form is conventionally `regexp(pattern, text)`. We implement the function in that order and the `operator` becomes correct automatically.
- We use Go’s `regexp` (RE2) for deterministic performance.

### Makefile targets

```74:84:go-sqlite-regexp/Makefile
# Build loadable SQLite extension (.so/.dylib) using c-shared
# Output goes to dist/regexp.$(EXT)
so: so-linux

so-linux:
	@echo "Building SQLite loadable extension for Linux (.so)..."
	@CGO_ENABLED=1 go build -buildmode=c-shared -o regexp.so ./extension

so-darwin:
	@echo "Building SQLite loadable extension for macOS (.dylib)..."
	@CGO_ENABLED=1 go build -buildmode=c-shared -o regexp.dylib ./extension
```

Build with:

- Linux: `make so-linux` → `./regexp.so`
- macOS: `make so-darwin` → `./regexp.dylib`

### Why split C and Go?

- The SQLite extension ABI expects a C symbol `sqlite3_<name>_init` and relies on macros `SQLITE_EXTENSION_INIT1/2`. These macros can’t be invoked from Go.
- Initially we tried to use these macros inside the cgo preamble of a Go file, which led to errors like “could not determine what C.SQLITE_EXTENSION_INIT2 refers to” and “multiple definition of `sqlite3_regexp_init`”. Moving the entrypoint to a separate `.c` file resolves this cleanly.
- Some SQLite API utilities are macros (e.g., `sqlite3_result_int`, `sqlite3_value_text`). We wrapped them in real C functions so Go can call them.

### Common pitfalls we encountered

1) Macro calls from Go
- Symptom: `call of non-function C.sqlite3_result_error`
- Fix: Wrap macro calls in C functions and call those from Go.

2) `SQLITE_EXTENSION_INIT2` in Go preamble
- Symptom: `could not determine what C.SQLITE_EXTENSION_INIT2 refers to`
- Fix: Implement the extension entrypoint in C and call a Go-exported `go_register_regexp` from there.

3) Multiple definitions during linking
- Symptom: `multiple definition of sqlite3_regexp_init` or `sqlite3_api`
- Cause: Including the same C code twice (once via preamble, once via separate C file).
- Fix: Keep the macro-based entrypoint only in the C file; in Go, include only the header (`#include "regexp_extension.h"`).

4) Missing C standard headers
- Symptom: `could not determine what C.free refers to`
- Fix: Add `#include <stdlib.h>` in the cgo preamble that uses `C.free`.

5) Argument order confusion
- We validated with sqlite3 CLI that the function form should be `regexp(pattern, text)` to make the operator `text REGEXP pattern` behave intuitively. We adjusted argument extraction accordingly and verified with tests.

### End-to-end test

Non-interactive CLI test on Linux:

```bash
sqlite3 -batch ":memory:" -cmd ".load ./regexp" \
  "SELECT regexp('h.llo','hello');" \
  "SELECT regexp('^h.*o$','hello');" \
  "SELECT regexp('d','abc');" \
  "SELECT 'hello' REGEXP 'h.llo';" \
  "SELECT 'hello' REGEXP '^h.*o$';" \
  "SELECT 'abc' REGEXP 'd';"
# Expected output:
# 1
# 1
# 0
# 1
# 1
# 0
```

### Clean recipe to repeat next time

1) Create `extension/regexp_extension.c` with:
- `SQLITE_EXTENSION_INIT1/2`
- `sqlite3_<name>_init` entrypoint that calls an exported Go registration function
- A trampoline that calls an exported Go function for the SQL callback
- Small C wrappers for SQLite API macros you need to call from Go

2) Create `extension/regexp_extension.h` declaring the C wrappers used from Go.

3) Create `extension/regexp_extension.go`:
- cgo includes the header and `stdlib.h` for `C.free`
- `//export go_register_regexp` that calls `create_regexp`
- `//export go_regexp` implementing the function logic (compile RE2, match, set int result)

4) Add `extension/main_dummy.go` with an empty `main()` to satisfy `-buildmode=c-shared`.

5) Add Make targets:
- `go build -buildmode=c-shared -o regexp.so ./extension`
- Optionally, a macOS target outputting `regexp.dylib`

6) Test in the sqlite3 CLI with `.load ./regexp` and sample queries.

### Future improvements

- Add Windows build target (`.dll`) via `-buildmode=c-shared` and appropriate sqlite linkage.
- Add simple CI matrix to produce `.so` and `.dylib` artifacts.
- Add caching to the extension version as in the Go library (today it recompiles regex per call). A small LRU in Go could mirror the package cache.
- Add parameter validation and clearer error messages in SQL.


