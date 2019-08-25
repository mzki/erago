--- bit module
local bit32 = require "bit32"

assert(bit32.band(1, 3) == 1)
assert(bit32.bor(1, 3) == 3)
assert(bit32.bxor(1, 3) == 2)
assert(bit32.lshift(1, 1) == 2)
assert(bit32.rshift(2, 1) == 1)
assert(bit32.set(0, 0) == 1)
assert(bit32.unset(2,1) == 0)
assert(bit32.get(3, 1) == 1)
assert(bit32.popcount(3) == 2)

--- log module
local log = require "log"
log.info "for information"
log.infof("for information, %s", "always outputted")
log.debug "for debug"
log.debugf("for debug, %s", "maybe not outputted")

-- time module
local time = require "time"
now = time.now()
now_t = time.now("*t")
assert(time.year(now) == now_t.year)
assert(time.month(now) ==now_t.month)
assert(time.day(now) == now_t.day)
assert(time.weekday(now) == now_t.wday)
assert(time.hour(now) == now_t.hour) -- maybe wrong by chance
_ = time.minute(now)
_ = time.second(now)
_ = now_t.minite
_ = now_t.second
_ = time.format(now)
_ = time.tostring(1*time.SECOND + 1*time.MILLISECOND + 1 *time.NANOSECOND)

-- csv module
-- DEPRECATED
--[[
local csv = require "csv"
csv.readFunc("csvdata.csv", function(i, records) 
  if i == 1 then
    assert(records[2] == "1")
  else 
    assert(records[2] == "2")
  end
end)
--]]

-- should fail.
local ok = pcall(require, "csv")
assert(not ok)

-- constants
assert(MAX_INTEGER)
assert(MAX_NUMBER)
assert(PRINTC_WIDTH)
assert(TEXTBAR_WIDTH)
assert(TEXTBAR_FG)
assert(TEXTBAR_BG)
assert(TEXTLINE_SYMBOL)

assert(_BASE_DIRECTORY)
assert(_DEBUG_ENABLE)
