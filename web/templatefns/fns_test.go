package templatefns

import (
	"bytes"
	"html/template"
	"testing"
)

func AssertEqual(t *testing.T, buffer *bytes.Buffer, testString string) {
	if buffer.String() != testString {
		t.Errorf("Expected %s, got %s", testString, buffer.String())
	}
	buffer.Reset()
}

func ParseTest(buffer *bytes.Buffer, body string, data interface{}) {
	tpl := template.New("test").Funcs(TemplateFnsMap)
	tpl.Parse(body)
	tpl.Execute(buffer, data)
}

func TestGtfFuncMap(t *testing.T) {
	var buffer bytes.Buffer

	ParseTest(&buffer, "{{ 21 | divisibleby 3 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ 21 | divisibleby 4 }}", "")
	AssertEqual(t, &buffer, "false")

	ParseTest(&buffer, "{{ 3.0 | divisibleby 3 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ 3.0 | divisibleby 1.5 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ . | divisibleby 1.5 }}", uint(300))
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ 12 | divisibleby . }}", uint(3))
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ 21 | divisibleby 4 }}", "")
	AssertEqual(t, &buffer, "false")

	ParseTest(&buffer, "{{ false | divisibleby 3 }}", "")
	AssertEqual(t, &buffer, "false")

	ParseTest(&buffer, "{{ 3 | divisibleby false }}", "")
	AssertEqual(t, &buffer, "false")

}
