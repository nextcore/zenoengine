
import { compile } from '../src/compiler.js';

function run(tpl, data, components) {
    const code = compile(tpl);
    const ctx = Object.assign({}, data, {
        $helpers: { classNames: () => '', styleNames: () => '' },
        renderComponent: (name, props, slots) => {
             let out = `[COMPONENT:${name} props=${JSON.stringify(props)}]`;
             if (slots.default) out += `[SLOT:default=${slots.default()}]`;
             return out;
        },
        renderLayout: (name, sections) => {
             let out = `[LAYOUT:${name}]`;
             if (sections.content) out += `[SECTION:content=${sections.content()}]`;
             return out;
        }
    });

    try {
        const fn = new Function(code);
        return fn.call(ctx);
    } catch(e) {
        console.error("Code:", code);
        throw e;
    }
}

const tests = [
    {
        name: "@extends basic",
        tpl: "@extends('master') @section('content') Hello @endsection",
        data: {},
        expect: "[LAYOUT:master][SECTION:content=Hello]"
    },
    {
        name: "@extends with multiple sections",
        tpl: "@extends('app') @section('header') Head @endsection @section('content') Body @endsection",
        data: {},
        expect: "[LAYOUT:app][SECTION:content=Body]" // Mock only renders content section
    }
];

let passed = 0;
for (const t of tests) {
    try {
        const res = run(t.tpl, t.data).trim();
        const normalize = s => s.replace(/\s+/g, ' ').trim();

        if (res.includes(t.expect) || normalize(res) === normalize(t.expect)) {
             console.log(`✅ ${t.name}`);
            passed++;
        } else {
             console.error(`❌ ${t.name}\nExpected: '${t.expect}'\nGot:      '${res}'`);
        }
    } catch (e) {
        console.error(`❌ ${t.name} (Error):`, e);
    }
}
