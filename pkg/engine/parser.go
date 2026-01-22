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
	root, err := parse(l, path)
	if err != nil {
		return nil, err
	}

	// [STATIC INCLUDE] Resolve includes to create monolithic AST for compilation
	if err := resolveIncludes(root); err != nil {
		return nil, err
	}

	return root, nil
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
					if len(valueParts) == 1 && valueParts[0].Type == TokenIdentifier {
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

// resolveIncludes recursively processes 'include' directives to create a monolithic AST
func resolveIncludes(node *Node) error {
	var newChildren []*Node
	for _, child := range node.Children {
		// Clean the name just in case
		childName := strings.ToLower(child.Name)

		if childName == "include" {
			// Resolve path from Value
			valStr := fmt.Sprintf("%v", child.Value)

			// [FIX] Strip Parser's Raw String Prefix (\x00) if present
			if strings.HasPrefix(valStr, "\x00") {
				valStr = valStr[1:]
			}

			// Remove quotes if present ("src/file.zl" -> src/file.zl)
			if len(valStr) >= 2 {
				if (strings.HasPrefix(valStr, "\"") && strings.HasSuffix(valStr, "\"")) ||
					(strings.HasPrefix(valStr, "'") && strings.HasSuffix(valStr, "'")) {
					valStr = valStr[1 : len(valStr)-1]
				}
			}
			path := strings.TrimSpace(valStr)

			// If path is empty (e.g. include: )
			if path == "" {
				continue
			}

			// Parse included file recursively
			// Note: This relies on parsed file paths being relative to CWD or absolute
			includedRoot, err := parseFile(path)
			if err != nil {
				return fmt.Errorf("failed to include '%s': %v", path, err)
			}

			// Flatten: Append included root's children to current node's new children list
			// We skip the 'root' node of the included file and take its content
			newChildren = append(newChildren, includedRoot.Children...)
		} else {
			// Recurse for nested includes in other blocks
			if err := resolveIncludes(child); err != nil {
				return err
			}
			newChildren = append(newChildren, child)
		}
	}
	node.Children = newChildren
	return nil
}
