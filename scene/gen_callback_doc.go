// +build ignore

package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
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

func main() {
	if err := ParseAST("./"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func ParseAST(dir string) error {
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

	err = writeAsMarkdown("callback_list.md", callbacks, keys)
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

	fmt.Fprintf(fp, "**revision %s**\n", revision)
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
		Template: scene_name + "_scene()",
		Doc: strings.Split(`この関数は、もし定義されていれば、シーンの最も始めに呼ばれ、
シーン全体の処理を置き換えます。
この関数内では、必ず次のシーンを指定しなければならないことに注意してください。`, "\n"),
	}

	event_decl := funcDecl{
		Template: scene_name + "_event_start()",
		Doc:      strings.Split(`この関数は、もし定義されていれば、シーンの始めに呼ばれます。`, "\n"),
	}

	return callbacks{scene_decl, event_decl}
}
