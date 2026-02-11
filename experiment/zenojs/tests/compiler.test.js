
import { compile } from '../src/compiler.js';

function run(tpl, data) {
    const code = compile(tpl);
    // Mock _helpers
    const ctx = Object.assign({}, data, {
        $helpers: {
            classNames: (arg) => JSON.stringify(arg), // Mock implementation
            styleNames: (arg) => JSON.stringify(arg)
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
        name: "@unless",
        tpl: "@unless(show) Hidden @endunless",
        data: { show: false },
        expect: "Hidden"
    },
    {
        name: "@unless (negative)",
        tpl: "@unless(show) Hidden @endunless",
        data: { show: true },
        expect: ""
    },
    {
        name: "@isset",
        tpl: "@isset(name) {{ name }} @endisset",
        data: { name: "Zeno" },
        expect: "Zeno"
    },
    {
        name: "@empty",
        tpl: "@empty(list) Empty @endempty",
        data: { list: [] },
        expect: "Empty"
    },
    {
        name: "@switch/@case",
        tpl: "@switch(role) @case('admin') Admin @break @case('user') User @break @endswitch",
        data: { role: 'admin' },
        expect: "Admin"
    },
    {
        name: "$loop variable",
        tpl: "@foreach(items as item) {{ loop.index }}:{{ item }} @endforeach",
        data: { items: ['a', 'b'] },
        expect: "0:a 1:b"
    },
    {
        name: "@json",
        tpl: "@json(data)",
        data: { data: { foo: 'bar' } },
        expect: '{"foo":"bar"}'
    },
    {
        name: "@checked",
        tpl: "<input @checked(isChecked)>",
        data: { isChecked: true },
        expect: '<input checked="checked">'
    }
];

let passed = 0;
for (const t of tests) {
    try {
        const res = run(t.tpl, t.data).trim();
        const normalize = s => s.replace(/\s+/g, ' ').trim();

        if (normalize(res) === normalize(t.expect)) {
            console.log(`✅ ${t.name}`);
            passed++;
        } else {
            console.error(`❌ ${t.name}\nExpected: '${t.expect}'\nGot:      '${res}'`);
        }
    } catch (e) {
        console.error(`❌ ${t.name} (Error):`, e);
    }
}

if (passed === tests.length) {
    console.log(`\nAll ${passed} compiler tests passed.`);
} else {
    process.exit(1);
}
