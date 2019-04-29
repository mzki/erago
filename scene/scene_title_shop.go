package scene

import (
	"context"
	"fmt"

	"github.com/mzki/erago/state/csv"
)

// TITLE SCENE
type titleScene struct {
	sceneCommon
}

func newTitleScene(sf *sceneFields) *titleScene {
	return &titleScene{newSceneCommon(SceneNameTitle, sf)}
}

func (scene titleScene) Name() string { return SceneNameTitle }

// +scene: title
const (
	// +callback: {{.Name}}()
	// It replaces builtin flow for loading existance savefile at title scene.
	// It is called when user select "loading game" at title scene.
	// If you want something to do after loading savefile, it should
	// be done at loadend scene, not to use this callbask.
	//
	// titleシーンの"続きから..."を選択した後の、
	// セーブファイルを読み込む処理を、この関数によって置き換えます。
	// セーブファイルを読み込んだ後に何らかの処理をしたい場合には、
	// この関数を用いるべきではありません。代わりにloadendシーンで
	// そのような処理を行うべきです。
	ScrSystemTitleLoadGame = "system_title_loadgame"
)

func (scene *titleScene) Next() (Scene, error) {
	if next, err := scene.atStart(); next != nil || err != nil {
		return next, err
	}

	io := scene.IO()
	io.SetSingleLayout(io.GetCurrentViewName())

	scenes := scene.Scenes()
	csvGameBase := scene.State().CSV.GameBase

	for !scenes.HasNext() {
		io.NewPage()
		io.SetAlignment(AlignmentCenter)

		io.PrintLine(DefaultLineSymbol)
		io.PrintL(csvGameBase.Title)
		io.PrintL(fmt.Sprintf("ver %v", csvGameBase.Version))
		io.PrintL(csvGameBase.Author)
		io.PrintL(csvGameBase.AdditionalInfo)

		io.PrintLine(DefaultLineSymbol)
		io.PrintC("[0] 最初から始める", DefaultPrintCWidth)
		io.Print("\n\n")
		io.PrintC("[1] 続きから始める", DefaultPrintCWidth)
		io.Print("\n\n")
		io.PrintC("[9] 終了", DefaultPrintCWidth)
		io.PrintL("")

		io.SetAlignment(AlignmentLeft)

		input, err := io.CommandNumberSelect(context.Background(), 0, 1, 9)
		if err != nil {
			return nil, err
		}
		switch input {
		case 0:
			scene.State().Clear()
			return scenes.GetScene(SceneNameNewGame)

		case 1:
			if called, err := scene.Script().checkCall(ScrSystemTitleLoadGame); called {
				if err != nil {
					return nil, err
				}
				if next := scenes.Next(); next != nil {
					return next, nil
				}
				break
			}

			if next, err := scene.loadGame(); next != nil || err != nil {
				return next, err
			}

		case 9:
			return nil, ErrorQuit
		}
	}
	return scenes.Next(), nil
}

func (tc titleScene) loadGame() (Scene, error) {
	return loadGameSceneProcess(tc.sceneFields)
}

// NEW GAME SCENE
type newGameScene struct {
	sceneCommon
}

func newNewGameScene(sf *sceneFields) *newGameScene {
	return &newGameScene{newSceneCommon(SceneNameNewGame, sf)}
}

func (s newGameScene) Name() string { return SceneNameNewGame }

// +scene: newgame
const (
	// +callback: {{.Name}}()
	// newgameシーンで、保存される全てのデータを初期化した後に呼ばれます。
	// ここで、新しくゲームを始める為に必要なデータを設定することを想定しています。
	ScrEventNewGameInit = "event_newgame_init"
)

func (fs *newGameScene) Next() (Scene, error) {
	if next, err := fs.atStart(); next != nil || err != nil {
		return next, err
	}

	if err := fs.Script().maybeCall(ScrEventNewGameInit); err != nil {
		return nil, err
	}

	if sc := fs.Scenes(); sc.HasNext() {
		return sc.Next(), nil
	} else {
		return sc.GetScene(SceneNameAutosave)
	}
}

// BASE SCENE
type baseScene struct {
	sceneCommon
}

func newBaseScene(sf *sceneFields) *baseScene {
	return &baseScene{newSceneCommon(SceneNameBase, sf)}
}

func (bs baseScene) Name() string { return SceneNameBase }

// +scene: base
const (
	// +callback: {{.Name}}()
	// baseシーンにおける、行動の選択肢を表示します。
	ScrShowBaseMenu = "show_base_menu"

	// +callback: handled = {{.Name}}(input_num)
	// 行動の選択肢を表示した後、ユーザーからの入力番号を得て、
	// その入力番号input_numと共に、この関数が呼ばれます。
	// もし、入力番号input_numに対して何らかの処理を行った場合、
	// この関数の戻り値としてtrueを返してください。
	// その場合、次のシーンの遷移先が決まっていれば、遷移します。
	// 決まっていなければ、再び、選択肢の表示から繰り返します。
	// 戻り値としてfalseを返した場合、ユーザーの入力待ちから繰り返します。
	ScrBaseMenuSelected = "base_menu_selected"
)

func (bs baseScene) Next() (Scene, error) {
	if next, err := bs.atStart(); next != nil || err != nil {
		return next, err
	}

	// loop for showing and selecting base menu
	scenes := bs.Scenes()
	for !scenes.HasNext() {
		if err := bs.Script().cautionCall(ScrShowBaseMenu); err != nil {
			return nil, err
		}
		if err := bs.inputLoop(); err != nil {
			return nil, err
		}
	}
	return scenes.Next(), nil
}

func (bs baseScene) inputLoop() error {
	for {
		input, err := bs.IO().CommandNumber()
		if err != nil {
			return err
		}
		handled, err := bs.Script().cautionCallBoolArgInt(ScrBaseMenuSelected, int64(input))
		if handled || err != nil {
			return err
		}
	}
}

// * SHOP SCENE
type shopScene struct {
	sceneCommon

	// userShop is whether user shows shop items?
	userShop bool
}

func newShopScene(sf *sceneFields) *shopScene {
	return &shopScene{sceneCommon: newSceneCommon(SceneNameShop, sf)}
}

func (sc shopScene) Name() string { return SceneNameShop }

// +scene: shop
const (
	// +callback: {{.Name}}()
	// Itemの一覧を表示する処理を置き換える。
	// もし定義されていなければ、CSVで定義されたItemを
	// クリックできるボタンの形式で全て表示する。
	ScrSystemShowShopMenu = "system_show_shop_menu"

	// +callback: handled = {{.Name}}(input_num)
	// 入力番号input_numと共に呼ばれ、それに対する処理を行う。
	// もし、何らかの処理を行った場合、この関数の戻り値としてtrueを
	// 返してください。その場合、次のシーンの遷移先が決まっていれば、
	// 遷移します。決まっていなければ、再び、選択肢の表示から繰り返します。
	// 戻り値としてfalseを返した場合、ユーザーの入力待ちから繰り返します。
	ScrShopMenuSelected = "shop_menu_selected"
)

func (sc shopScene) Next() (Scene, error) {
	if next, err := sc.atStart(); next != nil || err != nil {
		return next, err
	}

	scenes := sc.Scenes()
	io := sc.IO()
	for !scenes.HasNext() {
		called, err := sc.Script().checkCall(ScrSystemShowShopMenu)
		if err != nil {
			return nil, err
		}
		if !called {
			if err := sc.ShowItems(DefaultShowItemFormat); err != nil {
				return nil, err
			}
			io.PrintLine(DefaultLineSymbol)
			io.PrintC("[-1] 戻る", DefaultPrintCWidth)
			io.PrintL("")
		}

		sc.userShop = called

		if err := sc.inputLoop(); err != nil {
			return nil, err
		}
	}
	return scenes.Next(), nil
}

func (sc shopScene) inputLoop() error {
	for {
		// get user input
		input, err := sc.IO().CommandNumber()
		if err != nil {
			return err
		}

		// handle user input in script.
		handled, err := sc.Script().cautionCallBoolArgInt(ScrShopMenuSelected, int64(input))
		if handled || err != nil {
			return err
		}

		// default item prints: [-1] 戻る
		if !sc.userShop && input == -1 {
			s := sc.Scenes()
			s.SetNext(s.Prev())
			return nil
		}
	}
}

const DefaultShowItemFormat = "[%d] %s (%d圓)"

func (sc shopScene) ShowItems(fmtStr string) error {
	if fmtStr == "" {
		fmtStr = DefaultShowItemFormat
	}
	CSV := sc.State().CSV
	itemNames := CSV.Item.Names
	itemPrices := CSV.ItemPrices
	itemSold, _ := sc.State().SystemData.GetInt(csv.BuiltinItemSoldName)

	// itemNames and ItemPrices must have same length.
	// but ItemSold does not.
	var maxLen = itemNames.Len()
	if maxLen > itemSold.Len() {
		maxLen = itemSold.Len()
	}

	io := sc.IO()
	maxRuneWidth, err := io.WindowRuneWidth()
	if err != nil {
		return err
	}
	nColumn := maxRuneWidth / DefaultPrintCWidth

	cc := 0 // current column
	for i, item := range itemNames[:maxLen] {
		if len(item) == 0 {
			continue
		}
		if itemSold.Get(i) < 1 {
			continue
		}

		text := fmt.Sprintf(fmtStr, i, item, itemPrices[i])
		io.PrintC(text, DefaultPrintCWidth)
		cc += 1
		if cc == nColumn {
			cc = 0
			io.PrintL("")
		}
	}
	if cc != 0 {
		io.PrintL("")
	}
	return nil
}
