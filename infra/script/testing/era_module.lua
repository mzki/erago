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
_ = era.windowStrWidth()
_ = era.windowLineCount()
_ = era.currentStrWidth()
_ = era.lineCount()

_ = era.textWidth("こんにちは、世界")

local input = ""
input = era.input()
input = era.inputNum()
input = era.inputRange(0, 10)
input = era.inputSelect(0, 1)
era.wait()

local time_exceeded
local TIMEOUT = 1000000
time_exceeded = era.twait(TIMEOUT)
input, time_exceeded = era.tinput(TIMEOUT)
input, time_exceeded = era.tinputNum(TIMEOUT)

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

era.printl ""
assert(era.system)
assert(era.system.Number)
era.system.Number[0] = 10
assert(era.system.Number[0] == 10)
era.system.Number["数値１"] = 20
assert(era.system.Number["数値１"] == 20)
assert(era.system.Str)
era.system.Str[0] = "ABC"
assert(era.system.Str[0] == "ABC")
assert(era.share)

assert(era.csv.Train)
assert(era.csv.Item)
assert(era.csvindex.Train)
assert(era.csvindex.Item)
assert(era.csv.ItemPrice)

assert(era.csvfields.Item)
assert(era.csvfields.Item.price)
assert(era.csvfields.Item.price[0] == 200)
assert(era.csvfields.Number)
assert(era.csvfields.Number.desc)
assert(era.csvfields.Number.desc[0] == "こんにちは")
assert(#(era.csvfields.Number.desc) == 1)

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
assert(base[len-1] == 0)

local sliced_base = base:slice(0, 2)
sliced_base:fill(10)
assert(base:get(0) == 10)
assert(sliced_base:get(0) == 10)

local new_intparam = IntParam.new(100)
len = new_intparam:len()
new_intparam:slice(0,10):fill(100)
assert(new_intparam[9] == 100)
assert(new_intparam[10] == 0)
local ok, msg = pcall(IntParam.get, new_intparam, len)
assert(not ok)

-- pairs and ipairs
local nextop, t, val = ipairs(new_intparam:slice(0, 5))
assert(type(nextop) == "function")
assert(type(t) == "userdata")
assert(type(val) == "number")

local i, v = nextop(t, val) 
assert(i == 0)
assert(v == 100)
i, v = nextop(t, i) 
assert(i == 1)
assert(v == 100)

nextop, t, val = pairs(new_intparam:slice(0, 5))
assert(type(nextop) == "function")
assert(type(t) == "userdata")
assert(type(val) == "number")

i, v = nextop(t, val) 
assert(i == 0)
assert(v == 100)
i, v = nextop(t, i) 
assert(i == 1)
assert(v == 100)

-- character pairs
for i = 0, 1 do
	local added = era.chara:add(1)
	era.target[i] = added
	era.master[i] = added
	era.assi[i] = added
	era.player[i] = added
end

for i, list in ipairs {
	era.chara,
	era.target,
	era.master,
	era.assi,
	era.player,
} do
	local n = 0
	for j, c in ipairs(list) do
		n = n + 1
		assert(c, "current loop:"..j)
	end
	assert(n > 0, "current loop:"..i)

	n = 0
	for j, c in pairs(list) do
		assert(c, "current loop:"..j)
		n = n + 1
	end
	assert(n > 0, "current loop:"..i)
end


-- check pairs loops infinity?
local table = {}
table[0] = 1
table["true"] = true
table["false"] = false
table[10] = 10
for _, v in pairs(table) do
	era.print(v)
end
era.printl ""
