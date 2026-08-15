// Package schema describes callable surfaces above CIR.
package schema

import (
	"fmt"

	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
	"gopkg.in/yaml.v3"
)

// Schema describes a service: functions, objects, events, and streams.
type Schema struct {
	Service   string              `json:"service" yaml:"service"`
	Version   string              `json:"version,omitempty" yaml:"version,omitempty"`
	Inferred  bool                `json:"inferred,omitempty" yaml:"inferred,omitempty"`
	Functions map[string]Function `json:"functions,omitempty" yaml:"functions,omitempty"`
	Objects   map[string]Object   `json:"objects,omitempty" yaml:"objects,omitempty"`
	Events    map[string]Event    `json:"events,omitempty" yaml:"events,omitempty"`
	Streams   map[string]Stream   `json:"streams,omitempty" yaml:"streams,omitempty"`
	Errors    map[string]ErrorDef `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// Function is a callable entry point.
type Function struct {
	Args    map[string]TypeRef `json:"args,omitempty" yaml:"args,omitempty"`
	Returns TypeRef            `json:"returns,omitempty" yaml:"returns,omitempty"`
	Doc     string             `json:"doc,omitempty" yaml:"doc,omitempty"`
}

// Object describes a constructable foreign type.
type Object struct {
	Constructors map[string]Function `json:"constructors,omitempty" yaml:"constructors,omitempty"`
	Methods      map[string]Function `json:"methods,omitempty" yaml:"methods,omitempty"`
	Properties   map[string]TypeRef  `json:"properties,omitempty" yaml:"properties,omitempty"`
}

// Event is a named pub/sub surface.
type Event struct {
	Payload TypeRef `json:"payload,omitempty" yaml:"payload,omitempty"`
}

// Stream describes a streaming endpoint.
type Stream struct {
	Item TypeRef `json:"item,omitempty" yaml:"item,omitempty"`
}

// ErrorDef names a structured error.
type ErrorDef struct {
	Code    string `json:"code" yaml:"code"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// TypeRef names a CIR kind or a named object type.
type TypeRef struct {
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
	Nullable bool   `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Elem     string `json:"elem,omitempty" yaml:"elem,omitempty"`
}

// UnmarshalYAML allows args to be written as either a string kind or a mapping.
func (t *TypeRef) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		t.Type = value.Value
		return nil
	}
	type raw TypeRef
	var r raw
	if err := value.Decode(&r); err != nil {
		return err
	}
	*t = TypeRef(r)
	return nil
}

// ParseYAML loads a schema document.
func ParseYAML(data []byte) (*Schema, error) {
	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, czerr.Wrap(czerr.ErrSchemaInvalid, "parse yaml", err)
	}
	if s.Service == "" {
		return nil, czerr.New(czerr.ErrSchemaInvalid, "service name is required")
	}
	return &s, nil
}

// CIRType converts a type reference into a CIR validation type.
func (t TypeRef) CIRType() *cir.Type {
	if t.Type == "" {
		return nil
	}
	k, err := cir.ParseKind(t.Type)
	if err != nil {
		return &cir.Type{Name: t.Type, Nullable: t.Nullable}
	}
	out := &cir.Type{Kind: k, Nullable: t.Nullable, Name: t.Type}
	if t.Elem != "" {
		ek, eerr := cir.ParseKind(t.Elem)
		if eerr == nil {
			out.Elem = &cir.Type{Kind: ek}
		}
	}
	return out
}

// FunctionOf returns a function definition if present.
func (s *Schema) FunctionOf(name string) (Function, bool) {
	if s == nil {
		return Function{}, false
	}
	fn, ok := s.Functions[name]
	return fn, ok
}

// ValidateCall checks argument names and types when a schema is present.
func (s *Schema) ValidateCall(name string, args map[string]cir.Value) error {
	if s == nil {
		return nil
	}
	fn, ok := s.Functions[name]
	if !ok {
		if s.Inferred {
			return nil
		}
		return czerr.New(czerr.ErrSchemaMismatch, fmt.Sprintf("unknown function %q", name))
	}
	if s.Inferred {
		return nil
	}
	for argName, typ := range fn.Args {
		val, present := args[argName]
		if !present {
			if typ.Nullable {
				continue
			}
			return czerr.New(czerr.ErrSchemaMismatch, fmt.Sprintf("missing argument %q", argName))
		}
		if err := cir.Validate(val, typ.CIRType()); err != nil {
			return czerr.Wrap(czerr.ErrSchemaMismatch, "argument "+argName, err)
		}
	}
	return nil
}
