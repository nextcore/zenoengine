
import { reactive, effect } from './reactivity.js';
import { compile } from './compiler.js';

export class Zeno {
    static _components = {};
    static _layouts = {}; // Layouts registry

    static component(name, definition) {
        this._components[name] = definition;
    }

    // Register layout (which is just a component/template used as layout)
    static layout(name, definition) {
        this._layouts[name] = definition;
    }

    static create(options) {
        return new Zeno(options);
    }

    constructor(options) {
        this.data = reactive(options.data ? options.data() : {});
        this.methods = options.methods || {};
        this.template = options.template || '';
        this.props = options.props || [];
        this.slots = options.slots || {};
        this.sections = options.sections || {}; // For Layouts

        this.el = null;

        this.data.$helpers = {
            classNames: (arg) => {
                let classes = [];
                if (typeof arg === 'string') classes.push(arg);
                else if (Array.isArray(arg)) {
                    arg.forEach(a => {
                        if (typeof a === 'string') classes.push(a);
                        else if (typeof a === 'object') {
                            for (const k in a) if (a[k]) classes.push(k);
                        }
                    });
                } else if (typeof arg === 'object') {
                    for (const k in arg) if (arg[k]) classes.push(k);
                }
                return classes.join(' ');
            },
            styleNames: (arg) => {
                let styles = [];
                if (typeof arg === 'object') {
                    for (const k in arg) if (arg[k]) styles.push(`${k}: ${arg[k]}`);
                }
                return styles.join('; ');
            }
        };

        this.data.$slots = {};
        for (const k in this.slots) this.data.$slots[k] = true;
        this.data.$slot = (name = 'default') => this.slots[name] ? this.slots[name]() : '';

        // Sections Helper for Layouts
        this.data.$sections = this.sections;

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

        for (const key in this.methods) {
            this.methods[key] = this.methods[key].bind(this.data);
        }

        this.renderComponent = this.renderComponent.bind(this);
        this.renderLayout = this.renderLayout.bind(this);
    }

    renderComponent(name, props, slots) {
        const def = Zeno._components[name];
        if (!def) {
            console.warn(`Component '${name}' not found.`);
            return `[Component ${name} not found]`;
        }

        const dataFactory = def.data || (() => ({}));
        const componentData = dataFactory();
        Object.assign(componentData, props);

        const instance = new Zeno({
            data: () => componentData,
            methods: def.methods,
            template: def.template,
            render: def.render,
            props: def.props,
            slots: slots
        });

        try {
            return instance.renderFn.call(instance.data);
        } catch (e) {
            console.error(`Error rendering component ${name}:`, e);
            return `Error: ${e.message}`;
        }
    }

    renderLayout(name, sections) {
        // Layout is just a component but we inject sections instead of slots (or map sections to slots?)
        // In this implementation, layout template uses @yield('name') which compiles to `this.$sections['name']()`

        const def = Zeno._layouts[name];
        if (!def) {
            console.warn(`Layout '${name}' not found.`);
            return `[Layout ${name} not found]`;
        }

        // Layout usually doesn't have props from child, but shares data?
        // In Laravel, layout shares global data.
        // Here, we create a new instance for layout.

        const dataFactory = def.data || (() => ({}));
        const layoutData = dataFactory();

        const instance = new Zeno({
            data: () => layoutData,
            methods: def.methods,
            template: def.template,
            render: def.render,
            sections: sections // Pass sections!
        });

        try {
            return instance.renderFn.call(instance.data);
        } catch (e) {
            return `Error rendering layout ${name}: ${e.message}`;
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
        // Inject runtime helpers
        this.data.renderComponent = this.renderComponent;
        this.data.renderLayout = this.renderLayout;

        let html = '';
        try {
            html = this.renderFn.call(this.data);
        } catch (e) {
            console.error("Render Error:", e);
            html = `<div style="color:red">Render Error: ${e.message}</div>`;
        }

        if (this.el) {
            this.el.innerHTML = html;
            this.bindEvents();
        }
        return html;
    }

    bindEvents() {
        if (!this.el) return;
        const elements = this.el.querySelectorAll('[data-z-click]');
        elements.forEach(el => {
            const handlerName = el.getAttribute('data-z-click');
            if (this.methods[handlerName]) {
                el.addEventListener('click', (e) => this.methods[handlerName](e));
            }
            el.removeAttribute('data-z-click');
        });
    }
}
