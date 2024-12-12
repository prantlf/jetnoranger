package main

import (
	"bytes"
	"log"
	"reflect"

	"github.com/CloudyKit/jet"
	yaml "github.com/goccy/go-yaml"
)

// wrapper for yaml.MapSlice which implements jet.Ranger
type mapWrapper struct {
	entries yaml.MapSlice
	len     int
	i       int
}

// wrapper for values in yaml.MapSlice returned to the map range
type valueWrapper struct {
	Value   interface{}
	IsArray bool
	IsMap   bool
}

// context object for the template execution
type templateContext struct {
	Metadata interface{}
}

// implementing jet.Ranger for mapWrapper

func (m *mapWrapper) Range() (reflect.Value, reflect.Value, bool) {
	entries := m.entries
	for m.i < m.len {
		entry := &entries[m.i]
		key := entry.Key
		m.i++
		var mv *valueWrapper
		v := entry.Value
		if ms, ok := v.(yaml.MapSlice); ok {
			mv = &valueWrapper{
				Value: &mapWrapper{
					entries: ms,
					len:     len(ms),
				},
				IsMap: true,
			}
		} else if _, ok := v.([]interface{}); ok {
			mv = &valueWrapper{
				Value:   v,
				IsArray: true,
			}
		} else {
			mv = &valueWrapper{
				Value: v,
			}
		}
		return reflect.ValueOf(key), reflect.ValueOf(mv), false
	}
	return reflect.Value{}, reflect.Value{}, true
}

func (_ *mapWrapper) ProvidesIndex() bool {
	return false
}

// load the template
func getTemplate() *jet.Template {
	loader := jet.NewOSFileSystemLoader("")
	templateSet := jet.NewHTMLSetLoader(loader)
	templateSet.SetDevelopmentMode(true)

	t, err := templateSet.GetTemplate("template.jet")
	if err != nil {
		log.Fatalf("Getting template failed: %v", err)
	}
	return t
}

// parse input text to yaml.MapSlice
func getContext() interface{} {
	content := []byte(`
Name: Barbarian
Description: A fierce warrior of primitive background who can enter a battle rage
Proficiencies:
  Armor:
    - Light armor
    - Medium armor
    - Shields
  Weapons:
    - Simple weapons
    - Martial weapons
`)
	metadata := yaml.MapSlice{}
	if err := yaml.UnmarshalWithOptions(content, &metadata, yaml.UseOrderedMap()); err != nil {
		log.Fatalf("Unmarshaling metadata failed: %v", err)
	}

	wrapper := &mapWrapper{
		entries: metadata,
		len:     len(metadata),
	}
	context := &templateContext{
		Metadata: wrapper,
	}
	return context
}

// execute the template
func executeTemplate(template *jet.Template, context interface{}) string {
	var buf bytes.Buffer
	if err := template.Execute(&buf, nil, context); err != nil {
		log.Fatalf("Executing template failed: %v", err)
	}
	return buf.String()
}

// test the template execution
func main() {
	template := getTemplate()
	context := getContext()
	result := executeTemplate(template, context)
	println(result)
}
