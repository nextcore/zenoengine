
import { Zeno } from './zeno.js';
import { ZenoRouter } from './router.js';
import { store } from './store/index.js';

import App from './App.blade';
import Home from './pages/Home.blade';
import About from './pages/About.blade';
import Login from './pages/Login.blade';
import Dashboard from './pages/Dashboard.blade';
import Alert from './components/Alert.blade';
import Card from './components/Card.blade';

// Register Components
Zeno.component('Alert', Alert);
Zeno.component('Card', Card);

// Setup Router
const router = new ZenoRouter([
    { path: '/', component: Home },
    { path: '/about', component: About },
    { path: '/login', component: Login },
    {
        path: '/dashboard',
        component: Dashboard,
        meta: { requiresAuth: true }
    }
]);

// Route Guard
router.beforeEach((to, from) => {
    // Check Auth
    if (to.meta.requiresAuth) {
        const user = store.state.user;
        const isAuth = user && user.name !== 'Guest';
        if (!isAuth) {
            return '/login'; // Redirect
        }
    }
});

// Use Plugins
Zeno.use(router);
Zeno.use(store);

// Create instance
const app = Zeno.create(App);

app.mount('#app');
