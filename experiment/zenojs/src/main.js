
import { Zeno } from './zeno.js';
import { ZenoRouter } from './router.js';
import { store } from './store/index.js'; // Import Store

import App from './App.blade';
import Home from './pages/Home.blade';
import About from './pages/About.blade';
import Alert from './components/Alert.blade';
import Card from './components/Card.blade';

// Register Components
Zeno.component('Alert', Alert);
Zeno.component('Card', Card);

// Setup Router
const router = new ZenoRouter([
    { path: '/', component: Home },
    { path: '/about', component: About }
]);

// Use Plugins
Zeno.use(router);
Zeno.use(store);

// Create instance
const app = Zeno.create(App);

app.mount('#app');
