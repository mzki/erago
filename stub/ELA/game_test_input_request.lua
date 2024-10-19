-- This line should return immidiately
era.printl "game_test_input_request.lua"
local ONE_SEC = 1* 1000 * 1000 * 1000
local timeout = era.twait(ONE_SEC)
if timeout then 
  error("user input is timeout!!!")
else 
  era.printl "ok. some input returned"
end
