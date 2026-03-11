// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

import (
	"testing"
)

func TestQualifiedName_Encode_NilReceiver(t *testing.T) {
	var q *QualifiedName
	b, err := q.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if b != nil {
		t.Errorf("nil QualifiedName.Encode() should return nil bytes, got len=%d", len(b))
	}
}

func TestQualifiedName(t *testing.T) {
	cases := []CodecTestCase{
		{
			Name:   "normal",
			Struct: &QualifiedName{NamespaceIndex: 1, Name: "foobar"},
			Bytes: []byte{
				// namespace index
				0x01, 0x00,
				// name
				0x06, 0x00, 0x00, 0x00,
				0x66, 0x6f, 0x6f, 0x62, 0x61, 0x72,
			},
		},
		{
			Name:   "empty",
			Struct: &QualifiedName{NamespaceIndex: 1},
			Bytes: []byte{
				// namespace index
				0x01, 0x00,
				// name
				0xff, 0xff, 0xff, 0xff,
			},
		},
	}
	RunCodecTest(t, cases)
}
