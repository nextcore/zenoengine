
import { Zeno } from './zeno.js';
import App from './App.zeno'; // Import the SFC

// Create instance using the imported component definition
const app = Zeno.create(App);

app.mount('#app');
