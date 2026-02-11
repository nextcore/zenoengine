
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

        // Compiled Render Function
        this.renderFn = compile(this.template);

        // Bind methods to data context
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

        // Reactive Effect: Render when data changes
        effect(() => {
            this.render();
        });
    }

    render() {
        // Generate HTML string
        // Execute render function with 'this' context as data
        let html = '';
        try {
            html = this.renderFn.call(this.data);
        } catch (e) {
            console.error("Render Error:", e);
            html = `<div style="color:red">Render Error: ${e.message}</div>`;
        }

        // Diffing? No, just full replacement for now (simple implementation)
        // But we need to preserve input focus if possible? Not in MVP.
        this.el.innerHTML = html;

        // Re-bind Events
        this.bindEvents();
    }

    bindEvents() {
        // Scan for elements with 'data-z-click' attribute
        // The compiler transforms @click="method" into data-z-click="method"
        const elements = this.el.querySelectorAll('[data-z-click]');

        elements.forEach(el => {
            const handlerName = el.getAttribute('data-z-click');
            if (this.methods[handlerName]) {
                el.addEventListener('click', (e) => {
                    // Prevent default if modified? e.g. @click.prevent
                    // Standard click for now
                    this.methods[handlerName](e);
                });
            } else {
                 console.warn(`Method '${handlerName}' not found.`);
            }
            // Clean up attribute
            el.removeAttribute('data-z-click');
        });
    }
}
