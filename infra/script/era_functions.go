package script

import (
	"context"
	"time"

	lua "github.com/yuin/gopher-lua"

	attr "github.com/mzki/erago/attribute"
	"github.com/mzki/erago/scene"
	"github.com/mzki/erago/state"
	"github.com/mzki/erago/width"
)

const (
	// era functions are included in this module.
	EraModuleName = "era"

	// Redundancy for era module.
	eraModuleRegistryName = "_ERA"

	eraLayoutModName = "layout"
	eraFlowModName   = "flow"
)

// protect metatable by set it to table.__metatable.
const metaProtectObj = lua.LString("protected")

func (ip *Interpreter) registerEraModule(L *lua.LState, gamestate *state.GameState, game GameController) *lua.LTable {
	if era_mod, ok := L.GetGlobal(EraModuleName).(*lua.LTable); ok {
		return era_mod // already exist
	}

	era_module := L.NewTable()
	L.SetGlobal(EraModuleName, era_module)
	L.SetGlobal(eraModuleRegistryName, era_module)

	ft := &functor{game, gamestate}
	eraModFuncMap := map[string]lua.LGFunction{
		// TODO: move to module
		// state
		"clearSystem": ft.clearSystem,
		"saveSystem":  ft.saveSystem,
		"loadSystem":  ft.loadSystem,
		"clearShare":  ft.clearShare,
		"saveShare":   ft.saveShare,
		"loadShare":   ft.loadShare,
		// util
		"paramlv": ft.paramLv,
		"explv":   ft.expLv,
		// alignment functions
		"setAlignment": ft.setAlignment,
		"getAlignment": ft.getAlignment,

		// color functions
		"setColor":   ft.setColor,
		"getColor":   ft.getColor,
		"resetColor": ft.resetColor,

		// output functions
		"print":            ft.print,
		"printl":           ft.printL,
		"printc":           ft.printC,
		"printLine":        ft.printLine,
		"printBar":         ft.printBar,
		"textBar":          ft.textBar,
		"printButton":      ft.printButton,
		"printPlain":       ft.printPlain,
		"printImage":       ft.printImage,
		"measureImageSize": ft.measureImageSize,
		"newPage":          ft.newPage,
		"clearLineAll":     ft.clearLineAll,
		"clearLine":        ft.clearLine,
		"windowStrWidth":   ft.windowStrWidth,
		"windowLineCount":  ft.windowLineCount,
		"currentStrWidth":  ft.currentStrWidth,
		"lineCount":        ft.lineCount,
		"textWidth":        ft.textWidth,
		// "lastLineCount":
		"vprint":        ft.vprint,
		"vprintl":       ft.vprintL,
		"vprintc":       ft.vprintC,
		"vprintw":       ft.vprintW,
		"vprintLine":    ft.vprintLine,
		"vprintBar":     ft.vprintBar,
		"vprintButton":  ft.vprintButton,
		"vnewPage":      ft.vnewPage,
		"vclearLineAll": ft.vclearLineAll,
		"vclearLine":    ft.vclearLine,
	}
	eraModInputFuncMap := map[string]lua.LGFunction{
		// input functions
		"printw":      ft.printW,
		"wait":        ft.wait,
		"twait":       ft.twait,
		"input":       ft.inputStr,
		"tinput":      ft.tinputStr,
		"inputNum":    ft.inputNum,
		"tinputNum":   ft.tinputNum,
		"inputRange":  ft.inputNumRange,
		"inputSelect": ft.inputNumSelect,
		"rawInput":    ft.rawInput,
		"trawInput":   ft.trawInput,
	}
	// Input functions keep alive WDT since these interact with the user.
	for k, v := range eraModInputFuncMap {
		eraModInputFuncMap[k] = ip.watchDogTimer.WrapKeepAliveLG(v)
	}
	for k, v := range eraModInputFuncMap {
		eraModFuncMap[k] = v
	}
	// era functions consume pending tasks
	for k, v := range eraModFuncMap {
		eraModFuncMap[k] = ip.wrapConsumeTaskLG(v)
	}
	L.SetFuncs(era_module, eraModFuncMap)
	L.SetMetatable(era_module, getStrictTableMetatable(L))

	flowModFuncMap := map[string]lua.LGFunction{
		// Module for controling game or scene flow.
		"quit":          quitScript,
		"longReturn":    longReturnScript,
		"setNextScene":  ft.setNextScene,
		"gotoNextScene": ft.gotoNextScene,
		"saveScene":     ft.saveScene,
		"loadScene":     ft.loadScene,
		"doTrains":      ft.doTrains,
	}
	// flow module need not to wrap keep alive WDT since it backs controll to
	// platform side which stops WDT or just does additional subroutine which
	// still have possibility of the infinite loop.
	flowMod := L.SetFuncs(L.NewTable(), flowModFuncMap)
	L.SetMetatable(flowMod, getStrictTableMetatable(L))
	era_module.RawSetString(eraFlowModName, flowMod)

	// layout module will be deprecated.
	layoutModFuncMap := map[string]lua.LGFunction{
		// layouting
		"setCurrentView": ft.setCurrentView,
		"getCurrentView": ft.getCurrentViewName,
		"viewNames":      ft.getViewNames,
		"setSingle":      ft.setSingleLayout,
		"setVertical":    ft.setVerticalLayout,
		"setHorizontal":  ft.setHorizontalLayout,
		"setLayout":      ft.setLayout,
		"text":           singleTextLayout,
		"image":          singleImageLayout,
		"flowHorizontal": flowHorizontalLayout,
		"flowVertical":   flowVerticalLayout,
		"fixedSplit":     fixedSplitLayout,
		"withValue":      withLayoutValue,
	}
	layoutMod := L.SetFuncs(L.NewTable(), layoutModFuncMap)
	L.SetMetatable(layoutMod, getStrictTableMetatable(L))
	era_module.RawSetString(eraLayoutModName, layoutMod)

	return era_module
}

// functor projects game functions to lua functions
type functor struct {
	game  GameController
	state *state.GameState
}

// +gendoc.set_section "Era Module"

// //   Game State

// +gendoc "Era Module"
// * era.clearSystem()
//
// it clears the game system data, values under era.system.
// All the current values contained in the system data are set
// by 0 or empty string.
//
// 現在のゲームデータを初期化します。具体的には
// キャラはすべて削除し、全ての変数には、0や空文字列を代入します。
func (ft functor) clearSystem(L *lua.LState) int {
	ft.state.SystemData.Clear()
	return 0
}

// +gendoc "Era Module"
// * era.saveSystem(number, [comment])
//
// it saves current system data, values under era.system, into file
// specified by given number, i.e. save[number].sav.
// 2nd argument comment is optional, which is saved togather,
// and will be showed as header of available save file.
//
// 現在のゲームデータをsave[number].savに保存します。
// 2つ目の引数として、コメントを渡すことで、セーブデータの一覧表示時の
// コメントを設定できます。
// この操作の前後には、何の反応も起きないことに注意。
func (ft functor) saveSystem(L *lua.LState) int {
	no := L.CheckInt(1)
	comment := L.OptString(2, "")

	var err error
	if len(comment) > 0 {
		err = ft.state.SaveSystemWithComment(no, comment)
	} else {
		err = ft.state.SaveSystem(no)
	}
	raiseErrorIf(L, err)
	return 0
}

// +gendoc "Era Module"
// * era.loadSystem(number)
//
// it loads system data, values under era.system, from specified file directly.
// current system data is overwritten by loaded data.
// the name of load file is typically save[number].sav.
//
// 現在のゲームデータをsave[number].savから読みこんだデータで上書きします。
// この操作の前後には、何の反応も起きないことに注意。
func (ft functor) loadSystem(L *lua.LState) int {
	no := L.CheckInt(1)
	err := ft.state.LoadSystem(no)
	raiseErrorIf(L, err)
	// re-register loaded values, which may be changed its structure
	registerSystemParams(L, ft.state)
	registerCharaParams(L, ft.state)
	return 0
}

// +gendoc "Era Module"
// * era.clearShare()
//
// it clears the game share data, values under era.share.
// All the current values contained in the share data are set
// by 0 or empty string.
//
// 現在の共有ゲームデータを初期化します。具体的には
// 全ての変数に、0や空文字列を代入します。
func (ft functor) clearShare(L *lua.LState) int {
	ft.state.ShareData.Clear()
	return 0
}

// +gendoc "Era Module"
// * era.saveShare()
//
// it saves current share data, values under era.share, into file.
// share data are shared multiple independent savefile.
//
// 現在の共有ゲームデータをファイルに保存します。
// 共有ゲームデータは、個別のセーブファイル間で共有することを想定しています。
func (ft functor) saveShare(L *lua.LState) int {
	err := ft.state.SaveShare()
	raiseErrorIf(L, err)
	return 0
}

// +gendoc "Era Module"
// * era.loadShare()
//
// it loads share data, values under era.share, from file directly.
// current share data is overwritten by loaded data.
//
// 現在の共有ゲームデータを共有セーブファイルから読みこんだデータで上書きします。
// この操作の前後には、何の反応も起きないことに注意。
func (ft functor) loadShare(L *lua.LState) int {
	err := ft.state.LoadShare()
	raiseErrorIf(L, err)
	// re-register loaded values, which may be changed its structure
	registerSystemParams(L, ft.state)
	return 0
}

// //  alignment functions

// +gendoc "Era Module"
// * era.setAlignment(alignment)
//
// set alignment of current line by specified argument.
// The argument is "left", "center" or "right".
//
// 現在のアライメントを変更します。"left"、"center"、または"right"のいずれかを指定します。
func (ft functor) setAlignment(L *lua.LState) int {
	game := ft.game
	switch align := L.CheckString(1); align {
	case "left", "LEFT", "l", "L":
		game.SetAlignment(scene.AlignmentLeft)
	case "right", "RIGHT", "r", "R":
		game.SetAlignment(scene.AlignmentRight)
	case "center", "CENTER", "c", "C":
		game.SetAlignment(scene.AlignmentCenter)
	default:
		L.ArgError(1, "unknown alignment string: "+align)
	}
	return 0
}

// +gendoc "Era Module"
// * alignment = era.getAlignment()
//
// It gets alignment of current editing line.
// The returned value can use to set alignment.
//
// 現在のアライメントを取得します。
// 取得した値は、SetAlignment()に渡すことができます。
func (ft functor) getAlignment(L *lua.LState) int {
	var alignStr string
	align, err := ft.game.GetAlignment()
	if err != nil {
		raiseErrorf(L, "script.getAlignment(): %w", err)
	}
	switch align {
	case scene.AlignmentLeft:
		alignStr = "left"
	case scene.AlignmentCenter:
		alignStr = "center"
	case scene.AlignmentRight:
		alignStr = "right"
	default:
		raiseErrorf(L, "script.getAlignment(): unkown alignment")
	}
	L.Push(lua.LString(alignStr))
	return 1
}

func checkHexColor(L *lua.LState, pos int) uint32 {
	color := L.CheckNumber(pos)
	hex := uint32(color) & 0xffffff
	return hex
}

// +gendoc "Era Module"
// * era.setColor(color)
//
// set current text color. The argument color is represented as 0xRRGGBB.
//
// 現在のテキストカラーを変更します。カラーは0xRRGGBBの形式で渡します。
func (ft functor) setColor(L *lua.LState) int {
	color := checkHexColor(L, 1)
	ft.game.SetColor(color)
	return 0
}

func pushColor(L *lua.LState, color uint32) int {
	L.Push(lua.LNumber(color))
	return 1
}

// +gendoc "Era Module"
// * color = era.getColor()
//
// get current text color. The return color is represented as 0xRRGGBB.
//
// 現在のテキストカラーを返します。カラーは0xRRGGBBの形式で表されます。
func (ft functor) getColor(L *lua.LState) int {
	color, err := ft.game.GetColor()
	if err != nil {
		raiseErrorf(L, "script.getColor(): %w", err)
	}
	return pushColor(L, color)
}

// +gendoc "Era Module"
// * era.resetColor()
//
// reset current text color to default.
//
// 現在のテキストカラーをデフオルトのものに設定し直します。
func (ft functor) resetColor(L *lua.LState) int {
	ft.game.ResetColor()
	return 0
}

// // Output functions

func viewNameError(L *lua.LState, pos int, mes string, err error) {
	L.ArgError(pos, mes+err.Error())
}

func checkAnyString(L *lua.LState, pos int) string {
	return L.CheckAny(pos).String()
}

// +gendoc "Era Module"
// * era.print(text)
//
// print text to screen. it is ok to contain return code (\n),
// in which trailing string after return code is moved to next line.
//
// If certain pattern such as "[number] string" is appeard in the text,
// the pattern is converted to the button which is able to click and emits command
// as user input.
// This conversion is performed for every line, ends by return code \n, at once.
//
// テキストを画面に出力します。
// テキストには、改行文字(\n)が含まれていても構いません。
// 改行文字が含まれている場合、続く文字列は次の行に改行されます。
//
// また、テキストの中に、"[数字] 文字..." のようなパターンが
// 見つかると、その部分を、クリックすることでコマンドを発行するボタン、
// として扱います。
// このボタンの変換機能は、一行(\nまで)に一度のみ行われることに注意してください。
func (ft functor) print(L *lua.LState) int {
	text := checkAnyString(L, 1)
	ft.game.Print(text)
	return 0
}

func (ft functor) vprint(L *lua.LState) int {
	vname := L.CheckString(1)
	text := checkAnyString(L, 2)
	if err := ft.game.VPrint(vname, text); err != nil {
		L.ArgError(1, "vprint: "+err.Error())
		return 0
	}
	return 0
}

// +gendoc "Era Module"
// * era.printl(text)
//
// same as print(), but adding return code \n into end of text.
//
// print()と同じ動作をします。ただし、最後に改行文字(\n)を自動で付与します。
func (ft functor) printL(L *lua.LState) int {
	text := checkAnyString(L, 1)
	ft.game.PrintL(text)
	return 0
}

func (ft functor) vprintL(L *lua.LState) int {
	vname := L.CheckString(1)
	text := checkAnyString(L, 2)
	if err := ft.game.VPrintL(vname, text); err != nil {
		L.ArgError(1, "vprintl: "+err.Error())
		return 0
	}
	return 0
}

// +gendoc "Era Module"
// * era.printc(text, [count])
//
// The text, which is padding half space to fill the character count, is printed to screen.
// In this game, character count is counted by: single-byte character is 1 and multi-byte character
// is 2.
// the count is optional, calling with nothing count treats default count as 26.
// In this function, if the button pattern is found in the text,
// the conversion of the text to the button is performed to entire text, not part of the pattern only.
//
// Note that the part of previous return code of text is only treated.
// trailing string are ignored.
//
// count数を満たすように、半角スペースを追加したテキストを、スクリーンにプリントします。
// ここでの、countは、シングルバイト文字を1文字、マルチバイト文字を2文字として数えます。
// countの指定は省略することが可能です。その場合、デフォルトで26を用います。
// この関数では、ボタンパターンが現れた時、パターンの部分のみではなく、テキスト全体を
// ボタンに変換します。
//
// また、改行文字(\n)が現れた時、その前までの文字のみを扱い、それ以降の文字は
// 無視されることに注意してください。
func (ft functor) printC(L *lua.LState) int {
	text := checkAnyString(L, 1)
	count := L.OptInt(2, scene.DefaultPrintCWidth) // TODO: using ft.Config.PrintCSize as default?
	ft.game.PrintC(text, count)
	return 0
}

func (ft functor) vprintC(L *lua.LState) int {
	vname := L.CheckString(1)
	text := checkAnyString(L, 2)
	count := L.OptInt(3, scene.DefaultPrintCWidth) // TODO: using ft.Config.PrintCSize as default?
	if err := ft.game.VPrintC(vname, text, count); err != nil {
		L.ArgError(1, "vprintc: "+err.Error())
		return 0
	}
	return 0
}

// +gendoc "Era Module"
// * era.printw(text)
//
// textを出力し、ユーザーからの何らかの入力を待ちます。
//
//	era.print(text)
//	era.wait()
//
// と等価です。
func (ft functor) printW(L *lua.LState) int {
	text := checkAnyString(L, 1)
	ft.game.PrintW(text)
	return 0
}

func (ft functor) vprintW(L *lua.LState) int {
	vname := L.CheckString(1)
	text := checkAnyString(L, 2)
	if err := ft.game.VPrintW(vname, text); err != nil {
		L.ArgError(1, "vprintw: "+err.Error())
		return 0
	}
	return 0
}

// +gendoc "Era Module"
// * era.printLine([symbol])
//
// symbolを使って画面を横切る線を出力します。
// symbolを与えなかった場合には、デフォルトで"="を使用します。
func (ft functor) printLine(L *lua.LState) int {
	if L.GetTop() == 0 {
		ft.game.PrintLine(scene.DefaultLineSymbol)
		return 0
	}
	sym := L.CheckString(1)
	ft.game.PrintLine(sym)
	return 0
}

func (ft functor) vprintLine(L *lua.LState) int {
	vname := L.CheckString(1)
	if L.GetTop() == 1 {
		if err := ft.game.VPrintLine(vname, scene.DefaultLineSymbol); err != nil {
			L.ArgError(1, "vprintline: "+err.Error())
		}
		return 0
	}
	r := L.CheckString(2)
	if err := ft.game.VPrintLine(vname, r); err != nil {
		L.ArgError(1, "vprintline: "+err.Error())
	}
	return 0
}

// +gendoc "Era Module"
// * era.printButton(caption, command)
//
// captionを選択可能なテキストボタンとして出力します。
// このボタンを選択した場合には、commandが入力されます。
//
//	era.printButton("こんにちは", "10")
//	local input = era.input() -- "こんにちは"を選択
//	input == "10"
func (ft functor) printButton(L *lua.LState) int {
	caption := L.CheckString(1)
	cmd := L.CheckString(2)
	ft.game.PrintButton(caption, cmd)
	return 0
}

func (ft functor) vprintButton(L *lua.LState) int {
	vname := L.CheckString(1)
	caption := L.CheckString(2)
	cmd := L.CheckString(3)
	if err := ft.game.VPrintButton(vname, caption, cmd); err != nil {
		L.ArgError(1, "vprintline: "+err.Error())
	}
	return 0
}

// +gendoc "Era Module"
// * era.printPlain(text: string)
//
// text をそのまま画面に表示します。print() と異なり、クリックできるボタンへの自動変換を行いません。
func (ft functor) printPlain(L *lua.LState) int {
	text := L.CheckString(1)
	err := ft.game.PrintPlain(text)
	raiseErrorIf(L, err)
	return 0
}

// +gendoc "Era Module"
// * era.newPage()
//
// 画面に何も表示されなくなるまで、改行を繰り返します。
// 見た目的にはera.clearLineAll()と同様になりますが、
// 出力した履歴を消すことなく、また、次の出力は
// 画面の最下行から行います。
func (ft functor) newPage(L *lua.LState) int {
	ft.game.NewPage()
	return 0
}

func (ft functor) vnewPage(L *lua.LState) int {
	vname := L.CheckString(1)
	if err := ft.game.VNewPage(vname); err != nil {
		L.ArgError(1, "vnewpage: "+err.Error())
	}
	return 0
}

// +gendoc "Era Module"
// * era.clearLineAll()
//
// 画面に出力された全ての行を消去します。
// 次に出力するときは、画面の最上行から行います。
func (ft functor) clearLineAll(L *lua.LState) int {
	ft.game.ClearLineAll()
	return 0
}

func (ft functor) vclearLineAll(L *lua.LState) int {
	vname := L.CheckString(1)
	if err := ft.game.VClearLineAll(vname); err != nil {
		L.ArgError(1, "vclearlines: "+err.Error())
	}
	return 0
}

// +gendoc "Era Module"
// * era.clearLine(nline)
//
// 画面に出力された行をnlineの数だけ消去します。
// もしnlineが1であるなら、現在編集中の行を空にするのみです。
func (ft functor) clearLine(L *lua.LState) int {
	lines := L.OptInt(1, 1)
	ft.game.ClearLine(lines)
	return 0
}

func (ft functor) vclearLine(L *lua.LState) int {
	vname := L.CheckString(1)
	lines := L.OptInt(2, 1)
	if err := ft.game.VClearLine(vname, lines); err != nil {
		L.ArgError(1, "vclearline: "+err.Error())
	}
	return 0
}

// +gendoc "Era Module"
// * width = era.windowStrWidth()
//
// Return string width to fill the view's width.
// Here, a single byte character is counted 1, a multibyte is 2.
// For example, let windowStrWidth = 3, one single byte character
// and one multibyte character fill the view's width.
//
// Viewの横幅に収まる最大の文字幅を返します。
// ここでは、半角1文字を1、全角1文字を2として数えます。
// 例えば、windowStrWidth()が3ならば、半角文字1つと全角文字1つまでが
// 最大の横幅に収まります。
func (ft functor) windowStrWidth(L *lua.LState) int {
	width, err := ft.game.WindowRuneWidth()
	if err != nil {
		raiseErrorf(L, "script.windowStrWidth(): %w", err)
	}
	L.Push(lua.LNumber(width))
	return 1
}

// +gendoc "Era Module"
// * line_count = era.windowLineCount()
//
// return line count to fill the view's height.
//
// Viewの高さに収まる最大の行数を返します
func (ft functor) windowLineCount(L *lua.LState) int {
	width, err := ft.game.WindowLineCount()
	if err != nil {
		raiseErrorf(L, "script.windowLineCount(): %w", err)
	}
	L.Push(lua.LNumber(width))
	return 1
}

// +gendoc "Era Module"
// * width = era.currentStrWidth()
//
// Return string width in the currently editing line.
// Here, a single byte character is counted 1, a multibyte is 2.
//
// 現在、編集中の行の文字幅を返します。
// ここでは、半角1文字を1、全角1文字を2として数えます。
func (ft functor) currentStrWidth(L *lua.LState) int {
	width, err := ft.game.CurrentRuneWidth()
	if err != nil {
		raiseErrorf(L, "script.currentStrWidth(): %w", err)
	}
	L.Push(lua.LNumber(width))
	return 1
}

// +gendoc "Era Module"
// * count = era.lineCount()
//
// Return line count as it increases outputting new line.
//
// これまでに改行した数を返します。
// 明示的な改行（printlの使用や"\n"をprintする）が起こった数だけ、数値が増加します。
func (ft functor) lineCount(L *lua.LState) int {
	count, err := ft.game.LineCount()
	if err != nil {
		raiseErrorf(L, "script.lineCount(): %w", err)
	}
	L.Push(lua.LNumber(count))
	return 1
}

// +gendoc "Era Module"
// * width = era.textWidth(text)
//
// Return string width of given text.
// A single byte character is counted as 1, a multibyte one as 2.
//
// 引数として渡された文字列 text の文字幅を返します。
// 半角1文字を1、全角1文字を2として数えます。
func (ft functor) textWidth(L *lua.LState) int {
	text := L.CheckString(1)
	count := width.StringWidth(text) // TODO: use width.Condition istance to avoid global state?
	L.Push(lua.LNumber(count))
	return 1
}

// // Bar

func checkBarParams(L *lua.LState, base_pos int) (ret struct {
	now, max int64
	width    int
	fg, bg   string
}) {
	ret.now = L.CheckInt64(base_pos)
	ret.max = L.CheckInt64(base_pos + 1)
	ret.width = L.OptInt(base_pos+2, scene.DefaultTextBarWidth)
	ret.fg = L.OptString(base_pos+3, scene.DefaultTextBarFg)
	ret.bg = L.OptString(base_pos+4, scene.DefaultTextBarBg)
	return
}

// +gendoc "Era Module"
// * era.printBar(now, max, [width, fg, bg])
//
// nowとmaxの数値の比を、テキストの棒グラフで表示します。
// 例えば、"[###..]"のように出力されます。
// 文字数の幅widthでグラフの横幅を指定できますが、
// "["と"]"の幅を含んだ値であることに注意してください。
// また、文字列fgとbgでそれぞれグラフ部分および背景部分を指定できます。
//
// width, fg, bgについて指定がなかった場合、
// width=8, fg="#", bg="."をデフォルトで使用します。
//
//	era.printBar(10, 30)	             -- "[##....]"と出力
//	era.printBar(10, 30, 5, "=", " ") -- "[=  ]"と出力
func (ft functor) printBar(L *lua.LState) int {
	p := checkBarParams(L, 1)
	ft.game.PrintBar(p.now, p.max, p.width, p.fg, p.bg)
	return 0
}

func (ft functor) vprintBar(L *lua.LState) int {
	vname := L.CheckString(1)
	p := checkBarParams(L, 2)
	ft.game.VPrintBar(vname, p.now, p.max, p.width, p.fg, p.bg)
	return 0
}

// +gendoc "Era Module"
// * text_bar = era.textBar(now, max, [width, fg, bg])
//
// printBar()で表示されるテキスト形式の棒グラフを
// 文字列として受け取ります。画面への表示はされません。
// 詳細はprintBar()を参照
func (ft functor) textBar(L *lua.LState) int {
	p := checkBarParams(L, 1)
	res, err := ft.game.TextBar(p.now, p.max, p.width, p.fg, p.bg)
	if err != nil {
		raiseErrorf(L, "script.textBar(): %w", err)
	}
	L.Push(lua.LString(res))
	return 1
}

// +gendoc "Era Module"
// * era.printImage(image_path, width_in_tw, [height_in_lc])
//
// image_path で指定した画像データを表示します。
// image_path は スクリプト *.lua を配置するディレクトリからの相対パスで指定します。
// *.lua を配置するディレクトリより上の階層を指定した場合エラーとなります。
// サポートしている画像フォーマットは .png, .jpeg です。
// 存在しないファイルパスを指定した場合、異常を示す画像が変わりに表示されます。
// 画面上で画像が表示されるサイズを TextWidth分の幅、LineCount 分の高さで指定します。
// height_in_lc を 0 または省略した場合、width_in_tw に対して、元の画像のアスペクト比を
// 保った高さに自動調整されます。
//
//	-- 幅 TextWidth 30、高さアスペクト比追従で、 image.png を表示。
//	era.printImage("path/to/image.png", 30)
//	-- 30x30 のサイズで image2.png を表示。元画像のアスペクト比が 1:1 出ない場合、縦横いずれかに引き伸ばされて表示。
//	era.printImage("path/to/image2.png", 30, 30)
func (ft functor) printImage(L *lua.LState) int {
	imgPath, widthInTW, heightInLC := checkImageParams(L, 1)
	err := ft.game.PrintImage(imgPath, widthInTW, heightInLC)
	if err != nil {
		raiseErrorf(L, "script.printImage(): %w", err)
	}
	return 0
}

// +gendoc "Era Module"
// * w, h = era.measureImageSize(image_path, width_in_tw, [height_in_lc])
//
// image_path で指定した画像データの表示サイズを取得します。
// 返却された表示サイズ w, h の単位はそれぞれテキストの幅、行数です。
// この関数に渡すパラメータは printImage() と同じ仕様です。
// この関数は、例えば height_in_lc を省略した場合、自動的に計算された height がいくつになるのかを
// 知りたい場合に使用できます。
//
//	w, h = era.measureImageSize(image_path, width_in_tw)
//	era.printImage(image_path, width_in_tw)
//	-- 画像表示分だけ改行して、次の文字表示位置を画像の下の行に合わせる。
//	for i = 1, w do
//	  era.printl("")
//	end
func (ft functor) measureImageSize(L *lua.LState) int {
	imgPath, widthInTW, heightInLC := checkImageParams(L, 1)
	retW, retH, err := ft.game.MeasureImageSize(imgPath, widthInTW, heightInLC)
	if err != nil {
		raiseErrorf(L, "script.measureImageSize(): %w", err)
	}
	L.Push(lua.LNumber(retW))
	L.Push(lua.LNumber(retH))
	return 2
}

func checkImageParams(L *lua.LState, startPos int) (imgPath string, widthInTW, heightInLC int) {
	imgPath = L.CheckString(startPos)
	widthInTW = L.CheckInt(startPos + 1)
	heightInLC = L.OptInt(startPos+2, 0)

	imgPath, err := scriptPath(L, imgPath)
	if err != nil {
		L.ArgError(startPos, err.Error())
	}
	if widthInTW < 0 {
		widthInTW = 0
	}
	if heightInLC < 0 {
		heightInLC = 0
	}
	return
}

// // Input functions

// check whether error is context.DeadlineExceeded.
// if not Raise error and exit.
func checkTimeExceeded(L *lua.LState, err error) bool {
	switch err {
	case context.DeadlineExceeded:
		return true
	case nil:
		return false
	default:
		raiseErrorE(L, err)
		return false
	}
}

// +gendoc "Era Module"
// * era.wait()
//
// ユーザーからの何らかの入力を待ちます。
// 入力があるまで、スクリプトは停止し続けます。
func (ft functor) wait(L *lua.LState) int {
	err := ft.game.Wait()
	raiseErrorIf(L, err)
	return 0
}

// +gendoc "Era Module"
// * time_exceeded = era.twait(nanosec)
//
// ユーザーからの何らかの入力を待ちます。
// 入力があるか、時間がnanosec経過すると処理を再開します。
// nanosec経過した場合にはtime_exceededがtrueになり、
// ユーザからの入力があった場合にはtime_exceededがfalseになります。
func (ft functor) twait(L *lua.LState) int {
	timeout := time.Duration(L.CheckInt64(1))
	// TODO use application context?
	err := ft.game.WaitWithTimeout(context.Background(), timeout*time.Nanosecond)
	exceeded := checkTimeExceeded(L, err)
	L.Push(lua.LBool(exceeded))
	return 1
}

func pushIntError(L *lua.LState, num int, err error) int {
	if err != nil {
		raiseErrorE(L, err)
	}
	L.Push(lua.LNumber(num))
	return 1
}

// +gendoc "Era Module"
// * command_number = era.inputNum()
//
// ユーザーからの数字の入力を待ちます。
// 数字の入力があった場合、数値command_numberが返ってきます。
func (ft functor) inputNum(L *lua.LState) int {
	num, err := ft.game.CommandNumber()
	return pushIntError(L, num, err)
}

// +gendoc "Era Module"
// * command_number, time_exceeded = era.tinputNum(nanosec)
//
// ユーザーからの数字の入力を時間制限付きで待ちます。
// nanosec経過すると、time_exceededがtureになり、command_numberは0になります。
func (ft functor) tinputNum(L *lua.LState) int {
	timeout := time.Duration(L.CheckInt64(1))
	// TODO use application context?
	num, err := ft.game.CommandNumberWithTimeout(context.Background(), timeout*time.Nanosecond)
	exceeded := checkTimeExceeded(L, err)
	L.Push(lua.LNumber(num))
	L.Push(lua.LBool(exceeded))
	return 2
}

// +gendoc "Era Module"
// * command_number = era.inputRange(min, max)
//
// 数値minからmaxの範囲の数字の入力を待ちます。
// 結果をcommand_numberとして返します。
func (ft functor) inputNumRange(L *lua.LState) int {
	min := L.CheckInt(1)
	max := L.CheckInt(2)
	if min > max {
		L.ArgError(2, "1st argument is less than 2nd")
	}
	num, err := ft.game.CommandNumberRange(context.Background(), min, max)
	return pushIntError(L, num, err)
}

// +gendoc "Era Module"
// * command_number = era.inputSelect(num1, num2, ...)
//
// 渡した複数の数値num1, num2, ...に一致したものの入力を待ちます。
// 結果をcommand_numberとして返します。
func (ft functor) inputNumSelect(L *lua.LState) int {
	arg_size := L.GetTop()
	if arg_size == 0 {
		L.Error(lua.LString("at least 1 argument required"), 0)
	}
	candidates := make([]int, 0, arg_size)
	for i := 1; i <= arg_size; i++ {
		candidates = append(candidates, L.CheckInt(i))
	}
	num, err := ft.game.CommandNumberSelect(context.Background(), candidates...)
	return pushIntError(L, num, err)
}

func pushStringError(L *lua.LState, s string, err error) int {
	if err != nil {
		raiseErrorE(L, err)
	}
	L.Push(lua.LString(s))
	return 1
}

// +gendoc "Era Module"
// * command = era.input()
//
// ユーザーからの文字列の入力を待ちます。
// 結果をcommandとして返します。
// ユーザーからの入力が数字であったとしても、
// "10"のように文字列として返ってきます。
func (ft functor) inputStr(L *lua.LState) int {
	s, err := ft.game.Command()
	return pushStringError(L, s, err)
}

// +gendoc "Era Module"
// * command, time_exceeded = era.tinput(nanosec)
//
// nanosecの間、ユーザーからの文字列の入力を待ちます。
// 結果をcommandとして返します。
// nanosec経過した場合、time_exceededがtrueになります。
// ユーザーからの入力が数字であったとしても、
// "10"のように文字列として返ってきます。
func (ft functor) tinputStr(L *lua.LState) int {
	timeout := time.Duration(L.CheckInt64(1))
	// TODO use application context?
	s, err := ft.game.CommandWithTimeout(context.Background(), timeout*time.Nanosecond)
	exceeded := checkTimeExceeded(L, err)
	L.Push(lua.LString(s))
	L.Push(lua.LBool(exceeded))
	return 2
}

// +gendoc "Era Module"
// * ch = era.rawInput()
func (ft functor) rawInput(L *lua.LState) int {
	s, err := ft.game.RawInput()
	return pushStringError(L, s, err)
}

// +gendoc "Era Module"
// * ch, exceeded = era.trawInput(nanosec)
func (ft functor) trawInput(L *lua.LState) int {
	timeout := time.Duration(L.CheckInt64(1))
	// TODO use application context?
	s, err := ft.game.RawInputWithTimeout(context.Background(), timeout*time.Nanosecond)
	exceeded := checkTimeExceeded(L, err)
	L.Push(lua.LString(s))
	L.Push(lua.LBool(exceeded))
	return 2
}

// // Controll Flow

// +gendoc "Era Module"
// * var era.flow: flow

// +gendoc "Flow Module"
// * flow.quit()
//
// quit this game immediately, without any notification.
//
// ゲームを終了します。何の前触れもなく落ちるので注意。
func quitScript(L *lua.LState) int {
	L.Error(lua.LString(ScriptQuitMessage), 0)
	return 0
}

// +gendoc "Flow Module"
// * flow.longReturn()
//
// return system beyond current call stack.
// it is used to non-local exits of the function.
//
// 強制的にシステムのスクリプト呼び出し元まで戻ります。
// 大域脱出を想定しています。
func longReturnScript(L *lua.LState) int {
	L.Error(lua.LString(ScriptLongReturnMessage), 0)
	return 0
}

// error handling for interpreter.
func raiseErrorIf(L *lua.LState, err error) {
	if err != nil {
		raiseErrorE(L, err)
	}
}

// +gendoc "Flow Module"
// * flow.setNextScene(scene_name)
//
// it sets next scene detected by given scene name.
// passing unknown scene name will occurs error.
//
// 次のシーンを、その名前でシステムに伝えます。
// 次のシーンが決まらなければ、ゲームは進行していかないため、
// この関数は重要です。
// もし、存在しないシーンの名前を渡した場合、エラーを起こします。
func (ft functor) setNextScene(L *lua.LState) int {
	name := L.CheckString(1)
	err := ft.game.SetNextSceneByName(name)
	raiseErrorIf(L, err)
	return 0
}

// +gendoc "Flow Module"
// * flow.gotoNextScene(scene_name)
//
// go to next scene imidiately. current scene flow is interrupted
// and starting next scene specified by scene_name.
// passing unknown scene_name will occurs error.
//
// 強制的に次のシーンを開始します。
// 現在のシーンは中断され、scene_nameで指定したシーンが開始されます。
// もし、存在しないシーンの名前scene_nameを渡した場合、エラーを起こします。
func (ft functor) gotoNextScene(L *lua.LState) int {
	name := L.CheckString(1)
	if err := ft.game.SetNextSceneByName(name); err != nil {
		L.ArgError(1, err.Error())
	}
	L.Error(lua.LString(ScriptGoToNextSceneMessage), 0)
	return 0
}

// +gendoc "Flow Module"
// * flow.saveScene()
//
// it starts the save game scene to save current game state.
// the save game scene will show available save file lists, then
// return caller's position after do some action in the scene.
//
// ゲームデータをセーブするためのシーンを呼び出します。
// そのシーンでは、セーブデータの一覧が表示されます。
// セーブを行うか、戻るを選択すると、この関数を呼び出した
// 場所の直後から処理が再開されます。
func (ft functor) saveScene(L *lua.LState) int {
	err := ft.game.DoSaveGameScene()
	raiseErrorIf(L, err)
	return 0
}

// +gendoc "Flow Module"
// * flow.loadScene()
//
// it starts the load game scene to load game state into current state.
// the load game scene will show available save file lists.
// If any load file is done, then the load game scene starts
// scene "load_end" and never return the caller's position.
// But return caller's position, if selecting "return" in the load scene.
//
// ゲームデータをロードするためのシーンを呼び出します。
// そのシーンでは、セーブデータの一覧が表示されます。
// ロードを行うと、この関数を呼び出した場所には決して戻らず、
// "load_end"シーンを開始します。
// しかし、戻るを選択した場合には、この関数の呼び出し元に戻ってきます。
func (ft functor) loadScene(L *lua.LState) int {
	err := ft.game.DoLoadGameScene()
	if err == scene.ErrorSceneNext {
		// re-register loaded values, which may be changed its structure
		registerSystemParams(L, ft.state)
		registerCharaParams(L, ft.state)
		L.Error(lua.LString(ScriptGoToNextSceneMessage), 0)
		return 0
	}
	raiseErrorIf(L, err)
	return 0
}

func tableToInt64s(t *lua.LTable) []int64 {
	cmds := make([]int64, t.Len())
	t.ForEach(func(i lua.LValue, v lua.LValue) {
		// in lua, index starts 1, but in Go, starts 0 So -1
		cmds[int(lua.LVAsNumber(i))-1] = int64(lua.LVAsNumber(v))
	})
	return cmds
}

// +gendoc "Flow Module"
// * flow.doTrains(commands)
//
// Do multiple trains specified by commands which is sequence of command number.
// This function is callable in the scene train.
//
// 複数のTrainコマンド(commands)を実行します。コマンドは数字の配列として表現する必要があります。
// do_trains()はTrainシーンの中でのみ、実行可能であることに注意してください。
func (ft functor) doTrains(L *lua.LState) int {
	table := L.CheckTable(1)
	cmds := tableToInt64s(table)
	err := ft.game.DoTrainsScene(cmds)
	raiseErrorIf(L, err)
	return 0
}

// // Utils

func currentLv(param int64, lvs []int64) int {
	for lv, step := range lvs {
		if param < step {
			return lv
		}
	}
	return 0
}

func (ft functor) paramLv(L *lua.LState) int {
	param := L.CheckInt64(1)
	param_lvs := ft.state.CSV.ParamLvs
	L.Push(lua.LNumber(currentLv(param, param_lvs)))
	return 1
}

func (ft functor) expLv(L *lua.LState) int {
	param := L.CheckInt64(1)
	exp_lvs := ft.state.CSV.ExpLvs
	L.Push(lua.LNumber(currentLv(param, exp_lvs)))
	return 1
}

// // View and Screen Layout

// +gendoc "Layout Module"
// * layout.setCurrentView(vname)
func (ft functor) setCurrentView(L *lua.LState) int {
	vname := L.CheckString(1)
	if err := ft.game.SetCurrentView(vname); err != nil {
		viewNameError(L, 1, "set_current_view", err)
	}
	return 0
}

// +gendoc "Layout Module"
// * vname = layout.getCurrentView()
func (ft functor) getCurrentViewName(L *lua.LState) int {
	L.Push(lua.LString(ft.game.GetCurrentViewName()))
	return 1
}

// +gendoc "Layout Module"
// * vnames = layout.viewNames()
func (ft functor) getViewNames(L *lua.LState) int {
	vnames := ft.game.GetViewNames()
	table := L.CreateTable(len(vnames), 0)
	for i, vname := range vnames {
		// lua table starts 1 index.
		table.RawSetInt(i+1, lua.LString(vname))
	}
	L.Push(table)
	return 1
}

// +gendoc "Layout Module"
// * layout.setSingle(vname)
func (ft functor) setSingleLayout(L *lua.LState) int {
	s := L.OptString(1, ft.game.GetCurrentViewName())
	err := ft.game.SetSingleLayout(s)
	raiseErrorIf(L, err)
	return 0
}

func getVHLayoutArgs(L *lua.LState) (string, string, float64) {
	s1 := L.CheckString(1)
	s2 := L.CheckString(2)
	rate := L.OptNumber(3, lua.LNumber(0.5))
	return s1, s2, float64(rate)
}

// +gendoc "Layout Module"
// * layout.setVertical(vname1, vname2, [rate])
func (ft functor) setVerticalLayout(L *lua.LState) int {
	s1, s2, rate := getVHLayoutArgs(L)
	err := ft.game.SetVerticalLayout(s1, s2, rate)
	raiseErrorIf(L, err)
	return 0
}

// +gendoc "Layout Module"
// * layout.setHorizontal(vname1, vname2, [rate])
func (ft functor) setHorizontalLayout(L *lua.LState) int {
	s1, s2, rate := getVHLayoutArgs(L)
	err := ft.game.SetHorizontalLayout(s1, s2, rate)
	raiseErrorIf(L, err)
	return 0
}

// +gendoc "Layout Module"
// * layout.set(layout_data)
func (ft functor) setLayout(L *lua.LState) int {
	ld := checkLayoutData(L, 1)
	if ld == nil {
		raiseErrorf(L, "passing nil LayoutData")
		return 0
	}
	err := ft.game.SetLayout(ld)
	raiseErrorIf(L, err)
	return 0
}

func checkLayoutData(L *lua.LState, pos int) *attr.LayoutData {
	ud := L.CheckUserData(pos)
	if ld, ok := ud.Value.(*attr.LayoutData); ok {
		return ld
	}
	L.ArgError(pos, "require layout_data type")
	return nil
}

// +gendoc "Layout Module"
// * layout_data = layout.text(name)
func singleTextLayout(L *lua.LState) int {
	name := L.CheckString(1)
	text := attr.NewSingleText(name)
	L.Push(newUserData(L, text))
	return 1
}

// +gendoc "Layout Module"
// * layout_data = layout.image(src)
//
// the src path is under script directory, typically ELA/.
// for example, if src is "image/file.png" then
// the image is loaded from "ELA/image/file.png".
func singleImageLayout(L *lua.LState) int {
	src_path := checkFilePath(L, 1)
	img := attr.NewSingleImage(src_path)
	L.Push(newUserData(L, img))
	return 1
}

const nothingLayoutDataMessage = "require layout data more than or equal 1"

func checkMultipleLayoutData(L *lua.LState) []*attr.LayoutData {
	n := L.GetTop()
	if n <= 0 {
		raiseErrorf(L, nothingLayoutDataMessage)
		return nil
	}
	lds := make([]*attr.LayoutData, 0, n)
	for i := 1; i <= n; i++ {
		ld := checkLayoutData(L, i)
		lds = append(lds, ld)
	}
	return lds
}

// +gendoc "Layout Module"
// * layout_data = layout.flowHorizontal(layout_data...)
func flowHorizontalLayout(L *lua.LState) int {
	lds := checkMultipleLayoutData(L)
	flowH := attr.NewFlowHorizontal(lds...)
	L.Push(newUserData(L, flowH))
	return 1
}

// +gendoc "Layout Module"
// * layout_data = layout.flowVertical(layout_data...)
func flowVerticalLayout(L *lua.LState) int {
	lds := checkMultipleLayoutData(L)
	flowV := attr.NewFlowVertical(lds...)
	L.Push(newUserData(L, flowV))
	return 1
}

// +gendoc "Layout Module"
// * layout_data = layout.fixedSplit(edge, size, first_child, second_child)
func fixedSplitLayout(L *lua.LState) int {
	eStr := L.CheckString(1)
	var edge attr.Edge
	switch eStr {
	case "l", "left":
		edge = attr.EdgeLeft
	case "r", "right":
		edge = attr.EdgeRight
	case "t", "top":
		edge = attr.EdgeTop
	case "b", "bottom":
		edge = attr.EdgeBottom
	default:
		L.ArgError(1, "unknown edge: "+eStr)
	}

	size := L.CheckInt(2)

	first := attr.WithParentValue(checkLayoutData(L, 3), size)
	second := checkLayoutData(L, 4)
	fixed := attr.NewFixedSplit(edge, first, second)
	L.Push(newUserData(L, fixed))
	return 1
}

// +gendoc "Layout Module"
// * layout_data = layout.withValue(layout_data, flow_weight)
func withLayoutValue(L *lua.LState) int {
	ld := checkLayoutData(L, 1)
	v := L.CheckInt(2) // because only Flow uses this value, so limited to int.
	newLd := attr.WithParentValue(ld, v)
	L.Push(newUserData(L, newLd))
	return 1
}
