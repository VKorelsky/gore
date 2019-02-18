package main

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	printerPkgs = []struct {
		path string
		code string
	}{
		{"fmt", `fmt.Printf("%#v\n", x)`},
	}
}

func TestRun_import(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	s, err := NewSession(stdout, stderr)
	defer s.Clear()
	require.NoError(t, err)

	codes := []string{
		":import encoding/json",
		"b, err := json.Marshal(nil)",
		"string(b)",
	}

	for _, code := range codes {
		err := s.Eval(code)
		require.NoError(t, err)
	}

	require.Equal(t, `[]byte{0x6e, 0x75, 0x6c, 0x6c}
<nil>
"null"
`, stdout.String())
	require.Equal(t, "", stderr.String())
}

func TestRun_QuickFix_evaluated_but_not_used(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	s, err := NewSession(stdout, stderr)
	defer s.Clear()
	require.NoError(t, err)

	codes := []string{
		`[]byte("")`,
		`make([]int, 0)`,
		`1+1`,
		`func() {}`,
		`(4 & (1 << 1))`,
		`1`,
	}

	for _, code := range codes {
		err := s.Eval(code)
		require.NoError(t, err)
	}

	r := regexp.MustCompile(`0x[0-9a-f]+`)
	require.Equal(t, `[]byte{}
[]int{}
2
(func())(...)
0
1
`, r.ReplaceAllString(stdout.String(), "..."))
	require.Equal(t, "", stderr.String())
}

func TestRun_QuickFix_used_as_value(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	s, err := NewSession(stdout, stderr)
	defer s.Clear()
	require.NoError(t, err)

	codes := []string{
		`:import log`,
		`a := 1`,
		`log.SetPrefix("")`,
	}

	for _, code := range codes {
		err := s.Eval(code)
		require.NoError(t, err)
	}

	require.Equal(t, `1
`, stdout.String())
	require.Equal(t, "", stderr.String())
}

func TestRun_FixImports(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	s, err := NewSession(stdout, stderr)
	defer s.Clear()
	require.NoError(t, err)

	autoimport := true
	flagAutoImport = &autoimport

	codes := []string{
		`filepath.Join("a", "b")`,
	}

	for _, code := range codes {
		err := s.Eval(code)
		require.NoError(t, err)
	}

	require.Equal(t, `"a/b"
`, stdout.String())
	require.Equal(t, "", stderr.String())
}

func TestIncludePackage(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	s, err := NewSession(stdout, stderr)
	defer s.Clear()
	require.NoError(t, err)

	err = s.includePackage("github.com/motemen/gore/gocode")
	require.NoError(t, err)

	err = s.Eval("Completer{}")
	require.NoError(t, err)
}

func TestRun_Copy(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	s, err := NewSession(stdout, stderr)
	defer s.Clear()
	require.NoError(t, err)

	codes := []string{
		`a := []string{"hello", "world"}`,
		`b := []string{"goodbye", "world"}`,
		`copy(a, b)`,
		`if (a[0] != "goodbye") {
			panic("should be copied")
		}`,
	}

	for _, code := range codes {
		err := s.Eval(code)
		require.NoError(t, err)
	}

	require.Equal(t, `[]string{"hello", "world"}
[]string{"goodbye", "world"}
2
`, stdout.String())
	require.Equal(t, "", stderr.String())
}

func TestRun_Const(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	s, err := NewSession(stdout, stderr)
	defer s.Clear()
	require.NoError(t, err)

	codes := []string{
		`const ( a = iota; b )`,
		`a`,
		`b`,
	}

	for _, code := range codes {
		err := s.Eval(code)
		require.NoError(t, err)
	}

	require.Equal(t, `0
1
`, stdout.String())
	require.Equal(t, "", stderr.String())
}

func TestRun_Error(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	s, err := NewSession(stdout, stderr)
	defer s.Clear()
	require.NoError(t, err)

	codes := []string{
		`foo`,
		`len(100)`,
	}

	for _, code := range codes {
		err := s.Eval(code)
		require.Error(t, err)
	}

	require.Equal(t, "", stdout.String())
	require.Equal(t, `undefined: foo
invalid argument 100 (type int) for len
`, stderr.String())
}
