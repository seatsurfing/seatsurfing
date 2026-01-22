package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestEscapeStringsInStruct(t *testing.T) {
	type TestStruct struct {
		Name    string
		Comment string
		Age     int
	}

	input := TestStruct{
		Name:    `<script>alert("XSS")</script>`,
		Comment: `Hello & welcome to <b>Go</b>!`,
		Age:     30,
	}

	expected := TestStruct{
		Name:    `&lt;script&gt;alert(&#34;XSS&#34;)&lt;/script&gt;`,
		Comment: `Hello &amp; welcome to &lt;b&gt;Go&lt;/b&gt;!`,
		Age:     30,
	}

	err := EscapeStringsInStruct(&input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if input != expected {
		t.Errorf("expected %+v, got %+v", expected, input)
	}
}
