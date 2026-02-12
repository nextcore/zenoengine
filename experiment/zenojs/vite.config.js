
import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
    build: {
        lib: {
            entry: {
                zeno: resolve(__dirname, 'index.js'),
                plugin: resolve(__dirname, 'vite-plugin-zeno.js')
            },
            name: 'ZenoJS',
            formats: ['es', 'cjs']
        },
        rollupOptions: {
            // Externalize deps if any (none currently)
            external: ['fs', 'path'], // Plugin uses Node modules
        }
    }
});
