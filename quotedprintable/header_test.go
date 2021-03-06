// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quotedprintable

import (
	"fmt"
	"strings"
	"testing"
)

func ExampleEncodeHeader() {
	fmt.Println(StdHeaderEncoder.EncodeHeader("Cofee"))
	fmt.Println(StdHeaderEncoder.EncodeHeader("Café"))
	// Output:
	// Cofee
	// =?UTF-8?Q?Caf=C3=A9?=
}

func ExampleNewHeaderEncoder() {
	e, err := NewHeaderEncoder("UTF-8", B)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf(e.EncodeHeader("Caf\xc3"))
	// Output: =?UTF-8?B?Q2Fmww==?=
}

func ExampleDecodeHeader() {
	// text is not encoded in UTF-8 but in ISO-8859-1
	text, charset, err := DecodeHeader("=?ISO-8859-1?Q?Caf=C3?=")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("Text: %q, charset: %q", text, charset)
	// Output: Text: "Caf\xc3", charset: "ISO-8859-1"
}

func TestNewHeaderEncoder(t *testing.T) {
	_, err := NewHeaderEncoder("UTF-8", "A")
	if err == nil {
		t.Error(`NewHeaderEncoder("UTF-8", "A") should return an error`)
	}
}

func TestEncodeHeader(t *testing.T) {
	utf8, iso88591 := "UTF-8", "iso-8859-1"
	tests := []struct {
		charset, encoding, src, exp string
	}{
		{utf8, Q, "François-Jérôme", "=?UTF-8?Q?Fran=C3=A7ois-J=C3=A9r=C3=B4me?="},
		{utf8, B, "André", "=?UTF-8?B?QW5kcsOp?="},
		{iso88591, Q, "Rapha\xebl Dupont", "=?iso-8859-1?Q?Rapha=EBl_Dupont?="},
		{utf8, Q, "A", "A"},
		{utf8, Q, "An 'encoded-word' may not be more than 75 characters long, including 'charset', 'encoding', 'encoded-text', and delimiters. ©", "=?UTF-8?Q?An_'encoded-word'_may_not_be_more_than_75_characters_long,_incl?=\r\n =?UTF-8?Q?uding_'charset',_'encoding',_'encoded-text',_and_delimiters._?=\r\n =?UTF-8?Q?=C2=A9?="},
		{utf8, Q, strings.Repeat("0", 62) + "é", "=?UTF-8?Q?" + strings.Repeat("0", 62) + "?=\r\n =?UTF-8?Q?=C3=A9?="},
		{utf8, B, strings.Repeat("é", 23), "=?UTF-8?B?w6nDqcOpw6nDqcOpw6nDqcOpw6nDqcOpw6nDqcOpw6nDqcOpw6nDqcOpw6k=?=\r\n =?UTF-8?B?w6k=?="},
	}

	for _, test := range tests {
		e, err := NewHeaderEncoder(test.charset, test.encoding)
		if err != nil {
			t.Errorf("NewHeaderEncoder(%q, %q) = error %v, want %v", test.charset, test.encoding, err, error(nil))
		} else if s := e.EncodeHeader(test.src); s != test.exp {
			t.Errorf("EncodeHeader(%q) = %q, want %q", test.src, s, test.exp)
		}
	}
}

func TestDecodeHeader(t *testing.T) {
	tests := []struct {
		src, exp, charset string
		isError           bool
	}{
		{"=?UTF-8?Q?Fran=C3=A7ois-J=C3=A9r=C3=B4me?=", "François-Jérôme", "UTF-8", false},
		{"=?UTF-8?q?ascii?=", "ascii", "UTF-8", false},
		{"=?utf-8?B?QW5kcsOp?=", "André", "utf-8", false},
		{"=?ISO-8859-1?Q?Rapha=EBl_Dupont?=", "Rapha\xebl Dupont", "ISO-8859-1", false},
		{"Jean", "Jean", "", false},
		{"=?UTF-8?A?Test?=", "=?UTF-8?A?Test?=", "", false},
		{"=?UTF-8?Q?A=B?=", "=?UTF-8?Q?A=B?=", "", false},
		{"=?UTF-8?Q?=A?=", "=?UTF-8?Q?=A?=", "", false},
		{"=?UTF-8?A?A?=", "=?UTF-8?A?A?=", "", false},
		// Tests from RFC 2047
		{"=?ISO-8859-1?Q?a?=", "a", "ISO-8859-1", false},
		{"=?ISO-8859-1?Q?a?= b", "a b", "ISO-8859-1", false},
		{"=?ISO-8859-1?Q?a?= =?ISO-8859-1?Q?b?=", "ab", "ISO-8859-1", false},
		{"=?ISO-8859-1?Q?a?=  =?ISO-8859-1?Q?b?=", "ab", "ISO-8859-1", false},
		{"=?ISO-8859-1?Q?a?= \r\n\t =?ISO-8859-1?Q?b?=", "ab", "ISO-8859-1", false},
		{"=?ISO-8859-1?Q?a_b?=", "a b", "ISO-8859-1", false},
		{"=?ISO-8859-1?Q?a?= =?ISO-8859-2?Q?_b?=", "", "", true},
	}

	for _, test := range tests {
		s, charset, err := DecodeHeader(test.src)
		if test.isError && err == nil {
			t.Errorf("DecodeHeader(%q) should return an error", test.src)
		}
		if !test.isError && err != nil {
			t.Errorf("DecodeHeader(%q) = error %v, want %v", test.src, err, error(nil))
		}
		if s != test.exp || charset != test.charset {
			t.Errorf("DecodeHeader(%q) = %q (charset=%q), want %q (charset=%q)", test.src, s, charset, test.exp, test.charset)
		}
	}
}
