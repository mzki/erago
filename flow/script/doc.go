//go:generate  go run gendoc.go

package script

// +gendoc.set_section "Over View"

// +gendoc
// eragoでは、ゲームの動作をLua5.1というスクリプト言語を用いて記述していきます。
// 文字を出力するだけならば
//
//   era.print "こんにちは、世界"
//
// と記述し、ユーザーからの入力を受け付けるときには、
//
//   local user_input = era.input()
//
// のようにして結果をuser_inputで受け取ります。
