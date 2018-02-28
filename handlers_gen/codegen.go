package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"strings"
)

var (
	tplHeader = template.Must(template.New("tplHeader").Parse(`package {{.Package}}

import "net/http"
`))

	tpl = template.Must(template.New("tpl").Parse(`
// {{.FieldName}} comment was here
func (api *{{.StructName}}) {{.FieldName}}(http.ResponseWriter, *http.Request) {
	//TODO
}
`))
)

func loopFunc(currFunc *ast.FuncDecl, structures []string) []string {
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

	var structures []string
	vals := map[string]string{
		"FieldName":  "ServeHTTP",
		"StructName": "",
		"Package":    node.Name.Name,
	}

	dest := strings.ToLower(os.Args[2]) // here we can change filename or add some check&whatever
	out, _ := os.Create(dest)

	for _, n := range node.Decls {
		currFunc, ok := n.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if currFunc.Doc != nil {
			structures = loopFunc(currFunc, structures)
		}
	}

	tplHeader.Execute(out, vals)   // add package decklarations and imports
	for _, s := range structures { // TODO
		vals["StructName"] = s
		tpl.Execute(out, vals)
	}
	out.Close()
}

// go build handlers_gen/* && ./codegen api.go api_handlers.go

// идём по всем декларациям
// ищем структуры
// проверяем коммент на налииче нужнйо метки
// генерим метод если метка нашлась
