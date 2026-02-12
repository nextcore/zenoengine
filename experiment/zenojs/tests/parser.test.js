
import { parse } from '../src/parser.js';

const tests = [
    {
        name: "Simple Text",
        input: "Hello World",
        check: (ast) => ast.children.length === 1 && ast.children[0].type === 'Text' && ast.children[0].value === 'Hello World'
    },
    {
        name: "Echo",
        input: "{{ name }}",
        check: (ast) => ast.children.length === 1 && ast.children[0].type === 'Echo' && ast.children[0].value === 'name'
    },
    {
        name: "Directive Block",
        input: "@if(show) Yes @endif",
        check: (ast) => {
            const node = ast.children[0];
            return node.type === 'Directive' && node.name === 'if' && node.children.length > 0;
        }
    },
    {
        name: "Component",
        input: '<x-alert type="error">Body</x-alert>',
        check: (ast) => {
             const node = ast.children[0];
             return node.type === 'Component' && node.tagName === 'x-alert' && node.children.length > 0;
        }
    }
];

let passed = 0;
for (const t of tests) {
    try {
        const ast = parse(t.input);
        if (t.check(ast)) {
            console.log(`✅ ${t.name}`);
            passed++;
        } else {
            console.error(`❌ ${t.name} (Check Failed)`, JSON.stringify(ast, null, 2));
        }
    } catch (e) {
        console.error(`❌ ${t.name} (Error):`, e);
    }
}

if (passed === tests.length) {
    console.log(`\nAll ${passed} parser tests passed.`);
} else {
    process.exit(1);
}
