package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOk(t *testing.T) {
	input := `{"j":{"abc": false}}`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, input, result)
}

func TestDict1(t *testing.T) {
	input := `{"j":{"abc`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"j":{"abc<FIXED>":"<FIXED>"}}`, result)
}

func TestBraces1(t *testing.T) {
	input := `{"j":{"a": null, "b": ["1", {"b": 3}, "2"`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"j":{"a":null,"b":["1",{"b":3},"2"]}}`, result)
}

func TestFalse(t *testing.T) {
	input := `{"j":{"abc": f`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"j":{"abc": false}}`, result)
}

func TestString1(t *testing.T) {
	input := `{"j":{"abc": "pez`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"j":{"abc":"pez<FIXED>"}}`, result)
}

func TestString2(t *testing.T) {
	input := `{"j":{"ab": "c", "d`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"j":{"ab":"c","d<FIXED>":"<FIXED>"}}`, result)
}

func TestString3(t *testing.T) {
	input := `{"j":["a","b`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"j":["a","b<FIXED>"]}`, result)
}

func TestString4(t *testing.T) {
	input := `{"j":{"\`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"j":{"\\<FIXED>":"<FIXED>"}}`, result)
}

func TestString5(t *testing.T) {
	input := `{"a": "c", "j": "\u12`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"a":"c","j":"\u0000<FIXED>"}`, result)
}

func TestBrokenArrayOnComma(t *testing.T) {
	input := `{"j":["a",`
	result, err := parseAndFix(strings.NewReader(input))
	require.NoError(t, err, "parseAndFix")
	require.JSONEq(t, `{"j":["a"]}`, result)
}

func TestWrongBrokenArrayOnComma(t *testing.T) {
	input := `{"a": 1, ["b",`
	_, err := parseAndFix(strings.NewReader(input))
	require.Error(t, err, "parseAndFix")
	require.EqualError(t, err, `invalid character '[' looking for beginning of object key string at offset 9`)
}
