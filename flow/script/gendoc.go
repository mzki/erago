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
	return err
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
	Doc       []string
}

type docContext struct {
	Section string
}

var sectionPattern = regexp.MustCompile(`"(.*)"`)

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

	return docElement{
		GroupName: section,
		Doc:       doc,
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
