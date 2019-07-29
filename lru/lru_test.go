/*
 */

package lru

import (
	"fmt"
	"log"
	"runtime/debug"
	"testing"
)

/******************************************************************************
 *                            Constants
 ******************************************************************************/

// Possible operations to be performed on an LRU
const (
	Get       = "Get"
	Set       = "Set"
	Remove    = "Remove"
	Max       = "MaxStorage"
	Remaining = "RemainingStorage"
	Len       = "Len"
)

const operationFailMessage = `
***** Operation failed! *****
Command:  lru.%s(%s)
Expected: %s
Received: %s
`

const panicMessage = `Go panicked while executing student code!

Error: %s
Stacktrace:
%s

`

// Expected number of args for each method
var numArgs = map[string]int{
	Get:       1,
	Set:       2,
	Remove:    1,
	Max:       0,
	Remaining: 0,
	Len:       0,
}

/******************************************************************************
 *                            Structs
 ******************************************************************************/

// Binding is a key-value pair
type Binding struct {
	key string
	val []byte
}

type Record struct {
	val []byte
	ok  bool
}

func (a *Record) Equals(b *Record) bool {
	switch {
	case a.ok != b.ok:
		return false
	case len(a.val) != len(b.val):
		return false
	case a.val == nil && b.val != nil:
		return false
	}

	for i := range a.val {
		if a.val[i] != b.val[i] {
			return false
		}
	}
	return true
}

func (a *Record) String() string {
	if !a.ok {
		return "cache miss"
	}
	return fmt.Sprintf("cache hit:<'%s'>", a.val)
}

/******************************************************************************
 *                             Expected
 ******************************************************************************/
type Expected struct {
	exp interface{}
}

func (expected Expected) String() string {
	exp := expected.exp
	fstr := ""
	switch exp.(type) {
	case *Binding:
		fstr = "%s"
	case int, bool, string:
		fstr = "%v"
	default:
		fstr = "%v"
	}
	return fmt.Sprintf(fstr, exp)
}

func (expected Expected) Record() *Record {
	return expected.exp.(*Record)
}

func (expected Expected) Int() int {
	return expected.exp.(int)
}

func (expected Expected) Bool() bool {
	return expected.exp.(bool)
}

/******************************************************************************
 *                             Args
 ******************************************************************************/
type Args struct {
	args []interface{}
}

func (a *Args) String() string {
	switch len(a.args) {
	case 0:
		return ""
	case 1:
		// if only 1 arg, assume it to be the key
		return fmt.Sprintf("\"%s\"", a.args[0].(string))
	case 2:
		// if only 2 args, assume Set(key, val)
		//return fmt.Sprintf("\"%s\",'%s'==[% x]", a.args[0], a.args[1], a.args[1])
		return fmt.Sprintf("\"%s\",'%s'", a.args[0], a.args[1])
	default:
		return "???"
	}
}

func (a *Args) Len() int {
	return len(a.args)
}

func (a *Args) Key() string {
	if len(a.args) == 0 {
		return ""
	}
	return a.args[0].(string)
}

func (a *Args) Val() []byte {
	if len(a.args) < 2 {
		return nil
	}
	if a.args[1] == nil {
		return nil
	}
	return a.args[1].([]byte)
}

/******************************************************************************
 *                             Operation
 ******************************************************************************/
// Operation defines an operation on an LRU, like Get("key") or Set("key", "val")
// Methods: Get, Set, Remove, MaxStorage, RemainingStorage, Len
type Operation struct {
	method   string
	args     *Args
	expected Expected // ?
}

// NewOp constructs a New Operation, treating the first argument as the
// method, the final argument as the expected return value of the operation,
// and intervening arguments as the arguments to the function call
func NewOp(method string, extra ...interface{}) Operation {
	op := Operation{}
	op.method = method

	if len(extra) == 0 {
		log.Fatalln("Cannot make an operation without args or expected values")
	}

	op.args = &Args{extra[:len(extra)-1]}       // The first n-1 extras are arguments.
	op.expected = Expected{extra[len(extra)-1]} // The last extra is an expected value

	ValidateOperation(op)
	return op
}

// String returns a string representation of the operation
func (op Operation) String() string {
	return fmt.Sprintf("%s(%s)&%s", op.method, op.args, op.expected)
}

/******************************************************************************
 *                       Helper Functions
 ******************************************************************************/
// b is an alias for []byte() - it converts strings to []byte
// Not the best style, but saves 5 keystrokes over typing []byte()
func b(s string) []byte {
	return []byte(s)
}

// HasFactor returns true if num is evenly divisible by any of the candidates
// and that candidate is not num itself.
// Note that HasFactor does not count a number to be a factor of itself.
// Used for a fun application in the eviction order test
func HasFactor(num int, candidates []int) bool {
	for _, c := range candidates {
		if num != c && num%c == 0 {
			return true
		}
	}
	return false
}

// Wrapper for lru.NewLru. Only used when importing ./lru
// func NewLru(limit int) *LRU {
// 	return lru.NewLru(limit)
// }

func ValidateOperation(op Operation) {
	expArgs, mok := numArgs[op.method]
	if !mok {
		log.Fatalf("Unit Test Fatal Error: Unrecognized method %s\n", op.method)
	} else if expArgs != op.args.Len() {
		log.Fatalf("Unit Test Fatal Error: %s requires %d args, but found %d",
			op.method, expArgs, op.args.Len())
	}
}

func CatchPanic(t *testing.T, op Operation) {
	// If student code panicked, print stack trace and informative error message
	if e := recover(); e != nil {
		oldErrStr := e.(error).Error()
		trace := debug.Stack()
		panicMsg := fmt.Sprintf(panicMessage, oldErrStr, trace)
		t.Errorf(operationFailMessage, op.method, op.args, op.expected, panicMsg)
	}
}

func ExecuteOperation(t *testing.T, lru *LRU, op Operation) {
	ValidateOperation(op)

	fail := false
	var result interface{}

	// Catch panics raised by student code so all tests will finish running
	defer CatchPanic(t, op)

	switch op.method {
	case Get:
		key := op.args.Key()
		val, ok := lru.Get(key)

		result = &Record{val, ok}
		exp := op.expected.Record()

		if !exp.Equals(result.(*Record)) {
			fail = true
		}

	case Set:
		key := op.args.Key()
		val := op.args.Val()

		result = lru.Set(key, val)
		exp := op.expected.Bool()

		if result.(bool) != exp {
			fail = true
		}
	case Remove:
		key := op.args.Key()
		val, ok := lru.Remove(key)
		result = &Record{val, ok}
		exp := op.expected.Record()

		if !exp.Equals(result.(*Record)) {
			fail = true
		}

	case Max:
		result = lru.MaxStorage()
		exp := op.expected.Int()

		if result.(int) != exp {
			fail = true
		}

	case Remaining:
		result = lru.RemainingStorage()
		exp := op.expected.Int()

		if result.(int) != exp {
			fail = true
		}

	case Len:
		result = lru.Len()
		exp := op.expected.Int()

		if result.(int) != exp {
			fail = true
		}
	}

	if fail {
		// wrap result in Expected for smart printing
		t.Errorf(operationFailMessage, op.method, op.args, op.expected, Expected{result})
	}
}

// ExecuteOperations begins a new subtest and executes the given operations
// within it, asserting expected values to equal actual return values and
// failing the subtest if any unexpected values arise.
func ExecuteOperations(t *testing.T, lru *LRU, ops []Operation) {
	for _, op := range ops {
		name := op.String()
		t.Run(name, func(t *testing.T) {
			ExecuteOperation(t, lru, op)
		})
	}
}

func ExecuteOperationsNoSubtests(t *testing.T, lru *LRU, ops []Operation) {
	for _, op := range ops {
		ExecuteOperation(t, lru, op)
	}
}

// Construct a new LRU and try to add a single binding to it.
// Then verify that the add was successful if there was space for it,
// and unsuccessful otherwise
func CheckSingleBinding(t *testing.T, limit int, b Binding) {
	lru := NewLru(limit)

	rem := limit - len(b.key) - len(b.val)
	shouldFail := rem < 0
	len := 1
	rec := &Record{b.val, true}

	if shouldFail {
		rem = limit
		len = 0
		rec = &Record{nil, false}
	}

	ops := []Operation{
		NewOp(Set, b.key, b.val, !shouldFail),
		NewOp(Remaining, rem),
		NewOp(Len, len),
		NewOp(Max, limit),
		NewOp(Get, b.key, rec),
	}

	ExecuteOperations(t, lru, ops)
}

/******************************************************************************
 *                        TESTING SUITE
 ******************************************************************************/

// func TestTest(t *testing.T) {
// 	ops := []Operation{
// 		NewOp(Get, "key", &Record{nil, false}),
// 		NewOp(Set, "key", b("value"), true),
// 		NewOp(Get, "key", &Record{b("value"), true}),
// 	}
// 	lru := NewLru(1024)
// 	ExecuteOperations("name", t, lru, ops)
// }

/******************************************************************************
 *                             Basic tests
 ******************************************************************************/

func TestNewLRU(t *testing.T) {
	// desc := "Check that new LRUs are initialized with correct storage and size"
	for capacity := 16; capacity <= 1024; capacity <<= 2 {
		lru := NewLru(capacity)
		ops := []Operation{
			NewOp(Max, capacity),
			NewOp(Remaining, capacity),
			NewOp(Len, 0),
		}
		ExecuteOperations(t, lru, ops)
	}
}

func TestSmallLRU(t *testing.T) {
	// desc := "Test storage and size of a small LRU"
	key := "1234"
	val := b("1234")
	for capacity := 16; capacity <= 1024; capacity <<= 2 {
		lru := NewLru(capacity)
		ops := []Operation{
			NewOp(Set, key, val, true),
			NewOp(Max, capacity),
			NewOp(Remaining, capacity-len(key)-len(val)),
			NewOp(Len, 1),
		}
		ExecuteOperations(t, lru, ops)
	}
}

/******************************************************************************
 *                             Get/Set tests (no eviction)
 ******************************************************************************/

// Check that you cannot get bindigns that were never added to the LRU
func TestGetEmptyLRU(t *testing.T) {
	// desc := "Check that Get fails when called on an empty LRU"
	keys := []string{
		"hello world",
		"key",
		"value",
		"Get",
		"LRU",
	}

	lru := NewLru(1024)
	ops := make([]Operation, len(keys))
	for i, key := range keys {
		ops[i] = NewOp(Get, key, &Record{nil, false})
	}

	ExecuteOperations(t, lru, ops)
}

// TestSetBasic creates several new LRUs, adding one binding to each,
// and verifies that each binding is correctly added and uses the correct
// amount of storage
func TestSetBasic(t *testing.T) {
	// desc := "Add single binding to an LRU and check its validity"
	limit := 1024

	bindings := []Binding{
		{"Hello World", b("barbaz")},
		{"Abracadabra", b("Alakazam")},
		{"Key", b("Value")},
		{"Foo", b("bar")},
	}

	for _, kvp := range bindings {
		lru := NewLru(limit)
		key := kvp.key
		val := kvp.val
		ops := []Operation{
			NewOp(Remaining, limit),
			NewOp(Set, key, val, true),
			NewOp(Remaining, limit-len(key)-len(val)),
			NewOp(Get, key, &Record{val, true}),
		}

		ExecuteOperations(t, lru, ops)
	}
}

// TestSetMany adds several bindings to a single LRU and checks that
// the appropriate amount of storage was used, as well as checking that
// bindings were added successfully.
func TestSetMany(t *testing.T) {
	// desc := "Add many bindings to one LRU, check that resulting state is valid"
	expected := 10 * 1024
	lru := NewLru(expected)

	var op Operation // for printing errors in event of panic
	defer CatchPanic(t, op)
	ops := []Operation{NewOp(Remaining, expected)}

	value := []byte("barbaz")
	keyBase := "Hello World"
	totalStored := 0
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("%s%d", keyBase, i)
		totalStored += len(key)
		totalStored += len(value)
		ops = append(ops,
			NewOp(Set, key, value, true),
			NewOp(Remaining, expected-totalStored))
	}
	ops = append(ops,
		NewOp(Get, "Hello World22", &Record{value, true}),
		NewOp(Get, "Hello World44", &Record{value, true}),
		NewOp(Get, "Hello World88", &Record{value, true}),
	)

	// Too many operations -- don't open a subtest for each
	ExecuteOperationsNoSubtests(t, lru, ops)
}

// Check that items can continue to be added once the LRU becomes totally full
func TestSetFullLRU(t *testing.T) {
	// desc := "Check items can be added to a 'full' LRU if there's enough memory"
	lru := NewLru(30)
	ops := []Operation{NewOp(Len, 0)}

	for i := 0; i < 6; i++ {
		key := fmt.Sprintf("%5d", i)
		val := b(fmt.Sprintf("%5x", i))
		ops = append(ops, NewOp(Set, key, val, true))
		if i >= 3 {
			ops = append(ops, NewOp(Len, 3))
		}
	}

	ExecuteOperations(t, lru, ops)
}

func TestSetNotEnoughMemory(t *testing.T) {
	// desc := "Check that bindings too large for the LRU are rejected"
	lru := NewLru(10)
	ops := []Operation{}

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf(">%6d<", i)
		val := b(fmt.Sprintf(">%6x<", i))
		ops = append(ops,
			NewOp(Set, key, val, false),
			NewOp(Get, key, &Record{nil, false}),
		)
	}

	ExecuteOperations(t, lru, ops)
}

func TestSetZeroCapacity(t *testing.T) {
	// desc := "Attempt to construct and add bindings to a 0-capacity LRU"
	lru := NewLru(0)
	bindings := []Binding{
		{"hello", b("world")},
		{"abra", b("kadabra")},
		{"foo", b("bar")},
		{"key", b("val")},
	}

	ops := make([]Operation, len(bindings))
	for i, b := range bindings {
		ops[i] = NewOp(Set, b.key, b.val, false)
	}

	// Tricky - may decide this should not work
	ops = append(ops, NewOp(Set, "", []byte{}, true))

	ExecuteOperations(t, lru, ops)
}

func TestEmptyKey(t *testing.T) {
	// desc := "Check that the empty string can be used as a valid key"
	limit := 1024
	b := Binding{"", b("Value")}
	CheckSingleBinding(t, limit, b)
}

func TestEmptyValue(t *testing.T) {
	// desc := "Check that the empty []byte can be used as a valid value"
	limit := 1024
	b := Binding{"key", []byte{}}
	CheckSingleBinding(t, limit, b)
}

// Test that nil is an acceptable value.
// We may want to disallow this in the spec, in which case this test should
// be modified
func TestNilValue(t *testing.T) {
	// desc := "Check that nil can be used as a valid value"
	limit := 1024
	b := Binding{"key", nil}
	CheckSingleBinding(t, limit, b)
}

func TestBinaryValue(t *testing.T) {
	// desc := "Check that values can be non-ASCII (binary)"
	limit := 1024
	val := []byte{0x00, 0x01, 0xFF, 0x15, 0xEC}
	b := Binding{"key", val}
	CheckSingleBinding(t, limit, b)
}

func TestNonASCIIKeys(t *testing.T) {
	// desc := "Check that keys can be non-ASCII (Unicode)"
	limit := 1024
	// Various emoji and symbols
	bindings := []Binding{
		{"\xF0\x9F\x98\x82 \xF0\x9F\x9A\x80", b("\xE2\x9C\x94 \xF0\x9F\x9A\x97")},
		{"\xF0\x9F\x9A\xA9 \xF0\x9F\x86\x97", b("\xC2\xA9 \xE2\x98\x80")},
		{"\xE2\x98\x91 \xE2\x98\xBA", b("\xF0\x9F\x9A\x97 \xE2\x98\x94")},
	}
	for _, b := range bindings {
		CheckSingleBinding(t, limit, b)
	}
}

func TestSetSimpleOverwrite(t *testing.T) {
	// desc := "Test that values are overwritten when Set() called with same key"
	limit := 1024
	lru := NewLru(limit)

	key := "key"
	old := b("old")
	val := b("new")

	ops := []Operation{
		NewOp(Set, key, old, true),
		NewOp(Get, key, &Record{old, true}),
		NewOp(Set, key, val, true),
		NewOp(Get, key, &Record{val, true}),
	}

	ExecuteOperations(t, lru, ops)
}

func TestSetAdvancedOverwrite(t *testing.T) {
	// desc := "Test that internal state correctly updated when values overwritten"
	limit := 1024
	lru := NewLru(limit)

	key := "key"
	old := b("old")
	val := b("nw")

	oldSz := len(key) + len(old)
	newSz := len(key) + len(val)

	ops := []Operation{
		NewOp(Set, key, old, true),
		NewOp(Get, key, &Record{old, true}),
		NewOp(Max, limit),
		NewOp(Len, 1),
		NewOp(Remaining, limit-oldSz),
		NewOp(Set, key, val, true),
		NewOp(Get, key, &Record{val, true}),
		NewOp(Max, limit),
		NewOp(Len, 1),
		NewOp(Remaining, limit-newSz),
	}

	ExecuteOperations(t, lru, ops)
}

// Proof of concept - if this is something we wish to test for we can include it
// func TestDefensiveCopies(t *testing.T) {
// 	limit := 1024
// 	lru := NewLru(limit)
//
// 	foo := "foo"
// 	boo := "boo"
//
// 	pkey := new(string)
// 	*pkey = foo
//
// 	val := b("bar")
//
// 	lru.Set(*pkey, val)
// 	val[0] = 'f'
//
// 	v, _ := lru.Get(foo)
// 	fmt.Printf("%s\n", v)
//
// 	*pkey = boo
// 	v1, ok := lru.Get(foo)
// 	if ok {
// 		fmt.Printf("%s\n", v1)
// 	} else {
// 		fmt.Printf("%s\n", "cache miss")
// 	}
// }

/******************************************************************************
 *                             Remove tests
 ******************************************************************************/

func TestRemoveBasic(t *testing.T) {
	// desc := "Check that removed bindings are no longer accessible"
	limit := 1024
	lru := NewLru(limit)

	key := "key"
	val := b("value")

	ops := []Operation{
		NewOp(Set, key, val, true),
		NewOp(Get, key, &Record{val, true}),
		NewOp(Remove, key, &Record{val, true}),
		NewOp(Get, key, &Record{nil, false}),
	}

	ExecuteOperations(t, lru, ops)
}

func TestRemoveMemoryReleased(t *testing.T) {
	// desc := "Check that removed bindings no longer consume storage"
	limit := 1024
	lru := NewLru(limit)

	N := 4
	keys := make([]string, N)
	vals := make([][]byte, N)
	ops := []Operation{}

	// add several bindings
	for i := 0; i < N; i++ {
		keys[i] = fmt.Sprintf("%3d", i)
		vals[i] = b(fmt.Sprintf("%3x", i))
		ops = append(ops, NewOp(Set, keys[i], vals[i], true))
	}

	// remove some but not all
	for i := 0; i < 2; i++ {
		n := N - i - 1
		rem := limit - (n * (len(keys[0]) + len(vals[0])))
		ops = append(ops,
			NewOp(Remove, keys[i], &Record{vals[i], true}),
			NewOp(Len, n),
			NewOp(Remaining, rem),
		)
	}

	ExecuteOperations(t, lru, ops)
}

func TestRemoveOverwrite(t *testing.T) {
	// desc := "Check that overwriting values doesn't affect removal"
	limit := 1024
	lru := NewLru(limit)

	key := "key"
	old := b("old")
	val := b("value")

	ops := []Operation{
		NewOp(Set, key, old, true),
		NewOp(Get, key, &Record{old, true}),
		NewOp(Set, key, val, true),
		NewOp(Get, key, &Record{val, true}),
		NewOp(Remove, key, &Record{val, true}),
		NewOp(Get, key, &Record{nil, false}),
	}

	ExecuteOperations(t, lru, ops)
}

func TestRemoveEmpty(t *testing.T) {
	// desc := "Attempt to remove a binding from an empty LRU"
	lru := NewLru(1024)
	ops := []Operation{
		NewOp(Remove, "key", &Record{nil, false}),
		NewOp(Remove, "nada", &Record{nil, false}),
		NewOp(Remove, "foo", &Record{nil, false}),
		NewOp(Remove, "bar", &Record{nil, false}),
	}
	ExecuteOperations(t, lru, ops)
}

func TestRemoveNonexistant(t *testing.T) {
	// desc := "Attempt to remove a binding not in the LRU"
	limit := 1024
	lru := NewLru(limit)

	key := "key"
	val := b("value")

	ops := []Operation{
		NewOp(Remove, key, &Record{nil, false}),
		NewOp(Set, key, val, true),
		NewOp(Get, key, &Record{val, true}),
		NewOp(Remove, key, &Record{val, true}),
		NewOp(Remove, key, &Record{nil, false}),
	}

	ExecuteOperations(t, lru, ops)
}

/******************************************************************************
 *                             Eviction tests
 ******************************************************************************/

func TestSetEvict(t *testing.T) {
	// desc := "Overfill an LRU and check the correct binding is evicted"
	expected := 100
	lru := NewLru(expected)
	ops := make([]Operation, 11)
	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("%5d", i)
		value := []byte(fmt.Sprintf("%5x", i))
		ops[i] = NewOp(Set, key, value, true)
	}
	firstKey := fmt.Sprintf("%5d", 0)
	ops = append(ops,
		NewOp(Len, 10),
		NewOp(Get, firstKey, &Record{nil, false}),
	)

	ExecuteOperations(t, lru, ops)
}

func TestEvictAfterUse(t *testing.T) {
	// desc := "Overfill an LRU, Getting some items, then check for correct eviction"
	expected := 100
	lru := NewLru(expected)
	ops := make([]Operation, 10)
	keys := make([]string, 11)
	vals := make([][]byte, 11)
	for i := 0; i < 11; i++ {
		keys[i] = fmt.Sprintf("%5d", i)
		vals[i] = []byte(fmt.Sprintf("%5x", i))
		if i < 10 {
			ops[i] = NewOp(Set, keys[i], vals[i], true)
		}
	}
	ops = append(ops,
		NewOp(Len, 10),
		NewOp(Get, keys[0], &Record{vals[0], true}),
		NewOp(Set, keys[10], vals[10], true),
		NewOp(Len, 10),
		NewOp(Get, keys[1], &Record{nil, false}),
	)

	ExecuteOperations(t, lru, ops)
}

// Test that the entries are evicted in the appropriate order.
// At the end of the test the LRU contains primes less than 50 using
// a very strange implementation of the sieve of eratosthenes
func TestEvictionOrder(t *testing.T) {
	// desc := "Ensure that evictions occur in the appropriate order"
	if testing.Short() {
		t.Skip("Skipping eviction order test in short mode")
	}

	limit := 64 // 2 bytes/key, 2 bytes/val --> 15 values + 1 placeholder
	lru := NewLru(limit)

	N := 50
	primes := []int{2, 3, 5, 7, 11, 13, 17, 19, 23}

	keys := make([]string, N+1)
	vals := make([][]byte, N+1)
	ops := []Operation{}

	for i := 2; i <= 50; i++ {
		keys[i] = fmt.Sprintf("%2d", i)
		vals[i] = b(keys[i])
	}

	// Find primes
	for i := 2; i <= 50; i++ {
		ops = append(ops, NewOp(Set, keys[i], vals[i], true))
		// Touch all the possible primes
		for j := 2; j <= i; j++ {
			if !HasFactor(j, primes) {
				if keys[j] == "" {
					panic(j)
				}
				ops = append(ops, NewOp(Get, keys[j], &Record{vals[j], true}))
			}
		}
	}

	// Check result
	expected := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 50}
	for i, x := range expected {
		ops = append(ops,
			NewOp(Remove, keys[x], &Record{vals[x], true}),
			NewOp(Len, 16-i-1),
		)
	}

	// way too many ops - don't open a subtest for each
	ExecuteOperationsNoSubtests(t, lru, ops)
}

func TestPrematureEviction(t *testing.T) {
	// desc := "Make sure that values are not evicted before they need to be"
	limit := 4 // 2 bytes per binding, 3 bindings
	lru := NewLru(limit)

	keys := make([]string, 5)
	vals := make([][]byte, 5)

	ops := []Operation{}

	for i := 0; i < 5; i++ {
		keys[i] = fmt.Sprintf("%d", i)
		vals[i] = b(keys[i])
		ops = append(ops, NewOp(Set, keys[i], vals[i], true))
		if i >= 1 {
			ops = append(ops,
				NewOp(Get, keys[i-1], &Record{vals[i-1], true}),
				NewOp(Get, keys[i], &Record{vals[i], true}),
			)
		}
	}
	ExecuteOperations(t, lru, ops)
}

func TestEvictStorage(t *testing.T) {
	// desc := "Check that storage is freed correctly when eviction occurs"
	limit := 10
	lru := NewLru(limit)

	ops := []Operation{
		NewOp(Set, "12345", b("12345"), true),
		NewOp(Max, limit),
		NewOp(Len, 1),
		NewOp(Remaining, 0),
		NewOp(Set, "123", b("123"), true),
		NewOp(Len, 1),
		NewOp(Remaining, limit-len("123")-len(b("123"))),
	}

	ExecuteOperations(t, lru, ops)
}

func TestUnicodeEviction(t *testing.T) {
	// desc := "Check proper length is used when evicting Unicode strings"
	limit := 10
	lru := NewLru(limit)

	key := "\xF0\x9F\x98\x82"
	val := b("\xF0\x9F\x99\x88")

	key2 := "12"
	val2 := b("12")

	ops := []Operation{
		NewOp(Set, key, val, true),
		NewOp(Get, key, &Record{val, true}),
		NewOp(Set, key2, val2, true),
		NewOp(Len, 1),
		NewOp(Remaining, limit-len(key2)-len(val2)),
		NewOp(Get, key, &Record{nil, false}),
		NewOp(Get, key2, &Record{val2, true}),
	}

	ExecuteOperations(t, lru, ops)
}

func TestOverevictOnOverwrite(t *testing.T) {
	// desc := "Check that overeviction doesn't occur when updating existing key"
	limit := 20
	lru := NewLru(limit)

	ops := []Operation{
		NewOp(Max, limit),
		NewOp(Set, "abcd", b("efgh"), true),
		NewOp(Set, "1234", b("5678"), true),
		NewOp(Remaining, 4),
		NewOp(Set, "1234", b("12345678"), true), // should not need to evict "abcd"
		NewOp(Get, "abcd", &Record{b("efgh"), true}),
	}

	ExecuteOperations(t, lru, ops)
}

/******************************************************************************
 *                          Performance & Memory
 ******************************************************************************/
// Be careful with b *testing.B, as it shadows the []byte() alias b()

func BenchmarkSet(b *testing.B) {
	lru := NewLru(8192 * 10)

	for i := 0; i < b.N; i++ {
		key := string(i)
		val := []byte(key)
		ok := lru.Set(key, val)
		if !ok {
			b.FailNow()
		}
	}
}

func BenchmarkSetGet(b *testing.B) {
	lru := NewLru(8192 * 10)

	for i := 0; i < b.N; i++ {
		key := string(i)
		val := []byte(key)
		sok := lru.Set(key, val)
		_, gok := lru.Get(key)
		if !sok || !gok {
			b.FailNow()
		}
	}
}

// // Golang doesn't have a straightforward way of doing memory analysis that i've
// // been able to find
// func PrintMemStats(m1 runtime.MemStats, m2 runtime.MemStats) {
// 	fmt.Printf("Alloc:        %d\n", m2.Alloc-m1.Alloc)
// 	fmt.Printf("Sys:          %d\n", m2.Sys-m1.Sys)
// 	fmt.Printf("Mallocs:      %d\n", m2.Mallocs-m1.Mallocs)
// 	fmt.Printf("Frees:        %d\n", m2.Frees-m1.Frees)
// 	fmt.Printf("Malloc-Free:  %d\n", (m2.Mallocs-m1.Mallocs)-(m2.Frees-m1.Frees))
// 	fmt.Printf("Heap Objects: %d\n", m2.HeapObjects-m1.HeapObjects)
// 	fmt.Printf("Stack1 Inuse: %d\n", m1.StackInuse)
// 	fmt.Printf("Stack2 Inuse: %d\n", m2.StackInuse)
// }
//
// func TestMemory(t *testing.T) {
// 	runtime.GC()
//
// 	// try to force heap allocations for keys and lru
// 	newKvp := func(lru *LRU, i int) *Binding {
// 		k := fmt.Sprintf("%40d", i) // <32 may get stacked?
// 		s := new(Binding)
// 		s.key = k
// 		s.val = []byte(k)
//
// 		// this is only here to force LRU to escape to heap
// 		if lru.Len() < 0 {
// 			s.key = "oooooooooo"
// 		}
//
// 		return s
// 	}
//
// 	var m1 runtime.MemStats
// 	runtime.ReadMemStats(&m1)
//
// 	// possibly the LRU is stored on stack? unsure why allocs not behaving as expected
// 	lru := NewLru(8000000)
//
// 	for i := 0; i < 800000; i++ {
// 		kvp := newKvp(lru, i)
// 		lru.Set(kvp.key, kvp.val)
// 	}
//
// 	runtime.GC()
//
// 	var m2 runtime.MemStats
// 	runtime.ReadMemStats(&m2)
//
// 	PrintMemStats(m1, m2)
//
// }

// This doesn't work either as it measures total memory allocated, not
// the actual memory referenced by the LRU
// func BenchmarkMemory(b *testing.B) {
// 	N := 100000
//
// 	keys := make([]string, N)
// 	vals := make([][]byte, N)
// 	for i := 0; i < N; i++ {
// 		keys[i] = fmt.Sprintf("%20d", i)
// 		vals[i] = []byte(keys[i])
// 	}
//
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		// Make an LRU with fixed number of bindings
// 		lru := NewLru(200)
// 		for j := 0; j <= N; j++ {
// 			lru.Set(keys[i], vals[i])
// 		}
// 	}
// }

// This test is giving unexpected results - revisit at some point
// func TestMemory(t *testing.T) {
// 	runtime.GC()
// 	var orig runtime.MemStats
// 	runtime.ReadMemStats(&orig)
// 	fmt.Printf("Allocated bytes: %d\n", orig.Alloc)
//
// 	limit := 200
// 	lru := NewLru(limit)
//
// 	for i := 0; i < 1000000; i++ {
// 		key := fmt.Sprintf("%20d", i)
// 		val := b(key)
// 		lru.Set(key, val)
// 	}
//
// 	//	runtime.GC()
// 	var m runtime.MemStats
// 	runtime.ReadMemStats(&m)
// 	fmt.Printf("Allocated bytes: %d\n", m.Alloc)
// 	fmt.Printf("Delta: %d\n", m.Alloc-orig.Alloc)
// }

/******************************************************************************
 *                             ...
 ******************************************************************************/
/*

Ways LRU can fail:

Already being tested:

  Storage methods:
  - MaxStorage(), Len(), RemainingStorage() return wrong value
		- on empty LRUs
		- on partly full LRUs
		- (implicitly) on full/overfull LRUs

  Get & Set:
  - Get item that was never Set
  - Set item but cannot Get it
	- Still memory left, but cannot add items
	- Adds an item there is no space for
	- adding to a zero-capacity list
	- Check empty string as a key
	- Check empty []byte as a value
		- Was going to test nil as a value also, but it breaks some of the testing scripts
		- I don't think this corner case is super important, as it might be forbidden by spec anyway
	- Check non-string []byte as a value (i.e. binary)
	- Check non-ASCII keys
	- Test value overwriting when Set called with same key.
	- Check that size is not incremented when an old value is replaced with a new one

	Remove:
	- Basic add and then remove
	- Ensure remove updates memory
	- Can still get item after being Removed
	- Test adding an item, overwriting it, then removing
	- Remove an item not in the list
	- Try removing from an empty list

  Eviction:
	- Can still get item after being evicted
	- Items are evicted in incorrect order (e.g. LR added, MR added, MR used)
	- Items are evicted before they should be
	- Check that storage is not double counted (i.e. freed) when items are evicted
	- Overevicting for replacement values of an existing binding

  Performance & Memory:
  -

Not yet tested:
	New / Storage:
	- Attempt to construct a negative capacity LRU (presumably returns nil or panics)

	Get / Set:

	Remove:

	Evict:
	- test eviction corner cases with Unicode
	  - eg. Add 2-rune, 8-byte binding to a 10-byte LRU.
		- see if it gets evicted when we try to add 4 ASCII bytes.
		- this will check if the students are counting runes instead of bytes

	Memory / Performance:
	Note: So far I haven't been able to find any satisfactory way to measure
	      the memory usage of an LRU in Go
	- Check that memory is freed when items are evicted
    - i.e. a possible bad implementation might keep the old data around, but
      flag it as inaccessible or refuse to return it
	- Some test to discourage brute force solutions
    - Large capacity with many small blocks
    - Lots of worst case usage
	- Check that memory used is proportional to size, not capacity


// not exactly sure where these fit in:
// most are tested implicitly by other tests
	Various corner case / stress test scenarios
	  Test an LRU whose capacity is zero
	  Test an LRU whose capacity is nonzero but very small (<10)
	  Test an LRU that fills up exactly (size == capacity),
	     with nice even block sizes
	  Test an LRU that attempts to overfill slightly,
	     with constant block sizes
	  Test an LRU that attempts to overfill slightly,
	     with irregular block sizes

	Open Questions:
  - Should values be mutable or should there be defensive copies?
	- Confirm that spec asks for empty LRUs to have len=0, remaining=capacity
	- Confirm behavior for nil values
	- Confirm behavior for negative limits
*/
