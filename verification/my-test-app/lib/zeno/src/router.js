
import { reactive } from './reactivity.js';

export class ZenoRouter {
    constructor(routes) {
        this.routes = routes;
        this.current = reactive({
            path: window.location.pathname,
            component: null,
            params: {}
        });

        this.globalGuards = [];

        this._resolve(); // Initial resolve (without guards? or with?)
        // Standard SPA: initial load should also check guards.

        window.addEventListener('popstate', () => {
            this._navigate(window.location.pathname);
        });

        this._setupLinkInterception();
    }

    beforeEach(guard) {
        this.globalGuards.push(guard);
    }

    push(path) {
        // Run guards before pushing state
        this._navigate(path, true);
    }

    async _navigate(path, pushState = false) {
        // Find matching route
        const match = this.routes.find(r => r.path === path);
        const to = { path, meta: match ? match.meta || {} : {} };
        const from = { path: this.current.path };

        // 1. Run Global Guards
        for (const guard of this.globalGuards) {
            const result = await guard(to, from);
            if (result === false) return; // Cancel
            if (typeof result === 'string') {
                return this.push(result); // Redirect
            }
        }

        // 2. Run Per-Route Guard
        if (match && match.beforeEnter) {
            const result = await match.beforeEnter(to, from);
            if (result === false) return;
            if (typeof result === 'string') {
                return this.push(result);
            }
        }

        // 3. Commit Navigation
        if (pushState) {
            window.history.pushState({}, '', path);
        }

        this.current.path = path;

        if (match) {
            this.current.component = match.component;
        } else {
            this.current.component = {
                template: '<h1>404 Not Found</h1>',
                render: function() { return '<h1>404 Not Found</h1>'; }
            };
        }
    }

    _resolve() {
        // Just resolve current path (used in constructor)
        // Ideally reuse _navigate but without pushState and with guard checks.
        // We trigger navigation for initial state
        this._navigate(window.location.pathname);
    }

    _setupLinkInterception() {
        document.addEventListener('click', (e) => {
            const link = e.target.closest('a');
            if (!link) return;

            const href = link.getAttribute('href');
            if (href && href.startsWith('/') && link.origin === window.location.origin) {
                e.preventDefault();
                this.push(href);
            }
        });
    }

    install(Zeno) {
        Zeno.prototype.$router = this;

        Zeno.component('router-view', {
            template: `<!-- Router View -->`,
            render: function() {
                const routeComp = this.$router.current.component;
                if (!routeComp) return '';
                return this.renderDynamic(routeComp, {}, {});
            }
        });
    }
}
