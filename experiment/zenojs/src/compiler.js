
// ZenoBlade AST-Based Compiler
import { parse } from './parser.js';

export function compile(template) {
    const ast = parse(template);

    // Check for @extends
    // Ideally, @extends should be the first thing or at root level.
    // If @extends exists, we need to wrap the whole thing.

    // Scan AST for extends directive
    let layoutName = null;
    let sections = {}; // { name: AST_Node_List }

    // We need to traverse root children to find @extends and @section
    // BUT only at root level usually.

    const rootChildren = ast.children;
    const filteredChildren = [];

    for (const node of rootChildren) {
        if (node.type === 'Directive' && node.name === 'extends') {
            layoutName = node.args.replace(/['"]/g, ''); // strip quotes
        }
        else if (node.type === 'Directive' && node.name === 'section') {
            const sectionName = node.args.replace(/['"]/g, '');
            sections[sectionName] = node.children;
        }
        else {
            filteredChildren.push(node);
        }
    }

    // Code Gen
    let code = "let _out = '';\n";
    code += "const _helpers = this.$helpers || {};\n";
    code += "const _renderComponent = this.renderComponent || (() => '');\n";
    code += "const _renderLayout = this.renderLayout || (() => '');\n";

    code += "with(this) {\n";

    if (layoutName) {
        // Layout Mode
        // We compile sections into functions passed to layout

        // Default content (outside sections) - usually ignored in Blade if extends?
        // Actually, everything outside @section is ignored if @extends is present.
        // But for safety let's assume strict Blade: only @section matters.

        let sectionsCode = '{';
        for (const [name, children] of Object.entries(sections)) {
            const body = children.map(codegen).join('');
            sectionsCode += `\n"${name}": (() => {\nlet _out='';\n${body}\nreturn _out;\n}),`;
        }
        sectionsCode += '}';

        code += `_out += _renderLayout('${layoutName}', ${sectionsCode});\n`;

    } else {
        // Normal Mode
        code += codegen({ type: 'Root', children: rootChildren });
    }

    code += "}\n";
    code += "return _out;";
    return code;
}

function codegen(node) {
    let code = '';

    if (node.type === 'Root') {
        for (const child of node.children) {
            code += codegen(child);
        }
    }
    else if (node.type === 'Text') {
        if (node.value) {
            code += `_out += \`${escapeBackticks(node.value)}\`;\n`;
        }
    }
    else if (node.type === 'Echo') {
        code += `_out += ${node.value};\n`;
    }
    else if (node.type === 'Directive') {
        const name = node.name;
        const args = node.args ? node.args.trim() : '';
        const childrenCode = node.children.map(codegen).join('');

        if (name === 'if') {
            code += `if (${args}) {\n${childrenCode}\n}\n`;
        }
        else if (name === 'elseif') {
             code += `} else if (${args}) {\n${childrenCode}\n`;
        }
        else if (name === 'else') {
             code += `} else {\n${childrenCode}\n`;
        }
        else if (name === 'endif' || name === 'endunless' || name === 'endisset' || name === 'endempty' || name === 'endswitch' || name === 'endforeach' || name === 'endsection') {
             // Closing tags are handled by parent block structure in AST
        }
        else if (name === 'unless') {
             code += `if (!(${args})) {\n${childrenCode}\n}\n`;
        }
        else if (name === 'isset') {
             code += `if (typeof ${args} !== 'undefined' && ${args} !== null) {\n${childrenCode}\n}\n`;
        }
        else if (name === 'empty') {
             code += `if (!${args} || (Array.isArray(${args}) && ${args}.length === 0)) {\n${childrenCode}\n}\n`;
        }
        else if (name === 'switch') {
             code += `switch (${args}) {\n${childrenCode}\n}\n`;
        }
        else if (name === 'case') {
             code += `case ${args}:\n${childrenCode}\n`;
        }
        else if (name === 'break') {
             code += `break;\n`;
        }
        else if (name === 'default') {
             code += `default:\n${childrenCode}\n`;
        }
        else if (name === 'foreach') {
             const [list, item] = args.split(/\s+as\s+/).map(s => s.trim());
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
                    ${childrenCode}
                    _loopIndex++;
                }\n`;
        }
        else if (name === 'json') {
             code += `_out += JSON.stringify(${args});\n`;
        }
        else if (name === 'class') {
             code += `_out += 'class="' + _helpers.classNames(${args}) + '"';\n`;
        }
        else if (name === 'style') {
             code += `_out += 'style="' + _helpers.styleNames(${args}) + '"';\n`;
        }
        else if (['checked', 'selected', 'disabled', 'readonly', 'required'].includes(name)) {
             code += `if (${args}) { _out += '${name}="${name}"'; }\n`;
        }
        else if (name === 'click') {
             const handler = args.replace(/^["']|["']$/g, '');
             code += `_out += 'data-z-click="${handler}"';\n`;
        }
        else if (name === 'yield') {
             // @yield('name')
             // In layout mode, this outputs the section content passed from child
             // The compiled layout function will receive `sections` object.
             // But here we are in `with(this)`.
             // We need to inject `sections` into scope or `_renderLayout` logic.
             // Wait, `yield` is used IN THE LAYOUT.
             // When compiling layout, we need to access passed sections.
             // The runtime `renderLayout` will wrap the layout render call.
             // It should pass `sections` in `this` or as arg?
             // Let's assume `this.$sections` exists.
             const sectionName = args.replace(/['"]/g, '');
             code += `if (this.$sections && this.$sections['${sectionName}']) { _out += this.$sections['${sectionName}'](); }\n`;
        }
        else if (name === 'section') {
             // @section inside a normal view (not extending) is just a block?
             // Or output immediately?
             // Blade outputs immediately if not extending.
             // But if extending, it's captured.
             // Logic above handled @extends check.
             // If we are here, it means NO @extends was found (or we are recursing inside something else).
             // If no extends, @section usually outputs its content.
             code += childrenCode;
        }
        else {
             code += `_out += '@${name}';\n`;
        }
    }
    else if (node.type === 'Component') {
        const tagName = node.tagName.substring(2);
        const attrs = parseAttributes(node.attrs || '');

        const slots = {};
        let defaultSlotContent = [];

        for (const child of node.children) {
            if (child.type === 'Component' && child.tagName === 'x-slot') {
                const nameMatch = child.attrs.match(/name=["'](.*?)["']/);
                const slotName = nameMatch ? nameMatch[1] : 'default';
                slots[slotName] = child.children.map(codegen).join('');
            } else {
                defaultSlotContent.push(child);
            }
        }

        if (defaultSlotContent.length > 0) {
            slots['default'] = defaultSlotContent.map(codegen).join('');
        }

        let slotsCode = '{';
        for (const [name, body] of Object.entries(slots)) {
            slotsCode += `\n"${name}": (() => {\nlet _out='';\n${body}\nreturn _out;\n}),`;
        }
        slotsCode += '}';

        code += `_out += _renderComponent('${tagName}', {${attrs}}, ${slotsCode});\n`;
    }

    return code;
}

function escapeBackticks(str) {
    return str.replace(/`/g, '\\`').replace(/\$/g, '\\$');
}

function parseAttributes(str) {
    const attrs = [];
    const regex = /(:)?([a-zA-Z0-9_-]+)=["'](.*?)["']/g;
    let match;
    while ((match = regex.exec(str)) !== null) {
        const isDynamic = !!match[1];
        const key = match[2];
        const val = match[3];

        if (isDynamic) {
            attrs.push(`"${key}": ${val}`);
        } else {
            attrs.push(`"${key}": "${val}"`);
        }
    }
    return attrs.join(', ');
}
