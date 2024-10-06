assert(era.is_testing == true)

local tests = {}


function tests.no_flow_ctrl_func()
    local ok, msg = era.flow.pcall(function() return 1 * 2 - 3 / 4 end)
    assert(ok, msg)
    return true
end

function tests.other_error()
    local MARK_NEVER_REACH = "##-- never reach this --##"
    local err_msg = "##-- unknown error happened --##"
    local ok, msg = pcall(function()
        era.flow.pcall(function() error(err_msg) end)
        error(MARK_NEVER_REACH)
    end)
    assert(not ok)
    local found = string.find(msg, err_msg, nil, true)
    assert(found)
    return true
end

function tests.flow_goto_nextscene()
    local ok, msg = era.flow.pcall(era.flow.gotoNextScene, "title")
    assert(ok == false)
    assert(msg == "flow.gotoNextScene")
    return true
end

function tests.flow_longReturn()
    local ok, msg = era.flow.pcall(era.flow.longReturn)
    assert(ok == false, "falild by: "..msg)
    assert(msg == "flow.longReturn")
    return true
end

function tests.flow_quit()
    local ok, msg = era.flow.pcall(era.flow.quit)
    assert(ok == false)
    assert(msg == "flow.quit")
    return true
end

function tests.nested_flow_quit()
    local ok, msg = era.flow.pcall(function()
        local ret = 1 * 2 + 3 / 4
        era.flow.quit()
        return ret
    end)
    assert(ok == false)
    assert(msg == "flow.quit")
    return true
end

--[[
-- it will fail tests by quitError is raised at Interpreter caller. Just for confirmation.
function tests.flow_quit_without_pcall()
    era.flow.quit()
end
--]]

for k, fn in pairs(tests) do
    local ok, msg = pcall(fn)
    if ok then 
        assert(msg == true)
    else
        error("At tests."..k.."(): "..msg)
    end
end
