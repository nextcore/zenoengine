
import { reactive, effect } from '../src/reactivity.js';
import { compile } from '../src/compiler.js';

// Simple Test Runner
const tests = [];
function test(name, fn) {
    tests.push({ name, fn });
}

async function runTests() {
    let passed = 0;
    for (const t of tests) {
        try {
            await t.fn();
            console.log(`✅ ${t.name}`);
            passed++;
        } catch (e) {
            console.error(`❌ ${t.name}:`, e);
        }
    }
    console.log(`\nTests: ${passed}/${tests.length} passed`);
    if (passed !== tests.length) process.exit(1);
}

// 1. Reactivity Tests
test('Reactivity: basic', () => {
    const data = reactive({ count: 0 });
    let dummy;
    effect(() => {
        dummy = data.count;
    });

    if (dummy !== 0) throw new Error("Effect should run immediately");

    data.count = 1;
    if (dummy !== 1) throw new Error("Effect should update on change");
});

test('Reactivity: deep', () => {
    const data = reactive({ nested: { count: 0 } });
    let dummy;
    effect(() => {
        dummy = data.nested.count;
    });

    if (dummy !== 0) throw new Error("Initial value wrong");

    data.nested.count++;
    if (dummy !== 1) throw new Error("Deep update failed");
});

// 2. Compiler Tests
test('Compiler: interpolation', () => {
    const render = compile("Hello {{ name }}");
    const res = render.call({ name: "Zeno" });
    if (res !== "Hello Zeno") throw new Error(`Expected 'Hello Zeno', got '${res}'`);
});

test('Compiler: if/else', () => {
    const tpl = "@if(show) Yes @else No @endif";
    const render = compile(tpl);

    if (render.call({ show: true }).trim() !== "Yes") throw new Error("If true failed");
    if (render.call({ show: false }).trim() !== "No") throw new Error("Else failed");
});

test('Compiler: foreach', () => {
    const tpl = "@foreach(items as item) {{ item }} @endforeach";
    const render = compile(tpl);
    const res = render.call({ items: [1, 2] });

    // Output might have spaces/newlines
    if (!res.includes("1") || !res.includes("2")) throw new Error(`Foreach failed: ${res}`);
});

test('Compiler: click binding transformation', () => {
    const tpl = '<button @click="increment">Inc</button>';
    const render = compile(tpl);
    const res = render.call({});

    if (!res.includes('data-z-click="increment"')) throw new Error(`Click transformation failed: ${res}`);
});

runTests();
