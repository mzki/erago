# Erago

Erago is a clone of the [Emuera](https://ja.osdn.net/projects/emuera/wiki/FrontPage), which is the platform 
for creating and playing a text-based simulation game.
Erago is written by pure [Go](http://golang.org) and you can run Erago on multi platforms, Windows, Linux and MacOS!

Erago is a unofficial product. Please don't ask any issues about Erago to the maintainers of the original software.

## Motivation

* To use powerful script language (relative to the original one) to develop more complex game.
* To learn how to build Desktop application on Go. 

## Features 

**Erago is not compatibible with the original in some specification.**

* Use Lua5.1 as scripting language
* Resizable window
* Single binary and Cross platforms

## Getting started

If you just play the Erago, you can download pre-build binaries from [Release page (not yet)](#).
If you are developer or want to build manually, see next section.

### Starting Game 

To start Game
1. prepare extra resources, user scripts on `ELA` and game data schema table on `CSV`.
2. run the executable binary (click .exe file on Windows) on direcotry with above.

## Build binary from source

Requires `Go 1.15` or higher.

You can use `go get` to fetch the repository.

```sh
go get -u github.com/mzki/erago
```

Prepare dependencies for building.

```sh
cd $GOPATH/src/github.com/mzki/erago
go mod download
```

Build binary.

```sh
cd $GOPATH/src/github.com/mzki/erago/app/cmd
make linux # for windows use "windows" instead of "linux"
```

## For Mobile development

**Supports Android only.**

Because UI for Mobile and Desktop are optimized for each platform, 
Erago for mobile supplies only Model-level library, containing the script runtime and IO ports, not containing any UI.
Each mobile platform might implements UI with the Model library.

### Build model library for mobile.

To build Model-level library for mobile, `gomobile` and `Android NDK` are required.
Once prepared the build environment, run below to build the Model library.

```sh
cd $GOPATH/src/github.com/mzki/erago/mobile/model
bash build.sh
```

You can prepare the build environment using Docker and Dockerfile.
See `erago/dockerfiles/mobile/Dockerfile`. **Warning: This docker image is large, up to 3.6GB** 


## License

Erago is licensed under BSD 3-Clause license, the same as Go. See LICENSE file.

