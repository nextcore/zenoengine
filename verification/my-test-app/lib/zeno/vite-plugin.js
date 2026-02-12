
import { compile } from './src/compiler.js';

/**
 * Vite Plugin for ZenoJS
 * Transforms .blade files into JavaScript modules.
 */
export default function zeno() {
    return {
        name: 'vite-plugin-zeno',

        transform(code, id) {
            // Updated to check for .blade extension
            if (!id.endsWith('.blade')) return;

            // Simple parser for SFC
            // <script> ... </script>
            // <template> ... </template>

            const scriptMatch = code.match(/<script>([\s\S]*?)<\/script>/);
            const templateMatch = code.match(/<template>([\s\S]*?)<\/template>/);

            if (!scriptMatch || !templateMatch) {
                 return;
            }

            const scriptContent = scriptMatch[1].trim();
            const templateContent = templateMatch[1].trim();

            const renderFnBody = compile(templateContent);

            let finalCode = scriptContent.replace(/export\s+default/, 'const script =');
            finalCode += `\n\nscript.render = function() {\n${renderFnBody}\n};\n`;
            finalCode += `\nexport default script;\n`;

            return {
                code: finalCode,
                map: null
            };
        }
    };
}
