module github.com/mzki/erago

go 1.19

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/golang/mock v1.6.0
	github.com/mattn/go-runewidth v0.0.14
	github.com/ugorji/go/codec v1.2.7
	github.com/yuin/gopher-lua v0.0.0-20220504180219-658193537a64
	// I could not upgrade exp/shiny package from this versoin.
	// By upgrading exp package from this version, exp/shiny adds go.mod then
	// the versions of exp and exp/shiny are conflicts on gomobile which requires exp@2019.
	golang.org/x/exp v0.0.0-20220203164150-d4f80a91470e
	golang.org/x/image v0.9.0
	golang.org/x/mobile v0.0.0-20221110043201-43a038452099
	golang.org/x/sys v0.5.0
	golang.org/x/text v0.11.0
)

require github.com/fsnotify/fsnotify v1.6.0

require (
	dmitri.shuralyov.com/gpu/mtl v0.0.0-20201218220906-28db891af037 // indirect
	github.com/BurntSushi/xgb v0.0.0-20160522181843-27f122750802 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20221017161538-93cebf72946b // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
)
