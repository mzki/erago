//go:build ignore
// +build ignore

package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/util/errutil"
)

// This file generates list of callback funtions by
// parsing ast of files in this package.
//
// First, user defines documentation tag for the callback function like:
//
// // +scene :scenename
// // documentation for scene...
// const (
//	 // +function: {{.FuncName}}(arg)
//   // annotation text...
//	 Constant1 = "callback_name1"
//
//	 // +function: ret = {{.FuncName}}(number)
//   // annotation text...
//	 Constant2 = "callback_name2"
// )
//
// It genenrates callback list as:
//
//  # scenename
//
//    ![scenename_flow](scenename_flow.png))
//
//    documentation for scene...
//
//	## callback_name1(arg)
//
//    annotation text...
//
//	## ret = callback_name2(number)
//
//    annotation text...

type sceneDeclMap map[string]*sceneDecl

type sceneDecl struct {
	Name string
	Doc  []string

	callbacks callbacks
}

func newSceneDecl(name string) *sceneDecl {
	return &sceneDecl{
		Name:      name,
		Doc:       make([]string, 0, 4),
		callbacks: make(callbacks, 0, 4),
	}
}

func (decl sceneDecl) Callbacks() callbacks {
	return decl.callbacks
}

type callbacks []funcDecl

type funcDecl struct {
	// Suppose Scheme: ret = {{.Name}}(args) and Name: print,
	// FuncDecl generates complete definition of the funcion:
	//	 ret = print(args)

	Template string // template for the funcion definition
	Name     string // function name
	Doc      []string
}

const NamePlaceHolder = "{{.Name}}"

func (decl funcDecl) Definition() string {
	return strings.Replace(decl.Template, NamePlaceHolder, decl.Name, 1)
}

var (
	signaturePatternRet   = regexp.MustCompile(`(.*) = (.*)\((.*)?\)`)
	signaturePatternNoRet = regexp.MustCompile(`(.*)\((.*)?\)`)
)

func (decl funcDecl) findDefinitionParts() (retPart, funcPart, argPart string) {
	def := decl.Definition()
	if match := signaturePatternRet.FindStringSubmatch(def); match != nil {
		retPart, funcPart, argPart = match[1], match[2], match[3]
	} else if match := signaturePatternNoRet.FindStringSubmatch(def); match != nil {
		funcPart, argPart = match[1], match[2]
	}
	return
}

func (decl funcDecl) ArgList() []ParamDecl {
	_, _, argPart := decl.findDefinitionParts()
	params := make([]ParamDecl, 0, 2)
	for _, arg := range trimSplit(argPart, ",") {
		if len(arg) > 0 {
			params = append(params, parseNameAndType(arg))
		}
	}
	return params
}

func (decl funcDecl) RetList() []ParamDecl {
	retPart, _, _ := decl.findDefinitionParts()
	params := make([]ParamDecl, 0, 2)
	for _, arg := range trimSplit(retPart, ",") {
		if len(arg) > 0 {
			params = append(params, parseNameAndType(arg))
		}
	}
	return params
}

func trimSplit(s, sep string) []string {
	ss := strings.Split(s, sep)
	for i, field := range ss {
		ss[i] = strings.TrimSpace(field)
	}
	return ss
}

type ParamDecl struct {
	Name string
	Type string
}

func parseNameAndType(param string) ParamDecl {
	if len(param) == 0 {
		return ParamDecl{}
	}

	ss := trimSplit(param, ":")
	var name, typ string
	if l := len(ss); l >= 2 {
		name = ss[0]
		typ = ss[1]
	} else if l == 1 {
		name = ss[0]
		typ = "any"
	} else {
		name = "unknown"
		typ = "any"
	}
	return ParamDecl{Name: name, Type: typ}
}

func joinParamFunc(params []ParamDecl, fn func(p ParamDecl) string) string {
	ss := make([]string, 0, len(params))
	for _, p := range params {
		ss = append(ss, fn(p))
	}
	return strings.Join(ss, ", ")
}

func joinParamNames(params []ParamDecl) (names string) {
	return joinParamFunc(params, func(p ParamDecl) string { return p.Name })
}

func joinParamTypes(params []ParamDecl) (types string) {
	return joinParamFunc(params, func(p ParamDecl) string { return p.Type })
}

func joinParams(params []ParamDecl) (nameAndTypes string) {
	return joinParamFunc(params, func(p ParamDecl) string { return p.Name + ": " + p.Type })
}

// ==================== MAIN ====================

func main() {
	var outputDir string
	flag.StringVar(&outputDir, "outputdir", "./", "output directory for generated documents")
	flag.Parse()

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		panic("Can not create output direcotory: " + err.Error())
	}

	if err := ParseAST("./", outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func ParseAST(dir string, outputDir string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	callbacks := parseCallBacksFromAST(pkgs["scene"]) // NOTE: use package name directly

	// create sorted key list to fix output order.
	keys := make([]string, 0, len(callbacks))
	for k, _ := range callbacks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	err = checkNameConvention(callbacks, keys)
	if err != nil {
		return err
	}

	err = writeAsMarkdown(filepath.Join(outputDir, "erago-system-callbacks.md"), callbacks, keys)
	if err != nil {
		return err
	}
	err = writeAsLuaLSAddon(outputDir, callbacks, keys)
	return err
}

const SceneTag = "// +scene:"

func parseCallBacksFromAST(pkg *ast.Package) sceneDeclMap {
	callback_map := make(sceneDeclMap)

	for _, f := range pkg.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch decl := n.(type) {
			case *ast.GenDecl:
				if decl.Tok != token.CONST {
					break
				}

				doc := decl.Doc
				if doc == nil || len(doc.List) == 0 {
					break
				}

				tag := doc.List[0].Text
				if !strings.Contains(tag, SceneTag) {
					break
				}

				sceneName := strings.TrimPrefix(tag, SceneTag)
				sceneName = strings.TrimSpace(sceneName)
				if len(sceneName) == 0 {
					// TODO: notify error with declaration line.
				}
				if _, has := callback_map[sceneName]; !has {
					callback_map[sceneName] = newSceneDecl(sceneName)
				}

				callback_map[sceneName].Doc = parseDoc(doc.List[1:])
				callback_map[sceneName].callbacks = addCallBacksFromSpecs(callback_map[sceneName].callbacks, decl.Specs)
			}
			return true
		})
	}
	return callback_map
}

func parseDoc(comments []*ast.Comment) []string {
	doc := make([]string, 0, len(comments))
	for _, com := range comments {
		line := strings.TrimPrefix(com.Text, "//")
		line = strings.TrimSpace(line)
		doc = append(doc, line)
	}
	return doc
}

func addCallBacksFromSpecs(cs callbacks, specs []ast.Spec) callbacks {
	for _, spec := range specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		func_decl, err := parseCallbackDoc(vs.Doc)
		if err != nil {
			log.Println(vs.Names[0].NamePos, err)
			continue
		}

		func_name := vs.Values[0].(*ast.BasicLit).Value
		func_name = strings.Trim(func_name, "\"")

		func_decl.Name = func_name
		cs = append(cs, func_decl)
	}
	return cs
}

const (
	CallbackFuncTag = "// +callback:"
)

func parseCallbackDoc(comments *ast.CommentGroup) (funcDecl, error) {
	if comments == nil {
		return funcDecl{}, errors.New("documentation is not found")
	}

	firstLine := comments.List[0].Text
	if ok := strings.HasPrefix(firstLine, CallbackFuncTag); !ok {
		return funcDecl{}, errors.New(CallbackFuncTag + " is not found")
	}

	definition := strings.TrimSpace(strings.TrimPrefix(firstLine, CallbackFuncTag))
	return funcDecl{
		Template: definition,
		Doc:      parseDoc(comments.List[1:]),
	}, nil
}

func checkNameConvention(callbacks_list sceneDeclMap, keys []string) error {
	multiErr := errutil.NewMultiError()

	for _, sceneName := range keys {
		sceneDecl, ok := callbacks_list[sceneName]
		if !ok {
			continue
		}

		functions := sceneDecl.callbacks

		for _, f := range functions {
			var nameElems = strings.Split(f.Name, scene.ScrSep)

			if len(nameElems) < 2 {
				multiErr.Add(fmt.Errorf("%s: %q must starts [scene-name]_[callback-type]", sceneName, f.Name))
				continue // to next functions
			}

			// validate each elements
			var errTxt string = ""

			{
				if !strings.HasPrefix(nameElems[0], sceneName) {
					errTxt += fmt.Sprintf("%q should starts with %s; ", f.Name, sceneName)
				}

				var hasTyp = false
				for _, typ := range []string{
					scene.ScrEventPrefix,
					scene.ScrScenePrefix,
					scene.ScrReplacePrefix,
					scene.ScrUserPrefix,
				} {
					if strings.HasPrefix(nameElems[1], typ) {
						hasTyp = true
					}
				}
				if !hasTyp {
					errTxt += fmt.Sprintf("%q should contains any of callback type name at the 2nd place; ", f.Name)
				}
			}

			if len(errTxt) > 0 {
				multiErr.Add(fmt.Errorf("%s: %s", sceneName, errTxt))
			}
		}

	}

	return multiErr.Err()
}

const (
	DocIndentSpace = 2
)

func writeAsMarkdown(file string, callbacks_list sceneDeclMap, keys []string) error {
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	var revision string = "unknown"
	if err := exec.Command("git", "--help").Run(); err != nil {
		fmt.Fprintln(os.Stderr, "git command not found")
	} else {
		// parse revision from `git describe`
		cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
		out, err := cmd.Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "can not runs git rev-parse")
		} else {
			revision = strings.TrimSpace(string(out))
		}
	}

	// write final output

	fmt.Fprintf(fp, "**Version %s**\n", revision)
	fmt.Fprintln(fp, "")

	fmt.Fprintln(fp, `Generated by gen_callback_doc.go, by parsing pakcage erago/scene.

These callback functions must be prefixed "era." in the script file.
So callback function "event_title()" is defined "era.event_title()" in script file.

The naming convention for the callback function name is like [scene-name]_[callback-type]_[funcion-name].
e.g. title_event_start means, on the titile scene, start event is fired.

以下のcallback関数は、era.XXXという形式で、スクリプトファイル内に定義します。
例えば、"title_event()"という関数があったとき、スクリプト上では、"era.title_event()"と
いうように、定義します。
callback関数は、命名規則 [scene-name]_[callback-type]_[funcion-name] に従います。
例） title_event_start は、title シーンで、開始イベントが発生したことを示します。

`)

	indent := strings.Repeat(" ", DocIndentSpace)

	for _, scene := range keys {
		sceneDecl, ok := callbacks_list[scene]
		if !ok {
			continue
		}

		sceneImageName := scene + "_flow"

		fmt.Fprintln(fp, "# "+scene+"\n")
		fmt.Fprintln(fp, "!["+sceneImageName+"](images/"+sceneImageName+".png)\n")
		for _, line := range sceneDecl.Doc {
			fmt.Fprintln(fp, indent+line)
		}
		fmt.Fprintln(fp, "")

		functions := sceneDecl.callbacks
		functions = append(makeDefaultCallback(scene), functions...)
		for _, f := range functions {
			fmt.Fprintln(fp, "## "+f.Definition())
			fmt.Fprintln(fp, "")
			for _, line := range f.Doc {
				fmt.Fprintln(fp, indent+line)
			}
			fmt.Fprintln(fp, "")
		}
		fmt.Fprintln(fp, "")
	}
	return nil
}

func makeDefaultCallback(scene_name string) callbacks {
	scene_decl := funcDecl{
		Template: "{{.Name}}()",
		Name:     scene_name + "_scene",
		Doc: strings.Split(`この関数は、もし定義されていれば、シーンの最も始めに呼ばれ、
シーン全体の処理を置き換えます。
この関数内では、必ず次のシーンを指定しなければならないことに注意してください。`, "\n"),
	}

	event_decl := funcDecl{
		Template: "{{.Name}}()",
		Name:     scene_name + "_event_start",
		Doc:      strings.Split(`この関数は、もし定義されていれば、シーンの始めに呼ばれます。`, "\n"),
	}

	return callbacks{scene_decl, event_decl}
}

// --- LuaLS meta definitions ---------------------------------------------------------------------

var luaLSLuaArgEscapeList = []struct {
	name    string
	pattern *regexp.Regexp
	repr    string
}{
	{name: "arbitrary-arg", pattern: regexp.MustCompile(`.*(\.\.\.).*`), repr: `$1`},
	{name: "optional-arg", pattern: regexp.MustCompile(`(\[?)([^\[\]]+)(\]?)`), repr: `$2`},
	{name: "default-arg", pattern: regexp.MustCompile(`([^ ]+)[ ]*=[ ]*([^ ]+)`), repr: `$1`},
	{name: "string-literal", pattern: regexp.MustCompile(`"([^"]+)"`), repr: `str_literal`},
}

var luaLSMetaTmpl = template.Must(
	template.
		New("LuaLS-Meta").
		Funcs(template.FuncMap{
			"joinArgNames": func(params []ParamDecl) string {
				return joinParamNames(params)
			},
			"joinRetTypes": func(params []ParamDecl) string {
				s := joinParamTypes(params)
				if len(s) == 0 {
					return "nil"
				} else {
					return s
				}
			},
			"joinArgs": func(params []ParamDecl) string {
				return joinParams(params)
			},
		}).
		Parse(`
{{- define "CALLBACK_DEFINITION" -}}
{{- with $funcDecl := . }}
---{{ .Definition }}
{{ range $idx, $comment := $funcDecl.Doc -}}
---{{ $comment }}
{{ end -}}
{{- $argList := $funcDecl.ArgList -}}
{{- range $i, $arg := $argList -}}
{{- if gt (len $arg.Name) 0 -}}
---@param {{$arg.Name}} {{$arg.Type}}
{{ else -}}
{{- end -}}
{{ end -}}
{{- $retList := $funcDecl.RetList -}}
{{- range $i, $ret := $retList -}}
{{- if gt (len $ret.Name) 0 -}}
---@return {{$ret.Type}} {{$ret.Name}}
{{ else -}}
{{- end -}}
{{ end -}}
---@type fun({{joinArgs $argList}}): {{joinRetTypes $retList}}
era.{{$funcDecl.Name}} = nil
{{end -}}{{- /* end with func */ -}}
{{- end -}}

{{- /* ------------------- Content Body -------------------------- */ -}}
---@meta erago-callback
		
--generated by gendoc.go, parsing pakcage infra/script@{{ .Version }}.
--DO NOT EDIT MANUALLY.

era = {}
{{ $sceneDeclList := .SceneDeclList -}}
{{ range $sceneIdx, $sceneDecl := $sceneDeclList -}}
{{ range $funcIdx, $funcDecl := $sceneDecl.Callbacks -}}
{{ template "CALLBACK_DEFINITION" $funcDecl }}
{{ end -}}
{{ end -}}
`))

const luaLSConfigJSONContent = `
{
    "name": "Erago Callbacks",

    "settings": {
		"Lua.runtime.version": "Lua 5.1",

        "Lua.diagnostics.globals" : [
            "era",
        ]
    }
}
`

func getVersionStr() string {
	var revision string = "unknown"
	if err := exec.Command("git", "--help").Run(); err != nil {
		fmt.Fprintln(os.Stderr, "git command not found")
	} else {
		// parse revision from `git describe`
		cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
		out, err := cmd.Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "can not runs git rev-parse")
		} else {
			revision = strings.TrimSpace(string(out))
		}
	}
	return revision
}

func writeAsLuaLSAddon(outputDir string, callbacks_list sceneDeclMap, sceneOrder []string) error {
	// See Addon folder structure https://github.com/LuaLS/lua-language-server/wiki/Addons#manually-enabling
	addonDir := filepath.Join(outputDir, "addons", "erago-lua-callback")
	libraryDir := filepath.Join(addonDir, "library")
	if err := os.MkdirAll(libraryDir, os.ModePerm); err != nil {
		return err
	}

	{
		configFile := filepath.Join(addonDir, "config.json")
		if err := os.WriteFile(configFile, []byte(luaLSConfigJSONContent), os.ModePerm); err != nil {
			return err
		}
	}

	metaFile := filepath.Join(libraryDir, "erago-callback.lua")
	fp, err := os.Create(metaFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	var sceceDeclList []*sceneDecl
	for _, s := range sceneOrder {
		if sd, ok := callbacks_list[s]; ok {
			sceceDeclList = append(sceceDeclList, sd)
		}
	}

	// append default callbacks for each scene.
	for i, ss := range sceceDeclList {
		ss.callbacks = append(makeDefaultCallback(ss.Name), ss.callbacks...)
		sceceDeclList[i] = ss
	}

	templateContext := struct {
		Version       string
		SceneDeclList []*sceneDecl
	}{
		Version:       getVersionStr(),
		SceneDeclList: sceceDeclList,
	}
	err = luaLSMetaTmpl.Execute(fp, &templateContext)
	return err
}
