
// ZenoBlade Compiler
// Parses ZenoBlade syntax (@if, @foreach, {{ }}) into JavaScript logic.
// Simulates a "render function" by generating a string of code that builds DOM structure (Virtual DOM or HTML string).
// For simplicity in this experiment, we will compile to HTML string with inline JS logic (similar to PHP/Blade backend compilation),
// then re-evaluate it on every change.

/**
 * Compiles a ZenoBlade template string into a render function body (source code).
 * @param {string} template - The raw ZenoBlade template.
 * @returns {string} - The JavaScript source code for the render function.
 */
export function compile(template) {
    let code = "let _out = '';\n";
    // Setup helpers
    code += "const _helpers = this.$helpers || {};\n";
    code += "const _renderComponent = this.renderComponent || (() => '');\n";

    code += "with(this) {\n"; // Use 'with' to scope data properties (e.g. {{ message }} instead of {{ data.message }})

    let cursor = 0;
    const len = template.length;

    // Helper to find balanced parens
    function findBalancedParen(str, start) {
        let depth = 0;
        for (let i = start; i < str.length; i++) {
            if (str[i] === '(') depth++;
            if (str[i] === ')') {
                depth--;
                if (depth === 0) return i;
            }
        }
        return -1;
    }

    while (cursor < len) {
        // Find next tag start (@ or {{ or <x-)
        const nextTag = template.slice(cursor).search(/@|\{\{|<x-/);

        if (nextTag === -1) {
            // No more tags, append remaining text
            const text = template.slice(cursor);
            if (text) {
                code += `_out += \`${escapeBackticks(text)}\`;\n`;
            }
            break;
        }

        // Append text before tag
        if (nextTag > 0) {
            const text = template.slice(cursor, cursor + nextTag);
            if (text) {
                code += `_out += \`${escapeBackticks(text)}\`;\n`;
            }
        }

        cursor += nextTag;

        // Handle Tags
        if (template.startsWith('{{', cursor)) {
            // Echo: {{ expression }}
            const end = template.indexOf('}}', cursor);
            if (end === -1) break; // Error: Unclosed tag

            const expr = template.slice(cursor + 2, end).trim();
            // Basic XSS protection: we should escape, but for now raw.
            // code += `_out += String(${expr}).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');\n`;
            code += `_out += ${expr};\n`;

            cursor = end + 2;
        }
        else if (template.startsWith('<x-', cursor)) {
            // Component: <x-alert type="error">Body</x-alert>
            // Need to parse attributes and body.
            // Finding the closing tag is hard with nested components without a parser.
            // Using a simple stack-based approach or recursive regex might work for well-formed HTML.

            // 1. Parse Tag Name
            const tagStart = cursor;
            let tagNameEnd = template.indexOf(' ', tagStart);
            const tagClose = template.indexOf('>', tagStart);

            if (tagNameEnd === -1 || tagNameEnd > tagClose) tagNameEnd = tagClose;

            const rawTagName = template.slice(tagStart + 1, tagNameEnd); // x-alert
            const componentName = rawTagName.substring(2); // alert

            // 2. Parse Attributes
            const attrStr = template.slice(tagNameEnd, tagClose);
            const attrs = parseAttributes(attrStr);

            // 3. Check Self-Closing
            const isSelfClosing = template[tagClose - 1] === '/';

            if (isSelfClosing) {
                // Generate render call
                code += `_out += _renderComponent('${componentName}', {${attrs}}, {});\n`;
                cursor = tagClose + 1;
            } else {
                // 4. Parse Body (Slots)
                // Find matching closing tag </x-alert>
                const closingTag = `</${rawTagName}>`;
                const bodyStart = tagClose + 1;

                // Naive search for closing tag (doesn't handle nesting of same tag perfectly but works for simple cases)
                // Better: simple counter
                let depth = 1;
                let scan = bodyStart;
                let bodyEnd = -1;

                while (scan < len) {
                    if (template.startsWith(`<${rawTagName}`, scan)) {
                        depth++;
                        scan += rawTagName.length + 1;
                    } else if (template.startsWith(closingTag, scan)) {
                        depth--;
                        if (depth === 0) {
                            bodyEnd = scan;
                            break;
                        }
                        scan += closingTag.length;
                    } else {
                        scan++;
                    }
                }

                if (bodyEnd === -1) {
                    // Unclosed component, treat as text
                     code += `_out += '<x-${componentName}';\n`;
                     cursor = tagNameEnd;
                     continue;
                }

                const bodyContent = template.slice(bodyStart, bodyEnd);

                // 5. Parse Slots inside Body
                // If body contains <x-slot name="header">...</x-slot>, extract them.
                // Otherwise treat whole body as 'default' slot.

                // Helper to extract slots from string
                // We'll generate an object: { default: () => `...`, header: () => `...` }

                // Simplification: We assume slots are top-level children of the component body.
                // We won't parse recursively right now, just regex/split.

                const slotMap = parseSlots(bodyContent);
                let slotsCode = '{';
                for (const [name, content] of Object.entries(slotMap)) {
                    // Recursive compile for slot content!
                    // Slot content must be evaluated in PARENT scope (here).
                    // So we compile it into a sub-function or IIFE that returns string.
                    // But `compile` returns code for a function body.
                    // We can wrap it: `"${name}": () => { let _out=''; with(this){ ${compile(content)} } return _out; },`

                    // Crucial: Slots need access to PARENT scope variables.
                    // "with(this)" inside the arrow function should work if `this` is preserved?
                    // Arrow function preserves `this`. Yes.

                    // However, `compile` adds `let _out = '';` and `with(this)`.
                    // We can reuse `compile` but strip the wrapper if needed, or just let it be fully self-contained.
                    // The `compile` function output starts with `let _out...` and ends with `return _out`.
                    // Perfect for a function body.

                    const compiledSlot = compile(content);
                    // Indent for readability (optional)
                    slotsCode += `\n"${name}": (() => {\n${compiledSlot}\n}),`;
                }
                slotsCode += '}';

                code += `_out += _renderComponent('${componentName}', {${attrs}}, ${slotsCode});\n`;

                cursor = bodyEnd + closingTag.length;
            }
        }
        else if (template.startsWith('@if', cursor)) {
            // If: @if(cond)
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);

            if (openParen !== -1 && closeParen !== -1) {
                const cond = template.slice(openParen + 1, closeParen);
                code += `if (${cond}) {\n`;
                cursor = closeParen + 1;
            } else {
                code += `_out += '@if';\n`;
                cursor += 3;
            }
        }
        else if (template.startsWith('@elseif', cursor)) {
             const openParen = template.indexOf('(', cursor);
             const closeParen = findBalancedParen(template, openParen);
             if (openParen !== -1 && closeParen !== -1) {
                 const cond = template.slice(openParen + 1, closeParen);
                 code += `} else if (${cond}) {\n`;
                 cursor = closeParen + 1;
             } else {
                 cursor += 7;
             }
        }
        else if (template.startsWith('@else', cursor)) {
            code += `} else {\n`;
            cursor += 5;
        }
        else if (template.startsWith('@endif', cursor)) {
            code += `}\n`;
            cursor += 6;
        }
        else if (template.startsWith('@unless', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const cond = template.slice(openParen + 1, closeParen);
                code += `if (!(${cond})) {\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 7;
            }
        }
        else if (template.startsWith('@endunless', cursor)) {
            code += `}\n`;
            cursor += 10;
        }
        else if (template.startsWith('@isset', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const v = template.slice(openParen + 1, closeParen);
                code += `if (typeof ${v} !== 'undefined' && ${v} !== null) {\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 6;
            }
        }
        else if (template.startsWith('@endisset', cursor)) {
            code += `}\n`;
            cursor += 9;
        }
        else if (template.startsWith('@empty', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const v = template.slice(openParen + 1, closeParen);
                code += `if (!${v} || (Array.isArray(${v}) && ${v}.length === 0)) {\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 6;
            }
        }
        else if (template.startsWith('@endempty', cursor)) {
            code += `}\n`;
            cursor += 9;
        }
        else if (template.startsWith('@switch', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const val = template.slice(openParen + 1, closeParen);
                code += `switch (${val}) {\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 7;
            }
        }
        else if (template.startsWith('@case', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const val = template.slice(openParen + 1, closeParen);
                code += `case ${val}:\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 5;
            }
        }
        else if (template.startsWith('@break', cursor)) {
            code += `break;\n`;
            cursor += 6;
        }
        else if (template.startsWith('@default', cursor)) {
            code += `default:\n`;
            cursor += 8;
        }
        else if (template.startsWith('@endswitch', cursor)) {
            code += `}\n`;
            cursor += 10;
        }
        else if (template.startsWith('@foreach', cursor)) {
            // Foreach: @foreach(items as item)
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);

            if (openParen !== -1 && closeParen !== -1) {
                const raw = template.slice(openParen + 1, closeParen);
                const [list, item] = raw.split(/\s+as\s+/).map(s => s.trim());

                code += `
                let _loopIndex = 0;
                const _loopList = ${list} || [];
                const _loopCount = _loopList.length;
                for (let ${item} of _loopList) {
                    const loop = {
                        index: _loopIndex,
                        iteration: _loopIndex + 1,
                        remaining: _loopCount - (_loopIndex + 1),
                        count: _loopCount,
                        first: _loopIndex === 0,
                        last: _loopIndex === _loopCount - 1,
                        even: (_loopIndex + 1) % 2 === 0,
                        odd: (_loopIndex + 1) % 2 !== 0,
                        depth: 1
                    };
                `;

                cursor = closeParen + 1;
            } else {
                cursor += 8;
            }
        }
        else if (template.startsWith('@endforeach', cursor)) {
            code += `
                _loopIndex++;
            }\n`;
            cursor += 11;
        }
        else if (template.startsWith('@json', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const data = template.slice(openParen + 1, closeParen);
                code += `_out += JSON.stringify(${data});\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 5;
            }
        }
        else if (template.startsWith('@class', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const args = template.slice(openParen + 1, closeParen);
                code += `_out += 'class="' + _helpers.classNames(${args}) + '"';\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 6;
            }
        }
        else if (template.startsWith('@style', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const args = template.slice(openParen + 1, closeParen);
                code += `_out += 'style="' + _helpers.styleNames(${args}) + '"';\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 6;
            }
        }
        else if (template.startsWith('@checked', cursor)) {
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const cond = template.slice(openParen + 1, closeParen);
                code += `if (${cond}) { _out += 'checked="checked"'; }\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 8;
            }
        }
        else if (template.startsWith('@selected', cursor)) {
             const openParen = template.indexOf('(', cursor);
             const closeParen = findBalancedParen(template, openParen);
             if (openParen !== -1 && closeParen !== -1) {
                 const cond = template.slice(openParen + 1, closeParen);
                 code += `if (${cond}) { _out += 'selected="selected"'; }\n`;
                 cursor = closeParen + 1;
             } else {
                 cursor += 9;
             }
        }
        else if (template.startsWith('@disabled', cursor)) {
             const openParen = template.indexOf('(', cursor);
             const closeParen = findBalancedParen(template, openParen);
             if (openParen !== -1 && closeParen !== -1) {
                 const cond = template.slice(openParen + 1, closeParen);
                 code += `if (${cond}) { _out += 'disabled="disabled"'; }\n`;
                 cursor = closeParen + 1;
             } else {
                 cursor += 9;
             }
        }
        else if (template.startsWith('@readonly', cursor)) {
             const openParen = template.indexOf('(', cursor);
             const closeParen = findBalancedParen(template, openParen);
             if (openParen !== -1 && closeParen !== -1) {
                 const cond = template.slice(openParen + 1, closeParen);
                 code += `if (${cond}) { _out += 'readonly="readonly"'; }\n`;
                 cursor = closeParen + 1;
             } else {
                 cursor += 9;
             }
        }
        else if (template.startsWith('@required', cursor)) {
             const openParen = template.indexOf('(', cursor);
             const closeParen = findBalancedParen(template, openParen);
             if (openParen !== -1 && closeParen !== -1) {
                 const cond = template.slice(openParen + 1, closeParen);
                 code += `if (${cond}) { _out += 'required="required"'; }\n`;
                 cursor = closeParen + 1;
             } else {
                 cursor += 9;
             }
        }
        else if (template.startsWith('@click', cursor)) {
            const match = template.slice(cursor).match(/^@click=["'](.*?)["']/);
            if (match) {
                const handler = match[1];
                code += `_out += 'data-z-click="${handler}"';\n`;
                cursor += match[0].length;
            } else {
                 code += `_out += '@click';\n`;
                 cursor += 6;
            }
        }
        else {
             code += `_out += '@';\n`;
             cursor += 1;
        }
    }

    code += "}\n";
    code += "return _out;";

    return code;
}

function escapeBackticks(str) {
    return str.replace(/`/g, '\\`').replace(/\$/g, '\\$');
}

// Parses string like: type="error" :count="10"
function parseAttributes(str) {
    const attrs = [];
    const regex = /(:)?([a-zA-Z0-9_-]+)=["'](.*?)["']/g;
    let match;
    while ((match = regex.exec(str)) !== null) {
        const isDynamic = !!match[1];
        const key = match[2];
        const val = match[3];

        if (isDynamic) {
            // Key is 'count', Val is expression '10'
            attrs.push(`"${key}": ${val}`);
        } else {
            // Key is 'type', Val is string 'error'
            attrs.push(`"${key}": "${val}"`);
        }
    }
    return attrs.join(', ');
}

// Splits body into slots based on <x-slot name="..."> tags
// Returns { default: "content", header: "content" }
function parseSlots(body) {
    const slots = {};
    let defaultContent = "";

    // Find all <x-slot ...>...</x-slot>
    // Remove them from body, what remains is default slot.

    let cursor = 0;
    const len = body.length;
    let lastEnd = 0;

    // Simple regex for top-level slots?
    // Nested slots are hard with regex.
    // For now, let's assume no nested <x-slot> inside <x-slot>.

    const slotStartRegex = /<x-slot\s+name=["'](.*?)["']>/g;
    let match;

    while ((match = slotStartRegex.exec(body)) !== null) {
        // Found start
        const slotName = match[1];
        const startIdx = match.index;
        const innerStart = startIdx + match[0].length;

        // Find closing </x-slot>
        const endIdx = body.indexOf('</x-slot>', innerStart);
        if (endIdx === -1) break; // Error

        // Append previous content to default
        defaultContent += body.slice(lastEnd, startIdx);

        // Extract slot content
        slots[slotName] = body.slice(innerStart, endIdx);

        lastEnd = endIdx + 9; // length of </x-slot>

        // Advance regex index
        slotStartRegex.lastIndex = lastEnd;
    }

    defaultContent += body.slice(lastEnd);

    if (defaultContent.trim()) {
        slots['default'] = defaultContent;
    }

    return slots;
}
