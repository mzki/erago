assert(era.is_testing == true)

local inputQ = era.inputQueue()
inputQ:append {"3", "four"}
inputQ:prepend {"0", "one", "two"}
assert(era.inputNum() == 0)
assert(era.input() == "one") 
assert(era.inputNum() == 2) -- infinite loop!!
assert(era.inputNum() == 3)
assert(era.input() == "four")