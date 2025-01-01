# Demonstrate Unrecognised jet.Ranger Interface (NOW FIXED)

This repo demonstrates a problem with the recognition of rangeable interface in the jet template engine. The template execution context includes a rangeable map, which values can be also rangeable maps. When using `github.com/CloudyKit/jet` the template execution works well. When using `github.com/CloudyKit/jet/v6`, only the first-level map is recognised as rangeable. The second-level is not.

**UPDATE**: The problem was fixed by @sauerbraten in [#221](https://github.com/CloudyKit/jet/pull/221). Thank you very much! The latest commits in this repo upgraded the dependency on `github.com/CloudyKit/jet/v6@master` and verified that both `success` and `failure` binaries would produce the same correct output. Thanks again for the very quick fix!

## Testing

Build the testing binaries:

    ❯ make build
    cd success && go build success.go
    cd failure && go build failure.go

Check the successful one that depends on `github.com/CloudyKit/jet`:

    ❯ ./success/success
      Key Name:
        Scalar: Barbarian
      Key Description:
        Scalar: A fierce warrior of primitive background who can enter a battle rage
      Key Proficiencies:
          Key Armor:
            Array: [Light armor Medium armor Shields]
          Key Weapons:
            Array: [Simple weapons Martial weapons]

Check the failing one that depends on `github.com/CloudyKit/jet/v6`:

    ❯ ./failure/failure
    2024/12/12 18:19:37 Executing template failed: Jet Runtime Error ("/template.jet":1):
    value {[{Name Barbarian} {Description A fierce warrior of primitive background who can
    enter a battle rage} {Proficiencies [{Armor [Light armor Medium armor Shields]}
    {Weapons [Simple weapons Martial weapons]}]}] 3 0} (type interface {}) is not rangeable

The failure happens when trying to process the value of `Proficiencies` as a range.

## Input Data

Input is a YAML object, where the value of the `Proficiencies` property is another object:

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

The failing example doesn't see the `Proficiencies` object as rangeable.

## Rangeable Interface

Objects are unmarshaled by `github.com/goccy/go-yaml` to a ordered map `yaml.MapSlice` and wrapped in `mapWrapper` to pass them to the template engine:

    type mapWrapper struct {
      entries yaml.MapSlice
      len     int
      i       int
    }

    func (m *mapWrapper) Range() (reflect.Value, reflect.Value, bool)
    func (m *mapWrapper) ProvidesIndex() bool

The first-level object can be enumerated well. The second-level object cannot.

The `Range` enumerator returns the map values wrapped to `valueWrapper` to be able to recognise scalars, arrays and maps:

    type valueWrapper struct {
      Value   interface{}
      IsArray bool
      IsMap   bool
    }

The `Value` can be `string`, []interface{} or `mapWrapper`.

The first-level `mapWrapper` object is passed to the template execution as `Metadata` in the the context:

    type templateContext struct {
	    Metadata interface{}
    }

The context preparation and template execution:

    metadata := yaml.MapSlice{}
    if err := yaml.UnmarshalWithOptions(content, &metadata, yaml.UseOrderedMap()); err != nil {
      ...
    }
    wrapper := &mapWrapper{
      entries: metadata,
      len:     len(metadata),
    }
    context := &templateContext{
      Metadata: wrapper,
    }
    var buf bytes.Buffer
    if err := template.Execute(&buf, nil, context); err != nil {
      ...
    }

## Template

The template iterates ove the first level object and if there is a map on the second level, it iterates through it too:

    {{ range key, val := .Metadata }}
      Key {{ key }}:
      {{ if val.IsMap }}
        {{ range k, v := val.Value }}
          Key {{ k }}:
          {{ if v.IsMap }}
            Map: {{ v.Value }}
          {{ else if v.IsArray }}
            Array: {{ v.Value }}
          {{ else }}
            Scalar: {{ v.Value }}
          {{ end }}
        {{ end }}
      {{ else if val.IsArray }}
        Array: {{ val.Value }}
      {{ else }}
        Scalar: {{ val.Value }}
      {{ end }}
    {{ end }}

Although the first-level `.Metadata` and the second-level `v.Value` are moth `mapWrapper` instances, only the first-level one is recognised as rangeable.

Copyright (C) 2024 Ferdinand Prantl

Licensed under the [MIT License].

[MIT License]: http://en.wikipedia.org/wiki/MIT_License
