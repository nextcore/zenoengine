# ZenoJS

ZenoJS is an experimental Client-Side Rendering (CSR) framework that brings the elegance of **ZenoBlade** syntax to the browser. It aims to provide a lightweight, reactive SPA experience similar to Vue.js or Alpine.js, but using the familiar `@if`, `@foreach`, and `{{ }}` syntax from ZenoLang's server-side templating engine.

## Features

*   **ZenoBlade Syntax**: Use `@if`, `@else`, `@foreach` and `{{ variable }}` directly in your client-side templates.
*   **Reactive State**: Built-in reactivity system (using Proxies) that automatically updates the DOM when data changes.
*   **Component System**: Define components with `data`, `methods`, and `template`.
*   **Event Handling**: Bind events easily with `@click="method"`.

## Getting Started

1.  Open `index.html` in a modern browser (using a local server like `vite` or `http-server`).
2.  See the counter and todo list example in action.

## Example

```javascript
import { Zeno } from './zeno.js';

const app = Zeno.create({
    data() {
        return {
            count: 0
        };
    },
    methods: {
        increment() {
            this.count++;
        }
    },
    template: `
        <div>
            <h1>Count: {{ count }}</h1>
            <button @click="increment">+</button>

            @if(count > 5)
                <p>It's high!</p>
            @endif
        </div>
    `
});

app.mount('#app');
```

## Project Structure

*   `src/zeno.js`: Main component class.
*   `src/reactivity.js`: Core reactivity system (Proxy + Effect).
*   `src/compiler.js`: ZenoBlade to JavaScript compiler.
*   `src/main.js`: Demo application.

## Development

```bash
cd experiment/zenojs
npm install
npm run dev
```
