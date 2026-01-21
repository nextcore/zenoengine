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
	ValObject   // Interface{} / Map / Pointer
	ValFunction // [NEW] *Chunk
)

// Value represents any ZenoLang value in the VM.
// Designed to be compact and easily stored in an array (Stack).
type Value struct {
	Type  ValueType
	AsNum float64
	AsPtr interface{} // Used for strings and complex objects
}

// Serialize writes a Value to a binary stream.
func (v Value) Serialize(w io.Writer) error {
	// 1. Type
	if err := binary.Write(w, binary.LittleEndian, uint8(v.Type)); err != nil {
		return err
	}

	// 2. Data based on type
	switch v.Type {
	case ValNil:
		return nil
	case ValBool, ValNumber:
		return binary.Write(w, binary.LittleEndian, v.AsNum)
	case ValString:
		s := v.AsPtr.(string)
		if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
			return err
		}
		_, err := w.Write([]byte(s))
		return err
	case ValObject:
		return fmt.Errorf("cannot serialize complex objects yet")
	default:
		return fmt.Errorf("unknown value type")
	}
}

// DeserializeValue reads a Value from a binary stream.
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
		var n float64
		if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
			return Value{}, err
		}
		return NewBool(n > 0), nil
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
	default:
		return Value{}, fmt.Errorf("unsupported value type: %d", vt)
	}
}

func (v Value) String() string {
	switch v.Type {
	case ValNil:
		return "nil"
	case ValBool:
		if v.AsNum > 0 {
			return "true"
		}
		return "false"
	case ValNumber:
		return fmt.Sprintf("%g", v.AsNum)
	case ValString:
		return v.AsPtr.(string)
	case ValObject:
		return fmt.Sprintf("%v", v.AsPtr)
	default:
		return "unknown"
	}
}

// ToNative converts a VM Value back to a Go native type.
func (v Value) ToNative() interface{} {
	switch v.Type {
	case ValNil:
		return nil
	case ValBool:
		return v.AsNum > 0
	case ValNumber:
		return v.AsNum
	case ValString, ValObject, ValFunction:
		return v.AsPtr
	default:
		return nil
	}
}

// Helper constructors
func NewNumber(n float64) Value { return Value{Type: ValNumber, AsNum: n, AsPtr: n} }
func NewBool(b bool) Value {
	if b {
		return Value{Type: ValBool, AsNum: 1, AsPtr: true}
	}
	return Value{Type: ValBool, AsNum: 0, AsPtr: false}
}
func NewString(s string) Value      { return Value{Type: ValString, AsPtr: s} }
func NewNil() Value                 { return Value{Type: ValNil} }
func NewObject(o interface{}) Value { return Value{Type: ValObject, AsPtr: o} }
func NewFunction(c *Chunk) Value    { return Value{Type: ValFunction, AsPtr: c} }

// NewValue creates a Value from any native Go type, sniffing for numbers/strings/etc.
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
	case bool:
		return NewBool(val)
	case string:
		return NewString(val)
	case *Chunk:
		return NewFunction(val)
	default:
		return NewObject(val)
	}
}
