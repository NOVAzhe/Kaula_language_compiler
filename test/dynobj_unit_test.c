/*
 * dynobj_unit_test.c - C-level unit tests for std/obj/ runtime
 *
 * Covers coverage gaps left by the language-level .kl tests:
 *   - Boxing / unboxing (64-bit precision)
 *   - Type-check helpers (dynobj_is_*)
 *   - Hash-table edge cases: delete, contains, clear, rehash, collision chaining
 *   - dynobj_invoke with a real C function pointer
 *   - func_object_create / func_object_fnptr
 *   - Reference-count correctness on key overwrite
 *   - NULL-safety for every public entry point
 *   - DynIntObject / DynFloatObject vtable methods (equals, hash, to_string)
 *
 * Build:
 *   gcc -std=c11 -Wall -Wextra -I std -I std/base -I std/memory -I std/obj \
 *       test/dynobj_unit_test.c \
 *       std/obj/dynobj.c std/obj/object.c std/obj/string_object.c \
 *       std/obj/int_object.c std/obj/float_object.c std/obj/bool_object.c \
 *       -lm -o build/dynobj_unit_test
 *
 * Run:
 *   ./build/dynobj_unit_test
 */

#include <assert.h>
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* Provide minimal stubs for std_malloc/std_free to avoid linking the full
 * KMM V4 allocator. The dynobj module uses these via #define aliases. */
void* std_malloc(size_t size) { return malloc(size); }
void std_free(void* ptr) { free(ptr); }

#include "obj/dynobj.h"
#include "obj/object.h"
#include "obj/string_object.h"
#include "obj/int_object.h"
#include "obj/float_object.h"
#include "obj/bool_object.h"

/* ============================================================================
 * Test harness
 * ============================================================================ */

static int tests_run = 0;
static int tests_passed = 0;
static int tests_failed = 0;

#define RUN_TEST(fn) do {                          \
    tests_run++;                                   \
    printf("  [TEST] %-50s ", #fn);                \
    fflush(stdout);                                \
    fn();                                          \
    tests_passed++;                                \
    printf("PASS\n");                              \
} while (0)

#define ASSERT_TRUE(cond) do {                     \
    if (!(cond)) {                                 \
        printf("FAIL\n    assertion failed: %s\n"  \
               "    at %s:%d\n", #cond,            \
               __FILE__, __LINE__);                \
        tests_failed++;                            \
        tests_passed--;                            \
        return;                                    \
    }                                              \
} while (0)

#define ASSERT_FALSE(cond)   ASSERT_TRUE(!(cond))
#define ASSERT_NULL(p)       ASSERT_TRUE((p) == NULL)
#define ASSERT_NOT_NULL(p)   ASSERT_TRUE((p) != NULL)
#define ASSERT_EQ_INT(a, b)  ASSERT_TRUE((a) == (b))
#define ASSERT_EQ_STR(a, b)  ASSERT_TRUE(strcmp((a), (b)) == 0)

/* ============================================================================
 * 1. Boxing / Unboxing
 * ============================================================================ */

static void test_box_unbox_i64(void) {
    /* Positive, negative, zero, and values that overflow 32-bit int. */
    Object* o1 = dynobj_box_i64(42);
    ASSERT_NOT_NULL(o1);
    ASSERT_EQ_INT(dynobj_unbox_i64(o1), 42);
    object_release(o1);

    Object* o2 = dynobj_box_i64(-100);
    ASSERT_EQ_INT(dynobj_unbox_i64(o2), -100);
    object_release(o2);

    Object* o3 = dynobj_box_i64(0);
    ASSERT_EQ_INT(dynobj_unbox_i64(o3), 0);
    object_release(o3);

    /* 64-bit value that doesn't fit in 32-bit int. */
    int64_t big = (int64_t)1 << 40;
    Object* o4 = dynobj_box_i64(big);
    ASSERT_EQ_INT(dynobj_unbox_i64(o4), big);
    object_release(o4);
}

static void test_box_unbox_f64(void) {
    Object* o1 = dynobj_box_f64(3.14);
    ASSERT_NOT_NULL(o1);
    ASSERT_TRUE(fabs(dynobj_unbox_f64(o1) - 3.14) < 1e-9);
    object_release(o1);

    Object* o2 = dynobj_box_f64(-0.001);
    ASSERT_TRUE(fabs(dynobj_unbox_f64(o2) - (-0.001)) < 1e-9);
    object_release(o2);

    /* Integer boxed but unboxed as f64 -- should promote. */
    Object* o3 = dynobj_box_i64(7);
    ASSERT_TRUE(fabs(dynobj_unbox_f64(o3) - 7.0) < 1e-9);
    object_release(o3);
}

static void test_box_unbox_bool(void) {
    Object* t = dynobj_box_bool(1);
    Object* f = dynobj_box_bool(0);
    ASSERT_NOT_NULL(t);
    ASSERT_NOT_NULL(f);
    ASSERT_EQ_INT(dynobj_unbox_bool(t), 1);
    ASSERT_EQ_INT(dynobj_unbox_bool(f), 0);
    object_release(t);
    object_release(f);
}

static void test_box_unbox_cstr(void) {
    Object* o1 = dynobj_box_cstr("hello");
    ASSERT_NOT_NULL(o1);
    ASSERT_EQ_STR(dynobj_unbox_cstr(o1), "hello");
    object_release(o1);

    /* NULL input should produce empty string, not crash. */
    Object* o2 = dynobj_box_cstr(NULL);
    ASSERT_NOT_NULL(o2);
    ASSERT_EQ_STR(dynobj_unbox_cstr(o2), "");
    object_release(o2);
}

static void test_unbox_null_safety(void) {
    /* Every unbox must tolerate NULL without crashing. */
    ASSERT_EQ_INT(dynobj_unbox_i64(NULL), 0);
    ASSERT_TRUE(fabs(dynobj_unbox_f64(NULL) - 0.0) < 1e-9);
    ASSERT_EQ_INT(dynobj_unbox_bool(NULL), 0);
    ASSERT_EQ_STR(dynobj_unbox_cstr(NULL), "");
}

static void test_unbox_type_mismatch(void) {
    /* Unboxing a value with the wrong accessor must return the default. */
    Object* s = dynobj_box_cstr("text");
    ASSERT_EQ_INT(dynobj_unbox_i64(s), 0);
    ASSERT_TRUE(fabs(dynobj_unbox_f64(s) - 0.0) < 1e-9);
    object_release(s);

    Object* i = dynobj_box_i64(99);
    /* cstr on an int falls through to object_to_string. */
    const char* repr = dynobj_unbox_cstr(i);
    ASSERT_NOT_NULL(repr);
    object_release(i);
}

/* ============================================================================
 * 2. Type-check helpers
 * ============================================================================ */

static void test_type_checks(void) {
    Object* i = dynobj_box_i64(1);
    Object* f = dynobj_box_f64(1.0);
    Object* s = dynobj_box_cstr("x");

    ASSERT_TRUE(dynobj_is_int(i));
    ASSERT_FALSE(dynobj_is_float(i));
    ASSERT_FALSE(dynobj_is_string(i));

    ASSERT_FALSE(dynobj_is_int(f));
    ASSERT_TRUE(dynobj_is_float(f));
    ASSERT_FALSE(dynobj_is_string(f));

    ASSERT_FALSE(dynobj_is_int(s));
    ASSERT_FALSE(dynobj_is_float(s));
    ASSERT_TRUE(dynobj_is_string(s));

    /* NULL is none of the above. */
    ASSERT_FALSE(dynobj_is_int(NULL));
    ASSERT_FALSE(dynobj_is_float(NULL));
    ASSERT_FALSE(dynobj_is_string(NULL));

    object_release(i);
    object_release(f);
    object_release(s);
}

/* ============================================================================
 * 3. Hash-table operations: set/get/delete/contains/clear/size
 * ============================================================================ */

static void test_basic_set_get(void) {
    DynObject* d = dynobj_create();
    ASSERT_NOT_NULL(d);
    ASSERT_EQ_INT(dynobj_size((Object*)d), 0);

    Object* v = dynobj_box_i64(10);
    dynobj_set((Object*)d, "key", v);
    object_release(v);  /* set retains, so we release our reference. */

    ASSERT_EQ_INT(dynobj_size((Object*)d), 1);
    ASSERT_TRUE(dynobj_contains((Object*)d, "key"));
    ASSERT_FALSE(dynobj_contains((Object*)d, "missing"));

    Object* got = dynobj_get((Object*)d, "key");
    ASSERT_NOT_NULL(got);
    ASSERT_EQ_INT(dynobj_unbox_i64(got), 10);

    ASSERT_NULL(dynobj_get((Object*)d, "missing"));

    dynobj_clear((Object*)d);
    ASSERT_EQ_INT(dynobj_size((Object*)d), 0);
    ASSERT_FALSE(dynobj_contains((Object*)d, "key"));

    object_release((Object*)d);
}

static void test_overwrite_existing_key(void) {
    /* Overwriting a key must release the old value and retain the new one.
     * We verify this by checking that size stays at 1 and the new value wins. */
    DynObject* d = dynobj_create();

    Object* v1 = dynobj_box_i64(100);
    Object* v2 = dynobj_box_i64(200);
    dynobj_set((Object*)d, "k", v1);
    object_release(v1);
    dynobj_set((Object*)d, "k", v2);
    object_release(v2);

    ASSERT_EQ_INT(dynobj_size((Object*)d), 1);
    ASSERT_EQ_INT(dynobj_unbox_i64(dynobj_get((Object*)d, "k")), 200);

    object_release((Object*)d);
}

static void test_delete(void) {
    DynObject* d = dynobj_create();

    Object* v = dynobj_box_i64(5);
    dynobj_set((Object*)d, "a", v);
    object_release(v);

    ASSERT_TRUE(dynobj_delete((Object*)d, "a"));
    ASSERT_FALSE(dynobj_delete((Object*)d, "a"));  /* already gone */
    ASSERT_FALSE(dynobj_delete((Object*)d, "zz")); /* never existed */
    ASSERT_EQ_INT(dynobj_size((Object*)d), 0);

    object_release((Object*)d);
}

static void test_delete_null_safety(void) {
    ASSERT_FALSE(dynobj_delete(NULL, "x"));
    DynObject* d = dynobj_create();
    ASSERT_FALSE(dynobj_delete((Object*)d, NULL));
    object_release((Object*)d);
}

static void test_set_null_safety(void) {
    /* NULL self must not crash; returns self unchanged. */
    Object* v = dynobj_box_i64(1);
    ASSERT_NULL(dynobj_set(NULL, "k", v));
    object_release(v);

    DynObject* d = dynobj_create();
    /* NULL key must be rejected. */
    Object* before = dynobj_set((Object*)d, NULL, dynobj_box_i64(9));
    ASSERT_EQ_INT(dynobj_size((Object*)d), 0);
    ASSERT_TRUE(before == (Object*)d);
    object_release((Object*)d);
}

static void test_get_null_safety(void) {
    ASSERT_NULL(dynobj_get(NULL, "k"));
    DynObject* d = dynobj_create();
    ASSERT_NULL(dynobj_get((Object*)d, NULL));
    object_release((Object*)d);
}

static void test_contains_null_safety(void) {
    ASSERT_FALSE(dynobj_contains(NULL, "k"));
    DynObject* d = dynobj_create();
    ASSERT_FALSE(dynobj_contains((Object*)d, NULL));
    object_release((Object*)d);
}

static void test_size_null_safety(void) {
    ASSERT_EQ_INT(dynobj_size(NULL), 0);
}

/* ============================================================================
 * 4. Rehash under load
 * ============================================================================ */

static void test_rehash(void) {
    /* Default bucket count is 16; rehash triggers at count >= 32.
     * Insert 40 distinct keys and verify all survive. */
    DynObject* d = dynobj_create();
    char key[32];

    for (int i = 0; i < 40; i++) {
        snprintf(key, sizeof(key), "k%d", i);
        Object* v = dynobj_box_i64(i * 10);
        dynobj_set((Object*)d, key, v);
        object_release(v);
    }

    ASSERT_EQ_INT(dynobj_size((Object*)d), 40);

    for (int i = 0; i < 40; i++) {
        snprintf(key, sizeof(key), "k%d", i);
        Object* got = dynobj_get((Object*)d, key);
        ASSERT_NOT_NULL(got);
        ASSERT_EQ_INT(dynobj_unbox_i64(got), i * 10);
    }

    object_release((Object*)d);
}

/* ============================================================================
 * 5. Hash collision chaining
 * ============================================================================ */

/* Force collisions by using keys that hash to the same bucket.
 * FNV-1a is deterministic, so we just insert many keys and rely on
 * the pigeonhole principle: with 16 buckets and 50 keys, some bucket
 * will have multiple entries chained together. */
static void test_collision_chaining(void) {
    DynObject* d = dynobj_create();
    char key[32];

    for (int i = 0; i < 50; i++) {
        snprintf(key, sizeof(key), "col_%d", i);
        Object* v = dynobj_box_i64(i);
        dynobj_set((Object*)d, key, v);
        object_release(v);
    }

    /* Delete every other key -- exercises mid-chain deletion. */
    for (int i = 0; i < 50; i += 2) {
        snprintf(key, sizeof(key), "col_%d", i);
        ASSERT_TRUE(dynobj_delete((Object*)d, key));
    }

    ASSERT_EQ_INT(dynobj_size((Object*)d), 25);

    /* Verify surviving keys. */
    for (int i = 1; i < 50; i += 2) {
        snprintf(key, sizeof(key), "col_%d", i);
        Object* got = dynobj_get((Object*)d, key);
        ASSERT_NOT_NULL(got);
        ASSERT_EQ_INT(dynobj_unbox_i64(got), i);
    }

    object_release((Object*)d);
}

/* ============================================================================
 * 6. func_object and dynobj_invoke
 * ============================================================================ */

/* A simple method: returns self["x"] + argv[0], both as i64. */
static Object* test_add_method(Object* self, size_t nargs, Object** argv) {
    (void)nargs;
    Object* x = dynobj_get(self, "x");
    int64_t xv = x ? dynobj_unbox_i64(x) : 0;
    int64_t av = argv && argv[0] ? dynobj_unbox_i64(argv[0]) : 0;
    return dynobj_box_i64(xv + av);
}

static void test_func_object(void) {
    FuncObject* fo = func_object_create((void*)test_add_method);
    ASSERT_NOT_NULL(fo);

    void* fp = func_object_fnptr((Object*)fo);
    ASSERT_TRUE(fp == (void*)test_add_method);
    object_release((Object*)fo);

    /* func_object_fnptr on wrong type returns NULL. */
    Object* i = dynobj_box_i64(1);
    ASSERT_NULL(func_object_fnptr(i));
    object_release(i);

    /* func_object_fnptr on NULL returns NULL. */
    ASSERT_NULL(func_object_fnptr(NULL));
}

static void test_invoke(void) {
    DynObject* d = dynobj_create();

    /* Give the object a field "x" and a method "add". */
    Object* xv = dynobj_box_i64(10);
    dynobj_set((Object*)d, "x", xv);
    object_release(xv);

    FuncObject* fo = func_object_create((void*)test_add_method);
    dynobj_set((Object*)d, "add", (Object*)fo);
    object_release((Object*)fo);

    Object* arg = dynobj_box_i64(5);
    Object* result = dynobj_invoke((Object*)d, "add", 1, arg);
    object_release(arg);

    ASSERT_NOT_NULL(result);
    ASSERT_EQ_INT(dynobj_unbox_i64(result), 15);
    object_release(result);

    /* Invoke a non-existent method returns NULL. */
    ASSERT_NULL(dynobj_invoke((Object*)d, "nope", 0));

    /* Invoke with NULL self/method returns NULL. */
    ASSERT_NULL(dynobj_invoke(NULL, "add", 0));
    ASSERT_NULL(dynobj_invoke((Object*)d, NULL, 0));

    object_release((Object*)d);
}

/* ============================================================================
 * 7. DynIntObject / DynFloatObject vtable methods
 * ============================================================================ */

static void test_dyn_int_vtable(void) {
    Object* a = dynobj_box_i64(42);
    Object* b = dynobj_box_i64(42);
    Object* c = dynobj_box_i64(99);

    /* equals */
    ASSERT_TRUE(object_equals(a, b));
    ASSERT_FALSE(object_equals(a, c));
    ASSERT_FALSE(object_equals(a, NULL));
    ASSERT_TRUE(object_equals(a, a));  /* same pointer */

    /* hash: equal values must have equal hashes */
    ASSERT_EQ_INT(object_hash(a), object_hash(b));

    /* to_string */
    ASSERT_EQ_STR(object_to_string(a), "42");

    object_release(a);
    object_release(b);
    object_release(c);
}

static void test_dyn_float_vtable(void) {
    Object* a = dynobj_box_f64(2.5);
    Object* b = dynobj_box_f64(2.5);
    Object* c = dynobj_box_f64(3.0);

    ASSERT_TRUE(object_equals(a, b));
    ASSERT_FALSE(object_equals(a, c));

    /* to_string uses %g, so "2.5" */
    ASSERT_EQ_STR(object_to_string(a), "2.5");

    object_release(a);
    object_release(b);
    object_release(c);
}

/* ============================================================================
 * 8. Cross-type equals must return false (not crash)
 * ============================================================================ */

static void test_cross_type_equals(void) {
    Object* i = dynobj_box_i64(1);
    Object* f = dynobj_box_f64(1.0);
    Object* s = dynobj_box_cstr("1");

    ASSERT_FALSE(object_equals(i, f));
    ASSERT_FALSE(object_equals(i, s));
    ASSERT_FALSE(object_equals(f, s));

    object_release(i);
    object_release(f);
    object_release(s);
}

/* ============================================================================
 * main
 * ============================================================================ */

int main(void) {
    printf("=== dynobj unit tests ===\n");

    printf("\n-- Boxing / Unboxing --\n");
    RUN_TEST(test_box_unbox_i64);
    RUN_TEST(test_box_unbox_f64);
    RUN_TEST(test_box_unbox_bool);
    RUN_TEST(test_box_unbox_cstr);
    RUN_TEST(test_unbox_null_safety);
    RUN_TEST(test_unbox_type_mismatch);

    printf("\n-- Type checks --\n");
    RUN_TEST(test_type_checks);

    printf("\n-- Hash-table operations --\n");
    RUN_TEST(test_basic_set_get);
    RUN_TEST(test_overwrite_existing_key);
    RUN_TEST(test_delete);
    RUN_TEST(test_delete_null_safety);
    RUN_TEST(test_set_null_safety);
    RUN_TEST(test_get_null_safety);
    RUN_TEST(test_contains_null_safety);
    RUN_TEST(test_size_null_safety);

    printf("\n-- Rehash / collisions --\n");
    RUN_TEST(test_rehash);
    RUN_TEST(test_collision_chaining);

    printf("\n-- FuncObject / invoke --\n");
    RUN_TEST(test_func_object);
    RUN_TEST(test_invoke);

    printf("\n-- Vtable methods --\n");
    RUN_TEST(test_dyn_int_vtable);
    RUN_TEST(test_dyn_float_vtable);
    RUN_TEST(test_cross_type_equals);

    printf("\n=== Results: %d/%d passed", tests_passed, tests_run);
    if (tests_failed > 0) {
        printf(", %d FAILED", tests_failed);
    }
    printf(" ===\n");

    return tests_failed > 0 ? 1 : 0;
}
