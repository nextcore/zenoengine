package engine

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type ScriptCache struct {
	mu    sync.RWMutex
	files map[string]*CachedScript
}

type CachedScript struct {
	Root    *Node
	ModTime time.Time
}

var GlobalCache = &ScriptCache{files: make(map[string]*CachedScript)}

// LoadScript membaca dan memparsing file script, atau mengambil dari cache jika belum berubah
func LoadScript(path string) (*Node, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	GlobalCache.mu.RLock()
	cached, exists := GlobalCache.files[path]
	GlobalCache.mu.RUnlock()

	if exists && cached.ModTime.Equal(info.ModTime()) {
		return cached.Root, nil
	}

	root, err := parseFile(path)
	if err != nil {
		return nil, err
	}

	GlobalCache.mu.Lock()
	GlobalCache.files[path] = &CachedScript{Root: root, ModTime: info.ModTime()}
	GlobalCache.mu.Unlock()

	return root, nil
}

// ClearHandlerCache membersihkan semua cached handler dan metadata dari Node AST
// Fungsi ini harus dipanggil sebelum hot reload untuk mencegah panic akibat stale handlers
func (c *ScriptCache) ClearHandlerCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cached := range c.files {
		clearNodeCache(cached.Root)
	}
}

func clearNodeCache(node *Node) {
	if node == nil {
		return
	}
	node.cachedHandler = nil
	node.cachedMeta = nil
	for _, child := range node.Children {
		clearNodeCache(child)
	}
}

// ParseString memparsing string ZenoLang menjadi AST Node
func ParseString(data string) (*Node, error) {
	l := NewLexer(data)
	return parse(l, "string")
}

func parseFile(path string) (*Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	l := NewLexer(string(data))
	return parse(l, path)
}

func parse(l *Lexer, filename string) (*Node, error) {
	root := &Node{Name: "root"}
	stack := []*Node{root}
	var lastNode *Node

	for {
		tok := l.NextToken()
		if tok.Type == TokenEOF {
			break
		}

		switch tok.Type {
		case TokenIdentifier:
			// Identitas baru
			node := &Node{
				Name:     tok.Literal,
				Line:     tok.Line,
				Col:      tok.Column,
				Filename: filename,
			}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, node)
			node.Parent = parent
			lastNode = node

		case TokenColon:
			// Berpotensi diikuti Value (bisa beberapa token di baris yang sama) atau Blok
			currentLine := tok.Line
			var valueParts []Token

			for {
				peek := l.PeekToken()
				// Berhenti jika EOF, baris baru, atau ketemu pembuka blok { murni
				// TokenLBrace di Lexer baru skrg hanya murni standalone "{"
				if peek.Type == TokenEOF || peek.Line != currentLine || peek.Type == TokenLBrace || peek.Type == TokenColon {
					break
				}
				// Kasus penutup blok murni "}" di baris yang sama juga stop
				if peek.Type == TokenRBrace {
					break
				}

				// Ambil tokennya
				tok = l.NextToken()
				valueParts = append(valueParts, tok)
			}

			if len(valueParts) > 0 {
				if lastNode != nil {
					// Detect Raw String Literal
					if len(valueParts) == 1 && valueParts[0].Type == TokenString {
						lastNode.Value = "\x00" + valueParts[0].Literal
					} else {
						// Join dengan spasi agar "1 + 2" tetap "1 + 2"
						var fullVal string
						for i, p := range valueParts {
							if i > 0 {
								fullVal += " "
							}
							val := p.Literal
							// If it's a string token and doesn't have quotes, add them back
							// so the compiler knows it's a string literal.
							if p.Type == TokenString && !strings.HasPrefix(val, "\"") && !strings.HasPrefix(val, "'") {
								val = "\"" + val + "\""
							}
							fullVal += val
						}
						lastNode.Value = fullVal
					}
				}
			}

			// Cek apakah selanjutnya adalah blok {
			peek := l.PeekToken()
			if peek.Type == TokenLBrace {
				l.NextToken() // Konsumsi {
				if lastNode != nil {
					stack = append(stack, lastNode)
				}
			} else if peek.Type == TokenRBrace {
				// name: }  (Slot kosong)
				l.NextToken()
				if len(stack) > 1 {
					stack = stack[:len(stack)-1]
				}
			}

		case TokenLBrace:
			// Jika ada node sebelumnya, dia jadi parent
			if lastNode != nil {
				stack = append(stack, lastNode)
			} else {
				// Anonymous node
				node := &Node{
					Name:     "",
					Line:     tok.Line,
					Col:      tok.Column,
					Filename: filename,
				}
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
				node.Parent = parent
				stack = append(stack, node)
			}

		case TokenRBrace:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}

		case TokenComma:
			// Just consume it
			continue

		case TokenError:
			return nil, fmt.Errorf("lexical error at line %d, col %d in %s: %s", tok.Line, tok.Column, filename, tok.Literal)
		}
	}

	return root, nil
}
