Yes — but only if you build the extension as a **shared library** that SQLite can load, not as part of the statically linked `mattn/go-sqlite3` Go binary.

The Go driver’s built-in extensions (like `REGEXP`) exist only inside the Go program’s compiled binary. The stock `sqlite3` CLI can’t “see” them unless you separately compile your code into a `.so`/`.dylib` and load it at runtime using SQLite’s extension loading mechanism.

Here’s how you can make it work:

---

### 1. Write the Extension in C (or Cgo) Form

If your `mattn/go-sqlite3` REGEXP implementation is written in Go, you can:

* Either rewrite it in C so it compiles cleanly as a SQLite loadable extension.
* Or, use **Cgo** to build a `.so` from Go code, but this is more complex because SQLite extensions expect a C ABI entrypoint.

For example, a simple C extension entrypoint looks like:

```c
#include <sqlite3ext.h>
SQLITE_EXTENSION_INIT1

static void regexp_func(sqlite3_context *ctx, int argc, sqlite3_value **argv) {
    // your REGEXP logic here
}

int sqlite3_regexp_init(sqlite3 *db, char **pzErrMsg,
                        const sqlite3_api_routines *pApi) {
    SQLITE_EXTENSION_INIT2(pApi)
    sqlite3_create_function(db, "regexp", 2, SQLITE_UTF8, NULL, regexp_func, NULL, NULL);
    return SQLITE_OK;
}
```

---

### 2. Compile It as a Loadable Extension

Example:

```bash
gcc -fPIC -shared -o regexp.so regexp.c \
    -I/usr/include/sqlite3
```

---

### 3. Load It in the Standard SQLite CLI

Inside the SQLite CLI:

```sql
.load ./regexp
SELECT 'abc' REGEXP 'a.c';
```

If `sqlite3` on your system was compiled with `-DSQLITE_ENABLE_LOAD_EXTENSION`, you can do:

```sql
sqlite> .load ./regexp
sqlite> SELECT 'hello' REGEXP 'h.llo';
```

Otherwise, you need to enable it in the CLI first:

```sql
sqlite> .load ./regexp
```

---

### Key Points

* The Go driver’s extension code is **inside your Go binary**; the CLI won’t have it unless you extract and compile it as a standalone `.so`.
* You **cannot** just “point” the CLI to your Go program and expect it to pick up the REGEXP — you must repackage it in SQLite’s shared library extension format.
* Once built as `.so` (Linux) or `.dylib` (macOS), you can `.load` it in *any* SQLite CLI or any SQLite instance that supports extension loading.

---

If you want, I can walk you through taking your **exact Go-based REGEXP implementation** from `mattn/go-sqlite3` and building it as a `.so` so it works in `sqlite3`. That’s slightly tricky because of Go’s C ABI rules, but doable. Would you like me to detail that build process?
