package slots

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"time"

	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"html"

	"github.com/gorilla/csrf"
)

// Blade Template Cache
var (
	bladeCache sync.Map // map[string]*cachedTemplate
)

type cachedTemplate struct {
	ast     *engine.Node
	modTime time.Time
}

func ensureBladeExt(path string) string {
	if !strings.HasSuffix(path, ".blade.zl") {
		return path + ".blade.zl"
	}
	return path
}

func RegisterBladeSlots(eng *engine.Engine) {
	RegisterLogicSlots(eng)

	// 1. Helper Slot for Writing to Response (Internal)
	eng.Register("__native_write", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return nil
		}

		val := resolveValue(node.Value, scope)
		str := coerce.ToString(val)
		w.Write([]byte(str))
		return nil
	}, engine.SlotMeta{Description: "Internal write for native blade"})

	eng.Register("__native_write_safe", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return nil
		}

		val := resolveValue(node.Value, scope)
		str := coerce.ToString(val)
		safeStr := html.EscapeString(str)
		w.Write([]byte(safeStr))
		return nil
	}, engine.SlotMeta{Description: "Internal safe write for native blade"})

	// 2. Slot Utama: view.blade
	eng.Register("view.blade", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Inject CSRF if available
		if r, ok := ctx.Value("httpRequest").(*http.Request); ok {
			scope.Set("csrf_field", csrf.TemplateField(r))
			scope.Set("csrf_token", csrf.Token(r))
		}

		var viewFile string
		if node.Value != nil {
			viewFile = coerce.ToString(resolveValue(node.Value, scope))
		}

		// Fallback check children
		if viewFile == "" {
			for _, c := range node.Children {
				if c.Name == "file" {
					viewFile = coerce.ToString(resolveValue(c.Value, scope))
					continue
				}
				// Bind attribute to scope
				val := parseNodeValue(c, scope)
				scope.Set(c.Name, val)
			}
		} else {
			// Extract other children if viewFile was set by Value
			for _, c := range node.Children {
				// Bind attribute to scope
				val := parseNodeValue(c, scope)
				scope.Set(c.Name, val)
			}
		}

		if viewFile == "" {
			return fmt.Errorf("view.blade.native: file required")
		}

		fullPath := filepath.Join("views", ensureBladeExt(viewFile))

		// Use cache for performance
		programNode, err := getCachedOrParse(fullPath)
		if err != nil {
			return err
		}

		// Execute normally
		return eng.Execute(ctx, programNode, scope)

	}, engine.SlotMeta{Description: "Render Blade natively using ZenoLang AST."})

	// 3. Section System (Layouts)
	eng.Register("section.define", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		name := coerce.ToString(resolveValue(node.Value, scope))
		var body *engine.Node

		for _, c := range node.Children {
			if c.Name == "do" {
				body = c
			}
		}

		if body != nil {
			sectionsRaw, ok := scope.Get("__sections")
			var sections map[string]*engine.Node
			if !ok || sectionsRaw == nil {
				sections = make(map[string]*engine.Node)
			} else {
				sections = sectionsRaw.(map[string]*engine.Node)
			}
			sections[name] = body
			scope.Set("__sections", sections)
		}
		return nil
	}, engine.SlotMeta{Description: "Define a layout section"})

	eng.Register("section.yield", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		name := coerce.ToString(resolveValue(node.Value, scope))

		sectionsRaw, ok := scope.Get("__sections")
		if !ok || sectionsRaw == nil {
			return nil
		}

		sections := sectionsRaw.(map[string]*engine.Node)
		if body, found := sections[name]; found {
			for _, child := range body.Children {
				if err := eng.Execute(ctx, child, scope); err != nil {
					return err
				}
			}
		}
		return nil
	}, engine.SlotMeta{Description: "Yield a layout section"})

	eng.Register("view.extends", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		layoutFile := coerce.ToString(resolveValue(node.Value, scope))
		if layoutFile == "" {
			return fmt.Errorf("view.extends: file required")
		}

		fullPath := filepath.Join("views", ensureBladeExt(layoutFile))

		// Use cache for performance
		layoutRoot, err := getCachedOrParse(fullPath)
		if err != nil {
			return err
		}
		return eng.Execute(ctx, layoutRoot, scope)
	}, engine.SlotMeta{Description: "Extend a layout"})

	eng.Register("view.component", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Value: "component-name"
		compName := coerce.ToString(resolveValue(node.Value, scope))
		if compName == "" {
			return nil
		}

		// Map name to path: x-alert -> components/alert.blade.zl
		// x-user.profile -> components/user/profile.blade.zl
		// Logic: replace . with /
		compPath := strings.ReplaceAll(compName, ".", "/")
		fullPath := filepath.Join("views", "components", compPath+".blade.zl")

		// 1. Prepare Component Scope
		// Use empty scope for isolation (Components must declare props)
		newScope := engine.NewScope(nil)

		// 2. Process Attributes
		// Attributes are children with Name="attr" or similar?
		// In transpile, we will map attributes to children nodes.
		// e.g. <x-alert type="error"> -> Children: [ {Name:"type", Value:"error"} ]
		// Also slots.

		var slotContent string
		slots := make(map[string]string)

		for _, child := range node.Children {
			if child.Name == "slot" {
				// Named Slot
				// Render it to string
				// It shouldn't execute "do" immediately?
				// We need to capture the output of this block.
				// We can use a temporary recorder.
				rec := httptest.NewRecorder()
				subCtx := context.WithValue(ctx, "httpWriter", http.ResponseWriter(rec))

				// Execute children of slot
				for _, c := range child.Children {
					eng.Execute(subCtx, c, scope) // Use OUTER scope for slot content! (Important: lexical scoping)
				}
				slots[child.Value.(string)] = rec.Body.String()

			} else if child.Name == "default_slot" {
				rec := httptest.NewRecorder()
				subCtx := context.WithValue(ctx, "httpWriter", http.ResponseWriter(rec))
				for _, c := range child.Children {
					eng.Execute(subCtx, c, scope)
				}
				slotContent = rec.Body.String()

			} else {
				// Attribute
				// Name=VarName, Value=Val
				// Resolve Value
				val := resolveValue(child.Value, scope)
				newScope.Set(child.Name, val)
			}
		}

		// Bind Slots
		newScope.Set("slot", slotContent) // Htmlable technically, but string here.
		for k, v := range slots {
			newScope.Set(k, v) // $header, $footer
		}

		// 3. Render View (with caching)
		compRoot, err := getCachedOrParse(fullPath)
		if err != nil {
			return err
		}

		return eng.Execute(ctx, compRoot, newScope)

	}, engine.SlotMeta{Description: "Render a Blade Component"})

	eng.Register("view.include", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		viewFile := coerce.ToString(resolveValue(node.Value, scope))
		if viewFile == "" {
			return nil
		}

		fullPath := filepath.Join("views", ensureBladeExt(viewFile))

		// Parse Child Data
		var includeData map[string]interface{}

		if len(node.Children) > 0 {
			dataNode := node.Children[0]
			if dataNode.Name == "data_map" {
				includeData = make(map[string]interface{})
				for _, child := range dataNode.Children {
					valStr := coerce.ToString(child.Value)
					if strings.HasPrefix(valStr, "$") {
						resolved, _ := scope.Get(strings.TrimPrefix(valStr, "$"))
						includeData[child.Name] = resolved
					} else {
						includeData[child.Name] = valStr
					}
				}
			} else if dataNode.Name == "data_var" {
				val := resolveValue(dataNode.Value, scope)
				if m, ok := val.(map[string]interface{}); ok {
					includeData = m
				}
			}
		}

		// Create Inner Scope
		innerScope := scope.Clone()
		for k, v := range includeData {
			innerScope.Set(k, v)
		}

		// Use cache for performance
		transpiled, err := getCachedOrParse(fullPath)
		if err != nil {
			return err
		}

		return eng.Execute(ctx, transpiled, innerScope)
	}, engine.SlotMeta{Description: "Include a partial view"})

	eng.Register("view.push", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		name := coerce.ToString(resolveValue(node.Value, scope))
		var body *engine.Node
		for _, c := range node.Children {
			if c.Name == "do" {
				body = c
			}
		}

		if body != nil {
			stacksRaw, ok := scope.Get("__stacks")
			var stacks map[string][]*engine.Node
			if !ok || stacksRaw == nil {
				stacks = make(map[string][]*engine.Node)
			} else {
				stacks = stacksRaw.(map[string][]*engine.Node)
			}

			// Append Body to Stack List
			stacks[name] = append(stacks[name], body)
			scope.Set("__stacks", stacks)
		}
		return nil
	}, engine.SlotMeta{Description: "Push content to stack"})

	eng.Register("view.stack", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		name := coerce.ToString(resolveValue(node.Value, scope))

		stacksRaw, ok := scope.Get("__stacks")
		if !ok || stacksRaw == nil {
			return nil
		}

		stacks := stacksRaw.(map[string][]*engine.Node)
		if nodes, found := stacks[name]; found {
			for _, n := range nodes {
				// Execute the pushed block
				// Since push block is "do", iterate children
				for _, child := range n.Children {
					if err := eng.Execute(ctx, child, scope); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}, engine.SlotMeta{Description: "Render stack content"})
}

// ==========================================
// BLADE TEMPLATE CACHE HELPERS
// ==========================================

// getCachedOrParse retrieves a cached template or parses it if not cached/outdated
func getCachedOrParse(fullPath string) (*engine.Node, error) {
	// Check cache
	if cached, ok := bladeCache.Load(fullPath); ok {
		ct := cached.(*cachedTemplate)

		// In production, always use cache
		if os.Getenv("APP_ENV") == "production" {
			return ct.ast, nil
		}

		// In development, check if file changed
		fileInfo, err := os.Stat(fullPath)
		if err == nil && fileInfo.ModTime().Equal(ct.modTime) {
			return ct.ast, nil // Cache hit
		}
	}

	// Cache miss or file changed - parse
	return parseAndCache(fullPath)
}

// parseAndCache reads, parses, and caches a Blade template
func parseAndCache(fullPath string) (*engine.Node, error) {
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("view not found: %s", fullPath)
	}

	programNode, err := transpileBladeNative(string(contentBytes))
	if err != nil {
		return nil, err
	}

	// Get file mod time
	fileInfo, _ := os.Stat(fullPath)

	// Store in cache
	bladeCache.Store(fullPath, &cachedTemplate{
		ast:     programNode,
		modTime: fileInfo.ModTime(),
	})

	return programNode, nil
}

// ClearBladeCache clears all cached Blade templates
// This should be called when view files change during development
func ClearBladeCache() {
	bladeCache.Range(func(key, value interface{}) bool {
		bladeCache.Delete(key)
		return true
	})
}

// === NATIVE TRANSPILER ===
// Converts Blade string to *engine.Node (Root "do" block)
func transpileBladeNative(content string) (*engine.Node, error) {
	root := &engine.Node{
		Name:     "do",
		Children: []*engine.Node{},
	}

	// Simple: Tokenize based on {{ }} and @if/@endif
	// Since it's a nested structure, we need parser state.

	// Use a simple scanner approach.
	// 1. Scan Text until {{ or @ is found
	// 2. If {{ -> parse variable -> add __native_write node
	// 3. If @if -> parse condition -> recursive parse body -> add if node
	// 4. If @foreach -> parse loop -> recursive parse body -> add foreach/range node

	// Cursor position
	pos := 0
	length := len(content)

	// Extensions
	var extendsFile string

	for pos < length {
		// Find Logic Start
		nextTag := strings.IndexAny(content[pos:], "@{\n<")
		if nextTag == -1 {
			// Sisa text
			text := content[pos:]
			if text != "" {
				root.Children = append(root.Children, createWriteNode(text))
			}
			break
		}

		// Text before tag
		offset := pos + nextTag
		if offset > pos {
			text := content[pos:offset]
			root.Children = append(root.Children, createWriteNode(text))
		}

		pos = offset

		// Check tag type
		if strings.HasPrefix(content[pos:], "{{--") {
			// Comment
			endComment := strings.Index(content[pos:], "--}}")
			if endComment != -1 {
				pos += endComment + 4
			} else {
				pos = length // Unterminated
			}
		} else if strings.HasPrefix(content[pos:], "{!!") {
			// Unescaped Echo
			endEcho := strings.Index(content[pos:], "!!}")
			if endEcho != -1 {
				raw := content[pos+3 : pos+endEcho]
				varVal := strings.TrimSpace(raw)

				// Standard __native_write (Raw)
				root.Children = append(root.Children, &engine.Node{
					Name:  "__native_write",
					Value: varVal,
				})

				pos += endEcho + 3
			} else {
				root.Children = append(root.Children, createWriteNode("{!!"))
				pos += 3
			}

		} else if strings.HasPrefix(content[pos:], "{{") {
			// ... Echo (Existing) ...
			endEcho := strings.Index(content[pos:], "}}")
			if endEcho != -1 {
				raw := content[pos+2 : pos+endEcho]
				varVal := strings.TrimSpace(raw)

				// Check for function call: Name(...)
				if strings.Contains(varVal, "(") && strings.HasSuffix(varVal, ")") {
					// Parse Function Call
					funcName := varVal[:strings.Index(varVal, "(")]
					argsRaw := varVal[strings.Index(varVal, "(")+1 : len(varVal)-1]

					// Split args by comma (simple split, assumes no commas in strings for now)
					// TODO: Robust arg parsing
					args := strings.Split(argsRaw, ",")
					var children []*engine.Node

					for _, arg := range args {
						arg = strings.TrimSpace(arg)
						if arg == "" {
							continue
						}

						// Determine type
						if strings.HasPrefix(arg, "\"") || strings.HasPrefix(arg, "'") {
							// String literal
							val := strings.Trim(arg, "\"'")
							children = append(children, &engine.Node{Name: "arg", Value: val}) // Literal
						} else {
							// Variable
							children = append(children, &engine.Node{Name: "arg", Value: arg}) // Variable
						}
					}

					// If function, we MIGHT want to escape result too?
					// Standard Blade escapes result of function calls unless they return HtmlString.
					// For now, let's wrap result in Escape?
					// But we are returning an engine.Node which executes a slot.
					// If the slot writes to writer directly, we can't capture and escape easily here unless we wrap execution.
					// OR we assume Function Slots are responsible for safety if called directly?
					// Safest approach: The Function Slot executes logic.
					// Ideally we should use __native_write_safe with value as function execution result.
					// But current AST structure is: Node(Name=FuncName).
					// The FuncName slot executes.
					// If we want to escape, we need: Node(Name=__native_write_safe, Value=Node(Name=FuncName...))?
					// But our engine doesn't resolve Node as Value recursively like that easily for slots that don't return values but write to stream.
					// Many Zeno slots write to stream directly (echo, logic.json).
					// If the function slot writes to stream, we can't intercept easily without capturing.
					//
					// ALLOWANCE: For function calls inside {{ }}, we will KEEP existing behavior (direct execution) for now
					// to avoid breaking custom slots that expect to write HTML (like form builders).
					// Users should use __native_write_safe concept manually if needed or we upgrade engine later.
					// BUT standard variables MUST be escaped.

					root.Children = append(root.Children, &engine.Node{
						Name:     funcName,
						Children: children,
					})
				} else {
					root.Children = append(root.Children, &engine.Node{
						Name:  "__native_write_safe", // CHANGED TO SAFE
						Value: varVal,
					})
				}
				pos += endEcho + 2
			} else {
				root.Children = append(root.Children, createWriteNode("{{"))
				pos += 2
			}
		} else if strings.HasPrefix(content[pos:], "@csrf") {
			// @csrf -> {{ $.csrf_field }}
			// We can emit a write node that will be resolved at runtime
			// if we assume $.csrf_field is in scope.
			// Or explicitly fetch csrf token?
			// "view.blade" handler puts csrf_field in scope?
			// Let's check view.blade (formerly view.blade.native) implementation.
			// It calls `ctx.Value("httpRequest")`... wait?
			// The OLD view.blade did.
			// The NEW view.blade (native) does NOT seems to inject CSRF yet!
			// I need to fix that too.
			// For now, transpile to echo variable.
			root.Children = append(root.Children, &engine.Node{
				Name:  "__native_write", // CSRF field is HTML, must be raw
				Value: "$csrf_field",
			})
			pos += 5
		} else if strings.HasPrefix(content[pos:], "@include") {
			// @include('view.name', ['key' => 'val'])
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				argsRaw := content[pos+startParen+1 : pos+endParen]
				// Parse Args: ViewName, Data
				args := splitBladeArgs(argsRaw)

				viewNameRaw := strings.Trim(args[0], "'\" ")

				includeNode := &engine.Node{
					Name:  "view.include",
					Value: viewNameRaw,
				}

				if len(args) > 1 {
					dataRaw := strings.TrimSpace(args[1])
					// Parse Data Dict ['k'=>'v'] or variable $data
					dataNode := parseBladeData(dataRaw)
					if dataNode != nil {
						includeNode.Children = append(includeNode.Children, dataNode)
					}
				}

				root.Children = append(root.Children, includeNode)
				pos += endParen + 1
			} else {
				pos += 8 // @include
			}
		} else if strings.HasPrefix(content[pos:], "@extends") {
			// @extends('layout')
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])
			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]
				// Strip quotes
				valRaw = strings.Trim(valRaw, "'\"")
				extendsFile = valRaw
				pos += endParen + 1 // Skip )
			} else {
				pos += 8
			}
		} else if strings.HasPrefix(content[pos:], "@section") {
			// @section('name') ... @endsection
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				nameRaw := content[pos+startParen+1 : pos+endParen]
				nameRaw = strings.Trim(nameRaw, "'\"")

				blockStart := pos + endParen + 1
				blockEnd := findEndSection(content[blockStart:])

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @section")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				// section.define node
				sectionNode := &engine.Node{
					Name:  "section.define",
					Value: nameRaw,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}
				root.Children = append(root.Children, sectionNode)

				pos = absoluteBlockEnd + 11 // @endsection
			} else {
				pos += 8
			}

		} else if strings.HasPrefix(content[pos:], "@isset") {
			// @isset($var)
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]

				blockStart := pos + endParen + 1
				blockEnd := findEndIsset(content[blockStart:])

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @isset")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				node := &engine.Node{
					Name:  "isset",
					Value: valRaw,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}
				root.Children = append(root.Children, node)
				pos = absoluteBlockEnd + 9 // @endisset
			} else {
				pos += 6
			}
		} else if strings.HasPrefix(content[pos:], "@empty") {
			// @empty($var)
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]

				blockStart := pos + endParen + 1
				blockEnd := findEndEmpty(content[blockStart:])

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @empty")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				node := &engine.Node{
					Name:  "empty",
					Value: valRaw,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}
				root.Children = append(root.Children, node)
				pos = absoluteBlockEnd + 9 // @endempty
			} else {
				pos += 6
			}

		} else if strings.HasPrefix(content[pos:], "@auth") {
			// @auth ... @endauth
			// @auth('guard') supported? Usually yes.
			// Simple version for now

			// Optional parens
			// Check if ( follows immediately
			blockStart := pos + 5
			if strings.HasPrefix(content[pos:], "@auth(") {
				// Has guard arg, ignore for now but parse parens
				endParen := findBalancedParen(content[pos:])
				if endParen != -1 {
					blockStart = pos + endParen + 1
				}
			}

			blockEnd := findEndAuth(content[blockStart:])
			if blockEnd == -1 {
				return nil, fmt.Errorf("unclosed @auth")
			}

			absoluteBlockEnd := blockStart + blockEnd
			bodyContent := content[blockStart:absoluteBlockEnd]
			bodyNode, err := transpileBladeNative(bodyContent)
			if err != nil {
				return nil, err
			}

			node := &engine.Node{
				Name: "auth",
				Children: []*engine.Node{
					{Name: "do", Children: bodyNode.Children},
				},
			}
			root.Children = append(root.Children, node)
			pos = absoluteBlockEnd + 8 // @endauth

		} else if strings.HasPrefix(content[pos:], "@guest") {
			// @guest ... @endguest
			blockStart := pos + 6
			if strings.HasPrefix(content[pos:], "@guest(") {
				endParen := findBalancedParen(content[pos:])
				if endParen != -1 {
					blockStart = pos + endParen + 1
				}
			}

			blockEnd := findEndGuest(content[blockStart:])
			if blockEnd == -1 {
				return nil, fmt.Errorf("unclosed @guest")
			}

			absoluteBlockEnd := blockStart + blockEnd
			bodyContent := content[blockStart:absoluteBlockEnd]
			bodyNode, err := transpileBladeNative(bodyContent)
			if err != nil {
				return nil, err
			}

			node := &engine.Node{
				Name: "guest",
				Children: []*engine.Node{
					{Name: "do", Children: bodyNode.Children},
				},
			}
			root.Children = append(root.Children, node)
			pos = absoluteBlockEnd + 9 // @endguest

		} else if strings.HasPrefix(content[pos:], "@json") {
			// @json($data)
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]

				// Strip quotes? Variable mostly.
				// If literal?

				node := &engine.Node{
					Name:  "json",
					Value: valRaw, // logic.json will resolve this
				}
				root.Children = append(root.Children, node)
				pos = pos + endParen + 1
			} else {
				pos += 5
			}

		} else if strings.HasPrefix(content[pos:], "@for") {
			// Check for @foreach first!
			if strings.HasPrefix(content[pos:], "@foreach") {
				// ... Foreach Logic (Existing) ...
				startParen := strings.Index(content[pos:], "(")
				endParen := findBalancedParen(content[pos:])
				if startParen != -1 && endParen != -1 {
					defRaw := content[pos+startParen+1 : pos+endParen]
					parts := strings.Split(defRaw, " as ")
					if len(parts) == 2 {
						listVar := strings.TrimSpace(parts[0])
						itemVar := strings.TrimSpace(parts[1])
						blockStart := pos + endParen + 1
						blockEnd := findEndForeach(content[blockStart:])
						if blockEnd == -1 {
							return nil, fmt.Errorf("unclosed @foreach")
						}
						absoluteBlockEnd := blockStart + blockEnd
						bodyContent := content[blockStart:absoluteBlockEnd]
						bodyNode, err := transpileBladeNative(bodyContent)
						if err != nil {
							return nil, err
						}
						itemVarClean := strings.TrimPrefix(itemVar, "$")
						foreachNode := &engine.Node{Name: "for", Value: listVar, Children: []*engine.Node{{Name: "as", Value: itemVarClean}, {Name: "do", Children: bodyNode.Children}}}
						root.Children = append(root.Children, foreachNode)
						pos = absoluteBlockEnd + 11
					} else {
						pos += 8
					}
				} else {
					pos += 8
				}
			} else if strings.HasPrefix(content[pos:], "@forelse") {
				// @forelse($list as $item) ... @empty ... @endforelse
				startParen := strings.Index(content[pos:], "(")
				endParen := findBalancedParen(content[pos:])
				if startParen != -1 && endParen != -1 {
					defRaw := content[pos+startParen+1 : pos+endParen] // $list as $item

					// Parse definition similarly to foreach
					parts := strings.Split(defRaw, " as ")
					if len(parts) != 2 {
						return nil, fmt.Errorf("invalid @forelse format")
					}

					listVar := strings.TrimSpace(parts[0])
					itemVar := strings.TrimSpace(parts[1])
					itemVarClean := strings.TrimPrefix(itemVar, "$")

					blockStart := pos + endParen + 1
					blockEnd := findEndForelse(content[blockStart:])
					if blockEnd == -1 {
						return nil, fmt.Errorf("unclosed @forelse")
					}

					absoluteBlockEnd := blockStart + blockEnd
					fullBlockContent := content[blockStart:absoluteBlockEnd]

					// Split into Body and Empty
					emptyPos := -1

					// Custom scan for @empty
					d := 0
					scanPos := 0
					for scanPos < len(fullBlockContent) {
						if strings.HasPrefix(fullBlockContent[scanPos:], "@empty") {
							if d == 0 {
								emptyPos = scanPos
								break
							}
						} else if strings.HasPrefix(fullBlockContent[scanPos:], "@foreach") || strings.HasPrefix(fullBlockContent[scanPos:], "@forelse") {
							d++
						} else if strings.HasPrefix(fullBlockContent[scanPos:], "@endforeach") || strings.HasPrefix(fullBlockContent[scanPos:], "@endforelse") {
							d--
						}
						scanPos++
					}

					var bodyContent, emptyContent string
					if emptyPos != -1 {
						bodyContent = fullBlockContent[:emptyPos]
						emptyContent = fullBlockContent[emptyPos+6:]
					} else {
						bodyContent = fullBlockContent
					}

					bodyNode, err := transpileBladeNative(bodyContent)
					if err != nil {
						return nil, err
					}

					node := &engine.Node{
						Name:  "forelse",
						Value: listVar,
						Children: []*engine.Node{
							{Name: "as", Value: itemVarClean},
							{Name: "do", Children: bodyNode.Children},
						},
					}

					if emptyPos != -1 {
						emptyNode, err := transpileBladeNative(emptyContent)
						if err != nil {
							return nil, err
						}
						node.Children = append(node.Children, &engine.Node{
							Name:     "empty",
							Children: emptyNode.Children,
						})
					}

					root.Children = append(root.Children, node)
					pos = absoluteBlockEnd + 11 // @endforelse
				} else {
					pos += 8
				}
			} else {
				// It IS @for
				// @for($i=0;...)
				startParen := strings.Index(content[pos:], "(")
				endParen := findBalancedParen(content[pos:])

				if startParen != -1 && endParen != -1 {
					valRaw := content[pos+startParen+1 : pos+endParen]

					blockStart := pos + endParen + 1
					blockEnd := findEndFor(content[blockStart:])
					if blockEnd == -1 {
						return nil, fmt.Errorf("unclosed @for")
					}

					absoluteBlockEnd := blockStart + blockEnd
					bodyContent := content[blockStart:absoluteBlockEnd]
					bodyNode, err := transpileBladeNative(bodyContent)
					if err != nil {
						return nil, err
					}

					node := &engine.Node{
						Name:  "for", // The new C-style slot
						Value: valRaw,
						Children: []*engine.Node{
							{Name: "do", Children: bodyNode.Children},
						},
					}
					root.Children = append(root.Children, node)
					pos = absoluteBlockEnd + 7 // @endfor
				} else {
					pos += 4
				}
			}

		} else if strings.HasPrefix(content[pos:], "@unless") {
			// @unless($var)
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]

				blockStart := pos + endParen + 1
				blockEnd := findEndUnless(content[blockStart:])

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @unless")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				node := &engine.Node{
					Name:  "unless",
					Value: valRaw,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}
				root.Children = append(root.Children, node)
				pos = absoluteBlockEnd + 10 // @endunless
			} else {
				pos += 7
			}
		} else if strings.HasPrefix(content[pos:], "@switch") {
			// @switch($val)
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])
			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]

				blockStart := pos + endParen + 1
				blockEnd := findEndSwitch(content[blockStart:])
				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @switch")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]

				// Parse Cases
				cases := splitSwitchCases(bodyContent)

				switchNode := &engine.Node{
					Name:  "switch",
					Value: valRaw,
				}

				for _, c := range cases {
					caseBodyNode, err := transpileBladeNative(c.Body)
					if err != nil {
						return nil, err
					}

					caseNode := &engine.Node{
						Name:     c.Type, // "case" or "default"
						Value:    c.Value,
						Children: caseBodyNode.Children,
					}
					switchNode.Children = append(switchNode.Children, caseNode)
				}

				root.Children = append(root.Children, switchNode)
				pos = absoluteBlockEnd + 11 // @endswitch
			} else {
				pos += 7
			}

		} else if strings.HasPrefix(content[pos:], "@break") {
			if len(content) > pos+6 && content[pos+6] == '(' {
				startParen := strings.Index(content[pos:], "(")
				endParen := findBalancedParen(content[pos:])
				if startParen != -1 && endParen != -1 {
					valRaw := content[pos+startParen+1 : pos+endParen]
					root.Children = append(root.Children, &engine.Node{Name: "break", Value: valRaw})
					pos += endParen + 1
				} else {
					root.Children = append(root.Children, &engine.Node{Name: "break"})
					pos += 6
				}
			} else {
				root.Children = append(root.Children, &engine.Node{Name: "break"})
				pos += 6
			}
		} else if strings.HasPrefix(content[pos:], "@continue") {
			if len(content) > pos+9 && content[pos+9] == '(' {
				startParen := strings.Index(content[pos:], "(")
				endParen := findBalancedParen(content[pos:])
				if startParen != -1 && endParen != -1 {
					valRaw := content[pos+startParen+1 : pos+endParen]
					root.Children = append(root.Children, &engine.Node{Name: "continue", Value: valRaw})
					pos += endParen + 1
				} else {
					root.Children = append(root.Children, &engine.Node{Name: "continue"})
					pos += 9
				}
			} else {
				root.Children = append(root.Children, &engine.Node{Name: "continue"})
				pos += 9
			}
		} else if strings.HasPrefix(content[pos:], "@method") {
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])
			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]
				verb := strings.Trim(valRaw, "'\"")
				root.Children = append(root.Children, createWriteNode(fmt.Sprintf("<input type=\"hidden\" name=\"_method\" value=\"%s\">", verb)))
				pos += endParen + 1
			} else {
				pos += 7
			}
		} else if strings.HasPrefix(content[pos:], "@dd") {
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])
			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]
				root.Children = append(root.Children, &engine.Node{Name: "dd", Value: valRaw})
				pos += endParen + 1
			} else {
				pos += 3
			}
		} else if strings.HasPrefix(content[pos:], "@while") {
			// @while($cond)
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])
			if startParen != -1 && endParen != -1 {
				valRaw := content[pos+startParen+1 : pos+endParen]

				blockStart := pos + endParen + 1
				blockEnd := findEndWhile(content[blockStart:])
				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @while")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				node := &engine.Node{
					Name:  "while",
					Value: valRaw,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}
				root.Children = append(root.Children, node)
				pos = absoluteBlockEnd + 9 // @endwhile
			} else {
				pos += 6
			}
		} else if strings.HasPrefix(content[pos:], "@push") {
			// @push('name')
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				nameRaw := content[pos+startParen+1 : pos+endParen]
				nameRaw = strings.Trim(nameRaw, "'\"")

				blockStart := pos + endParen + 1
				blockEnd := findEndPush(content[blockStart:])

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @push")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				pushNode := &engine.Node{
					Name:  "view.push",
					Value: nameRaw,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}
				root.Children = append(root.Children, pushNode)

				pos = absoluteBlockEnd + 8 // @endpush
			} else {
				pos += 5
			}
		} else if strings.HasPrefix(content[pos:], "@error") {
			// @error('fieldname')
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				fieldRaw := content[pos+startParen+1 : pos+endParen]
				fieldName := strings.Trim(fieldRaw, "'\"")

				blockStart := pos + endParen + 1
				blockEnd := findEndError(content[blockStart:])

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @error")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				errorNode := &engine.Node{
					Name:  "error",
					Value: fieldName,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}
				root.Children = append(root.Children, errorNode)

				pos = absoluteBlockEnd + 9 // @enderror
			} else {
				pos += 6
			}
		} else if strings.HasPrefix(content[pos:], "@cannot") {
			// @cannot('ability', $resource)
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				argsRaw := content[pos+startParen+1 : pos+endParen]

				parts := strings.Split(argsRaw, ",")
				ability := strings.Trim(strings.TrimSpace(parts[0]), "'\"")

				var resourceValue interface{}
				if len(parts) > 1 {
					resourceValue = strings.TrimSpace(parts[1])
				}

				blockStart := pos + endParen + 1
				blockEnd := findEndCannot(content[blockStart:])

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @cannot")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				cannotNode := &engine.Node{
					Name:  "cannot",
					Value: ability,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}

				if resourceValue != nil {
					cannotNode.Children = append([]*engine.Node{
						{Name: "resource", Value: resourceValue},
					}, cannotNode.Children...)
				}

				root.Children = append(root.Children, cannotNode)
				pos = absoluteBlockEnd + 10 // @endcannot
			} else {
				pos += 7
			}
		} else if strings.HasPrefix(content[pos:], "@can") {
			// @can('ability', $resource) or @can('ability')
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])

			if startParen != -1 && endParen != -1 {
				argsRaw := content[pos+startParen+1 : pos+endParen]

				// Parse arguments: 'ability', $resource
				parts := strings.Split(argsRaw, ",")
				ability := strings.Trim(strings.TrimSpace(parts[0]), "'\"")

				var resourceValue interface{}
				if len(parts) > 1 {
					resourceValue = strings.TrimSpace(parts[1])
				}

				blockStart := pos + endParen + 1
				blockEnd := findEndCan(content[blockStart:])

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @can")
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				canNode := &engine.Node{
					Name:  "can",
					Value: ability,
					Children: []*engine.Node{
						{Name: "do", Children: bodyNode.Children},
					},
				}

				// Add resource if provided
				if resourceValue != nil {
					canNode.Children = append([]*engine.Node{
						{Name: "resource", Value: resourceValue},
					}, canNode.Children...)
				}

				root.Children = append(root.Children, canNode)
				pos = absoluteBlockEnd + 7 // @endcan
			} else {
				pos += 4
			}
		} else if strings.HasPrefix(content[pos:], "@stack") {
			// @stack('name')
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])
			if startParen != -1 && endParen != -1 {
				nameRaw := content[pos+startParen+1 : pos+endParen]
				nameRaw = strings.Trim(nameRaw, "'\"")

				stackNode := &engine.Node{
					Name:  "view.stack",
					Value: nameRaw,
				}
				root.Children = append(root.Children, stackNode)
				pos += endParen + 1
			} else {
				pos += 6
			}
		} else if strings.HasPrefix(content[pos:], "@yield") {
			// @yield('name')
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])
			if startParen != -1 && endParen != -1 {
				nameRaw := content[pos+startParen+1 : pos+endParen]
				nameRaw = strings.Trim(nameRaw, "'\"")

				yieldNode := &engine.Node{
					Name:  "section.yield",
					Value: nameRaw,
				}
				root.Children = append(root.Children, yieldNode)
				pos += endParen + 1
			} else {
				pos += 6
			}
		} else if strings.HasPrefix(content[pos:], "@if") {
			// ... If Logic (Existing) ...
			startParen := strings.Index(content[pos:], "(")
			endParen := findBalancedParen(content[pos:])
			if startParen != -1 && endParen != -1 {

				condRaw := content[pos+startParen+1 : pos+endParen]
				blockStart := pos + endParen + 1

				// Find End or Else
				blockEnd, matchType := findEndIf(content[blockStart:])
				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed @if")
				}

				absoluteBlockEnd := blockStart + blockEnd
				trueContent := content[blockStart:absoluteBlockEnd]
				trueNode, err := transpileBladeNative(trueContent)
				if err != nil {
					return nil, err
				}

				var elseNode *engine.Node

				if matchType == "else" {
					// We found an @else, now find the @endif
					elseStart := absoluteBlockEnd + 5 // length of @else
					elseEnd, matchType2 := findEndIf(content[elseStart:])
					if matchType2 != "endif" {
						return nil, fmt.Errorf("unclosed @else (expected @endif)")
					}
					absoluteElseEnd := elseStart + elseEnd
					elseContent := content[elseStart:absoluteElseEnd]
					elseBody, err := transpileBladeNative(elseContent)
					if err != nil {
						return nil, err
					}

					elseNode = &engine.Node{
						Name:     "else",
						Children: elseBody.Children,
					}

					// Update pos to end of endif
					pos = absoluteElseEnd + 6 // @endif
				} else {
					// Found @endif directly
					pos = absoluteBlockEnd + 6 // @endif
				}

				ifNode := &engine.Node{
					Name:  "if",
					Value: condRaw,
					Children: []*engine.Node{
						{Name: "then", Children: trueNode.Children},
					},
				}
				if elseNode != nil {
					ifNode.Children = append(ifNode.Children, elseNode)
				}

				root.Children = append(root.Children, ifNode)
			} else {
				pos += 3
			}

		} else if strings.HasPrefix(content[pos:], "<x-") {
			// <x-name attr="val" :attr="expr">
			// <x-slot name="header">

			// 1. Extract Tag Content until > or />
			// Simple scan?
			// Caveat: > inside attribute quotes?
			tagEnd := -1
			inQuote := false
			var quoteChar rune

			for i, c := range content[pos:] {
				if inQuote {
					if c == quoteChar {
						inQuote = false
					}
				} else {
					if c == '"' || c == '\'' {
						inQuote = true
						quoteChar = c
					}
					if c == '>' {
						tagEnd = i
						break
					}
				}
			}

			if tagEnd == -1 {
				// Malformed, treat as text
				root.Children = append(root.Children, createWriteNode("<"))
				pos++
				continue
			}

			fullTag := content[pos : pos+tagEnd+1] // <x-name ... > or <x-name ... />
			isSelfClosing := strings.HasSuffix(fullTag, "/>")

			// Parse Name
			// <x-name ...
			// Cut <, >/ />
			inner := fullTag
			if isSelfClosing {
				inner = strings.TrimSuffix(inner, "/>")
			} else {
				inner = strings.TrimSuffix(inner, ">")
			}
			inner = strings.TrimPrefix(inner, "<")
			inner = strings.TrimSpace(inner)

			// Split name from attrs
			parts := strings.SplitN(inner, " ", 2)
			tagName := parts[0]

			var attrsRaw string
			if len(parts) > 1 {
				attrsRaw = parts[1]
			}

			// Determine Node Type
			nodeName := "view.component"
			nodeValue := strings.TrimPrefix(tagName, "x-") // user.profile

			if tagName == "x-slot" {
				nodeName = "slot"
				// Value will be extracted from name attribute later, or default to "" if missing?
				// x-slot MUST have name?
			}

			compNode := &engine.Node{Name: nodeName, Value: nodeValue}

			// Parse Attributes and append as children
			// Helper: parseAttributes(attrsRaw) -> []*engine.Node
			attrNodes := parseBladeAttributes(attrsRaw)
			compNode.Children = append(compNode.Children, attrNodes...)

			if nodeName == "slot" {
				// Fix Value from name attribute
				for _, c := range attrNodes {
					if c.Name == "name" {
						// Evaluate literal value
						// If key="val", Value is "val".
						// We need "val" string.
						// resolveValue?
						// But here we are building the AST. Value currently holds the raw string if literal?
						// parseBladeAttributes returns Node{Name, Value}.
						// If Value is expression (starts with $?), we can't fully resolve name at parse time if dynamic?
						// Blade x-slot name usually static.
						// Assuming value is literal string.
						compNode.Value = c.Value
					}
				}
			}

			// Handle Body
			if isSelfClosing {
				root.Children = append(root.Children, compNode)
				pos += tagEnd + 1
			} else {
				// Find closing tag </x-name> (or </x-slot>)
				// name for searching: tagName (x-name)
				blockStart := pos + tagEnd + 1
				blockEnd := findEndComponent(content[blockStart:], tagName)

				if blockEnd == -1 {
					return nil, fmt.Errorf("unclosed component %s", tagName)
				}

				absoluteBlockEnd := blockStart + blockEnd
				bodyContent := content[blockStart:absoluteBlockEnd]

				// Recursive Parse Body
				bodyNode, err := transpileBladeNative(bodyContent)
				if err != nil {
					return nil, err
				}

				if nodeName == "slot" {
					compNode.Children = append(compNode.Children, bodyNode.Children...)
				} else {
					// For component, body is Default Slot
					// Wrap in "default_slot" node?
					// Or just append children that are NOT slots?
					// view.component slot handles "default_slot" or just children.
					// Implementation (Step 1436) handled `child.Name == "default_slot"`.
					// So wrap body in `default_slot` node.

					// But wait, <x-slot> children inside body will be parsed as "slot" nodes.
					// We should separate them?
					// Implementation iterates children. If "slot", puts in map. If "default_slot", puts in slotContent.
					// So if we put <x-slot> nodes as direct children of compNode, they are handled.
					// What about non-slot content? That's default slot.
					// So we should verify body children.

					defaultSlotNode := &engine.Node{Name: "default_slot"}

					for _, c := range bodyNode.Children {
						if c.Name == "slot" {
							compNode.Children = append(compNode.Children, c)
						} else {
							defaultSlotNode.Children = append(defaultSlotNode.Children, c)
						}
					}
					// Only append default slot if has children
					if len(defaultSlotNode.Children) > 0 {
						compNode.Children = append(compNode.Children, defaultSlotNode)
					}
				}

				root.Children = append(root.Children, compNode)

				// Advance past closing tag </x-name>
				// findEndComponent returns index of start of closing tag.
				// length of </x-name> = 2 + len(name) + 1
				closingLen := 2 + len(tagName) + 1

				// Verify closing tag match? findEndComponent guarantees it find the prefix.
				// Assume simple match.
				// What about whitespace in closing? </x-name >?
				// Just scan until >
				endTagScan := content[absoluteBlockEnd:]
				endTagClose := strings.Index(endTagScan, ">")
				if endTagClose != -1 {
					pos = absoluteBlockEnd + endTagClose + 1
				} else {
					// Fallback
					pos = absoluteBlockEnd + closingLen
				}
			}

		} else {
			root.Children = append(root.Children, createWriteNode(string(content[pos])))
			pos++
		}
	}

	// If extends exists, execute it AT THE END
	if extendsFile != "" {
		root.Children = append(root.Children, &engine.Node{
			Name:  "view.extends",
			Value: extendsFile,
		})
	}

	return root, nil
}

func findEndSection(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@section") {
			depth++
			pos += 8
		} else if strings.HasPrefix(s[pos:], "@endsection") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 11
		} else {
			pos++
		}
	}
	return -1
}

// Simple Helpers for parsing

func findEndIf(s string) (int, string) {
	// Find @endif OR @else respecting nested @if
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@if") {
			depth++
			pos += 3
		} else if strings.HasPrefix(s[pos:], "@endif") {
			if depth == 0 {
				return pos, "endif" // Found match
			}
			depth--
			pos += 6
		} else if strings.HasPrefix(s[pos:], "@else") {
			if depth == 0 {
				return pos, "else"
			}
			pos += 5
		} else {
			pos++
		}
	}
	return -1, ""
}

func findEndForeach(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@foreach") {
			depth++
			pos += 8
		} else if strings.HasPrefix(s[pos:], "@endforeach") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 11
		} else {
			pos++
		}
	}
	return -1
}

func splitBladeArgs(s string) []string {
	// Simple split by comma, respecting parens/brackets/quotes?
	// For now VERY simple: split by comma if not in brackets.
	// Valid: 'view', ['a'=>1]
	// Valid: 'view', $data

	var args []string
	depth := 0
	lastSplit := 0

	for i, c := range s {
		if c == '[' || c == '(' {
			depth++
		}
		if c == ']' || c == ')' {
			depth--
		}
		if c == ',' && depth == 0 {
			args = append(args, s[lastSplit:i])
			lastSplit = i + 1
		}
	}
	args = append(args, s[lastSplit:])
	return args
}

func parseBladeData(s string) *engine.Node {
	s = strings.TrimSpace(s)
	// Check if it's a bracketed dict ['k' => 'v']
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := s[1 : len(s)-1]
		// Split by comma
		pairs := splitBladeArgs(inner)
		dataNode := &engine.Node{Name: "data_map"}

		for _, pair := range pairs {
			parts := strings.Split(pair, "=>")
			if len(parts) == 2 {
				key := strings.Trim(strings.TrimSpace(parts[0]), "'\"") // Key is usually string
				valRaw := strings.TrimSpace(parts[1])

				// Value can be literal or variable
				var valNode *engine.Node
				if strings.HasPrefix(valRaw, "$") {
					valNode = &engine.Node{Name: "var", Value: valRaw}
				} else {
					valNode = &engine.Node{Name: "literal", Value: strings.Trim(valRaw, "'\"")}
				}

				valNode.Name = key // Encode Key as Node Name for simplicity in this specific "data_map" structure
				dataNode.Children = append(dataNode.Children, valNode)
			}
		}
		return dataNode
	}

	// Variable $data
	if strings.HasPrefix(s, "$") {
		return &engine.Node{Name: "data_var", Value: s}
	}

	return nil
}

func findEndSwitch(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@switch") {
			depth++
			pos += 7
		} else if strings.HasPrefix(s[pos:], "@endswitch") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 10
		} else {
			pos++
		}
	}
	return -1
}

type switchCase struct {
	Type  string // "case" or "default"
	Value string // Raw value for case
	Body  string
}

func splitSwitchCases(body string) []switchCase {
	var cases []switchCase

	// Scan for @case, @default at scanning depth 0 (respecting nested @switch)
	depth := 0
	pos := 0
	length := len(body)

	// Track start of current case body
	currentStart := -1
	currentType := ""
	currentVal := ""

	for pos < length {
		if strings.HasPrefix(body[pos:], "@switch") {
			depth++
			pos += 7
		} else if strings.HasPrefix(body[pos:], "@endswitch") {
			depth--
			pos += 10
		} else if depth == 0 {
			// Check for @case / @default
			if strings.HasPrefix(body[pos:], "@case") {
				// Close previous
				if currentStart != -1 {
					cases = append(cases, switchCase{
						Type:  currentType,
						Value: currentVal,
						Body:  body[currentStart:pos],
					})
				}

				// Start new
				startParen := strings.Index(body[pos:], "(")
				endParen := findBalancedParen(body[pos:])
				if startParen != -1 && endParen != -1 {
					currentVal = body[pos+startParen+1 : pos+endParen]
					currentType = "case"
					currentStart = pos + endParen + 1
					pos = currentStart
					continue
				} else {
					pos += 5
				}
			} else if strings.HasPrefix(body[pos:], "@default") {
				// Close previous
				if currentStart != -1 {
					cases = append(cases, switchCase{
						Type:  currentType,
						Value: currentVal,
						Body:  body[currentStart:pos],
					})
				}
				currentType = "default"
				currentVal = ""
				currentStart = pos + 8
				pos += 8
				continue
			} else {
				pos++
			}
		} else {
			pos++
		}
	}

	// Close last
	if currentStart != -1 {
		cases = append(cases, switchCase{
			Type:  currentType,
			Value: currentVal,
			Body:  body[currentStart:],
		})
	}

	return cases
}

func findEndPush(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@push") {
			depth++
			pos += 5
		} else if strings.HasPrefix(s[pos:], "@endpush") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 8
		} else {
			pos++
		}
	}
	return -1
}

func findEndIsset(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@isset") {
			depth++
			pos += 6
		} else if strings.HasPrefix(s[pos:], "@endisset") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 9
		} else {
			pos++
		}
	}
	return -1
}

func findEndEmpty(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@empty") {
			depth++
			pos += 6
		} else if strings.HasPrefix(s[pos:], "@endempty") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 9
		} else {
			pos++
		}
	}
	return -1
}

func findEndUnless(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@unless") {
			depth++
			pos += 7
		} else if strings.HasPrefix(s[pos:], "@endunless") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 10
		} else {
			pos++
		}
	}
	return -1
}

func findEndAuth(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@auth") {
			depth++
			pos += 5
		} else if strings.HasPrefix(s[pos:], "@endauth") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 8
		} else {
			pos++
		}
	}
	return -1
}

func findEndGuest(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@guest") {
			depth++
			pos += 6
		} else if strings.HasPrefix(s[pos:], "@endguest") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 9
		} else {
			pos++
		}
	}
	return -1
}

func findEndFor(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@for") {
			if strings.HasPrefix(s[pos:], "@foreach") {
				pos += 8
			} else if strings.HasPrefix(s[pos:], "@for") {
				depth++
				pos += 4
			}
		} else if strings.HasPrefix(s[pos:], "@endfor") {
			if strings.HasPrefix(s[pos:], "@endforeach") {
				pos += 11
			} else {
				if depth == 0 {
					return pos
				}
				depth--
				pos += 7
			}
		} else {
			pos++
		}
	}
	return -1
}

func createWriteNode(text string) *engine.Node {
	return &engine.Node{
		Name:  "__native_write",
		Value: text,
	}
}

func findBalancedParen(s string) int {
	depth := 0
	for i, c := range s {
		if c == '(' {
			depth++
		}
		if c == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func findEndWhile(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@while") {
			depth++
			pos += 6
		} else if strings.HasPrefix(s[pos:], "@endwhile") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 9
		} else {
			pos++
		}
	}
	return -1
}

func findEndForelse(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@forelse") {
			depth++
			pos += 8
		} else if strings.HasPrefix(s[pos:], "@endforelse") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 11
		} else {
			pos++
		}
	}
	return -1
}

func findEndComponent(s, tagName string) int {
	// tagName: x-alert
	closing := "</" + tagName
	start := "<" + tagName

	depth := 0
	pos := 0

	for pos < len(s) {
		// Check start tag <x-name (ensure word boundary?)
		// Simplification: HasPrefix
		if strings.HasPrefix(s[pos:], start) {
			depth++
			pos += len(start)
		} else if strings.HasPrefix(s[pos:], closing) {
			if depth == 0 {
				return pos
			}
			depth--
			pos += len(closing)
		} else {
			pos++
		}
	}
	return -1
}

func parseBladeAttributes(raw string) []*engine.Node {
	// raw: class="btn" :type="$type" dismissible
	var nodes []*engine.Node

	raw = strings.TrimSpace(raw)
	n := len(raw)
	i := 0

	for i < n {
		// Skip spaces
		for i < n && raw[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}

		// Read Key
		keyStart := i
		for i < n && raw[i] != '=' && raw[i] != ' ' {
			i++
		}
		key := raw[keyStart:i]

		val := ""

		if i < n && raw[i] == '=' {
			i++ // skip =
			// Read Value (Quoted)
			if i < n && (raw[i] == '"' || raw[i] == '\'') {
				quote := raw[i]
				i++
				valStart := i
				for i < n && raw[i] != quote {
					i++
				}
				val = raw[valStart:i]
				if i < n {
					i++
				} // skip quote
			} else {
				// Unquoted? Not standard HTML but verify
				valStart := i
				for i < n && raw[i] != ' ' {
					i++
				}
				val = raw[valStart:i]
			}
		} else {
			// Boolean attribute
			val = "true"
		}

		// Create Node
		nodes = append(nodes, &engine.Node{
			Name:  key,
			Value: val,
		})
	}
	return nodes
}

func findEndError(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@error") {
			depth++
			pos += 6
		} else if strings.HasPrefix(s[pos:], "@enderror") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 9
		} else {
			pos++
		}
	}
	return -1
}

func findEndCan(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@can") {
			depth++
			pos += 4
		} else if strings.HasPrefix(s[pos:], "@endcan") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 7
		} else {
			pos++
		}
	}
	return -1
}

func findEndCannot(s string) int {
	depth := 0
	pos := 0
	for pos < len(s) {
		if strings.HasPrefix(s[pos:], "@cannot") {
			depth++
			pos += 7
		} else if strings.HasPrefix(s[pos:], "@endcannot") {
			if depth == 0 {
				return pos
			}
			depth--
			pos += 10
		} else {
			pos++
		}
	}
	return -1
}
