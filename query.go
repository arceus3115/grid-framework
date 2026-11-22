package grid

import (
	"context"
	"fmt"
	"strings"
)

// Arg represents a typed GraphQL argument bound to a variable in a query.
type Arg struct {
	Name        string
	GraphQLType string
	Value       any
}

// QuerySpec describes a single root GraphQL query field and its typed arguments.
type QuerySpec struct {
	OperationName string
	FieldName     string
	Args          []Arg
}

// BuildQuery constructs a parameterized GraphQL query string and its variables map.
// It is a pure function with no side effects.
func BuildQuery(spec QuerySpec, selection Selection) (string, map[string]any, error) {
	if strings.TrimSpace(spec.FieldName) == "" {
		return "", nil, fmt.Errorf("grid query: field name is required")
	}

	var varDecls []string
	var fieldArgs []string
	var variables map[string]any

	for _, arg := range spec.Args {
		if arg.Name == "" || arg.GraphQLType == "" {
			continue
		}
		if variables == nil {
			variables = make(map[string]any, len(spec.Args))
		}
		varDecls = append(varDecls, fmt.Sprintf("$%s: %s", arg.Name, arg.GraphQLType))
		fieldArgs = append(fieldArgs, fmt.Sprintf("%s: $%s", arg.Name, arg.Name))
		variables[arg.Name] = arg.Value
	}

	var b strings.Builder
	b.WriteString("query")
	if op := strings.TrimSpace(spec.OperationName); op != "" {
		b.WriteString(" ")
		b.WriteString(op)
	}
	if len(varDecls) > 0 {
		b.WriteString("(")
		b.WriteString(strings.Join(varDecls, ", "))
		b.WriteString(")")
	}
	b.WriteString(" {\n  ")
	b.WriteString(spec.FieldName)
	if len(fieldArgs) > 0 {
		b.WriteString("(")
		b.WriteString(strings.Join(fieldArgs, ", "))
		b.WriteString(")")
	}

	if !selection.IsEmpty() {
		b.WriteString(" {\n")
		for _, line := range strings.Split(selection.Render(), "\n") {
			b.WriteString("    ")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("  }")
	}
	b.WriteString("\n}")

	return b.String(), variables, nil
}

// Execute runs a QuerySpec against the specified GRID API using the provided modular Form[T].
func Execute[T any](
	ctx context.Context,
	client *Client,
	api API,
	spec QuerySpec,
	form Form[T],
) (T, error) {
	var zero T
	if client == nil {
		return zero, fmt.Errorf("grid execute: client is nil")
	}

	query, variables, err := BuildQuery(spec, form.Selection)
	if err != nil {
		return zero, err
	}

	data, err := client.DoRaw(ctx, api, spec.OperationName, query, variables)
	if err != nil {
		return zero, err
	}

	raw, ok := data[spec.FieldName]
	if !ok {
		return zero, fmt.Errorf("grid execute: response missing field %q", spec.FieldName)
	}

	return form.decodeRaw(raw)
}

// String returns a pointer to v.
func String(v string) *string { return &v }

// Int returns a pointer to v.
func Int(v int) *int { return &v }

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// Float64 returns a pointer to v.
func Float64(v float64) *float64 { return &v }

func appendPageArgs(args []Arg, p PageParams) []Arg {
	if p.First != nil {
		args = append(args, Arg{Name: "first", GraphQLType: "Int", Value: *p.First})
	}
	if p.After != nil && *p.After != "" {
		args = append(args, Arg{Name: "after", GraphQLType: "String", Value: *p.After})
	}
	if p.Last != nil {
		args = append(args, Arg{Name: "last", GraphQLType: "Int", Value: *p.Last})
	}
	if p.Before != nil && *p.Before != "" {
		args = append(args, Arg{Name: "before", GraphQLType: "String", Value: *p.Before})
	}
	return args
}
