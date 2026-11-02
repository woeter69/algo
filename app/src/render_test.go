package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestAllTemplatesParse(t *testing.T) {
	initTemplates()

	files, err := os.ReadDir("../templates")
	if err != nil {
		t.Fatalf("Failed to read templates: %v", err)
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".html") {
			tmpl, err := templateSet.FromFile(f.Name())
			if err != nil {
				t.Errorf("Template %s failed to parse: %v", f.Name(), err)
			} else {
				var buf bytes.Buffer
				ctx := pongo2.Context{
					"url_for": func(endpoint string, args ...interface{}) string {
						return "/" + endpoint
					},
					"session": SessionDict{values: map[interface{}]interface{}{"user_id": 1, "username": "test"}},
				}
				_ = tmpl.ExecuteWriter(ctx, &buf)
				t.Logf("✅ Template %s parsed and rendered successfully!", f.Name())
			}
		}
	}
}
