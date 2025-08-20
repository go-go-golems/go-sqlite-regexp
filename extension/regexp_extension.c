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


