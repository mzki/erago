// +build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"text/template"
)

// This program collects and lists documentation for scripting era game.
//
// Parsed documentation shceme is:
//	// +gendoc "section_name"
//	// text...
//
//
// Subcommand
//
// if the following comment is appeard:
//   // +gendoc.set_section "section_name"
//
// trailing default section is auto-set by "section_name",
// but each gendoc declaration can have own section by // +gendoc "section_name"
//
// the context of section is valid only in a file scope.
//
// if documentation text starts "* " then it is treated as definition
// then outputs directly.
// otherwise text is treated as plain text
// then outputs with indent 2 space.
//
// Example:
//
// comments above declaration of go function:
//
//   // +gendoc "era_func"
//   // * color = get_color()
//   //
//	 //   get color from system.
//	 //
//	 func getColor() {
//		...
//
// is converts to script documentation:
//
//   # era_func
//
//	 * color = get_color()
//
//		 get color from system.
//
// Another example;
//
// 	 // +gendoc.section "era_func"
// 	 // * set_color(color)
// 	 //
// 	 //   set color to system.
// 	 //
// 	 func setColor() {
// 	 	...
//
// is converts to the documentation:
//
//	 * set_color(color)
//		 set color to system.
//
// and trailing documentation is set default section "era_func".

var sortingOrder = []string{
	"Over View",
	"Era Module",
	"Lua Character",
	"Characters",
	"XXXParam",
	"Builtin Module: bit32",
	"Builtin Module: time",
	"Builtin Module: csv",
	"Builtin Module: log",
	"Constant Value",
}

func main() {
	if err := ParseAST("./"); err != nil {
		panic(err)
	}
}

const parsingPkgName = "script"

func ParseAST(dir string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	docMap := make(map[string][]docElement, 4)
	for _, doc := range parseDocElements(pkgs[parsingPkgName]) {
		docMap[doc.GroupName] = append(docMap[doc.GroupName], doc)
	}
	// manual order. Section names are not auto collected.
	sections := sort_by(docMap, sortingOrder)
	err = writeTxt("document.txt", sections)
	if err != nil {
		return fmt.Errorf("Failed writeTxt(): %w", err)
	}
	err = writeVSCodeSnippet("era-lua.json.code-snippets", sections)
	if err != nil {
		return fmt.Errorf("Failed writeVSCodeSnippet(): %w", err)
	}
	return nil
}

func parseDocElements(pkg *ast.Package) []docElement {
	docElements := make([]docElement, 0, 256)
	docCtx := &docContext{Section: "Others"}
	for _, f := range pkg.Files {
		// NOTE: it handles only file scope comments.
		for _, cgroup := range f.Comments {
			tag_i := findMetaTagIndex(cgroup.List)
			if tag_i == -1 {
				continue
			}
			comments := cgroup.List[tag_i:]
			if docCtx.subCommand(comments[0].Text) {
				continue
			}
			docElements = append(docElements, docCtx.parseDocElement(comments))
		}
		docCtx.reset()
	}
	return docElements
}

const metaTag = "// +gendoc"

// find index of declaration of script document.
func findMetaTagIndex(comments []*ast.Comment) int {
	for i, comment := range comments {
		if strings.HasPrefix(comment.Text, metaTag) {
			return i
		}
	}
	return -1
}

// sort by sectionOrder for docMap and return ordered sequence of docGroup.
func sort_by(docMap map[string][]docElement, sectionOrder []string) []docGroup {
	groups := make([]docGroup, 0, len(docMap))
	touched := make(map[string]struct{}, len(docMap))
	for _, sec := range sectionOrder {
		docs, ok := docMap[sec]
		if !ok {
			fmt.Println("gendoc: unknown section:", sec)
			continue
		}
		touched[sec] = struct{}{}
		groups = append(groups, docGroup{
			Section: docs[0].GroupName,
			List:    docs,
		})
	}

	// check untouched docMap, and push it to groups
	for sec, docs := range docMap {
		if _, ok := touched[sec]; ok {
			continue
		}
		fmt.Println("gendoc: untouched in sorting section:", sec)
		groups = append(groups, docGroup{
			Section: docs[0].GroupName,
			List:    docs,
		})
	}
	return groups
}

// group by section name
type docGroup struct {
	Section string
	List    []docElement
}

// docElement is a each element of raw parsed document
type docElement struct {
	GroupName string
	Doc       []string // raw text

	Signature docSignature
	Comments  []string
}

type docSignature struct {
	RetList  []string
	FuncName string
	ArgList  []string
}

type docContext struct {
	Section string
}

var (
	sectionPattern        = regexp.MustCompile(`"(.*)"`)
	signaturePatternRet   = regexp.MustCompile(`(.*) = (.*)\((.*)?\)`)
	signaturePatternNoRet = regexp.MustCompile(`(.*)\((.*)?\)`)
)

// if header has subcommand then executes it and return true,
// otherwise return false.
func (ctx *docContext) subCommand(header string) bool {
	header = strings.TrimPrefix(header, metaTag)
	if !strings.HasPrefix(header, ".") { // unmatch any subcommand
		return false
	}
	switch {
	// sub command to change section context
	case strings.HasPrefix(header, ".set_section"):
		if match := sectionPattern.FindStringSubmatch(header); match != nil {
			ctx.Section = match[1]
		}
		return true
	default:
		fmt.Println("gendoc: unknown subcommand:", header)
		return false
	}
}

func (ctx *docContext) reset() {
	ctx.Section = "Others"
}

func (ctx *docContext) parseDocElement(comments []*ast.Comment) docElement {
	header := comments[0].Text
	var section string
	if match := sectionPattern.FindStringSubmatch(header); match != nil {
		section = match[1] // manually specified section
	} else {
		section = ctx.Section
	}

	doc_part := comments[1:]
	doc := make([]string, 0, len(doc_part))
	for _, comment := range doc_part {
		line := strings.TrimPrefix(comment.Text, "//")
		line = strings.TrimPrefix(line, " ")
		doc = append(doc, line)
	}

	// parse function signature
	var (
		retList     []string
		funcName    string = ""
		argList     []string
		commentList []string
	)
	trimmedSplit := func(s, sep string) []string {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			return []string{}
		}
		parts := strings.Split(s, sep)
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts
	}
	for _, line := range doc {
		if i := strings.Index(line, "*"); i == -1 {
			commentList = append(commentList, line)
		} else {
			line = line[i+1:]
			var retPart, funcPart, argPart string
			if match := signaturePatternRet.FindStringSubmatch(line); match != nil {
				retPart, funcPart, argPart = match[1], match[2], match[3]
			} else if match = signaturePatternNoRet.FindStringSubmatch(line); match != nil {
				funcPart, argPart = match[1], match[2]
			} else {
				// TODO: error message?
				fmt.Printf("unmatched to function signature: %v\n", line)
				continue
			}
			retList = trimmedSplit(retPart, ",")
			funcName = strings.TrimSpace(funcPart)
			argList = trimmedSplit(argPart, ",")
		}
	}

	return docElement{
		GroupName: section,
		Doc:       doc,
		Signature: docSignature{
			RetList:  retList,
			FuncName: funcName,
			ArgList:  argList,
		},
		Comments: commentList,
	}
}

func writeTxt(file string, docGroups []docGroup) error {
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	// print header
	for _, s := range []string{
		"generated by gendoc.go, parsing pakcage script.",
		"",
		"",
	} {
		fmt.Fprintln(fp, s)
	}

	const indent = "  " // 2 space to indent
	newLine := func() { fmt.Fprintln(fp, "") }

	// print section list
	fmt.Fprintln(fp, "## Section List")
	newLine()
	for _, group := range docGroups {
		fmt.Fprintln(fp, "*", group.Section)
	}
	newLine()
	newLine()

	// print content
	for _, group := range docGroups {
		fmt.Fprintln(fp, "##", group.Section)
		newLine()

		for _, doc := range group.List {
			for _, line := range doc.Doc {
				if !strings.HasPrefix(line, "* ") { // line not starting by "* " is indented
					line = indent + line
				}
				fmt.Fprintln(fp, line)
			}
			newLine()
			newLine()
		}
		newLine()
	}
	return nil
}

/* Create below for each docElement
"era.print": {
	"scope": "lua",
	"prefix": "era.print",
	"body": [
		"era.print(\"$1\")",
		"$2"
	],
	"description": [
		"Print text at screen.",
		"",
		"This is second lines..."
	],
},
*/
var vscodeSnippetTmpl = template.Must(
	template.
		New("vscode-snippets").
		Funcs(template.FuncMap{
			"custom_add":       func(a, b int) int { return a + b },
			"custom_is_last":   func(i int, list []string) bool { return i+1 >= len(list) },
			"custom_is_last_n": func(i, length int) bool { return i+1 >= length },
			"custom_esc": func(s string) string { return strings.Replace(s, `"`, `\"`, -1) },
		}).
		Parse(`
{{- define "ARG_LIST_BODY" -}}
  {{$args := . -}}
  {{range $idx, $arg := $args -}}
    ${ {{- custom_add $idx 1}}:{{$arg -}} }{{ if custom_is_last $idx $args }}{{else}}, {{end}}
  {{- end }}
{{- end -}}

{{- define "ARG_LIST" -}}
  {{$args := . -}}
  {{range $idx, $arg := $args -}}
    {{$arg}}{{ if custom_is_last $idx $args }}{{else}}, {{end}}
  {{- end }}
{{- end -}}

{{- define "RET_LIST" -}}
  {{$rets := . -}}
  {{- if gt (len $rets) 0 -}}
    {{template "ARG_LIST" $rets}}
  {{- else -}}
    nil
  {{- end -}}
{{ end -}}

{
	// generated by gendoc.go, parsing pakcage infra/script.
	// DO NOT EDIT MANUALLY.
	{{- $modules := . }}
	{{- range $moduleIdx, $mod := $modules }}
	{{- range $funcIdx, $func := $mod.List }}
	{{- with $sig := $func.Signature }}
	"{{$sig.FuncName}}": {
		"scope": "lua",
		"prefix": "{{$sig.FuncName}}",
		"body": [
			"{{$sig.FuncName}}({{template "ARG_LIST_BODY" $sig.ArgList}})",
			"${{custom_add (len $sig.ArgList) 1}}"
		],
		"description": [
			"{{$sig.FuncName}}({{template "ARG_LIST" $sig.ArgList}}) -> {{template "RET_LIST" $sig.RetList}}",
			{{- range $idx, $comment := $func.Comments }}
			"{{custom_esc $comment}}"{{if custom_is_last $idx $func.Comments}}{{else}},{{end}}
			{{- end }}
		]
	}
	  {{- if and (custom_is_last_n $moduleIdx (len $modules)) (custom_is_last_n $funcIdx (len $mod.List)) -}}
	  {{else -}}
	  ,
	  {{ end -}}
	{{end -}}{{- /* end with */ -}}
	{{end -}}
	{{end -}}
}
`))

func writeVSCodeSnippet(file string, docGroups []docGroup) error {
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	var docGs []docGroup
	for _, docG := range docGroups {
		newDocG := docG
		elems := make([]docElement, 0, len(docG.List))
		for _, elem := range docG.List {
			if len(elem.Signature.FuncName) > 0 {
				elems = append(elems, elem)
			}
		}
		newDocG.List = elems
		docGs = append(docGs, newDocG)
	}
	err = vscodeSnippetTmpl.Execute(fp, &docGs)
	return err
}
