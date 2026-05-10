// Package gorp provides the root-package application startup surface for gorp framework.
// This file exposes the public version marker of the root package.
// Provides stable constant for build/version display for release tooling.
//
// Gorp 包提供 gorp 框架的根包层应用启动入口。
// 本文件暴露根包层的公共版本标记。
// 为发布工具和业务代码提供稳定的构建/版本展示常量。
package gorp

// Release is the current public root-package version marker.
// Release 是当前根包公开版本标记。
//
// It defaults to "dev" in source builds and can be overridden at release time if needed.
// 源码构建时默认值为 "dev"，发布时可按需覆盖。
const Release = "dev"
