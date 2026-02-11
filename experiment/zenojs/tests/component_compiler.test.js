
import { compile } from '../src/compiler.js';

function run(tpl, data, components) {
    const code = compile(tpl);
    const ctx = Object.assign({}, data, {
        $helpers: { classNames: () => '', styleNames: () => '' },
        renderComponent: (name, props, slots) => {
             // Mock renderComponent
             let out = `[COMPONENT:${name} props=${JSON.stringify(props)}]`;
             if (slots.default) out += `[SLOT:default=${slots.default()}]`;
             if (slots.header) out += `[SLOT:header=${slots.header()}]`;
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
        name: "<x-component> self-closing",
        tpl: '<x-alert type="error" />',
        data: {},
        expect: '[COMPONENT:alert props={"type": "error"}]'
    },
    {
        name: "<x-component> with body",
        tpl: '<x-alert type="info">Message</x-alert>',
        data: {},
        expect: '[COMPONENT:alert props={"type": "info"}][SLOT:default=Message]'
    },
    {
        name: "<x-component> with dynamic prop",
        tpl: '<x-alert :count="1+1" />',
        data: {},
        expect: '[COMPONENT:alert props={"count": 2}]'
    },
     {
        name: "<x-component> with named slot",
        // Note: The mock parser in test doesn't do deep parsing of slots unless we implement it.
        // The compiler.js `parseSlots` logic is used inside the compiled code?
        // Wait, `compile` generates the `slots` object structure.
        // The slot content is compiled recursively.
        tpl: '<x-card><x-slot name="header">Head</x-slot>Body</x-card>',
        data: {},
        expect: '[COMPONENT:card props={}][SLOT:default=Body][SLOT:header=Head]'
    }
];

let passed = 0;
for (const t of tests) {
    try {
        const res = run(t.tpl, t.data).trim();
        const normalize = s => s.replace(/\s+/g, ' ').trim();

        // Loose comparison for slots order?
        // Our mock renderer appends default then header.
        // Compiler output order depends on `parseSlots`.

        if (res.includes(t.expect) || normalize(res) === normalize(t.expect)) {
             console.log(`✅ ${t.name}`);
            passed++;
        } else {
             // Try to match parts
             console.error(`❌ ${t.name}\nExpected: '${t.expect}'\nGot:      '${res}'`);
        }
    } catch (e) {
        console.error(`❌ ${t.name} (Error):`, e);
    }
}
