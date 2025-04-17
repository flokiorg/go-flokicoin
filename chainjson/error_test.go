// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainjson_test

import (
	"testing"

	"github.com/flokiorg/go-flokicoin/chainjson"
)

// TestErrorCodeStringer tests the stringized output for the ErrorCode type.
func TestErrorCodeStringer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   chainjson.ErrorCode
		want string
	}{
		{chainjson.ErrDuplicateMethod, "ErrDuplicateMethod"},
		{chainjson.ErrInvalidUsageFlags, "ErrInvalidUsageFlags"},
		{chainjson.ErrInvalidType, "ErrInvalidType"},
		{chainjson.ErrEmbeddedType, "ErrEmbeddedType"},
		{chainjson.ErrUnexportedField, "ErrUnexportedField"},
		{chainjson.ErrUnsupportedFieldType, "ErrUnsupportedFieldType"},
		{chainjson.ErrNonOptionalField, "ErrNonOptionalField"},
		{chainjson.ErrNonOptionalDefault, "ErrNonOptionalDefault"},
		{chainjson.ErrMismatchedDefault, "ErrMismatchedDefault"},
		{chainjson.ErrUnregisteredMethod, "ErrUnregisteredMethod"},
		{chainjson.ErrNumParams, "ErrNumParams"},
		{chainjson.ErrMissingDescription, "ErrMissingDescription"},
		{0xffff, "Unknown ErrorCode (65535)"},
	}

	// Detect additional error codes that don't have the stringer added.
	if len(tests)-1 != int(chainjson.TstNumErrorCodes) {
		t.Errorf("It appears an error code was added without adding an " +
			"associated stringer test")
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}

// TestError tests the error output for the Error type.
func TestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   chainjson.Error
		want string
	}{
		{
			chainjson.Error{Description: "some error"},
			"some error",
		},
		{
			chainjson.Error{Description: "human-readable error"},
			"human-readable error",
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.Error()
		if result != test.want {
			t.Errorf("Error #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}
