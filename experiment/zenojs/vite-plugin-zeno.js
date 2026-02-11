
import { compile } from './src/compiler.js';

/**
 * Vite Plugin for ZenoJS
 * Transforms .zeno files into JavaScript modules.
 */
export default function zeno() {
    return {
        name: 'vite-plugin-zeno',

        transform(code, id) {
            if (!id.endsWith('.zeno')) return;

            // Simple parser for SFC
            // <script> ... </script>
            // <template> ... </template>

            const scriptMatch = code.match(/<script>([\s\S]*?)<\/script>/);
            const templateMatch = code.match(/<template>([\s\S]*?)<\/template>/);

            if (!scriptMatch || !templateMatch) {
                 return; // Invalid component?
            }

            const scriptContent = scriptMatch[1].trim();
            const templateContent = templateMatch[1].trim();

            // Compile Template
            // This returns the FUNCTION BODY string.
            // We need to wrap it in "function() { ... }" string.
            const renderFnBody = compile(templateContent);

            // Transform Script:
            // "export default { ... }" -> "const script = { ... }"
            // But user might write "const data = ...; export default ..."
            // We assume standard "export default" for now.
            // Replace first occurrence of 'export default' with 'const script ='
            let finalCode = scriptContent.replace(/export\s+default/, 'const script =');

            // Add Render Function
            // Note: renderFnBody is code inside the function block.
            finalCode += `\n\nscript.render = function() {\n${renderFnBody}\n};\n`;

            // Re-export
            finalCode += `\nexport default script;\n`;

            return {
                code: finalCode,
                map: null // TODO: Source Map
            };
        }
    };
}
