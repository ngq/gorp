// Package sentinel 提供 Sentinel 熔断降级实现。
// 本文件内联了 native.As，消除对 contrib/internal 的依赖，
// 使本包成为可独立引用的模块。
package sentinel

import "reflect"

// As 尝试通过 reflect 将 source 转换为 target 指向的类型。
// 支持直接赋值、接口实现和类型转换三种路径。
// 当 target 为 nil、非指针、nil 指针或不可设置时返回 false。
func As(source any, target any) bool {
	if target == nil {
		return false
	}
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return false
	}
	sourceValue := reflect.ValueOf(source)
	if !sourceValue.IsValid() {
		return false
	}
	destination := targetValue.Elem()
	if !destination.CanSet() {
		return false
	}
	sourceType := sourceValue.Type()
	destinationType := destination.Type()
	if sourceType.AssignableTo(destinationType) {
		destination.Set(sourceValue)
		return true
	}
	if destinationType.Kind() == reflect.Interface && sourceType.Implements(destinationType) {
		destination.Set(sourceValue)
		return true
	}
	if sourceType.ConvertibleTo(destinationType) {
		destination.Set(sourceValue.Convert(destinationType))
		return true
	}
	return false
}
