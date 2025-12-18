// Command gorunpy-gen generates typed Go client code from Python function signatures.
//
// Usage:
//
//	gorunpy-gen -binary ./dist/myapp -package myapp -output client.go
//
// This will introspect the Python executable and generate typed Go wrappers
// for all exported functions.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var (
	binaryPath  = flag.String("binary", "", "Path to Python executable")
	packageName = flag.String("package", "main", "Go package name")
	outputFile  = flag.String("output", "", "Output file (default: stdout)")
	modulePath  = flag.String("module", "github.com/younseoryu/gorunpy/gorunpy", "GoRunPy module import path")
)

// FunctionInfo represents metadata about an exported Python function.
type FunctionInfo struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters"` // name -> type
	ReturnType string            `json:"return_type"`
}

// ParamInfo for ordered parameters
type ParamInfo struct {
	Name   string
	GoName string
	GoType string
	PyType string
}

// IntrospectResponse is the response from the __introspect__ call.
type IntrospectResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		Value struct {
			Functions []FunctionInfo `json:"functions"`
		} `json:"value"`
	} `json:"result"`
}

func main() {
	flag.Parse()

	if *binaryPath == "" {
		fmt.Fprintln(os.Stderr, "Error: -binary is required")
		flag.Usage()
		os.Exit(1)
	}

	// Introspect the Python executable
	functions, err := introspect(*binaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error introspecting binary: %v\n", err)
		fmt.Fprintln(os.Stderr, "Note: The Python executable must support the __introspect__ function.")
		fmt.Fprintln(os.Stderr, "Make sure you're using gorunpy Python SDK.")
		os.Exit(1)
	}

	// Filter out internal functions
	var publicFunctions []FunctionInfo
	for _, f := range functions {
		if !strings.HasPrefix(f.Name, "_") {
			publicFunctions = append(publicFunctions, f)
		}
	}

	// Generate code
	code, err := generateCode(*packageName, *modulePath, publicFunctions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	// Output
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(code), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Generated %s with %d functions\n", *outputFile, len(publicFunctions))
	} else {
		fmt.Print(code)
	}
}

func introspect(binaryPath string) ([]FunctionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request := map[string]any{
		"function": "__introspect__",
		"args":     map[string]any{},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, binaryPath)
	cmd.Stdin = bytes.NewReader(requestJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run binary: %v\nstderr: %s", err, stderr.String())
	}

	var resp IntrospectResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v\nstdout: %s", err, stdout.String())
	}

	if !resp.OK {
		return nil, fmt.Errorf("introspection failed")
	}

	return resp.Result.Value.Functions, nil
}

func generateCode(pkg, modulePath string, functions []FunctionInfo) (string, error) {
	funcMap := template.FuncMap{
		"goName":       toGoName,
		"goType":       pythonTypeToGo,
		"isSimpleType": isSimpleReturnType,
		"needsPointer": needsPointerReturn,
		"zeroValue":    goZeroValue,
		"getParams":    getOrderedParams,
		"hasParams":    func(f FunctionInfo) bool { return len(f.Parameters) > 0 },
	}

	tmpl := template.Must(template.New("client").Funcs(funcMap).Parse(clientTemplate))

	var buf bytes.Buffer
	data := struct {
		Package    string
		ModulePath string
		Functions  []FunctionInfo
	}{
		Package:    pkg,
		ModulePath: modulePath,
		Functions:  functions,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted if formatting fails (for debugging)
		return buf.String(), nil
	}

	return string(formatted), nil
}

func toGoName(name string) string {
	parts := strings.Split(name, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			result.WriteString(part[1:])
		}
	}
	return result.String()
}

func pythonTypeToGo(pyType string) string {
	pyType = strings.TrimSpace(pyType)

	switch pyType {
	case "int":
		return "int"
	case "float":
		return "float64"
	case "str":
		return "string"
	case "bool":
		return "bool"
	case "None", "NoneType":
		return ""
	case "Any", "any":
		return "any"
	}

	// List[T]
	if strings.HasPrefix(pyType, "List[") || strings.HasPrefix(pyType, "list[") {
		inner := pyType[5 : len(pyType)-1]
		return "[]" + pythonTypeToGo(inner)
	}

	// Dict[str, T]
	if strings.HasPrefix(pyType, "Dict[") || strings.HasPrefix(pyType, "dict[") {
		inner := pyType[5 : len(pyType)-1]
		parts := splitTypeArgs(inner)
		if len(parts) == 2 {
			return "map[" + pythonTypeToGo(parts[0]) + "]" + pythonTypeToGo(parts[1])
		}
		return "map[string]any"
	}

	// Optional[T]
	if strings.HasPrefix(pyType, "Optional[") {
		inner := pyType[9 : len(pyType)-1]
		return "*" + pythonTypeToGo(inner)
	}

	// Union - just use any for now
	if strings.HasPrefix(pyType, "Union[") {
		return "any"
	}

	return "any"
}

func isSimpleReturnType(pyType string) bool {
	goType := pythonTypeToGo(pyType)
	switch goType {
	case "int", "float64", "string", "bool", "any", "":
		return true
	}
	if strings.HasPrefix(goType, "[]") {
		return true
	}
	return false
}

func needsPointerReturn(pyType string) bool {
	goType := pythonTypeToGo(pyType)
	// Complex types like maps and structs should be returned as pointers
	return strings.HasPrefix(goType, "map[")
}

func goZeroValue(pyType string) string {
	goType := pythonTypeToGo(pyType)
	switch goType {
	case "int":
		return "0"
	case "float64":
		return "0"
	case "string":
		return `""`
	case "bool":
		return "false"
	case "":
		return ""
	default:
		return "nil"
	}
}

func getOrderedParams(f FunctionInfo) []ParamInfo {
	var params []ParamInfo
	for name, pyType := range f.Parameters {
		params = append(params, ParamInfo{
			Name:   name,
			GoName: toGoParamName(name),
			GoType: pythonTypeToGo(pyType),
			PyType: pyType,
		})
	}
	// Sort alphabetically for consistent output
	sort.Slice(params, func(i, j int) bool {
		return params[i].Name < params[j].Name
	})
	return params
}

func toGoParamName(name string) string {
	// Convert snake_case to camelCase for parameters
	parts := strings.Split(name, "_")
	var result strings.Builder
	for i, part := range parts {
		if len(part) > 0 {
			if i == 0 {
				result.WriteString(strings.ToLower(part))
			} else {
				result.WriteString(strings.ToUpper(string(part[0])))
				result.WriteString(part[1:])
			}
		}
	}
	s := result.String()
	// Handle reserved words
	if isGoReserved(s) {
		return s + "_"
	}
	return s
}

func isGoReserved(s string) bool {
	reserved := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	return reserved[s]
}

func splitTypeArgs(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, r := range s {
		switch r {
		case '[':
			depth++
			current.WriteRune(r)
		case ']':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			if !unicode.IsSpace(r) || current.Len() > 0 {
				current.WriteRune(r)
			}
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

const clientTemplate = `// Code generated by gorunpy-gen. DO NOT EDIT.

package {{.Package}}

import (
	"context"

	"{{.ModulePath}}"
)

// Client provides typed methods for calling Python functions.
type Client struct {
	*gorunpy.Client
}

// NewClient creates a new Client for the Python executable at binaryPath.
func NewClient(binaryPath string) *Client {
	return &Client{Client: gorunpy.NewClient(binaryPath)}
}

{{range .Functions}}
{{$params := getParams .}}
{{$goName := goName .Name}}
{{$returnType := goType .ReturnType}}
{{$isSimple := isSimpleType .ReturnType}}
{{$needsPtr := needsPointer .ReturnType}}
{{$zero := zeroValue .ReturnType}}
// {{$goName}} calls the Python function "{{.Name}}".
{{if eq $returnType ""}}func (c *Client) {{$goName}}(ctx context.Context{{range $params}}, {{.GoName}} {{.GoType}}{{end}}) error {
{{else if $needsPtr}}func (c *Client) {{$goName}}(ctx context.Context{{range $params}}, {{.GoName}} {{.GoType}}{{end}}) ({{$returnType}}, error) {
{{else}}func (c *Client) {{$goName}}(ctx context.Context{{range $params}}, {{.GoName}} {{.GoType}}{{end}}) ({{$returnType}}, error) {
{{end}}	args := map[string]any{
{{- range $params}}
		"{{.Name}}": {{.GoName}},
{{- end}}
	}
{{if eq $returnType ""}}
	return c.Call(ctx, "{{.Name}}", args, nil)
{{else}}
	var result {{$returnType}}
	if err := c.Call(ctx, "{{.Name}}", args, &result); err != nil {
		return {{$zero}}, err
	}
	return result, nil
{{end}}}
{{end}}
`
