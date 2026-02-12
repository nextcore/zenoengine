
import { Zeno } from '../lib/zeno/src/zeno.js';
import { ZenoRouter } from '../lib/zeno/src/router.js';

import App from './App.blade';
import DocsLayout from './layouts/DocsLayout.blade';
import CodeBlock from './components/CodeBlock.blade';

import Home from './pages/Home.blade';
import Installation from './pages/Installation.blade';
import BladeSyntax from './pages/BladeSyntax.blade';

// Register Global Components/Layouts
Zeno.layout('DocsLayout', DocsLayout);
Zeno.component('CodeBlock', CodeBlock);

// Setup Router
const router = new ZenoRouter([
    { path: '/', component: Home },
    { path: '/installation', component: Installation },
    { path: '/blade-syntax', component: BladeSyntax },
    // Fallback
    { path: '/concepts', component: { template: '<h1>Coming Soon</h1><p>Concepts documentation pending.</p>' } },
    { path: '/components', component: { template: '<h1>Coming Soon</h1><p>Components documentation pending.</p>' } },
    { path: '/store', component: { template: '<h1>Coming Soon</h1><p>Store documentation pending.</p>' } },
    { path: '/router', component: { template: '<h1>Coming Soon</h1><p>Router documentation pending.</p>' } }
]);

Zeno.use(router);

// Mount
const app = Zeno.create(App);
app.mount('#app');
