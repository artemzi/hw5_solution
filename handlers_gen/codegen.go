package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

var (
	tplHeader = template.Must(template.New("tplHeader").Parse(`package {{.Package}}

import (
	"net/http"
	"encoding/json"
	"fmt"
)
`))

	tpl = template.Must(template.New("tpl").Parse(`
// ServeHTTP comment was here
func (api *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})
	switch r.URL.Path {
	{{.Cases}}
	default:
		w.WriteHeader(http.StatusNotFound)
		resp["error"] = "unknown method"
		body, _ := json.Marshal(resp)
		w.Write(body)
	}
}
`))

	caseTemplate = template.Must(template.New("caseTemplate").Parse(`
	case "{{.Path}}":
		api.handler{{.Handler}}(w, r)
`))

	handlerTemplate = template.Must(template.New("handlerTemplate").Parse(`
func (api *{{.StructName}}) handler{{.Method}}(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data := r.URL.Query()
		fmt.Printf("{{.StructName}} GET %v\n", data)
	case "POST":
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		data := r.Form
		fmt.Printf("{{.StructName}} POST %v\n", data)
	}
}
`))
)

type apigen struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type caseTpl struct {
	Path    string
	Handler string
	Method  string
	Struct  string
}

type structTpl struct {
	StructName string
	Package    string
	Cases      string
	Method     string
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	var (
		structName *ast.Ident
		cases      []caseTpl
		structures []string
	)
	output := new(bytes.Buffer)

	dest := strings.ToLower(os.Args[2]) // here we can change filename or add some check&whatever
	out, _ := os.Create(dest)

	tplHeader.Execute(out, structTpl{Package: node.Name.Name}) // add package declarations and imports
	for _, n := range node.Decls {
		currFunc, ok := n.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if currFunc.Doc != nil {
			apigen := &apigen{Method: "GET" /*Default value*/}
			for _, i := range currFunc.Recv.List {
				structName = i.Type.(*ast.StarExpr).X.(*ast.Ident)

				for _, doc := range currFunc.Doc.List {
					if strings.Contains(doc.Text, "apigen:api") {
						start := strings.Index(doc.Text, "{")
						end := strings.Index(doc.Text, "}") + 1
						json.Unmarshal([]byte(doc.Text[start:end]), apigen)
						cases = append(cases, caseTpl{apigen.URL, currFunc.Name.Name, apigen.Method, structName.Name})
						if len(structures) > 0 {
							for _, i := range structures {
								if i == structName.Name {
									break
								}
								structures = append(structures, structName.Name)
							}
						} else {
							structures = append(structures, structName.Name)
						}
					}
				}
			}
		}
	}

	for _, s := range structures {
		for _, c := range cases {
			if c.Struct == s {
				caseTemplate.Execute(output, c)
				handlerTemplate.Execute(out, structTpl{StructName: s, Method: c.Handler})
			}
		}
		tpl.Execute(out, structTpl{StructName: s, Cases: output.String()})
		output = new(bytes.Buffer)
	}
	out.Close()
}

// go build handlers_gen/* && ./codegen api.go api_handlers.go

// идём по всем декларациям
// ищем структуры
// проверяем коммент на налииче нужнйо метки
// генерим метод если метка нашлась
