package scene

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"local/erago/flow"
	"local/erago/state/csv"
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

func (ts trainScene) Name() string { return SceneNameTrain }

func (ts *trainScene) setCanDoTrain(ok bool) { ts.can_do_train = ok }

// +scene: train
const (
	// +callback: {{.Name}}()
	// 調教対象のステータスの表示をこの関数で行います。
	ScrTrainShowStatus = "train_show_status"

	// +callback: ok = {{.Name}}(input_num)
	// 選択番号input_numに対応するコマンドが、現在実行可能であるかを
	// true/falseで返します。ここで実行可能であったコマンド群が、
	// 画面に表示され、train_com()が実行されます。
	ScrTrainCmdAble = "train_com_able"

	// +callback: {{.Name}}()
	// ユーザー定義のコマンド群の表示をこの関数で行います。
	ScrTrainShowUserCmd = "train_show_user_com"

	// +callback: handled = {{.Name}}(input_num)
	// 選択番号input_numに対応するコマンドを実行します。
	// コマンドを処理した場合、trueを返してください。
	// trueを返した場合、train_check_source()を実行し、
	// 定義されていればevent_train_com_end()が実行されます。
	ScrTrainCmd = "train_com"

	// +callback: handled = {{.Name}}(input_num)
	// 通常のコマンドが実行不能のとき代わりに呼ばれます。
	// 選択番号input_numに対応するユーザー定義のコマンドを実行します。
	// コマンドを処理した場合、trueを返してください。
	// trueを返した場合、train_check_source()を実行し、
	// 定義されていればevent_train_com_end()が実行されます。
	ScrTrainUserCmd = "train_user_com"

	// +callback: {{.Name}}(input_num)
	// train_com()または、train_user_com()でtrueを返したとき、
	// 実行したコマンド番号input_numとともに、この関数が呼ばれます。
	// ここでは、コマンドの実行によって入手したSource変数の値を、
	// 調教対象のパラメータに変換します。
	ScrTrainCheckSource = "train_check_source"

	// +callback: {{.Name}}(input_num)
	// 選択番号input_numに対応するコマンドが実行された後に、呼びだされます。
	// ここで、調教に対する口上の表示を行います。
	ScrEventTrainCmdEnd = "event_train_com_end"
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
		if err := script.cautionCall(ScrTrainShowStatus); err != nil {
			return nil, err
		}

		// check commands are available?
		if err := ts.CheckAllTrainCommands(); err != nil {
			return nil, err
		}

		// show train menus.
		ts.showTrainCommands()
		if err := script.maybeCall(ScrTrainShowUserCmd); err != nil {
			return nil, err
		}

		// get user command
		num, err := ts.Game().CommandNumber()
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

	ok, err := ts.Script().maybeCallBoolArgInt(ScrTrainCmdAble, int64(cmd_no))
	ts.command_ables[cmd_no] = ok
	return err
}

func (ts trainScene) showTrainCommands() {
	entryColumn := ts.Game().MaxRuneWidth() / flow.DefaultPrintCWidth
	n := 0
	for cmd_no, name := range ts.command_names {
		if !ts.command_ables[cmd_no] {
			continue
		}

		ts.Game().PrintC(fmt.Sprintf(ts.command_format, cmd_no, name), flow.DefaultPrintCWidth)
		n += 1
		// insert return code to format line.
		if n == entryColumn {
			n = 0
			ts.Game().PrintL("")
		}
	}
	// not end by "\n", so add it.
	if n != 0 {
		ts.Game().PrintL("")
	}
}

// do train which is given by command No.
func (ts trainScene) DoTrain(cmd_no int64) error {
	if !ts.can_do_train {
		return errorExternalDoTrainNotAllowed
	}

	script := ts.Script()

	// Actual Execution of Train Command
	var cmd_handled bool
	var err error
	if ts.isExecutable(int(cmd_no)) {
		cmd_handled, err = script.cautionCallBoolArgInt(ScrTrainCmd, cmd_no)
	} else {
		cmd_handled, err = script.maybeCallBoolArgInt(ScrTrainUserCmd, cmd_no)
	}
	if !cmd_handled || err != nil {
		return err
	}

	// check soruce for executing result of a command.
	_, err = script.cautionCallBoolArgInt(ScrTrainCheckSource, cmd_no)
	if err != nil {
		return err
	}
	// event command end.
	_, err = script.maybeCallBoolArgInt(ScrEventTrainCmdEnd, cmd_no)
	return err
}

func (ts trainScene) isExecutable(cmd int) bool {
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

func (scene trainEndScene) Name() string { return SceneNameTrainEnd }

// +scene: trainend
const (
// ScrSceneTrainEnd = "scene_trainend"
// ScrEventTrainEnd = "event_trainend"
)

func (scene trainEndScene) Next() (Scene, error) {
	if next, err := scene.atStart(); next != nil || err != nil {
		return next, err
	}
	return scene.Scenes().GetScene(SceneNameTurnEnd)
}

// ABLUP SCENE
type ablUpScene struct {
	sceneCommon
}

func newAblUpScene(sf *sceneFields) *ablUpScene {
	return &ablUpScene{newSceneCommon(SceneNameAblUp, sf)}
}

func (aus ablUpScene) Name() string { return SceneNameAblUp }

// +scene: ablup
const (
	// +callback: {{.Name}}()
	ScrShowAblUpJuel = "show_ablup_juel"

	// +callback: {{.Name}}()
	ScrShowAblUpMenu = "show_ablup_menu"

	// +callback: ok = {{.Name}}(input_num)
	ScrAblUpMenuSelected = "ablup_menu_selected" // +number -> bool
)

func (aus ablUpScene) Next() (Scene, error) {
	if next, err := aus.atStart(); next != nil || err != nil {
		return next, err
	}

	script := aus.Script()
	for !aus.Scenes().HasNext() {
		if err := script.cautionCall(ScrShowAblUpJuel); err != nil {
			return nil, err
		}

		if err := script.cautionCall(ScrShowAblUpMenu); err != nil {
			return nil, err
		}

		if err := aus.inputLoop(); err != nil {
			return nil, err
		}
	}
	return aus.Scenes().Next(), nil
}

func (aus ablUpScene) inputLoop() error {
	for {
		input, err := aus.Game().CommandNumber()
		if err != nil {
			return err
		}

		used, err := aus.Script().cautionCallBoolArgInt(ScrAblUpMenuSelected, int64(input))
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

func (tes turnEndScene) Name() string { return SceneNameTurnEnd }

// +scene: turnend
const (
// ScrSceneTurnEnd = "scene_turnend"
// ScrEventTurnEnd = "event_turnend"
)

func (tes *turnEndScene) Next() (Scene, error) {
	if next, err := tes.atStart(); next != nil || err != nil {
		return next, err
	}
	return tes.Scenes().GetScene(SceneNameBase)
}
