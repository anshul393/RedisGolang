package resp

import (
	"reflect"
	"testing"
)

func TestRespSerializer(t *testing.T) {

	input := []*RespDataType{{Value: []*RespDataType{{Value: 1, Type: INTEGERS}, {Value: 2, Type: INTEGERS}, {Value: 3, Type: INTEGERS}}, Type: ARRAYS}, {Value: []*RespDataType{{Value: "Hello", Type: SIMPLESTRING}, {Value: "World", Type: ERRORS}}, Type: ARRAYS}}

	testSerializer := NewRespSerializer()
	r := testSerializer.Arrays(input)

	t.Logf("%#v", string(r))

	expected := []byte("*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n-World\r\n")
	if len(r) != len(expected) {
		t.Fatalf("expected length: %d, got: %d", len(expected), len(r))
	}

	for i, exp := range expected {
		if exp != r[i] {
			t.Fatalf("expected byte: %d, got: %d at index %d", exp, r[i], i)
		}
	}

}

func TestRespDeserializer(t *testing.T) {
	inputs := []struct {
		Data string
		Type byte
	}{
		{Data: "+OK\r\n", Type: SIMPLESTRING},
		{Data: "-Error message\r\n", Type: ERRORS},
		{Data: "$0\r\n\r\n", Type: BULKSTRINGS},
		{Data: "+hello world\r\n", Type: SIMPLESTRING},
		{Data: "*1\r\n$4\r\nping\r\n", Type: ARRAYS},
		{Data: "*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n", Type: ARRAYS},
		{Data: "*2\r\n$3\r\nget\r\n$3\r\nkey\r\n", Type: ARRAYS},
		{Data: "*2\r\n\r\n", Type: ARRAYS},
		{Data: "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n-World\r\n", Type: ARRAYS},
	}

	expectedResult := []struct {
		Data interface{}
		Type byte
	}{
		{Data: "OK", Type: SIMPLESTRING},
		{Data: "Error message", Type: ERRORS},
		{Data: "", Type: BULKSTRINGS},
		{Data: "hello world", Type: SIMPLESTRING},
		{Data: []*RespDataType{{Value: "ping", Type: BULKSTRINGS}}, Type: ARRAYS},
		{Data: []*RespDataType{{Value: "echo", Type: BULKSTRINGS}, {Value: "hello world", Type: BULKSTRINGS}}, Type: ARRAYS},
		{Data: []*RespDataType{{Value: "get", Type: BULKSTRINGS}, {Value: "key", Type: BULKSTRINGS}}, Type: ARRAYS},
		{Data: nil, Type: ARRAYS},
		{Data: []*RespDataType{{Value: []*RespDataType{{Value: 1, Type: INTEGERS}, {Value: 2, Type: INTEGERS}, {Value: 3, Type: INTEGERS}}, Type: ARRAYS}, {Value: []*RespDataType{{Value: "Hello", Type: SIMPLESTRING}, {Value: "World", Type: ERRORS}}, Type: ARRAYS}}, Type: ARRAYS},
	}

	testDeserializer := NewRespDeserializer()

	deserializeFuncs := map[byte]func([]byte, int) (*RespDataType, int, error){
		SIMPLESTRING: testDeserializer.SimpleString,
		ERRORS:       testDeserializer.SimpleErrors,
		INTEGERS:     testDeserializer.Integers,
		BULKSTRINGS:  testDeserializer.BulkString,
		ARRAYS:       testDeserializer.Arrays,
	}

	for i, input := range inputs {
		t.Log(i)

		var result *RespDataType
		var err error

		switch input.Type {
		case SIMPLESTRING:
			result, _, err = deserializeFuncs[SIMPLESTRING]([]byte(input.Data), 0)
		case ERRORS:
			result, _, err = deserializeFuncs[ERRORS]([]byte(input.Data), 0)
		case BULKSTRINGS:
			result, _, err = deserializeFuncs[BULKSTRINGS]([]byte(input.Data), 0)
		case ARRAYS:
			result, _, err = deserializeFuncs[ARRAYS]([]byte(input.Data), 0)

		}

		if result == nil {
			result = &RespDataType{Value: nil, Type: input.Type}
		}

		if err != nil {
			t.Logf("Error deserializing at index %v,  %v: %v", i, input.Data, err)
		}

		if !reflect.DeepEqual(expectedResult[i].Data, result.Value) {
			t.Fatalf("For input %v, expected %v, got %v", input.Data, expectedResult[i].Data, result.Value)
		}

		if result.Type != expectedResult[i].Type {
			t.Fatalf("For input %v, expected type %v, got %v", input.Data, expectedResult[i].Type, result.Type)
		}

	}

}
