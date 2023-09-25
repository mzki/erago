function era.testquit()
	era.flow.quit()
end

function era.testgoto() 
	era.flow.gotoNextScene "undefined"
end

function era.testlongreturn() 
	era.flow.longReturn()
end

--- pcall pass through special error for runtime caller

local flow = assert(era.flow)

function era.testpcall_gotoNextScene()
    local ok, msg = pcall(flow.gotoNextScene, "train")
    assert(false, "gotoNextScene never reach this")
end

function era.testpcall_longReturn()
    local ok, msg = pcall(flow.longReturn)
    assert(false, "longReturn never reach this")
end

function era.testpcall_quit()
    local ok, msg = pcall(flow.quit)
    assert(false, "quit never reach this")
end

local SOMETHING_WRONG_MSG = "script something wrong"

local function something_wrong()
	error(SOMETHING_WRONG_MSG)
end

function era.testpcall_something()
	local ok, msg = pcall(something_wrong)
	assert(not ok)
end

--- xpcall pass through special error for runtime caller

local function make_catch_error_func(funcName)
	return function (err)
		error(funcName .." never reach this")
	end
end

function era.testxpcall_gotoNextScene()
    local ok, msg = xpcall(function() flow.gotoNextScene "train" end, make_catch_error_func("gotoNextScene"))
    error("gotoNextScene never reach this")
end

function era.testxpcall_longReturn()
	local ok, msg = xpcall(flow.longReturn, make_catch_error_func("longReturn"))
    error("longReturn never reach this")
end

function era.testxpcall_quit()
    local ok, msg = xpcall(flow.quit, make_catch_error_func("quit"))
    error("quit never reach this")
end

function era.testxpcall_something()
	local SUFFIX = " at error handler"
	local ok, msg = xpcall(something_wrong, function(err) return (err .. SUFFIX) end)
	assert(not ok)
	-- era.printl("debug: " .. msg)
	assert(msg:find(SOMETHING_WRONG_MSG .. SUFFIX) ~= nil)
end

function era.testxpcall_something2()
	local SUFFIX = " at error handler"
	local ok, msg = xpcall(something_wrong, function(err) return (err .. SUFFIX), "dummy" end)
	assert(not ok)
	-- era.printl("debug: " .. msg)
	assert(msg:find(SOMETHING_WRONG_MSG .. SUFFIX) ~= nil)
end

function era.testxpcall_something_error_handler()
	local SUFFIX = " at error handler"
	local ok, msg = xpcall(something_wrong, function(err) error(err .. SUFFIX) end)
	assert(not ok)
	-- era.printl("debug: " .. msg)
	assert(msg:find(SOMETHING_WRONG_MSG .. SUFFIX) ~= nil)
end
