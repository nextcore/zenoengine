
// ZenoBlade AST-Based Compiler
import { parse } from './parser.js';

export function compile(template) {
    const ast = parse(template);

    // Scan AST for extends directive
    let layoutName = null;
    let sections = {};
    let pushes = {};

    const rootChildren = ast.children;
    const filteredChildren = [];

    for (const node of rootChildren) {
        if (node.type === 'Directive' && node.name === 'extends') {
            layoutName = node.args.replace(/['"]/g, '');
        }
        else if (node.type === 'Directive' && node.name === 'section') {
            const sectionName = node.args.replace(/['"]/g, '');
            sections[sectionName] = node.children;
        }
        else {
            filteredChildren.push(node);
        }
    }

    let code = "let _out = '';\n";
    code += "const _helpers = this.$helpers || {};\n";
    code += "const _renderComponent = this.renderComponent || (() => '');\n";
    code += "const _renderLayout = this.renderLayout || (() => '');\n";
    code += "const _renderInclude = this.renderInclude || (() => '');\n";

    code += "with(this) {\n";

    if (layoutName) {
        const sideEffects = filteredChildren.filter(n => n.type === 'Directive' && (n.name === 'push' || n.name === 'inject'));
        for (const effect of sideEffects) {
             code += codegen(effect);
        }

        let sectionsCode = '{';
        for (const [name, children] of Object.entries(sections)) {
            const body = children.map(codegen).join('');
            sectionsCode += `\n"${name}": (() => {\nlet _out='';\n${body}\nreturn _out;\n}),`;
        }
        sectionsCode += '}';

        code += `_out += _renderLayout('${layoutName}', ${sectionsCode});\n`;

    } else {
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
        else if (['endif', 'endunless', 'endisset', 'endempty', 'endswitch', 'endforeach', 'endsection', 'endpush'].includes(name)) {
             // Closed
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
        else if (name === 'model') {
             // @model('var')
             // Outputs: value="..." data-z-model="var"
             const modelVar = args.replace(/^["']|["']$/g, '');
             code += `_out += 'value="' + (${modelVar}) + '" data-z-model="${modelVar}"';\n`;
        }
        else if (name === 'yield') {
             const sectionName = args.replace(/['"]/g, '');
             code += `if (this.$sections && this.$sections['${sectionName}']) { _out += this.$sections['${sectionName}'](); }\n`;
        }
        else if (name === 'section') {
             code += childrenCode;
        }
        else if (name === 'push') {
             const stackName = args.replace(/['"]/g, '');
             code += `
             {
                 let _pushOut = '';
                 this.$push('${stackName}', (() => { let _out=''; ${childrenCode} return _out; })());
             }
             \n`;
        }
        else if (name === 'stack') {
             const stackName = args.replace(/['"]/g, '');
             code += `_out += this.$stack('${stackName}');\n`;
        }
        else if (name === 'include') {
             code += `_out += _renderInclude(${args});\n`;
        }
        else if (name === 'inject') {
             const [varName, serviceName] = args.split(',').map(s => s.trim().replace(/['"]/g, ''));
             code += `var ${varName} = this.$services['${serviceName}'];\n`;
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
