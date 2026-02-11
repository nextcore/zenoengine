
// ZenoBlade Compiler
// Parses ZenoBlade syntax (@if, @foreach, {{ }}) into JavaScript logic.
// Simulates a "render function" by generating a string of code that builds DOM structure (Virtual DOM or HTML string).
// For simplicity in this experiment, we will compile to HTML string with inline JS logic (similar to PHP/Blade backend compilation),
// then re-evaluate it on every change.

/**
 * Compiles a ZenoBlade template string into a render function body.
 * @param {string} template - The raw ZenoBlade template.
 * @returns {Function} - A function that takes `data` and returns an HTML string.
 */
export function compile(template) {
    let code = "let _out = '';\n";
    code += "with(this) {\n"; // Use 'with' to scope data properties (e.g. {{ message }} instead of {{ data.message }})

    let cursor = 0;
    const len = template.length;

    while (cursor < len) {
        // Find next tag start (@ or {{)
        const nextTag = template.slice(cursor).search(/@|\{\{/);

        if (nextTag === -1) {
            // No more tags, append remaining text
            const text = template.slice(cursor);
            code += `_out += \`${escapeBackticks(text)}\`;\n`;
            break;
        }

        // Append text before tag
        if (nextTag > 0) {
            const text = template.slice(cursor, cursor + nextTag);
            code += `_out += \`${escapeBackticks(text)}\`;\n`;
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
                // Invalid syntax, treat as text
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
        else if (template.startsWith('@foreach', cursor)) {
            // Foreach: @foreach(items as item)
            // We need to parse "items as item"
            const openParen = template.indexOf('(', cursor);
            const closeParen = findBalancedParen(template, openParen);

            if (openParen !== -1 && closeParen !== -1) {
                const raw = template.slice(openParen + 1, closeParen);
                const [list, item] = raw.split(/\s+as\s+/).map(s => s.trim());

                // Using standard JS for loop or map?
                // For loop is easier to append to string buffer.
                // We need to support index too? @foreach($list as $key => $val)?
                // Simple version: for (let item of list)
                code += `for (let ${item} of ${list}) {\n`;

                cursor = closeParen + 1;
            } else {
                cursor += 8;
            }
        }
        else if (template.startsWith('@endforeach', cursor)) {
            code += `}\n`;
            cursor += 11;
        }
        else if (template.startsWith('@click', cursor)) {
            // This usually appears inside HTML attributes, handled differently in standard parsers.
            // BUT here we are generating HTML string.
            // HTML string doesn't support binding functions directly.
            // We need a strategy.
            // 1. Generate unique ID for element.
            // 2. Add event listener after render.
            // OR
            // 3. Use inline 'onclick' calling global handler? (Bad practice)
            //
            // BETTER APPROACH:
            // Since we are building a string, we can't easily attach events.
            // The "Compiler" should ideally produce DOM nodes or VDOM.
            //
            // ALTERNATIVE:
            // We produce HTML with special markers data-z-click="methodName".
            // Then after setting innerHTML, we scan for these attributes and bind events.

            // To do this, we need to detect if we are inside a tag?
            // "Compiling to String" has this limitation.
            // Let's assume the user writes standard HTML: <button @click="fn">
            // Our regex replacement in `zeno.js` did this.
            // The compiler needs to preserve this or transform it.
            // Let's output it as a custom attribute `data-z-click`?

            // If we encounter `@click="fn"` in the stream:
            // It will be parsed as text because it doesn't start with @ at the block level?
            // Wait, my scanner searches for `@`.
            // So `<button @click` -> `@click` is found.

            // We need to parse the attribute value.
            // Expecting: ="method" or ="method()"
            // Let's assume standard attribute syntax.
            // We'll replace `@click="foo"` with `data-z-click="foo"` in the output string.

            // Look ahead for ="..."
            const match = template.slice(cursor).match(/^@click=["'](.*?)["']/);
            if (match) {
                const handler = match[1];
                code += `_out += 'data-z-click="${handler}"';\n`;
                cursor += match[0].length;
            } else {
                // Just text
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

    // console.log("COMPILED CODE:", code);

    try {
        return new Function(code);
    } catch (e) {
        console.error("Compilation Error:", e);
        return () => "Error compiling template";
    }
}

function escapeBackticks(str) {
    return str.replace(/`/g, '\\`').replace(/\$/g, '\\$'); // Escape backticks and template literal interpolation
}

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
