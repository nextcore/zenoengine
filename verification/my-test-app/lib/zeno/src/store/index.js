
import { createStore } from '../store.js';
import { http } from '../http.js';

// Setup listener for auto-logout
window.addEventListener('zeno:unauthorized', () => {
    // We can't access store instance directly here easily unless exported/global.
    // Ideally this listener is setup inside app init.
    // For now, let's rely on the action to handle cleanup.
    console.warn("Session expired. Please login again.");
});

export const store = createStore({
    state: {
        counter: 0,
        user: null // null if guest
    },
    mutations: {
        INCREMENT(state) {
            state.counter++;
        },
        SET_USER(state, user) {
            state.user = user;
        }
    },
    actions: {
        increment({ commit }) {
            commit('INCREMENT');
        },

        async login({ commit }, credentials) {
            // Simulate API Call
            // const data = await http.post('/api/login', credentials);

            // Mock Response
            return new Promise((resolve) => {
                setTimeout(() => {
                    const mockToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.mock";
                    const mockUser = { name: credentials.username || 'User', role: 'admin' };

                    // Save Token
                    http.setToken(mockToken);

                    // Set State
                    commit('SET_USER', mockUser);

                    resolve(true);
                }, 500);
            });
        },

        logout({ commit }) {
            http.removeToken();
            commit('SET_USER', null);
        },

        // Restore session on page load
        async checkAuth({ commit }) {
            const token = http.getToken();
            if (token) {
                // Verify token with API
                // const user = await http.get('/api/me');
                // For mock:
                console.log("Token found, restoring session...");
                commit('SET_USER', { name: 'Restored User', role: 'admin' });
            }
        }
    },
    getters: {
        doubleCounter: (state) => state.counter * 2,
        isAuthenticated: (state) => !!state.user
    }
});
