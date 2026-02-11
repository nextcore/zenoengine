
import { defineConfig } from 'vite';
import zeno from './vite-plugin-zeno.js';

export default defineConfig({
    plugins: [
        zeno()
    ],
    // Ensure we can import compiler.js in the plugin which runs in Node
    // Node handles ESM natively now.
});
