package test

import (
	"strings"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestValidEmails(t *testing.T) {
	inputs := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.org",
		"user_name@sub.domain.de",
		"u@example.io",
		"confluence-user@company.com",
	}
	for _, input := range inputs {
		CheckTestBool(t, ValidateEmail(input), true)
	}
}

func TestInvalidEmails(t *testing.T) {
	inputs := []string{
		"",
		"notanemail",
		"missing@tld",
		"@nodomain.com",
		"no space@example.com",
		"double@@example.com",
		"user@",
		strings.Repeat("a", 250) + "@example.com",
	}
	for _, input := range inputs {
		CheckTestBool(t, ValidateEmail(input), false)
	}
}

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

func TestValidHumanNames(t *testing.T) {
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
		CheckTestBool(t, IsValidHumanName(input), true)
	}
}

func TestInvalidHumanNames(t *testing.T) {
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
		CheckTestBool(t, IsValidHumanName(input), false)
	}
}

func TestValidOrgNames(t *testing.T) {
	inputs := []string{
		"Acme Corp",
		"Müller & Söhne",
		"Tech+Media GmbH",
		"Company (Int'l)",
		"Startup #42",
		"user@company",
		"Best-Org!",
		"Division_A",
		"Ångström Lab",
	}
	for _, input := range inputs {
		CheckTestBool(t, IsValidOrgName(input), true)
	}
}

func TestInvalidOrgNames(t *testing.T) {
	inputs := []string{
		"",
		"A",
		strings.Repeat("A", 65),
		"Name\nWithNewline",
		"Name\tWithTab",
	}
	for _, input := range inputs {
		CheckTestBool(t, IsValidOrgName(input), false)
	}
}
