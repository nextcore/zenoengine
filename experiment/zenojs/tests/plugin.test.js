
import zenoPlugin from '../vite-plugin-zeno.js';
import fs from 'fs';
import path from 'path';

// Simulate Vite Transform Hook
const plugin = zenoPlugin();
const transform = plugin.transform;

const appZenoPath = path.resolve('experiment/zenojs/src/App.zeno');
const code = fs.readFileSync(appZenoPath, 'utf-8');

console.log("--- Input Code ---");
console.log(code.substring(0, 100) + "...");

const result = transform(code, appZenoPath);

if (!result) {
    console.error("Transform returned null!");
    process.exit(1);
}

console.log("\n--- Transformed Code ---");
console.log(result.code);

// Verification: Check if render function exists and has code
if (!result.code.includes('script.render = function() {')) {
    console.error("FAIL: Render function not injected");
    process.exit(1);
}

if (!result.code.includes('let _out =')) {
    console.error("FAIL: Compiler output not found");
    process.exit(1);
}

if (!result.code.includes('export default script')) {
    console.error("FAIL: Export default not found");
    process.exit(1);
}

console.log("\n✅ Plugin Test Passed!");
