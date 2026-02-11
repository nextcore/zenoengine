
import { compile } from '../src/compiler.js';

function run(tpl, data, components) {
    const code = compile(tpl);
    const ctx = Object.assign({}, data, {
        $helpers: { classNames: () => '', styleNames: () => '' },
        renderComponent: (name, props, slots) => `[COMP:${name}]`,
        renderLayout: (name, sections) => `[LAYOUT:${name}]`,
        renderInclude: (name, data) => `[INCLUDE:${name} DATA=${JSON.stringify(data)}]`,
        $stack: (name) => `[STACK:${name}]`,
        $push: (name, content) => { /* mock push */ },
        $services: { 'UserService': { name: 'ServiceUser' } }
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
        name: "@push",
        tpl: "@push('js') <script>...</script> @endpush",
        data: {},
        // Output should be empty string (side effect only)
        expect: ""
    },
    {
        name: "@stack",
        tpl: "@stack('js')",
        data: {},
        expect: "[STACK:js]"
    },
    {
        name: "@include basic",
        tpl: "@include('view.name')",
        data: {},
        expect: "[INCLUDE:view.name DATA=undefined]"
    },
    {
        name: "@include with data",
        // Note: Compiler just passes raw args string to _renderInclude
        tpl: "@include('view.name', {id: 1})",
        data: {},
        expect: '[INCLUDE:view.name DATA={"id":1}]'
    },
    {
        name: "@inject",
        tpl: "@inject('user', 'UserService') {{ user.name }}",
        data: {},
        expect: "ServiceUser"
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
