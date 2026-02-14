package event

import (
	"reflect"
	"sync"
)

// History stores recent events in memory.
type History struct {
	mu      sync.RWMutex
	maxSize int
	events  []*Event
}

func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 1
	}
	return &History{
		maxSize: maxSize,
		events:  make([]*Event, 0, maxSize),
	}
}

func (h *History) Add(ev *Event) {
	if h == nil || ev == nil {
		return
	}
	cloned := cloneEvent(ev)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, cloned)
	if len(h.events) > h.maxSize {
		h.events = h.events[len(h.events)-h.maxSize:]
	}
}

// Recent returns recent events from newest to oldest.
func (h *History) Recent(limit int) []*Event {
	if h == nil {
		return nil
	}
	h.mu.RLock()
	if len(h.events) == 0 {
		h.mu.RUnlock()
		return nil
	}
	if limit <= 0 || limit > len(h.events) {
		limit = len(h.events)
	}
	selected := make([]*Event, 0, limit)
	for i := len(h.events) - 1; i >= len(h.events)-limit; i-- {
		selected = append(selected, h.events[i])
	}
	h.mu.RUnlock()

	out := make([]*Event, 0, len(selected))
	for _, ev := range selected {
		out = append(out, cloneEvent(ev))
	}
	return out
}

func cloneEvent(in *Event) *Event {
	if in == nil {
		return nil
	}
	out := *in
	if in.Data != nil {
		out.Data = make(map[string]interface{}, len(in.Data))
		for k, v := range in.Data {
			out.Data[k] = cloneValue(v)
		}
	}
	return &out
}

func cloneValue(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}
	return cloneValueReflect(rv).Interface()
}

func cloneValueReflect(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		return cloneValueReflect(v.Elem())
	case reflect.Pointer:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.New(v.Type().Elem())
		out.Elem().Set(cloneValueReflect(v.Elem()))
		return out
	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			// Keep map keys as-is to preserve key identity/semantics (especially pointer keys).
			out.SetMapIndex(iter.Key(), cloneValueReflect(iter.Value()))
		}
		return out
	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			out.Index(i).Set(cloneValueReflect(v.Index(i)))
		}
		return out
	case reflect.Array:
		out := reflect.New(v.Type()).Elem()
		for i := 0; i < v.Len(); i++ {
			out.Index(i).Set(cloneValueReflect(v.Index(i)))
		}
		return out
	case reflect.Struct:
		out := reflect.New(v.Type()).Elem()
		out.Set(v) // copy all fields first (including unexported)
		for i := 0; i < v.NumField(); i++ {
			field := out.Field(i)
			if !field.CanSet() {
				continue
			}
			field.Set(cloneValueReflect(v.Field(i)))
		}
		return out
	default:
		return v
	}
}
