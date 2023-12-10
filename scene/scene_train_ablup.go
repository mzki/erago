package scene

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/mzki/erago/state/csv"
)

// TRAIN SCENE
type trainScene struct {
	sceneCommon

	command_names  csv.Names
	command_ables  []bool
	command_format string

	// whether can do train command?
	//
	// it is used for controling of external execution of DoTrain().
	can_do_train bool

	// whether user shows other commands?
	user_shown_other_commands bool
}

var errorExternalDoTrainNotAllowed = errors.New("It is allowed only in train scene")

func newTrainScene(sf *sceneFields) *trainScene {
	// detect fmt for print command.
	cmd_names := sf.State().CSV.MustConst(csv.BuiltinTrainName).Names
	cmd_n := cmd_names.Len()
	width := math.Floor(math.Log10(float64(cmd_n))) + 1
	cmd_fmt := `[%` + strconv.Itoa(int(width)) + `d] %s`

	return &trainScene{
		sceneCommon:    newSceneCommon(SceneNameTrain, sf),
		command_names:  cmd_names,
		command_ables:  make([]bool, cmd_n),
		command_format: cmd_fmt,
	}
}

func (ts *trainScene) Name() string { return SceneNameTrain }

func (ts *trainScene) setCanDoTrain(ok bool) { ts.can_do_train = ok }

// +scene: train
// 調教のシーンです。
// ここでは、調教コマンドの実行およびコマンドの結果の反映を行うことを想定しています。
const (
	// +callback: {{.Name}}()
	// 調教対象のステータスの表示をこの関数で行います。
	ScrTrainUserShowStatus = "train_user_show_status"

	// +callback: ok: boolean = {{.Name}}(input_num: integer)
	// 選択番号input_numに対応するコマンドが、現在実行可能であるかを
	// true/falseで返します。ここで実行可能であったコマンド群が、
	// 画面に表示され、train_user_cmd()が実行されます。
	ScrTrainReplaceCmdAble = "train_replace_cmd_able"

	// +callback: {{.Name}}()
	// ユーザー定義のコマンド群の表示をこの関数で行います。
	ScrTrainReplaceShowOtherCmd = "train_replace_show_other_cmd"

	// +callback: handled: boolean = {{.Name}}(input_num: integer)
	// 選択番号input_numに対応するコマンドを実行します。
	// コマンドを処理した場合、trueを返してください。
	// trueを返した場合、train_user_check_source()を実行し、
	// 定義されていれば train_event_cmd_end()が実行されます。
	ScrTrainUserCmd = "train_user_cmd"

	// +callback: handled: boolean = {{.Name}}(input_num: integer)
	// 通常のコマンドが実行不能のとき代わりに呼ばれます。
	// 選択番号input_numに対応するユーザー定義のコマンドを実行します。
	// コマンドを処理した場合、trueを返してください。
	// trueを返した場合、train_user_check_source()を実行し、
	// 定義されていれば train_event_cmd_end()が実行されます。
	ScrTrainUserOtherCmd = "train_user_other_cmd"

	// +callback: {{.Name}}(input_num)
	// train_user_cmd()または、train_user_other_cmd()でtrueを返したとき、
	// 実行したコマンド番号input_numとともに、この関数が呼ばれます。
	// ここでは、コマンドの実行によって入手したSource変数の値を、
	// 調教対象のパラメータに変換します。
	ScrTrainUserCheckSource = "train_user_check_source"

	// +callback: {{.Name}}(input_num)
	// 選択番号input_numに対応するコマンドが実行された後に、呼びだされます。
	// ここで、調教に対する口上の表示を行います。
	ScrTrainEventTrainCmdEnd = "train_event_cmd_end"
)

// Go To Next Scene
func (ts *trainScene) Next() (Scene, error) {
	if next, err := ts.atStart(); next != nil || err != nil {
		return next, err
	}
	return ts.trainCycle()
}

// main cycle of train scene.
func (ts *trainScene) trainCycle() (Scene, error) {
	ts.setCanDoTrain(true)
	defer ts.setCanDoTrain(false)

	script := ts.Script()

	for !ts.Scenes().HasNext() {

		// show status for train target.
		if err := script.cautionCall(ScrTrainUserShowStatus); err != nil {
			return nil, err
		}

		// check commands are available?
		if err := ts.CheckAllTrainCommands(); err != nil {
			return nil, err
		}

		// show train menus.
		if err := ts.showTrainCommands(); err != nil {
			return nil, err
		}
		userShownOtherCommands, err := ts.showOtherTrainCommands()
		if err != nil {
			return nil, err
		}
		ts.user_shown_other_commands = userShownOtherCommands

		// get user command
		num, err := ts.IO().CommandNumber()
		if err != nil {
			return nil, err
		}

		// executes command
		if err := ts.DoTrain(int64(num)); err != nil {
			return nil, err
		}
	}
	return ts.Scenes().Next(), nil
}

// check weather all train commands can be performed?
// checked result is stored into trainScene to use later.
func (ts *trainScene) CheckAllTrainCommands() error {
	for cmd_no, name := range ts.command_names {
		if err := ts.CheckTrainCommand(cmd_no, name); err != nil {
			return err
		}
	}
	return nil
}

// check weather specified train command can be performed?
// checked result is stored into trainScene to use later.
func (ts *trainScene) CheckTrainCommand(cmd_no int, name string) error {
	if len(name) == 0 {
		ts.command_ables[cmd_no] = false
		return nil
	}

	ret, err := ts.Script().checkCallBoolArgInt(ScrTrainReplaceCmdAble, int64(cmd_no))
	if ret.Called {
		ts.command_ables[cmd_no] = ret.Return
	} else {
		ts.command_ables[cmd_no] = true // always true on builtin flow
	}
	return err
}

func (ts *trainScene) showTrainCommands() error {
	maxRuneWidth, err := ts.IO().WindowRuneWidth()
	if err != nil {
		return err
	}
	entryColumn := maxRuneWidth / DefaultPrintCWidth

	n := 0
	for cmd_no, name := range ts.command_names {
		if !ts.command_ables[cmd_no] {
			continue
		}

		ts.IO().PrintC(fmt.Sprintf(ts.command_format, cmd_no, name), DefaultPrintCWidth)
		n += 1
		// insert return code to format line.
		if n == entryColumn {
			n = 0
			ts.IO().PrintL("")
		}
	}
	// not end by "\n", so add it.
	if n != 0 {
		ts.IO().PrintL("")
	}
	return nil
}

func (ts *trainScene) showOtherTrainCommands() (userCalled bool, err error) {
	called, err := ts.Script().checkCall(ScrTrainReplaceShowOtherCmd)
	if err != nil {
		return
	}

	// buitin flow
	if !called {
		io := ts.IO()
		io.PrintL("")
		io.PrintC("[-1] "+DefaultOrString("Back", ts.ReplaceText().ReturnMenu), DefaultPrintCWidth)
		io.PrintL("")
	}
	return called, err
}

// do train which is given by command No.
func (ts *trainScene) DoTrain(cmd_no int64) error {
	if !ts.can_do_train {
		return errorExternalDoTrainNotAllowed
	}

	script := ts.Script()

	// Actual Execution of Train Command
	var cmd_handled bool
	var err error
	if ts.isExecutable(int(cmd_no)) {
		// do train command
		cmd_handled, err = script.cautionCallBoolArgInt(ScrTrainUserCmd, cmd_no)
	} else if ts.user_shown_other_commands {
		// do user defined other command
		cmd_handled, err = script.cautionCallBoolArgInt(ScrTrainUserOtherCmd, cmd_no)
	} else {
		// do builtin other command
		if cmd_no == -1 {
			// select `[-1] Back`
			err = ts.Scenes().SetNextByName(SceneNameTrainEnd)
			cmd_handled = false // indeed command handled but set false to stop rest of execution.
		}
	}
	if !cmd_handled || err != nil {
		return err
	}

	// check soruce for executing result of a command.
	_, err = script.cautionCallBoolArgInt(ScrTrainUserCheckSource, cmd_no)
	if err != nil {
		return err
	}
	// event command end.
	_, err = script.maybeCallBoolArgInt(ScrTrainEventTrainCmdEnd, cmd_no)
	return err
}

func (ts *trainScene) isExecutable(cmd int) bool {
	if 0 <= cmd && cmd < len(ts.command_ables) {
		return ts.command_ables[cmd]
	}
	return false
}

// * TRAIN END SCENE
type trainEndScene struct {
	sceneCommon
}

func newTrainEndScene(sf *sceneFields) *trainEndScene {
	return &trainEndScene{newSceneCommon(SceneNameTrainEnd, sf)}
}

func (scene *trainEndScene) Name() string { return SceneNameTrainEnd }

// +scene: trainend
// 調教終了時の後処理シーンです。
// 終了時のデータの後片付けなどを行うことを想定しています。
const (
// ScrSceneTrainEnd = "scene_trainend"
// ScrEventTrainEnd = "event_trainend"
)

func (scene *trainEndScene) Next() (Scene, error) {
	if next, err := scene.atStart(); next != nil || err != nil {
		return next, err
	}

	if ss := scene.Scenes(); ss.HasNext() {
		return ss.Next(), nil
	} else {
		return ss.GetScene(SceneNameAblUp)
	}
}

// ABLUP SCENE
type ablUpScene struct {
	sceneCommon
}

func newAblUpScene(sf *sceneFields) *ablUpScene {
	return &ablUpScene{newSceneCommon(SceneNameAblUp, sf)}
}

func (aus *ablUpScene) Name() string { return SceneNameAblUp }

// +scene: ablup
// 能力上昇のシーンです。
// Trainシーンの実行結果を、成長の結果として反映することを想定しています。
const (
	// +callback: {{.Name}}()
	ScrAblUpUserShowJuel = "ablup_user_show_juel"

	// +callback: {{.Name}}()
	ScrAblUpUserShowMenu = "ablup_user_show_menu"

	// +callback: ok: boolean = {{.Name}}(input_num: integer)
	ScrAblUpUserMenuSelected = "ablup_user_menu_selected" // +number -> bool
)

func (aus *ablUpScene) Next() (Scene, error) {
	if next, err := aus.atStart(); next != nil || err != nil {
		return next, err
	}

	script := aus.Script()
	for !aus.Scenes().HasNext() {
		if err := script.cautionCall(ScrAblUpUserShowJuel); err != nil {
			return nil, err
		}

		if err := script.cautionCall(ScrAblUpUserShowMenu); err != nil {
			return nil, err
		}

		if err := aus.inputLoop(); err != nil {
			return nil, err
		}
	}
	return aus.Scenes().Next(), nil
}

func (aus *ablUpScene) inputLoop() error {
	for {
		input, err := aus.IO().CommandNumber()
		if err != nil {
			return err
		}

		used, err := aus.Script().cautionCallBoolArgInt(ScrAblUpUserMenuSelected, int64(input))
		if used || err != nil {
			return err
		}
	}
}

// * TURN END SCENE
type turnEndScene struct {
	sceneCommon
}

func newTurnEndScene(sf *sceneFields) *turnEndScene {
	return &turnEndScene{newSceneCommon(SceneNameTurnEnd, sf)}
}

func (tes *turnEndScene) Name() string { return SceneNameTurnEnd }

// +scene: turnend
// 1ターンの終了シーンです。
// ここでは、ゲームとしての１ターンが終了したときの処理を実施することを想定しています。
const (
// ScrSceneTurnEnd = "scene_turnend"
// ScrEventTurnEnd = "event_turnend"
)

func (tes *turnEndScene) Next() (Scene, error) {
	if next, err := tes.atStart(); next != nil || err != nil {
		return next, err
	}

	if ss := tes.Scenes(); ss.HasNext() {
		return ss.Next(), nil
	} else {
		return ss.GetScene(SceneNameAutosave)
	}
}
