assert(era.is_testing == true)
local inputQ = era.inputQueue()
assert(inputQ:size() == 0)

local nQ = 0
nQ = inputQ:append {"3", "4", "5", ""}
assert(inputQ:size() == nQ)
nQ = inputQ:prepend {"0", "1", "two"}
assert(inputQ:size() == nQ)

assert(era.inputNum() == 0)
assert(era.input() == "1")
assert(era.input() == "two")
assert(era.inputNum() == 3)
assert(era.inputRange(0, 6) == 4)
assert(era.inputSelect(5) == 5)
assert(era.input() == "")
-- empty
assert(inputQ:size() == 0)
assert(era.input() == "")

-- clear at middle
local _ = 0
_ = inputQ:prepend {"0", "one", "2"}
_ = era.inputNum()
inputQ:clear()
assert(inputQ:size() == 0)
assert(era.input() == "")