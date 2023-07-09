# Devcontainer for Erago

Image is based on Mobile build environment which also includes desktop build capability. 
Note that it can be around 4GB size.

* `cd app/cmd && make linux` to build s linux desktop binary only. (`make windows` for windows)
* `bash scripts/cross -o <output-dir>` to build cross platform (only for windows, linux) binaries.
* `bash scripts/build-mobile-v2 <output-dir>` to build mobile artifacts. 