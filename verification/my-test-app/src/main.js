
import { Zeno } from '../lib/zeno/src/zeno.js';
import { ZenoRouter } from '../lib/zeno/src/router.js';
import { createStore } from '../lib/zeno/src/store.js';

import App from './App.blade';
import Home from './pages/Home.blade';

// Router
const router = new ZenoRouter([
    { path: '/', component: Home }
]);

// Store
const store = createStore({
    state: { count: 0 }
});

Zeno.use(router);
Zeno.use(store);

// Mount
const app = Zeno.create(App);
app.mount('#app');
