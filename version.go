// Application scenarios:
// - Expose the public version marker of the root package.
// - Give release tooling and business code one stable constant for build/version display.
// - Keep source-build and release-build version semantics explicit.
//
// 适用场景：
// - 暴露根包层的公共版本标记。
// - 为发布工具和业务代码提供稳定的构建/版本展示常量。
// - 显式表达源码构建与正式发布的版本语义。
package gorp

// Release is the current public root-package version marker.
// Release 是当前根包公开版本标记。
//
// It defaults to "dev" in source builds and can be overridden at release time if needed.
// 源码构建时默认值为 "dev"，发布时可按需覆盖。
const Release = "dev"
