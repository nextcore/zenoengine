
import { reactive } from './reactivity.js';

export class ZenoRouter {
    constructor(routes) {
        this.routes = routes;
        this.current = reactive({
            path: window.location.pathname,
            component: null,
            params: {}
        });

        this._resolve();

        window.addEventListener('popstate', () => {
            this.current.path = window.location.pathname;
            this._resolve();
        });

        // Intercept links
        this._setupLinkInterception();
    }

    _setupLinkInterception() {
        document.addEventListener('click', (e) => {
            // Find closest anchor tag
            const link = e.target.closest('a');
            if (!link) return;

            // Check if it's internal link and has wire:navigate (or just internal?)
            // Let's use `z-link` attribute or assume all relative links are SPA links.
            // Convention: `href` starts with `/` and origin matches.

            const href = link.getAttribute('href');
            if (href && href.startsWith('/') && link.origin === window.location.origin) {
                // Prevent default navigation
                e.preventDefault();
                this.push(href);
            }
        });
    }

    push(path) {
        window.history.pushState({}, '', path);
        this.current.path = path;
        this._resolve();
    }

    _resolve() {
        const path = this.current.path;
        const match = this.routes.find(r => r.path === path);
        if (match) {
            this.current.component = match.component;
        } else {
            this.current.component = {
                template: '<h1>404 Not Found</h1>',
                render: function() { return '<h1>404 Not Found</h1>'; } // pre-compile fallback
            };
        }
    }

    install(Zeno) {
        Zeno.prototype.$router = this;

        Zeno.component('router-view', {
            // Placeholder template required for compiler to be happy?
            // Or just render function is enough if pre-compiled.
            template: `<!-- Router View -->`,

            render: function() {
                // Access router via data scope (injected in Zeno constructor)
                // this.$router is available on data proxy
                const routeComp = this.$router.current.component;
                if (!routeComp) return '';

                return this.renderDynamic(routeComp, {}, {});
            }
        });
    }
}
