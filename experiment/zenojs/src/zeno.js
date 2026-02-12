
import { reactive, effect } from './reactivity.js';
import { compile } from './compiler.js';

let globalStacks = {};
let currentInstance = null;

export function onMounted(fn) {
    if (currentInstance) {
        currentInstance.hooks.mounted.push(fn);
    }
}

export function onUnmounted(fn) {
    if (currentInstance) {
        currentInstance.hooks.unmounted.push(fn);
    }
}

export class Zeno {
    static _components = {};
    static _layouts = {};
    static _services = {};
    static _views = {};
    static _plugins = [];

    static use(plugin) {
        this._plugins.push(plugin);
        if (plugin.install) plugin.install(Zeno);
    }

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
        this.hooks = { mounted: [], unmounted: [] };

        const prevInstance = currentInstance;
        currentInstance = this;

        try {
            const dataFactory = options.data || (() => ({}));
            const rawData = typeof dataFactory === 'function' ? dataFactory() : dataFactory;

            if (options.mounted) this.hooks.mounted.push(options.mounted.bind(null));
            if (options.unmounted) this.hooks.unmounted.push(options.unmounted);

            this.data = reactive(rawData);
        } finally {
            currentInstance = prevInstance;
        }

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

        this.data.$push = (name, content) => {
            if (!globalStacks[name]) globalStacks[name] = [];
            globalStacks[name].push(content);
        };
        this.data.$stack = (name) => {
            return globalStacks[name] ? globalStacks[name].join('') : '';
        };

        this.data.$services = Zeno._services;

        // Inject Plugins
        if (Zeno.prototype.$router) this.data.$router = Zeno.prototype.$router;
        if (Zeno.prototype.$store) {
            this.data.$store = Zeno.prototype.$store;

            // Auth Helper Integration
            // If store has 'user' state, alias it to `this.user` or `this.auth`
            // for Blade directives @auth / @guest compatibility.
            // Assumption: User object is in `$store.state.user`.
            // Check if user is logged in: user !== null && user.name !== 'Guest' (convention)
            // Or better: $store.getters.isAuthenticated

            // We define a getter 'auth' on the data proxy?
            // Actually, `with(this)` will look for 'auth'.
            // Let's define a computed property for it?

            // Simpler: Just rely on developers defining `user` in data OR
            // mapping it here.

            // Convention: ZenoJS maps `this.user` to `$store.state.user` automatically if not defined.
            if (this.data.user === undefined && this.data.$store.state.user) {
                // Not easily reactive if we just assign value.
                // We need `get user() { return this.$store.state.user }`.
                // But `this.data` is a Proxy target (plain object usually).
                // We can't define property easily on the instance.

                // Workaround: In Compiler, @auth checks `this.user` OR `this.$store.state.user`?
                // No, standard Blade checks `auth()` helper or `$user`.

                // Let's inject an `auth` helper object.
                this.data.auth = {
                    user: () => this.data.$store.state.user,
                    check: () => {
                        const u = this.data.$store.state.user;
                        return u && u.name !== 'Guest' && u.name !== null;
                    },
                    guest: () => {
                        const u = this.data.$store.state.user;
                        return !u || u.name === 'Guest' || u.name === null;
                    }
                };

                // Also alias `user` for convenience?
                // this.data.user = this.data.$store.state.user; // Initial value only :(
            }
        }

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

        this.hooks.mounted = this.hooks.mounted.map(fn => fn.bind(this.data));
        this.hooks.unmounted = this.hooks.unmounted.map(fn => fn.bind(this.data));

        this.renderComponent = this.renderComponent.bind(this);
        this.renderLayout = this.renderLayout.bind(this);
        this.renderInclude = this.renderInclude.bind(this);
        this.renderDynamic = this.renderDynamic.bind(this);
    }

    renderComponent(name, props, slots) {
        const def = Zeno._components[name];
        if (!def) return `[Component ${name} not found]`;
        return this.renderDynamic(def, props, slots);
    }

    renderDynamic(def, props = {}, slots = {}) {
        const dataFactory = def.data || (() => ({}));
        const componentData = dataFactory();

        const declaredProps = def.props || [];
        const passedProps = {};
        const passedAttrs = {};
        for (const key in props) {
            if (declaredProps.includes(key)) passedProps[key] = props[key];
            else passedAttrs[key] = props[key];
        }
        Object.assign(componentData, passedProps);

        const instance = new Zeno({
            data: () => componentData,
            methods: def.methods,
            template: def.template,
            render: def.render,
            props: def.props,
            attributes: passedAttrs,
            slots: slots,
            mounted: def.mounted,
            unmounted: def.unmounted
        });

        try {
            return instance.renderFn.call(instance.data);
        } catch (e) {
            return `Error: ${e.message}`;
        }
    }

    renderLayout(name, sections) {
        const def = Zeno._layouts[name];
        if (!def) return `[Layout ${name} not found]`;
        const dataFactory = def.data || (() => ({}));
        const layoutData = dataFactory();
        const instance = new Zeno({
            data: () => layoutData,
            methods: def.methods,
            template: def.template,
            render: def.render,
            sections: sections
        });
        try { return instance.renderFn.call(instance.data); } catch (e) { return e.message; }
    }

    renderInclude(name, data) {
        const tpl = Zeno._views[name];
        if (!tpl) return `[View ${name} not found]`;
        const mergedData = { ...this.data, ...data };
        const instance = new Zeno({ data: () => mergedData, template: tpl });
        try { return instance.renderFn.call(instance.data); } catch (e) { return e.message; }
    }

    mount(selector) {
        this.el = document.querySelector(selector);
        if (!this.el) {
            console.error(`Element ${selector} not found`);
            return;
        }

        let isMounted = false;

        effect(() => {
            this.render();
            if (!isMounted) {
                isMounted = true;
                this.hooks.mounted.forEach(fn => fn());
            }
        });
    }

    render() {
        globalStacks = {};

        this.data.renderComponent = this.renderComponent;
        this.data.renderLayout = this.renderLayout;
        this.data.renderInclude = this.renderInclude;
        this.data.renderDynamic = this.renderDynamic;

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

        const clickElements = this.el.querySelectorAll('[data-z-click]');
        clickElements.forEach(el => {
            const handlerName = el.getAttribute('data-z-click');
            if (this.methods[handlerName]) {
                el.addEventListener('click', (e) => this.methods[handlerName](e));
            }
            el.removeAttribute('data-z-click');
        });

        const modelElements = this.el.querySelectorAll('[data-z-model]');
        modelElements.forEach(el => {
            const varName = el.getAttribute('data-z-model');
            el.addEventListener('input', (e) => {
                const val = e.target.type === 'checkbox' ? e.target.checked : e.target.value;
                if (varName.includes('.')) {
                    const parts = varName.split('.');
                    let target = this.data;
                    for (let i = 0; i < parts.length - 1; i++) {
                        target = target[parts[i]];
                    }
                    target[parts[parts.length - 1]] = val;
                } else {
                    this.data[varName] = val;
                }
            });
            el.removeAttribute('data-z-model');
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
