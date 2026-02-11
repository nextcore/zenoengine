
import { reactive, effect } from './reactivity.js';
import { compile } from './compiler.js';

export class Zeno {
    static _components = {};

    // Register component globally
    static component(name, definition) {
        this._components[name] = definition;
    }

    static create(options) {
        return new Zeno(options);
    }

    constructor(options) {
        // Reactive State
        this.data = reactive(options.data ? options.data() : {});
        this.methods = options.methods || {};
        this.template = options.template || '';
        this.props = options.props || []; // Props array: ['title', 'type']

        // Pass parent slots or default
        this.slots = options.slots || {};

        // Store element ref
        this.el = null;

        // Runtime Helpers for Blade Directives
        this.data.$helpers = {
            classNames: (arg) => {
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
                let styles = [];
                if (typeof arg === 'object') {
                    for (const k in arg) {
                         if (arg[k]) styles.push(`${k}: ${arg[k]}`);
                    }
                }
                return styles.join('; ');
            }
        };

        // Helper: $slots for checking if slot exists
        this.data.$slots = {};
        for (const k in this.slots) {
            this.data.$slots[k] = true; // Just existence check
        }
        // Slot renderer helper (injects slot content)
        // Usage in component: {{ $slot('header') }} or {{ $slot() }} for default
        this.data.$slot = (name = 'default') => {
            if (this.slots[name]) {
                // Execute the slot render function
                // The slot function is bound to PARENT scope (closure),
                // but we might want to pass props? usually slots just render parent content.
                // However, in string-based rendering, if we call it, it returns HTML string.
                return this.slots[name]();
            }
            return '';
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

        // Bind renderComponent to instance context so it can access Zeno.components
        this.renderComponent = this.renderComponent.bind(this);
    }

    // Internal method called by compiled code: _out += this.renderComponent('alert', {type: 'error'}, {default: ...})
    renderComponent(name, props, slots) {
        const def = Zeno._components[name];
        if (!def) {
            console.warn(`Component '${name}' not found.`);
            return `<div style="border:1px solid red">Component ${name} not found</div>`;
        }

        // Create component instance
        // We need to merge props into data
        // But data() is a factory function.
        // We modify the factory? Or the result?

        const dataFactory = def.data || (() => ({}));

        // Wrap data factory to inject props
        const componentData = dataFactory();

        // Inject props
        // In Vue, props are separate from data, but accessible via `this`.
        // Here, for simplicity, we merge props into data (so {{ type }} works directly).
        // Props take precedence or data? Usually props override data init.
        Object.assign(componentData, props);

        const instance = new Zeno({
            data: () => componentData,
            methods: def.methods,
            template: def.template,
            render: def.render, // Pre-compiled render function
            props: def.props,
            slots: slots
        });

        // Render to string (synchronously)
        // Note: This creates a new reactive instance every render?
        // Yes, this is inefficient "Re-create World" strategy for this experiment.
        // In a real VDOM, we would diff/patch/update existing instance.
        // Here, we just produce HTML string.

        try {
            return instance.renderFn.call(instance.data);
        } catch (e) {
            console.error(`Error rendering component ${name}:`, e);
            return `Error: ${e.message}`;
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
            // We need to expose renderComponent on data scope too?
            // "with(this)" in renderFn uses `this` which is `this.data` usually?
            // Wait, in `zeno.js` render(), we call `renderFn.call(this.data)`.
            // So `this` inside render function is `this.data`.
            // `this.data` does NOT have `renderComponent`.
            // We must attach it!
            this.data.renderComponent = this.renderComponent;

            html = this.renderFn.call(this.data);
        } catch (e) {
            console.error("Render Error:", e);
            html = `<div style="color:red">Render Error: ${e.message}</div>`;
        }

        if (this.el) {
            this.el.innerHTML = html;
            this.bindEvents();
        }
        return html; // For recursive calls
    }

    bindEvents() {
        if (!this.el) return;

        const elements = this.el.querySelectorAll('[data-z-click]');
        elements.forEach(el => {
            const handlerName = el.getAttribute('data-z-click');
            // Check if handler is in methods
            if (this.methods[handlerName]) {
                el.addEventListener('click', (e) => {
                    this.methods[handlerName](e);
                });
            } else {
                 if (handlerName.includes('(')) {
                     console.warn(`Complex event handlers like '${handlerName}' are not fully supported in ZenoJS yet.`);
                 } else {
                     console.warn(`Method '${handlerName}' not found.`);
                 }
            }
            el.removeAttribute('data-z-click');
        });
    }
}
