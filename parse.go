package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

func parseAndFix(r io.Reader) (string, error) {
	var buf bytes.Buffer
	decoder := json.NewDecoder(io.TeeReader(r, &buf))

	var stat = &parseStat{}
	var output strings.Builder // This will store the reconstructed JSON
	//var lastToken json.Token
	var failOffset int64

	// Read tokens until the end or an error occurs
	for {
		// Read each token (key, value, brackets, etc.)
		token, err := decoder.Token()
		if err != nil {
			failOffset = decoder.InputOffset()
			if _, err := io.Copy(io.Discard, r); err != nil {
				fmt.Println("Error discarding input:", err)
			}
			stat.err = err
			stat.failOffset = failOffset
			break
		}

		switch t := token.(type) {
		case json.Delim:
			switch t {
			case '{':
				if len(stat.braces) > 0 {
					if !stat.isFirst && (stat.IsArray() || stat.currentKey == "") {
						output.WriteString(",")
					}
				}

				stat.braces = append(stat.braces, '{')
				stat.isFirst = true
			case '}':
				lastIdx := len(stat.braces) - 1
				if stat.braces[lastIdx] != '{' {
					return "", fmt.Errorf("unexpected '}' at offset %d", decoder.InputOffset())
				}
				stat.braces = stat.braces[:lastIdx]
			case '[':
				if !stat.isFirst && (stat.IsArray() || stat.currentKey == "") {
					output.WriteString(",")
				}
				stat.braces = append(stat.braces, '[')
				stat.isFirst = true
			case ']':
				lastIdx := len(stat.braces) - 1
				if stat.braces[lastIdx] != '[' {
					return "", fmt.Errorf("unexpected ']' at offset %d", decoder.InputOffset())
				}
				stat.braces = stat.braces[:lastIdx]
			default:
				panic("Unknown delimiter: " + string(t))
			}
			output.WriteRune(rune(t))
			stat.currentKey = ""

		case string:
			// This is either a key or a string value
			if stat.currentKey == "" && !stat.IsArray() {
				stat.writeValue(&output, fmt.Sprintf("\"%s\":", t))
				stat.currentKey = t
			} else {
				stat.writeValue(&output, fmt.Sprintf("\"%s\"", t))
			}
		case float64:
			stat.writeValue(&output, fmt.Sprintf("%v", t))
		case bool:
			stat.writeValue(&output, fmt.Sprintf("%t", t))
		case nil:
			stat.writeValue(&output, "null")
		}
	}
	if stat.err == nil {
		// impossible situation?
		return "", errors.New("expected error, at least EOF")
	}
	if len(stat.braces) == 0 {
		if !errors.Is(stat.err, io.EOF) {
			return "", stat.err
		}
		return output.String(), nil
	}
	var synErr *json.SyntaxError
	if errors.As(stat.err, &synErr) {
		return "", stat.errorContext
	}

	restoreJSON(buf.Bytes()[failOffset:], *stat, &output)
	return output.String(), nil
}

func restoreJSON(rest []byte, stat parseStat, output *strings.Builder) {
	if len(rest) > 0 {
		var result = string(rest)
		result = strings.TrimSuffix(result, "\n")
		var (
			brokenUnicode = regexp.MustCompile(`\\u[0-9]{1,3}$`)
			boolTrueRe    = regexp.MustCompile(`^tr?u?$`)
			boolFalseRe   = regexp.MustCompile(`^fa?l?s?$`)
		)

		switch {
		// unfinished string
		case rest[0] == '"':
			if !stat.isFirst && (stat.currentKey == "" || stat.IsArray()) {
				output.WriteString(",")
			}

			result = brokenUnicode.ReplaceAllString(result, `\u0000`)
			output.WriteString(result)

			if result[len(result)-1] == '\\' {
				output.WriteString(`\`)
			}

			output.WriteString(`<FIXED>"`)
			if stat.currentKey == "" && !stat.IsArray() {
				output.WriteString(`:"<FIXED>"`)
			}
		default:
			result = boolTrueRe.ReplaceAllString(result, "true")
			result = boolFalseRe.ReplaceAllString(result, "false")
			output.WriteString(result)
		}
	}

	for i := len(stat.braces) - 1; i >= 0; i-- {
		switch stat.braces[i] {
		case '{':
			output.WriteString("}")
		case '[':
			output.WriteString("]")
		}
	}
}

type errorContext struct {
	err        error
	failOffset int64
}

func (ec errorContext) Error() string {
	return fmt.Sprintf("%s at offset %d", ec.err, ec.failOffset)
}

type parseStat struct {
	isFirst    bool
	braces     []rune
	currentKey string
	errorContext
}

func (stat *parseStat) writeValue(b *strings.Builder, s string) {
	if !stat.isFirst && (stat.IsArray() || stat.currentKey == "") {
		b.WriteString(",")
	}
	b.WriteString(s)
	stat.currentKey = ""
	stat.isFirst = false
}

func (stat *parseStat) IsArray() bool {
	return stat.braces[len(stat.braces)-1] == '['
}
