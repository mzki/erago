
local function N_loop(n) 
	local a = 1
	while a < n do
		a = a+1
	end
end

local N = 100000

function era.bench1() 
	N_loop(N)
end

function era.bench2() 
	N_loop(N)
end
