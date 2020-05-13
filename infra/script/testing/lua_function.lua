local function should_error_code(code)
  local fn = assert(loadstring(code, "should_error_code"))
  local ok = pcall(fn)
  assert(not ok, string.format("the code should get error but not, code: %s", code))
end
  
should_error_code [[io.open("anonymous.file")]]
should_error_code [[os.open("anonymous.file")]]
should_error_code [[_G["io"].open("anonymous.file")]]
should_error_code [[_G["os"].open("anonymous.file")]]

local fn = assert(loadstring [[ return arg_env.a + arg_env.b ]])
local load_env = {
  arg_env = {
    a = 1,
    b = 2
  }
}

fn = setfenv(fn, load_env)
assert(fn() == 3)
