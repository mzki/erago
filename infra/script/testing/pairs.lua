local next_, t, i = ipairs(era.system.Number)
local j, v = next_(t, i)
assert(type(j) == "number")
assert(type(v) == "number")
i = j
j, v = next_(t, i) -- exceeds length
assert(type(j) == "nil")
assert(type(v) == "nil")

for i, v in ipairs(era.system.Number) do
    if type(i) ~= "number" then error(i.."-th element is not number") end
    assert(type(v) == "number")
end

local next_, t, i = ipairs(era.system.Str)
local j, v = next_(t, i)
assert(type(j) == "number")
assert(type(v) == "string")
i = j
j, v = next_(t, i)
assert(type(j) == "number")
assert(type(v) == "string")

for i, v in ipairs(era.system.Str) do
    if type(i) ~= "number" then error(i.."-th element is not number") end
    assert(type(v) == "string")
end

next_, t, i = ipairs(era.csv.Base)
j, v = next_(t, i)
assert(type(j) == "number")
assert(type(v) == "string")
i = j
j, v = next_(t, i)
assert(type(j) == "number")
assert(type(v) == "string")

for i, v in ipairs(era.csv.Base) do
    if type(i) ~= "number" then error(i.."-th element is not number") end
    assert(type(v) == "string")
end
