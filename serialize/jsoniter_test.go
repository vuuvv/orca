package serialize

import "testing"

type jsonTypeA struct {
	Name string `json:"name"`
}

const testJsonText = `
{
	"name": "john"
}
`

func TestJsonParse(t *testing.T) {
	a, err := JsonParse[jsonTypeA](testJsonText)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(a)
}
