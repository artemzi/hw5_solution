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
	// method {{.Method}}
	case "{{.Path}}":
		api.handler{{.Handler}}(w, r)
`))

	httpWrapperTemplate = template.Must(template.New("httpWrapperHeaderTemplate").Parse(`
func (srv *{{.StructName}}) handler{{.FuncName}}(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})
	resp["error"] = ""
	{{.Method}}
	{{.Auth}}
	var v url.Values
	switch r.Method {
	case "POST":
		v = parseCrutchyBody(r.Body)
	default:
		v = r.URL.Query()
	}
	var params {{.FuncName}}Params
	if err := params.validateAndFill{{.FuncName}}Params(v); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp["error"] = err.Error()
		body, _ := json.Marshal(resp)
		w.Write(body)
		return
	}
	user, err := srv.{{.FuncName}}(r.Context(), params)
	if err != nil {
		switch err.(type) {
		case ApiError:
			w.WriteHeader(err.(ApiError).HTTPStatus)
			resp["error"] = err.Error()
		default:
			w.WriteHeader(http.StatusInternalServerError)
			resp["error"] = "bad user"
		}
		body, _ := json.Marshal(resp)
		w.Write(body)
		return
	}
	resp["response"] = user
	body, _ := json.Marshal(resp)
	w.Write(body)
}
`))
)

type MethodSignature struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type caseTpl struct {
	Path      string
	Handler   string
	Structure string
	Method    string
}

type structTpl struct {
	StructName string
	Package    string
	Cases      string
	Method     string
}

func getStructures(currFunc *ast.FuncDecl, structures []string) []string {
	for _, doc := range currFunc.Doc.List {
		if strings.Contains(doc.Text, "apigen:api") {
		LOOP:
			for _, i := range currFunc.Recv.List {
				structName := i.Type.(*ast.StarExpr).X.(*ast.Ident)
				for _, i := range structures {
					if i == structName.Name {
						break LOOP
					}
				}
				structures = append(structures, structName.Name)
			}
		}
	}
	return structures
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	var (
		structures []string
		cases      caseTpl
	)
	output := new(bytes.Buffer)

	dest := strings.ToLower(os.Args[2]) // here we can change filename or add some check&whatever
	out, _ := os.Create(dest)

	for _, n := range node.Decls {
		currFunc, ok := n.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if currFunc.Doc != nil {
			structures = getStructures(currFunc, structures)
			methodSignature := new(MethodSignature)
			for _, doc := range currFunc.Doc.List {
				if strings.Contains(doc.Text, "apigen:api") {
					start := strings.Index(doc.Text, "{")
					end := strings.Index(doc.Text, "}") + 1
					json.Unmarshal([]byte(doc.Text[start:end]), methodSignature)
					cases = caseTpl{methodSignature.Url, currFunc.Name.Name, "", methodSignature.Method}
				}
			}
		}
	}

	tplHeader.Execute(out, structTpl{Package: node.Name.Name}) // add package decklarations and imports
	for _, s := range structures {                             // TODO
		cases.Structure = s
		caseTemplate.Execute(output, cases)
		tpl.Execute(out, structTpl{StructName: s, Cases: output.String()})
		output = new(bytes.Buffer) // clear last assigment
	}
	out.Close()
}

// go build handlers_gen/* && ./codegen api.go api_handlers.go

// идём по всем декларациям
// ищем структуры
// проверяем коммент на налииче нужнйо метки
// генерим метод если метка нашлась
