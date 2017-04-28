-- builtin functions

assert(ipairs)
assert(pairs)


-- function access

era.print "test"
era.printl "test"
era.printc "test"
era.printw "test"
era.printLine "test"
era.printBar(0, 10, 5)
_ = era.textBar(0, 10, 5)

era.clearLineAll()
era.clearLine(3)
era.newPage()
_ = era.maxStrWidth()
_ = era.currentStrWidth()
_ = era.lineCount()

local input = ""
input = era.input()
input = era.inputNum()
input = era.inputRange(0, 10)
input = era.inputSelect(0, 1)
era.wait()

era.setColor(0x00ffff)
era.getColor()
era.resetColor()
era.setAlignment "left"
era.getAlignment()

era.clearSystem()
era.saveSystem(1)
era.loadSystem(1)
era.clearShare()
era.saveShare()
era.loadShare()

era.flow.setNextScene "test"
era.flow.saveScene()
era.flow.loadScene()
era.flow.doTrains({1, 2})

era.layout.setCurrentView("test")
era.layout.getCurrentView()
era.layout.viewNames()  
era.layout.setSingle()    
era.layout.setVertical("top", "bottom")
era.layout.setHorizontal("left", "bottom")

local layout = era.layout
layout.setLayout(
	layout.flowVertical(
		layout.text("1"),
		layout.text("2"),
		layout.flowHorizontal(
			layout.withValue(layout.text("3"), 1), layout.withValue(layout.text("4"), 3)
		),
		layout.fixedSplit("top", 30, 
			layout.text("5"),
			layout.text("6")
		)
	)
)

-- data access

assert(era.system)
assert(era.share)

assert(era.csv.Train)
assert(era.csv.Item)
assert(era.csvindex.Train)
assert(era.csvindex.Item)
assert(era.csv.ItemPrice)

assert(era.chara)
assert(era.master)
assert(era.assi)
assert(era.target)
assert(era.player)

-- chara access

local chara = era.chara:add(1)
-- builtins
assert(chara.id)
assert(chara.uid)
assert(chara.is_assi)
assert(chara.name)
assert(chara.nick_name)
assert(chara.master_name)
assert(chara.call_name)

-- user defined variable
local key = "体力"
local hp = assert(chara.Base[key])
chara.Base[key] = hp + 300
assert(chara.Base[key] == hp + 300)

-- XXXParam
local base = chara.Base
local len = base:len()
base:set(key, 100)
assert(base:get(key) == 100)

local sliced_base = base:slice(0, 2)
sliced_base:fill(10)
assert(base:get(0) == 10)
assert(sliced_base:get(0) == 10)

local new_intparam = IntParam.new(100)
_ = new_intparam:len()
new_intparam:slice(0,10):fill(100)

-- check pairs loops infinity?
local table = {}
table[0] = 1
table["true"] = true
table["false"] = false
table[10] = 10
for _, v in pairs(table) do
	era.print(v)
end
