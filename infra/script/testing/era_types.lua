
-- era types
assert(era.typeof(era.system.Number) == "IntParam")
assert(era.typeof(era.system.Str) == "StrParam")
assert(era.typeof(era.saveinfo) == "SaveInfo")
assert(era.typeof(era.chara) == "CharaList")
assert(era.typeof(era.master) == "CharaRefList")
assert(era.typeof(era.assi) == "CharaRefList")
assert(era.typeof(era.target) == "CharaRefList")
assert(era.typeof(era.player) == "CharaRefList")

local c = era.chara:addEmpty()
assert(era.typeof(c) == "Chara")

-- csv types
assert(era.typeof(era.csv.Number) == "CsvNames")
assert(era.typeof(era.csvindex.Number) == "CsvIndex")
assert(era.typeof(era.csvfields.Number) == "CsvFields")
assert(era.typeof(era.csvfields.Number.desc) == "CsvFieldStrings")

-- builtin types
assert(era.typeof(1) == "builtin")
assert(era.typeof("") == "builtin")
assert(era.typeof({}) == "builtin")
assert(era.typeof(function() end) == "builtin")
assert(era.typeof(nil) == "builtin")