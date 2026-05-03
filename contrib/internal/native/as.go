package native

import "reflect"

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
