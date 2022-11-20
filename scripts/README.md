# scripts

This directory contains utility scripts for build, archive and release procedure.

## Condition

* Every scripts are inteded for launch at repository root directory, e.g. launched by `bash scripts/cross`. 
* `*-on-docker` scripts are launch scripts on docker container to have consistency for build environment. 
  * Developer need to build docker images from `dockerfiles/*` before launch those scripts.