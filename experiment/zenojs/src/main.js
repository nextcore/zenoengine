
import { Zeno } from './zeno.js';

// Counter App
const counterApp = Zeno.create({
    data() {
        return {
            count: 0,
            todos: [
                { id: 1, text: 'Learn ZenoJS', done: false },
                { id: 2, text: 'Build Something', done: false }
            ]
        };
    },
    methods: {
        increment() {
            this.count++;
        },
        decrement() {
            this.count--;
        },
        addTodo() {
            const input = document.getElementById('new-todo');
            if (input && input.value) {
                this.todos.push({
                    id: Date.now(),
                    text: input.value,
                    done: false
                });
                input.value = '';
            }
        },
        toggle(todo) {
            todo.done = !todo.done;
        },
        remove(todo) {
             const index = this.todos.indexOf(todo);
             if (index > -1) {
                 this.todos.splice(index, 1);
             }
        }
    },
    template: `
        <div class="box">
            <h1>ZenoJS Counter</h1>
            <p>Count is: <strong>{{ count }}</strong></p>
            <button @click="increment">+</button>
            <button @click="decrement">-</button>

            @if(count > 5)
                <p style="color: red">Count is high!</p>
            @endif
        </div>

        <div class="box">
            <h2>Todo List</h2>
            <ul>
                @foreach(todos as todo)
                    <li>
                        <span style="{{ todo.done ? 'text-decoration: line-through' : '' }}">
                            {{ todo.text }}
                        </span>
                        <button @click="toggle(todo)">Toggle</button>
                        <button @click="remove(todo)">X</button>
                    </li>
                @endforeach
            </ul>

            <input type="text" id="new-todo" placeholder="Add Todo">
            <button @click="addTodo">Add</button>
        </div>
    `
});

counterApp.mount('#app');
