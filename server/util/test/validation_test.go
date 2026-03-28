package test

import (
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
