
import { Zeno } from './zeno.js';
import App from './App.zeno';
import Alert from './components/Alert.zeno';
import Card from './components/Card.zeno';

// Register Components
Zeno.component('Alert', Alert);
Zeno.component('Card', Card);

// Create instance
const app = Zeno.create(App);

app.mount('#app');
