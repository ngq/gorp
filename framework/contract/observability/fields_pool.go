// Package observability provides framework-wide logging, metrics, tracing, and health contracts.
package observability

import "sync"

var smallFieldsPool = sync.Pool{
	New: func() any {
		s := make([]Field, 0, 8)
		return &s
	},
}

var mediumFieldsPool = sync.Pool{
	New: func() any {
		s := make([]Field, 0, 32)
		return &s
	},
}

// AcquireFields borrows a Field slice with capacity from the sync pool.
// The returned slice has length 0 and capacity >= capacity.
//
// AcquireFields 从对象池获取一个初始长度为 0 的 Field 切片。
func AcquireFields(capacity int) []Field {
	if capacity <= 8 {
		ptr := smallFieldsPool.Get().(*[]Field)
		return (*ptr)[:0]
	}
	if capacity <= 32 {
		ptr := mediumFieldsPool.Get().(*[]Field)
		return (*ptr)[:0]
	}
	return make([]Field, 0, capacity)
}

// ReleaseFields returns a Field slice back to the sync pool for reuse.
//
// ReleaseFields 将 Field 切片归还到对象池。
func ReleaseFields(fields []Field) {
	capSize := cap(fields)
	if capSize == 8 {
		fields = fields[:0]
		smallFieldsPool.Put(&fields)
	} else if capSize == 32 {
		fields = fields[:0]
		mediumFieldsPool.Put(&fields)
	}
}

// FieldOf creates a structured log Field without heap escape when passed by value.
func FieldOf(key string, val any) Field {
	return Field{Key: key, Value: val}
}
