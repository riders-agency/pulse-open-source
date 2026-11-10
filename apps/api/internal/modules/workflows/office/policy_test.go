package office

import (
	"bytes"
	"testing"
)

func TestDemoExecutableGraphCannotAddToolsOrRewireApproval(t *testing.T) {
	def, _ := Definition()
	if err := ValidateDefinition(def); err != nil {
		t.Fatal(err)
	}
	disabled := bytes.ReplaceAll(def, []byte(`"enabled":true`), []byte(`"enabled":false`))
	if err := ValidateDefinition(disabled); err != nil {
		t.Fatal("disabling actions must be allowed", err)
	}
	for _, bad := range [][]byte{
		bytes.Replace(def, []byte(`office.classify`), []byte(`http.request`), 1),
		bytes.Replace(def, []byte(`"to":"approval"`), []byte(`"to":"apply"`), 1),
		bytes.Replace(def, []byte(`"action_key":"website"`), []byte(`"action_key":"classify"`), 1),
	} {
		if ValidateDefinition(bad) == nil {
			t.Fatal("unsafe graph accepted")
		}
	}
}
