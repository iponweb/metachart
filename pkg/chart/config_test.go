package chart

import (
	"encoding/json"
	"testing"
)

func TestResourceDefinitionDefaults(t *testing.T) {
	var definition ResourceDefinition
	if err := json.Unmarshal([]byte(`{"apiVersion":"v1","kind":"ConfigMap"}`), &definition); err != nil {
		t.Fatal(err)
	}
	if !definition.Template || !definition.Root || !definition.Defaults {
		t.Errorf("template, root and defaults must default to true, got %+v", definition)
	}
	if definition.Global {
		t.Errorf("global must default to false, got %+v", definition)
	}
}

func TestResourceDefinitionGlobal(t *testing.T) {
	var definition ResourceDefinition
	if err := json.Unmarshal([]byte(`{"template":false,"defaults":false,"global":true}`), &definition); err != nil {
		t.Fatal(err)
	}
	if definition.Template || definition.Defaults || !definition.Root || !definition.Global {
		t.Errorf("unexpected flags: %+v", definition)
	}
}
