
import { reactive, effect } from './reactivity.js';
import { compile } from './compiler.js';

// Global Stack Container (shared across render pass)
let globalStacks = {};

export class Zeno {
    static _components = {};
    static _layouts = {};
    static _services = {};
    static _views = {}; // For @include

    static component(name, definition) {
        this._components[name] = definition;
    }

    static layout(name, definition) {
        this._layouts[name] = definition;
    }

    static service(name, instance) {
        this._services[name] = instance;
    }

    static view(name, template) {
        this._views[name] = template;
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
        this.sections = options.sections || {};

        const rawAttrs = options.attributes || {};
        this.data.$attributes = {
            ...rawAttrs,
            merge: (defaults) => {
                const merged = { ...defaults, ...rawAttrs };
                if (defaults.class && rawAttrs.class) {
                    merged.class = `${defaults.class} ${rawAttrs.class}`;
                }
                return new AttributeBag(merged);
            },
            toString: () => new AttributeBag(rawAttrs).toString()
        };

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
        this.data.$sections = this.sections;

        // Stack Helpers
        this.data.$push = (name, content) => {
            if (!globalStacks[name]) globalStacks[name] = [];
            globalStacks[name].push(content);
        };
        this.data.$stack = (name) => {
            return globalStacks[name] ? globalStacks[name].join('') : '';
        };

        // Services Helper
        this.data.$services = Zeno._services;

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
        this.renderInclude = this.renderInclude.bind(this);
    }

    renderComponent(name, props, slots) {
        const def = Zeno._components[name];
        if (!def) {
            console.warn(`Component '${name}' not found.`);
            return `[Component ${name} not found]`;
        }

        const dataFactory = def.data || (() => ({}));
        const componentData = dataFactory();

        const declaredProps = def.props || [];
        const passedProps = {};
        const passedAttrs = {};

        for (const key in props) {
            if (declaredProps.includes(key)) {
                passedProps[key] = props[key];
            } else {
                passedAttrs[key] = props[key];
            }
        }

        Object.assign(componentData, passedProps);

        const instance = new Zeno({
            data: () => componentData,
            methods: def.methods,
            template: def.template,
            render: def.render,
            props: def.props,
            attributes: passedAttrs,
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
        const def = Zeno._layouts[name];
        if (!def) {
            console.warn(`Layout '${name}' not found.`);
            return `[Layout ${name} not found]`;
        }

        const dataFactory = def.data || (() => ({}));
        const layoutData = dataFactory();

        const instance = new Zeno({
            data: () => layoutData,
            methods: def.methods,
            template: def.template,
            render: def.render,
            sections: sections
        });

        try {
            return instance.renderFn.call(instance.data);
        } catch (e) {
            return `Error rendering layout ${name}: ${e.message}`;
        }
    }

    renderInclude(name, data) {
        const tpl = Zeno._views[name];
        if (!tpl) {
             return `[View ${name} not found]`;
        }

        // Merge data with current data?
        // Usually include inherits scope + passed data.
        // We can create a new instance with merged data.

        const mergedData = { ...this.data, ...data }; // Flatten proxy?
        // Actually this.data is proxy.
        // We want new instance to have access to helpers too.

        const instance = new Zeno({
            data: () => mergedData,
            template: tpl
        });

        try {
            return instance.renderFn.call(instance.data);
        } catch (e) {
            return `Error including ${name}: ${e.message}`;
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
        // Reset global stacks on root render
        globalStacks = {};

        this.data.renderComponent = this.renderComponent;
        this.data.renderLayout = this.renderLayout;
        this.data.renderInclude = this.renderInclude;

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

class AttributeBag {
    constructor(attrs) {
        this.attrs = attrs;
    }

    toString() {
        return Object.entries(this.attrs)
            .map(([k, v]) => {
                if (v === true) return k;
                if (v === false || v === null || v === undefined) return '';
                return `${k}="${v}"`;
            })
            .filter(s => s)
            .join(' ');
    }
}
