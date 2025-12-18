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
)

var (
	binaryPath  = flag.String("binary", "", "Path to Python executable")
	packageName = flag.String("package", "main", "Go package name")
	outputFile  = flag.String("output", "", "Output file (default: stdout)")
	modulePath  = flag.String("module", "github.com/younseoryu/gorunpy/gorunpy", "GoRunPy module path")
)

type FunctionInfo struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters"`
	ReturnType string            `json:"return_type"`
}

type ParamInfo struct {
	Name, GoName, GoType string
}

func main() {
	flag.Parse()
	if *binaryPath == "" {
		fmt.Fprintln(os.Stderr, "Error: -binary required")
		os.Exit(1)
	}

	functions, err := introspect(*binaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var publicFuncs []FunctionInfo
	for _, f := range functions {
		if !strings.HasPrefix(f.Name, "_") {
			publicFuncs = append(publicFuncs, f)
		}
	}

	code, _ := generateCode(*packageName, *modulePath, publicFuncs)

	if *outputFile != "" {
		os.WriteFile(*outputFile, []byte(code), 0644)
		fmt.Fprintf(os.Stderr, "Generated %s\n", *outputFile)
	} else {
		fmt.Print(code)
	}
}

func introspect(binary string) ([]FunctionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqJSON, _ := json.Marshal(map[string]any{"function": "__introspect__", "args": map[string]any{}})
	cmd := exec.CommandContext(ctx, binary)
	cmd.Stdin = bytes.NewReader(reqJSON)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %s", err, stderr.String())
	}

	var resp struct {
		OK     bool `json:"ok"`
		Result struct {
			Value struct {
				Functions []FunctionInfo `json:"functions"`
			} `json:"value"`
		} `json:"result"`
	}
	json.Unmarshal(stdout.Bytes(), &resp)
	return resp.Result.Value.Functions, nil
}

func generateCode(pkg, mod string, funcs []FunctionInfo) (string, error) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"goName":    toGoName,
		"goType":    pyTypeToGo,
		"zeroValue": goZero,
		"getParams": getParams,
	}).Parse(tmplStr))

	var buf bytes.Buffer
	tmpl.Execute(&buf, map[string]any{"Package": pkg, "Module": mod, "Functions": funcs})
	formatted, _ := format.Source(buf.Bytes())
	return string(formatted), nil
}

func toGoName(s string) string {
	var b strings.Builder
	for _, p := range strings.Split(s, "_") {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return b.String()
}

func pyTypeToGo(t string) string {
	switch t {
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
	}
	if strings.HasPrefix(t, "List[") {
		return "[]" + pyTypeToGo(t[5:len(t)-1])
	}
	if strings.HasPrefix(t, "Dict[") {
		return "map[string]" + pyTypeToGo(strings.Split(t[5:len(t)-1], ", ")[1])
	}
	if strings.HasPrefix(t, "Optional[") {
		return "*" + pyTypeToGo(t[9:len(t)-1])
	}
	return "any"
}

func goZero(t string) string {
	switch pyTypeToGo(t) {
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
	}
	return "nil"
}

func getParams(f FunctionInfo) []ParamInfo {
	var ps []ParamInfo
	for n, t := range f.Parameters {
		gn := strings.ToLower(n[:1]) + n[1:]
		ps = append(ps, ParamInfo{n, gn, pyTypeToGo(t)})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].Name < ps[j].Name })
	return ps
}

const tmplStr = `package {{.Package}}

import (
	"context"
	"{{.Module}}"
)

type Client struct {
	*gorunpy.Client
}

func NewClient(binaryPath string) *Client {
	return &Client{Client: gorunpy.NewClient(binaryPath)}
}
{{range .Functions}}{{$ps := getParams .}}{{$ret := goType .ReturnType}}{{$zero := zeroValue .ReturnType}}
func (c *Client) {{goName .Name}}(ctx context.Context{{range $ps}}, {{.GoName}} {{.GoType}}{{end}}) ({{if $ret}}{{$ret}}, {{end}}error) {
	args := map[string]any{ {{range $ps}}"{{.Name}}": {{.GoName}},{{end}} }
{{if $ret}}	var result {{$ret}}
	if err := c.Call(ctx, "{{.Name}}", args, &result); err != nil {
		return {{$zero}}, err
	}
	return result, nil{{else}}	return c.Call(ctx, "{{.Name}}", args, nil){{end}}
}
{{end}}`
