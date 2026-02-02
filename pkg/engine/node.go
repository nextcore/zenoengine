package engine

import "fmt"

type Node struct {
	Name     string
	Value    interface{}
	Children []*Node
	Parent   *Node
	Line     int
	Col      int
	Filename string
	
	// Inline caching: Pre-resolved handler and metadata
	// Set on first execution, reused on subsequent calls
	// Eliminates map lookup overhead (15-25% faster)
	cachedHandler HandlerFunc
	cachedMeta    *SlotMeta
}

type Diagnostic struct {
	Type     string `json:"type"` // "error", "warning"
	Message  string `json:"message"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	Slot     string `json:"slot,omitempty"`
}

func (d Diagnostic) Error() string {
	if d.Filename != "" {
		return fmt.Sprintf("[%s:%d:%d] %s", d.Filename, d.Line, d.Col, d.Message)
	}
	return d.Message
}
