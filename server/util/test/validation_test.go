package test

import (
	"strings"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestValidPasswords(t *testing.T) {
	inputs := []string{
		"Abc!1234",
		"Abcdefg!1",
		"aB!aaa1a",
		"Ä#üöaa1aaa",
		"Sea!surf1ng",
	}
	for _, input := range inputs {
		CheckTestBool(t, ValidatePassword(input), true)
	}
}

func TestInvalidPasswords(t *testing.T) {
	inputs := []string{
		"SHORT1!",
		"alllowercase!",
		"ALLUPPERCASE!",
		"NoSpecialChar1",
		"",
	}
	for _, input := range inputs {
		CheckTestBool(t, ValidatePassword(input), false)
	}
}

func TestValidNames(t *testing.T) {
	inputs := []string{
		"John",
		"Jane Doe",
		"O'Brien",
		"Anne-Marie",
		"St. John",
		"Müller",
		"José",
		"山田太郎",
		"Ångström Lab",
		"Room 42",
	}
	for _, input := range inputs {
		CheckTestBool(t, IsValidName(input), true)
	}
}

func TestInvalidNames(t *testing.T) {
	inputs := []string{
		"",
		"A",
		strings.Repeat("A", 65),
		"Invalid<Name>",
		"Name@Domain",
		"Name\nWithNewline",
		"Name\tWithTab",
		"Name!Exclamation",
		"Name#Hash",
	}
	for _, input := range inputs {
		CheckTestBool(t, IsValidName(input), false)
	}
}
