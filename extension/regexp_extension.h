#pragma once
#include <sqlite3ext.h>

// Helpers exposed to Go via cgo
sqlite3_value* value_at(sqlite3_value **argv, int idx);
const unsigned char* value_text(sqlite3_value* v);
void result_null(sqlite3_context* ctx);
void result_error(sqlite3_context* ctx, const char* msg);
void result_int(sqlite3_context* ctx, int v);
int create_regexp(sqlite3* db);


