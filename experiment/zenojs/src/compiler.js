
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
        // Find next tag start (@ or {{)
        const nextTag = template.slice(cursor).search(/@|\{\{/);

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
            // Unless: @unless(cond) -> if (!cond)
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
            // Isset: @isset(var) -> if (typeof var !== 'undefined' && var !== null)
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
            // Empty: @empty(var) -> if (!var || (Array.isArray(var) && var.length === 0))
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
            // Switch: @switch(val)
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
            // Case: @case(val)
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
            // Enhanced Loop: injecting $loop variable
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);

            if (openParen !== -1 && closeParen !== -1) {
                const raw = template.slice(openParen + 1, closeParen);
                const [list, item] = raw.split(/\s+as\s+/).map(s => s.trim());

                // Generate loop code
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
                        depth: 1 // TODO: nested loop depth tracking
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
            // @json(data) -> JSON.stringify(data)
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
            // @class(['p-4', 'bg-red' => hasError])
            // We need a helper function for this complex logic.
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);
            if (openParen !== -1 && closeParen !== -1) {
                const args = template.slice(openParen + 1, closeParen);
                // We emit class="..." attribute
                code += `_out += 'class="' + _helpers.classNames(${args}) + '"';\n`;
                cursor = closeParen + 1;
            } else {
                cursor += 6;
            }
        }
        else if (template.startsWith('@style', cursor)) {
            // @style(['color: red' => hasError])
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
            // @checked(cond) -> if(cond) out += 'checked="checked"'
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
            // @click="method" -> data-z-click="method"
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
             // Unknown directive, treat as text
             code += `_out += '@';\n`;
             cursor += 1;
        }
    }

    code += "}\n"; // close with
    code += "return _out;";

    return code;
}

function escapeBackticks(str) {
    return str.replace(/`/g, '\\`').replace(/\$/g, '\\$'); // Escape backticks and template literal interpolation
}
