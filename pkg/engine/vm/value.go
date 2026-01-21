package vm

import (
	"encoding/binary"
	"fmt"
	"io"
)

type ValueType int

const (
	ValNil ValueType = iota
	ValBool
	ValNumber
	ValString
	ValList
	ValMap
	ValFunction
)

// Value represents any ZenoLang value in the VM.
// Designed to be compact and easily stored in an array (Stack).
//
// OWNERSHIP: Value owns its heap-allocated data (strings, lists, maps).
// MEMORY: Primitives (nil, bool, number) are stack-allocated.
//
//	Complex types use pointers for heap allocation.
type Value struct {
	Type ValueType

	// Primitive types (stack-allocated)
	numVal  float64
	boolVal bool

	// Complex types (heap-allocated pointers)
	stringVal   *string
	listVal     *[]Value
	mapVal      *map[string]Value
	functionVal *Chunk
}

// Type-safe accessors

func (v Value) AsNumber() (float64, bool) {
	if v.Type == ValNumber {
		return v.numVal, true
	}
	return 0, false
}

func (v Value) AsBool() (bool, bool) {
	if v.Type == ValBool {
		return v.boolVal, true
	}
	return false, false
}

func (v Value) AsString() (string, bool) {
	if v.Type == ValString && v.stringVal != nil {
		return *v.stringVal, true
	}
	return "", false
}

func (v Value) AsList() ([]Value, bool) {
	if v.Type == ValList && v.listVal != nil {
		return *v.listVal, true
	}
	return nil, false
}

func (v Value) AsMap() (map[string]Value, bool) {
	if v.Type == ValMap && v.mapVal != nil {
		return *v.mapVal, true
	}
	return nil, false
}

func (v Value) AsFunction() (*Chunk, bool) {
	if v.Type == ValFunction && v.functionVal != nil {
		return v.functionVal, true
	}
	return nil, false
}

// String returns a string representation for debugging
func (v Value) String() string {
	switch v.Type {
	case ValNil:
		return "nil"
	case ValBool:
		if v.boolVal {
			return "true"
		}
		return "false"
	case ValNumber:
		return fmt.Sprintf("%g", v.numVal)
	case ValString:
		if v.stringVal != nil {
			return *v.stringVal
		}
		return ""
	case ValList:
		if v.listVal != nil {
			return fmt.Sprintf("List[%d items]", len(*v.listVal))
		}
		return "List[0 items]"
	case ValMap:
		if v.mapVal != nil {
			return fmt.Sprintf("Map[%d keys]", len(*v.mapVal))
		}
		return "Map[0 keys]"
	case ValFunction:
		return "Function"
	default:
		return "unknown"
	}
}

// ToNative converts a VM Value back to a Go native type (interface{})
// This is needed for compatibility with engine.Scope and slots
func (v Value) ToNative() interface{} {
	switch v.Type {
	case ValNil:
		return nil
	case ValBool:
		return v.boolVal
	case ValNumber:
		return v.numVal
	case ValString:
		if v.stringVal != nil {
			return *v.stringVal
		}
		return ""
	case ValList:
		if v.listVal != nil {
			// Convert []Value to []interface{}
			result := make([]interface{}, len(*v.listVal))
			for i, item := range *v.listVal {
				result[i] = item.ToNative()
			}
			return result
		}
		return []interface{}{}
	case ValMap:
		if v.mapVal != nil {
			// Convert map[string]Value to map[string]interface{}
			result := make(map[string]interface{})
			for k, val := range *v.mapVal {
				result[k] = val.ToNative()
			}
			return result
		}
		return map[string]interface{}{}
	case ValFunction:
		return v.functionVal
	default:
		return nil
	}
}

// Helper constructors

func NewNumber(n float64) Value {
	return Value{Type: ValNumber, numVal: n}
}

func NewBool(b bool) Value {
	return Value{Type: ValBool, boolVal: b}
}

func NewString(s string) Value {
	return Value{Type: ValString, stringVal: &s}
}

func NewNil() Value {
	return Value{Type: ValNil}
}

func NewList(items []Value) Value {
	return Value{Type: ValList, listVal: &items}
}

func NewMap(m map[string]Value) Value {
	return Value{Type: ValMap, mapVal: &m}
}

func NewFunction(c *Chunk) Value {
	return Value{Type: ValFunction, functionVal: c}
}

// NewValue creates a Value from any native Go type
// This is the bridge from Go interface{} to typed Value
func NewValue(v interface{}) Value {
	if v == nil {
		return NewNil()
	}
	switch val := v.(type) {
	case float64:
		return NewNumber(val)
	case int:
		return NewNumber(float64(val))
	case int64:
		return NewNumber(float64(val))
	case int32:
		return NewNumber(float64(val))
	case bool:
		return NewBool(val)
	case string:
		return NewString(val)
	case []interface{}:
		// Convert to []Value
		items := make([]Value, len(val))
		for i, item := range val {
			items[i] = NewValue(item)
		}
		return NewList(items)
	case map[string]interface{}:
		// Convert to map[string]Value
		m := make(map[string]Value)
		for k, item := range val {
			m[k] = NewValue(item)
		}
		return NewMap(m)
	case *Chunk:
		return NewFunction(val)
	default:
		// Last resort: try to handle as object
		// This maintains backward compatibility with ValObject
		return Value{Type: ValMap, mapVal: &map[string]Value{}}
	}
}

// Serialize writes a Value to a binary stream
func (v Value) Serialize(w io.Writer) error {
	// 1. Type
	if err := binary.Write(w, binary.LittleEndian, uint8(v.Type)); err != nil {
		return err
	}

	// 2. Data based on type
	switch v.Type {
	case ValNil:
		return nil
	case ValBool:
		var b byte
		if v.boolVal {
			b = 1
		}
		return binary.Write(w, binary.LittleEndian, b)
	case ValNumber:
		return binary.Write(w, binary.LittleEndian, v.numVal)
	case ValString:
		s := ""
		if v.stringVal != nil {
			s = *v.stringVal
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
			return err
		}
		_, err := w.Write([]byte(s))
		return err
	case ValList:
		list := []Value{}
		if v.listVal != nil {
			list = *v.listVal
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(len(list))); err != nil {
			return err
		}
		for _, item := range list {
			if err := item.Serialize(w); err != nil {
				return err
			}
		}
		return nil
	case ValMap:
		m := map[string]Value{}
		if v.mapVal != nil {
			m = *v.mapVal
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(len(m))); err != nil {
			return err
		}
		for k, val := range m {
			if err := writeString(w, k); err != nil {
				return err
			}
			if err := val.Serialize(w); err != nil {
				return err
			}
		}
		return nil
	case ValFunction:
		return fmt.Errorf("cannot serialize function chunks as constants")
	default:
		return fmt.Errorf("unknown value type: %d", v.Type)
	}
}

// DeserializeValue reads a Value from a binary stream
func DeserializeValue(r io.Reader) (Value, error) {
	var t uint8
	if err := binary.Read(r, binary.LittleEndian, &t); err != nil {
		return Value{}, err
	}

	vt := ValueType(t)
	switch vt {
	case ValNil:
		return NewNil(), nil
	case ValBool:
		var b byte
		if err := binary.Read(r, binary.LittleEndian, &b); err != nil {
			return Value{}, err
		}
		return NewBool(b > 0), nil
	case ValNumber:
		var n float64
		if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
			return Value{}, err
		}
		return NewNumber(n), nil
	case ValString:
		var l uint32
		if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
			return Value{}, err
		}
		buf := make([]byte, l)
		if _, err := io.ReadFull(r, buf); err != nil {
			return Value{}, err
		}
		return NewString(string(buf)), nil
	case ValList:
		var l uint32
		if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
			return Value{}, err
		}
		items := make([]Value, l)
		for i := uint32(0); i < l; i++ {
			item, err := DeserializeValue(r)
			if err != nil {
				return Value{}, err
			}
			items[i] = item
		}
		return NewList(items), nil
	case ValMap:
		var l uint32
		if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
			return Value{}, err
		}
		m := make(map[string]Value)
		for i := uint32(0); i < l; i++ {
			key, err := readString(r)
			if err != nil {
				return Value{}, err
			}
			val, err := DeserializeValue(r)
			if err != nil {
				return Value{}, err
			}
			m[key] = val
		}
		return NewMap(m), nil
	default:
		return Value{}, fmt.Errorf("unsupported value type: %d", vt)
	}
}

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func readString(r io.Reader) (string, error) {
	var l uint32
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return "", err
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
