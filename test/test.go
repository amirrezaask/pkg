package test

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/mitchellh/mapstructure"
)

func Failf(t *testing.T, msgAndArgs ...any) {
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Failed")
	}
	if _, isString := msgAndArgs[0].(string); !isString {
		msgAndArgs[0] = fmt.Sprint(msgAndArgs[0])
	}
	var args []any
	if len(msgAndArgs) >= 2 {
		args = msgAndArgs[1:]
	}
	caller := strings.Join(callerInfo(), "\n\t")

	fmt.Printf("\n\n❌ Failed %s => %s\n\nStack: \n\t%s\n\n\n", t.Name(), fmt.Sprintf(msgAndArgs[0].(string), args...), caller)
	t.FailNow()
}

func AreNotEqual(t *testing.T, expected any, have any, msgAndArgs ...any) {
	if !ObjectsAreEqualValues(expected, have) {
		return
	}
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Assertion failed due to both arguments are equal")
	}
	if _, isString := msgAndArgs[0].(string); !isString {
		msgAndArgs[0] = fmt.Sprint(msgAndArgs[0])
	}
	var args []any
	if len(msgAndArgs) >= 2 {
		args = msgAndArgs[1:]
	}
	caller := strings.Join(callerInfo(), "\n\t")

	fmt.Printf("\n\n❌ Failed %s\n\tAsserts => %s\nexpected not '%v'\nhave '%v'\nStack: \n\t%s\n\n\n", t.Name(), fmt.Sprintf(msgAndArgs[0].(string), args...), expected, have, caller)

	t.FailNow()
}

func AssertEq(t *testing.T, expected, have any, msgAndArgs ...any) {
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Assertion failed due to type mismatch")
	}
	if reflect.TypeOf(expected) != reflect.TypeOf(have) {
		fmt.Printf("[%s] %s: expected('%T') have('%T')", t.Name(), fmt.Sprint(msgAndArgs[0]), expected, have)
	}
	AssertWeakEq(t, expected, have, msgAndArgs...)
}
func AssertWeakEq(t *testing.T, expected any, have any, msgAndArgs ...any) {

	if ObjectsAreEqualValues(expected, have) {
		return
	}
	caller := strings.Join(callerInfo(), "\n\t")

	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Assertion failed")
	}
	if _, isString := msgAndArgs[0].(string); !isString {
		msgAndArgs[0] = fmt.Sprint(msgAndArgs[0])
	}
	var args []any
	if len(msgAndArgs) > 1 {
		args = msgAndArgs[1:]
	}
	fmt.Printf("\n\n❌ Failed %s\n\tAsserts => %s\nexpected '%v'\nhave '%v'\nStack: \n\t%s\n\n\n", t.Name(), fmt.Sprintf(msgAndArgs[0].(string), args...), expected, have, caller)
	t.FailNow()
}

func NotEmpty(t *testing.T, obj any, msgAndArgs ...any) {
	v := reflect.ValueOf(obj)
	ty := reflect.TypeOf(obj)
	if ty.Kind() != reflect.Array && ty.Kind() != reflect.Slice && ty.Kind() != reflect.Map {
		panic(fmt.Sprintf("invalid argument to test.Empty, only Array|Slice|Map not %s", reflect.TypeOf(obj).Kind().String()))
		t.FailNow()
	}
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "expected empty")
	}

	AreNotEqual(t, 0, v.Len(), msgAndArgs...)
}

func IsEmpty(t *testing.T, obj any, msgAndArgs ...any) {
	v := reflect.ValueOf(obj)
	ty := reflect.TypeOf(obj)
	if ty.Kind() != reflect.Array && ty.Kind() != reflect.Slice && ty.Kind() != reflect.Map {
		panic(fmt.Sprintf("invalid argument to test.Empty, only Array|Slice|Map not %s", reflect.TypeOf(obj).Kind().String()))
		t.FailNow()
	}
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "expected empty")
	}

	AssertEq(t, 0, v.Len(), msgAndArgs...)
}

func ObjectsAreEqualValues(expected, actual interface{}) bool {
	if ObjectsAreEqual(expected, actual) {
		return true
	}

	expectedValue := reflect.ValueOf(expected)
	actualValue := reflect.ValueOf(actual)
	if !expectedValue.IsValid() || !actualValue.IsValid() {
		return false
	}

	expectedType := expectedValue.Type()
	actualType := actualValue.Type()
	if !expectedType.ConvertibleTo(actualType) {
		return false
	}

	if !isNumericType(expectedType) || !isNumericType(actualType) {
		// Attempt comparison after type conversion
		return reflect.DeepEqual(
			expectedValue.Convert(actualType).Interface(), actual,
		)
	}

	// If BOTH values are numeric, there are chances of false positives due
	// to overflow or underflow. So, we need to make sure to always convert
	// the smaller type to a larger type before comparing.
	if expectedType.Size() >= actualType.Size() {
		return actualValue.Convert(expectedType).Interface() == expected
	}

	return expectedValue.Convert(actualType).Interface() == actual
}

// isNumericType returns true if the type is one of:
// int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
// float32, float64, complex64, complex128
func isNumericType(t reflect.Type) bool {
	return t.Kind() >= reflect.Int && t.Kind() <= reflect.Complex128
}

func ObjectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

func NoError(t *testing.T, err error, msgAndArgs ...any) {
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Expected error to be nil")
	}
	AssertWeakEq(t, nil, err, msgAndArgs...)
}
func HasError(t *testing.T, err error, msgAndArgs ...any) {
	if len(msgAndArgs) < 1 {
		msgAndArgs = append(msgAndArgs, "Expected error to not be nil")
	}
	AreNotEqual(t, nil, err, msgAndArgs...)
}

func IsFalse(t *testing.T, b bool, msgAndArgs ...any) {
	AssertEq(t, false, b, msgAndArgs...)
}
func IsTrue(t *testing.T, b bool, msgAndArgs ...any) {
	AssertEq(t, true, b, msgAndArgs...)
}

func eqaulValues(t *testing.T, expected any, have any) bool { //TODO:
	eql := reflect.DeepEqual(expected, have)

	return eql
}

func AreSimilar(t *testing.T, expected any, have any) {
	hv := reflect.ValueOf(have)
	ev := reflect.ValueOf(expected)
	AssertWeakEq(t, ev.Kind(), hv.Kind(), "expected and have must have similar kind")

	AssertWeakEq(t, ev.Len(), hv.Len(), "expected and have must have equal len")
	if ev.Kind() == reflect.Slice {
		if ev.Type().Elem().Kind() != reflect.Map {
			m := []map[string]any{}
			NoError(t, mapstructure.Decode(expected, &m))
			expected = m
			ev = reflect.ValueOf(ev)
		}
		if hv.Type().Elem().Kind() != reflect.Map {
			m := []map[string]any{}
			NoError(t, mapstructure.Decode(have, &m))
			have = m
			hv = reflect.ValueOf(have)
		}
		for i := 0; i < ev.Len(); i++ {
			ei := ev.Index(i)
			hi := hv.Index(i)

			for _, k := range ev.Index(i).MapKeys() {
				haveValue := hi.MapIndex(k)
				IsTrue(t, haveValue.IsValid(), "expected key '%s' exist but it wasn't:", k.Interface())
				expectedValue := ei.MapIndex(k)
				AssertWeakEq(t, expectedValue.Interface(), haveValue.Interface(), "expected same value for key '%s'", k.Interface())
			}
		}
	}

}

func callerInfo() []string {
	isTest := func(name, prefix string) bool {
		if !strings.HasPrefix(name, prefix) {
			return false
		}
		if len(name) == len(prefix) { // "Test" is ok
			return true
		}
		r, _ := utf8.DecodeRuneInString(name[len(prefix):])
		return !unicode.IsLower(r)
	}

	var pc uintptr
	var ok bool
	var file string
	var line int
	var name string

	callers := []string{}
	for i := 0; ; i++ {
		pc, file, line, ok = runtime.Caller(i)
		if !ok {
			// The breaks below failed to terminate the loop, and we ran off the
			// end of the call stack.
			break
		}

		// This is a huge edge case, but it will panic if this is the case, see #180
		if file == "<autogenerated>" {
			break
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		name = f.Name()

		// testing.tRunner is the standard library function that calls
		// tests. Subtests are called directly by tRunner, without going through
		// the Test/Benchmark/Example function that contains the t.Run calls, so
		// with subtests we should break when we hit tRunner, without adding it
		// to the list of callers.
		if name == "testing.tRunner" {
			break
		}

		parts := strings.Split(file, "/")
		if len(parts) > 1 {
			filename := parts[len(parts)-1]
			dir := parts[len(parts)-2]
			if (dir != "assert" && dir != "mock" && dir != "require") || filename == "mock_test.go" {
				callers = append(callers, fmt.Sprintf("%s:%d", file, line))
			}
		}

		// Drop the package
		segments := strings.Split(name, ".")
		name = segments[len(segments)-1]
		if isTest(name, "Test") ||
			isTest(name, "Benchmark") ||
			isTest(name, "Example") {
			break
		}
	}
	slices.Reverse(callers)
	return callers
}
