
// ZenoBlade AST Parser
// Parses a ZenoBlade template string into an Abstract Syntax Tree (AST).
//
// Key Concepts:
// - Lexer: Breaks the string into a stream of tokens.
// - Parser: Consumes tokens and builds a tree structure.
// - AST Node Types:
//   - Root: The root of the AST.
//   - Text: Raw text content.
//   - Echo: {{ expression }}
//   - Directive: @name(args) ... @endname (Block) or @name(args) (Inline)
//   - Component: <x-name ...> ... </x-name>
//   - Slot: <x-slot name="..."> ... </x-slot>

// --- LEXER ---

const TOKEN_TYPES = {
    TEXT: 'TEXT',
    ECHO_START: 'ECHO_START', // {{
    ECHO_END: 'ECHO_END',     // }}
    DIRECTIVE: 'DIRECTIVE',   // @name
    PAREN_START: 'PAREN_START', // (
    PAREN_END: 'PAREN_END',     // )
    TAG_OPEN: 'TAG_OPEN',     // <x-name or <tag
    TAG_CLOSE: 'TAG_CLOSE',    // >
    TAG_END_OPEN: 'TAG_END_OPEN', // </x-name
    EOF: 'EOF'
};

class Lexer {
    constructor(input) {
        this.input = input;
        this.pos = 0;
        this.len = input.length;
    }

    nextToken() {
        if (this.pos >= this.len) return { type: TOKEN_TYPES.EOF };

        const char = this.input[this.pos];

        // 1. Check for {{ (Echo)
        if (char === '{' && this.input[this.pos + 1] === '{') {
            this.pos += 2;
            return { type: TOKEN_TYPES.ECHO_START };
        }

        // 2. Check for }} (Echo End)
        if (char === '}' && this.input[this.pos + 1] === '}') {
            this.pos += 2;
            return { type: TOKEN_TYPES.ECHO_END };
        }

        // 3. Check for @ (Directive)
        if (char === '@') {
            // Read Directive Name
            let start = this.pos + 1; // skip @
            let end = start;
            while (end < this.len && /[a-zA-Z0-9_]/.test(this.input[end])) {
                end++;
            }
            if (end > start) {
                const name = this.input.slice(start, end);
                this.pos = end;
                return { type: TOKEN_TYPES.DIRECTIVE, value: name };
            }
            // Just a raw @ if no name
        }

        // 4. Check for <x- (Component Start) or </x- (Component End)
        if (char === '<') {
            if (this.input.startsWith('<x-', this.pos)) {
                // Component Start Tag
                // Read until space or >
                let start = this.pos + 1; // skip <
                let end = start;
                while (end < this.len && /[a-zA-Z0-9\-\.]/.test(this.input[end])) {
                    end++;
                }
                const tagName = this.input.slice(start, end);
                this.pos = end;
                return { type: TOKEN_TYPES.TAG_OPEN, value: tagName }; // x-alert
            }

            if (this.input.startsWith('</x-', this.pos)) {
                // Component End Tag
                let start = this.pos + 2; // skip </
                let end = start;
                while (end < this.len && /[a-zA-Z0-9\-\.]/.test(this.input[end])) {
                    end++;
                }
                const tagName = this.input.slice(start, end);

                // Expect >
                if (this.input[end] === '>') {
                    this.pos = end + 1;
                    return { type: TOKEN_TYPES.TAG_END_OPEN, value: tagName };
                }
                // Malformed?
            }
        }

        // 5. Check for > (Tag Close) - context sensitive but lexer is dumb
        if (char === '>') {
             this.pos++;
             return { type: TOKEN_TYPES.TAG_CLOSE };
        }

        // 6. Check for ( )
        if (char === '(') {
            this.pos++;
            return { type: TOKEN_TYPES.PAREN_START };
        }
        if (char === ')') {
            this.pos++;
            return { type: TOKEN_TYPES.PAREN_END };
        }

        // 7. Text (Read until next special char)
        let end = this.pos + 1;
        while (end < this.len) {
            const c = this.input[end];
            if (c === '{' && this.input[end+1] === '{') break;
            if (c === '}' && this.input[end+1] === '}') break;
            if (c === '@') break;
            if (c === '<' && this.input.startsWith('<x-', end)) break;
            if (c === '<' && this.input.startsWith('</x-', end)) break;
            if (c === '(' || c === ')') break; // Parens are significant for directives
            if (c === '>') break; // Tag close
            end++;
        }

        const text = this.input.slice(this.pos, end);
        this.pos = end;
        return { type: TOKEN_TYPES.TEXT, value: text };
    }
}


// --- PARSER ---

export function parse(input) {
    const lexer = new Lexer(input);
    const root = { type: 'Root', children: [] };
    const stack = [root]; // Stack of parent nodes

    let token = lexer.nextToken();

    while (token.type !== TOKEN_TYPES.EOF) {
        const parent = stack[stack.length - 1];

        if (token.type === TOKEN_TYPES.TEXT) {
            parent.children.push({ type: 'Text', value: token.value });
        }

        else if (token.type === TOKEN_TYPES.ECHO_START) {
            // Read until ECHO_END
            // Simple approach: Assume balanced or just find }}
            // Lexer gives tokens, so we consume TEXT/PARENS until ECHO_END
            let content = '';
            token = lexer.nextToken();
            while (token.type !== TOKEN_TYPES.ECHO_END && token.type !== TOKEN_TYPES.EOF) {
                if (token.value) content += token.value;
                // Reconstruct exact string if tokens were split
                if (token.type === TOKEN_TYPES.PAREN_START) content += '(';
                if (token.type === TOKEN_TYPES.PAREN_END) content += ')';
                // Directives inside echo? e.g. {{ @json(...) }} - lexer splits @json
                if (token.type === TOKEN_TYPES.DIRECTIVE) content += '@' + token.value;
                token = lexer.nextToken();
            }
            parent.children.push({ type: 'Echo', value: content.trim() });
        }

        else if (token.type === TOKEN_TYPES.DIRECTIVE) {
            const name = token.value;

            // Check if it's a closing directive (@endif, @endforeach)
            if (name.startsWith('end')) {
                // Pop stack until matching open directive
                // e.g. @endif matches @if
                // logic: stack top should be @if
                const expectedStart = name.substring(3); // if, foreach
                if (parent.type === 'Directive' && (
                    parent.name === expectedStart ||
                    (expectedStart === 'unless' && parent.name === 'unless') || // unless/endunless
                    (expectedStart === 'empty' && parent.name === 'empty') ||
                    (expectedStart === 'isset' && parent.name === 'isset') ||
                    (expectedStart === 'switch' && parent.name === 'switch')
                )) {
                    stack.pop();
                } else {
                    // Mismatch or stray closing tag, treat as text
                    parent.children.push({ type: 'Text', value: '@' + name });
                }
            }
            else if (name === 'else' || name === 'elseif' || name === 'case' || name === 'default' || name === 'break') {
                // Control flow siblings
                // For @else/@elseif: They effectively close the previous block and start a new one AT THE SAME LEVEL
                // But structure-wise:
                // If
                //   Then-Block
                //   Else-Block
                // The parser should attach 'else' to the parent 'if'.

                // Current simplified stack:
                // Root -> If
                // When we see @else, we are inside If.
                // We should close the current 'children' accumulation and switch to 'alternate'?
                // Or just treat @else as a child node?

                // Let's treat them as Directive nodes.
                // But @else doesn't have an @endelse. It is closed by @endif.

                // Special handling:
                // If parent is @if, @else closes the current block?
                // Actually, Blade structure:
                // @if
                //    ...
                // @else
                //    ...
                // @endif

                // We can model this as:
                // Directive(if) -> children: [ ... content ..., Directive(else, children: [...]) ]
                // This nesting is tricky with a simple stack.

                // Alternative: Flat list of children for @if, containing Text, Directive(else), Text...
                // Code generator handles the logic.
                // So @else is just a Directive node inside @if's children.
                // And @else starts its own block? No, @else is a single point.

                // Wait, @else has a body.
                // So @else should be a Node that takes over collection of children until @endif?
                // This suggests @if should have `branches`.

                // SIMPLEST AST for Code Gen:
                // Just push @else as a node. Code gen outputs `} else {`.
                // This relies on the structure being correct.

                let args = null;
                // Check for args ( )
                // Lexer peek?
                // We need to consume args if present.
                // But nextToken() advances.
                // Check next token.
                const next = lexer.nextToken();
                if (next.type === TOKEN_TYPES.PAREN_START) {
                    args = parseBalancedParens(lexer); // Consumes until )
                } else {
                    // Push back or process as text/space?
                    // Lexer is simple linear.
                    // If not paren, it's start of body.
                    // We need to push 'next' back or handle it.
                    // Let's handle it: it's the first token of body.
                    // But we are constructing the Directive node first.

                    // Hack: Store 'next' to be processed next loop?
                    // Or recursively call processToken?
                    // Let's implement `peek` or `pushBack` in parser (not lexer)
                    // For now: assume arguments are immediate.
                    // If no args, we start body.

                    // We need to process `next` as content.
                    // But we haven't pushed the Directive yet.
                }

                const node = { type: 'Directive', name, args, children: [] };
                parent.children.push(node);

                // Does this directive start a block?
                // @else -> Yes?
                // Actually, code gen `} else {` implies it doesn't nest in AST, it separates siblings.
                // But for AST purity, `@if` should contain `@else`?
                // Let's keep it flat in the parent's children list.
                // The Parent is the @if (or Root).
                // So @if children = [ Text, Directive(else), Text ].
                // CodeGen: if() { ... } else { ... }
                // This works!

                if (args !== null) {
                    // If we consumed args, we are good.
                } else {
                   // If we read a token that wasn't parens, we need to add it to children!
                   // But which children? The `parent`'s children.
                   // Wait, `next` was popped.
                   if (next.type !== TOKEN_TYPES.PAREN_START) {
                       // We must process `next` as a child of `parent` (the @if block).
                       // We can't easily push back to lexer.
                       // Just add it now.
                       addTokenToNode(parent, next);
                   }
                }
            }
            else {
                // Regular Directive (@if, @foreach, @class, @json)
                let args = null;
                const next = lexer.nextToken();
                if (next.type === TOKEN_TYPES.PAREN_START) {
                    args = parseBalancedParens(lexer);
                } else {
                    if (next.type !== TOKEN_TYPES.EOF) {
                         // Directive without args (e.g. @break, @default)
                         // Or @class with space? Blade requires no space for args? Usually yes.
                         // If space, it's not args.
                         addTokenToNode(parent, next);
                    }
                }

                const node = { type: 'Directive', name, args, children: [] };
                parent.children.push(node);

                // Check if it's a Block Directive (needs closing)
                const isBlock = ['if', 'unless', 'isset', 'empty', 'switch', 'foreach', 'push', 'component'].includes(name);
                // Also <x-component> is handled by TAG_OPEN.
                // @component is legacy blade.

                if (isBlock) {
                    stack.push(node);
                }
            }
        }

        else if (token.type === TOKEN_TYPES.TAG_OPEN) {
            // <x-name
            const tagName = token.value; // x-alert

            // Parse Attributes until > or />
            // We need to consume tokens until TAG_CLOSE.
            // Attributes can be complex text with spaces.
            // Lexer gives TEXT tokens.

            let attrStr = "";
            let isSelfClosing = false;

            token = lexer.nextToken();
            while (token.type !== TOKEN_TYPES.TAG_CLOSE && token.type !== TOKEN_TYPES.EOF) {
                // If text ends with /, it might be self closing
                if (token.type === TOKEN_TYPES.TEXT) {
                    let val = token.value;
                    if (val.trim().endsWith('/')) {
                        isSelfClosing = true;
                        val = val.replace('/', ''); // Strip /
                    }
                    attrStr += val;
                } else if (token.type === TOKEN_TYPES.ECHO_START) {
                    // Echo in attribute? <x-alert :msg="{{ $val }}">
                    // We need to consume echo
                    attrStr += "{{";
                    // ... consume echo content ...
                    // This creates complexity.
                    // Let's assume attributes are parsed as raw string for now.
                    // Lexer splits {{.
                    // We reconstruct.
                } else if (token.value) {
                    attrStr += token.value;
                }
                token = lexer.nextToken();
            }

            const node = {
                type: 'Component',
                tagName,
                attrs: attrStr.trim(),
                children: []
            };
            parent.children.push(node);

            if (!isSelfClosing) {
                stack.push(node);
            }
        }

        else if (token.type === TOKEN_TYPES.TAG_END_OPEN) {
            // </x-name
            const tagName = token.value;
            // Expect TAG_CLOSE > check done in lexer? Lexer returns TAG_END_OPEN only if > follows?
            // Lexer logic: "Expect >". If found, returns TAG_END_OPEN.

            // Close matching component
            // Stack top should be Component with same tagName
            if (parent.type === 'Component' && parent.tagName === tagName) {
                stack.pop();
            } else {
                // Mismatch, treat as text
                parent.children.push({ type: 'Text', value: `</${tagName}>` });
            }
        }

        token = lexer.nextToken();
    }

    return root;
}

// Helper to consume tokens until balanced paren end
function parseBalancedParens(lexer) {
    let content = '';
    let depth = 1;
    let token = lexer.nextToken();

    while (depth > 0 && token.type !== TOKEN_TYPES.EOF) {
        if (token.type === TOKEN_TYPES.PAREN_START) depth++;
        if (token.type === TOKEN_TYPES.PAREN_END) {
            depth--;
            if (depth === 0) break;
        }

        // Reconstruct content
        if (token.value) content += token.value;
        else if (token.type === TOKEN_TYPES.ECHO_START) content += '{{';
        else if (token.type === TOKEN_TYPES.ECHO_END) content += '}}';
        else if (token.type === TOKEN_TYPES.PAREN_START) content += '(';
        else if (token.type === TOKEN_TYPES.PAREN_END) content += ')';
        // Add spaces if needed? Lexer strips spaces between tokens?
        // Lexer implementation above:
        // TEXT reads until special char.
        // It PRESERVES spaces in TEXT token.
        // But between TEXT and PAREN?
        // Lexer loop: `end` starts at `pos + 1`.
        // If `(` is at `pos`, it returns PAREN.
        // It does NOT skip whitespace.
        // So whitespace should be in TEXT tokens.

        token = lexer.nextToken();
    }
    return content;
}

function addTokenToNode(node, token) {
    if (token.type === TOKEN_TYPES.TEXT) node.children.push({ type: 'Text', value: token.value });
    else if (token.type === TOKEN_TYPES.ECHO_START) node.children.push({ type: 'Text', value: '{{' }); // Broken echo
    else if (token.type === TOKEN_TYPES.PAREN_START) node.children.push({ type: 'Text', value: '(' });
    // ... handle others as text fallback
}
