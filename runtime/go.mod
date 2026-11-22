// 本目录是独立的嵌套模块：go get github.com/kui-engine/kui-go/v3 下载的模块 zip 不含有自己 go.mod 的子目录，
// 所以三平台运行库（会让模块 zip 接近或超过 Go 的 500 MiB 上限）只随 git clone 与 Release 上的 SDK 包分发。
// 本模块没有 Go 代码，也不需要被任何人 import。
// A nested module boundary: the module zip of github.com/kui-engine/kui-go/v3 (what go get downloads) excludes
// any subdirectory with its own go.mod, so the runtimes ship only with git clone and the SDK packages on the Release.
module github.com/kui-engine/kui-go/runtime

go 1.22
