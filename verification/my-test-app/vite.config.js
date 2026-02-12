
import { defineConfig } from 'vite';
import zeno from './lib/zeno/vite-plugin.js';

export default defineConfig({
    plugins: [
        zeno()
    ]
});
