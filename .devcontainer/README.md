# Devcontainer for Erago

Image is used vscode go image to minimize image size. Since mobile build is optional, its build environment is created on demand by
`scripts/build-docker-image-mobile`, which image size can be around 4GB size.

* `cd app/cmd && make linux` to build s linux desktop binary only. (`make windows` for windows)
* `bash scripts/cross -o <output-dir>` to build cross platform (only for windows, linux) binaries.

For mobile build, need to use docker image mentioned at first. Once mobile build image is built, run 
`bash scripts/build-mobile-v2-on-docker <output-dir>` to build mobile artifacts. 