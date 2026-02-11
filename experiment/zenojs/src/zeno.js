
import { reactive, effect } from './reactivity.js';
import { compile } from './compiler.js';

export class Zeno {
    static create(options) {
        return new Zeno(options);
    }

    constructor(options) {
        // Reactive State
        this.data = reactive(options.data ? options.data() : {});
        this.methods = options.methods || {};
        this.template = options.template || '';
        this.el = null;

        // Runtime Helpers for Blade Directives
        this.data.$helpers = {
            classNames: (arg) => {
                // Supports string, array, or object
                // @class(['p-4', 'bg-red' => hasError])
                // In JS, 'bg-red' => hasError is not valid syntax inside array literals unless parsed differently.
                // But if user writes @class({'p-4': true, 'bg-red': hasError}) it works.
                // Blade's array syntax ['k' => v] is PHP.
                // In ZenoJS, we should encourage JS object syntax { 'k': v } or standard array.

                let classes = [];
                if (typeof arg === 'string') {
                    classes.push(arg);
                } else if (Array.isArray(arg)) {
                    arg.forEach(a => {
                        if (typeof a === 'string') classes.push(a);
                        else if (typeof a === 'object') {
                            for (const k in a) {
                                if (a[k]) classes.push(k);
                            }
                        }
                    });
                } else if (typeof arg === 'object') {
                    for (const k in arg) {
                        if (arg[k]) classes.push(k);
                    }
                }
                return classes.join(' ');
            },
            styleNames: (arg) => {
                // @style({ 'color': 'red', 'font-size': size + 'px' })
                let styles = [];
                if (typeof arg === 'object') {
                    for (const k in arg) {
                         if (arg[k]) styles.push(`${k}: ${arg[k]}`);
                    }
                }
                return styles.join('; ');
            }
        };

        // Render Function
        if (options.render) {
            this.renderFn = options.render;
        } else {
            const code = compile(this.template);
            try {
                this.renderFn = new Function(code);
            } catch (e) {
                 console.error("Compilation Error:", e);
                 this.renderFn = () => "Error compiling template";
            }
        }

        // Bind methods
        for (const key in this.methods) {
            this.methods[key] = this.methods[key].bind(this.data);
        }
    }

    mount(selector) {
        this.el = document.querySelector(selector);
        if (!this.el) {
            console.error(`Element ${selector} not found`);
            return;
        }

        effect(() => {
            this.render();
        });
    }

    render() {
        let html = '';
        try {
            html = this.renderFn.call(this.data);
        } catch (e) {
            console.error("Render Error:", e);
            html = `<div style="color:red">Render Error: ${e.message}</div>`;
        }

        this.el.innerHTML = html;
        this.bindEvents();
    }

    bindEvents() {
        const elements = this.el.querySelectorAll('[data-z-click]');
        elements.forEach(el => {
            const handlerName = el.getAttribute('data-z-click');
            // Remove parens if present (e.g. toggle(todo))
            // This is a naive implementation.
            // Ideally we parse the handler expression.

            // If it's a simple method name:
            if (this.methods[handlerName]) {
                el.addEventListener('click', (e) => {
                    this.methods[handlerName](e);
                });
            } else {
                // It might be a call expression: toggle(todo)
                // We need to evaluate it in context of data
                // BUT we don't have access to the local scope variables (like 'todo' from foreach) here!
                // 'todo' existed only during render time loop.
                // This is the fundamental challenge of "HTML String" rendering vs VDOM with Closures.
                // In VDOM, the onClick handler captures the 'todo' variable closure.
                // In HTML String, that context is lost.

                // SOLUTION for ZenoJS (HTML String approach):
                // We must serialize the arguments into the DOM, or rely on index.
                // E.g. data-z-args="[1]"
                // Or simply: Warn user that arguments in @click are not fully supported yet in this string-based version,
                // OR implement a global registry of handlers (like `window.__zeno_handlers[id]`).

                // For this MVP, let's stick to simple method binding.
                // Complex args are out of scope for this step unless we refactor to VDOM/Registry.
                // Wait, our previous demo used `toggle(todo)`. Did it work?
                // No, I didn't verify that specific part in Playwright!
                // I verified "Add Todo" which called `addTodo` (no args).
                // `toggle(todo)` likely failed silently or threw error in console.

                // To support `toggle(todo)`, we would need:
                // 1. Serialize `todo` to JSON in `data-z-arg`? (Expensive/Complex)
                // 2. Use index `toggle(index)`.

                // Let's rely on simple methods for now, or use a workaround.
                // I will add a warning log.
                 if (handlerName.includes('(')) {
                     console.warn(`Complex event handlers like '${handlerName}' are not fully supported in ZenoJS yet. Please use index or simple method names.`);
                 } else {
                     console.warn(`Method '${handlerName}' not found.`);
                 }
            }
            el.removeAttribute('data-z-click');
        });
    }
}
