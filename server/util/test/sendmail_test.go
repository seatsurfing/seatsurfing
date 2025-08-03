package test

import (
	"path/filepath"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestGetEmailTemplatePathExists(t *testing.T) {
	res, err := GetEmailTemplatePath(GetEmailTemplatePathResetpassword(), "de")
	CheckStringNotEmpty(t, res)
	CheckTestBool(t, true, err == nil)
}

func TestGetEmailTemplatePathFallback(t *testing.T) {
	res, err := GetEmailTemplatePath(GetEmailTemplatePathResetpassword(), "notexists")
	CheckStringNotEmpty(t, res)
	CheckTestBool(t, true, err == nil)
}

func TestGetEmailTemplatePathNotExists(t *testing.T) {
	path, _ := filepath.Abs("./res/notexisting.txt")
	res, err := GetEmailTemplatePath(path, "en")
	CheckTestString(t, "", res)
	CheckTestBool(t, true, err != nil)
}

func TestGetLocalPartFromEmailAddress(t *testing.T) {
	CheckTestString(t, "test", GetLocalPartFromEmailAddress("test@domain.com"))
	CheckTestString(t, "test", GetLocalPartFromEmailAddress("test@domain"))
	CheckTestString(t, "test", GetLocalPartFromEmailAddress("test"))
	CheckTestString(t, "\"a@b\"", GetLocalPartFromEmailAddress("\"a@b\"@example.com"))
}

func TestReplaceVarsInTemplate(t *testing.T) {
	template := "Hello {{name}}, your code is {{code}}."
	vars := map[string]string{
		"name": "John",
		"code": "123456",
	}
	result := ReplaceVarsInTemplate(template, vars)
	expected := "Hello John, your code is 123456."
	CheckTestString(t, expected, result)

	vars["code"] = ""
	result = ReplaceVarsInTemplate(template, vars)
	expected = "Hello John, your code is ."
	CheckTestString(t, expected, result)

	vars["name"] = ""
	result = ReplaceVarsInTemplate(template, vars)
	expected = "Hello , your code is ."
	CheckTestString(t, expected, result)
}

func TestReplaceVarsInTemplateConditions(t *testing.T) {
	template := "Hello {{name}}, {{if showCode}}your code is {{code}}{{end}}{{if !showCode}}you don't have a code{{end}}."
	vars := map[string]string{
		"name": "John",
		"code": "123456",
	}

	vars["showCode"] = "1"
	result := ReplaceVarsInTemplate(template, vars)
	expected := "Hello John, your code is 123456."
	CheckTestString(t, expected, result)

	vars["showCode"] = "0"
	result = ReplaceVarsInTemplate(template, vars)
	expected = "Hello John, you don't have a code."
	CheckTestString(t, expected, result)
}
