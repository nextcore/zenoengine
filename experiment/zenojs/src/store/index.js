
import { createStore } from '../store.js';

export const store = createStore({
    state: {
        counter: 0,
        user: { name: 'Guest' }
    },
    mutations: {
        INCREMENT(state) {
            state.counter++;
        },
        SET_USER(state, name) {
            state.user.name = name;
        }
    },
    actions: {
        increment({ commit }) {
            commit('INCREMENT');
        },
        login({ commit }, name) {
            setTimeout(() => {
                commit('SET_USER', name);
            }, 500); // Simulate API
        }
    },
    getters: {
        doubleCounter: (state) => state.counter * 2
    }
});
