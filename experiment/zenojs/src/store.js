
import { reactive, effect } from './reactivity.js';

export function createStore(options = {}) {
    // 1. Reactive State
    const state = reactive(options.state || {});

    // 2. Getters
    // We use a proxy to evaluate getters on access?
    // Or just define properties on the store instance.
    const getters = {};
    const rawGetters = options.getters || {};

    // For each getter, we want it to be computed.
    // Since we don't have a computed() implementation in reactivity.js yet,
    // we can just run the function. The effect wrapping the component render
    // will track dependencies accessed inside the getter.
    // So simple property accessors work fine for now!

    for (const key in rawGetters) {
        Object.defineProperty(getters, key, {
            get: () => rawGetters[key](state),
            enumerable: true
        });
    }

    // 3. Mutations (Sync)
    const mutations = options.mutations || {};
    function commit(type, payload) {
        const handler = mutations[type];
        if (!handler) {
            console.error(`[ZenoStore] Mutation ${type} not found`);
            return;
        }
        handler(state, payload);
    }

    // 4. Actions (Async)
    const actions = options.actions || {};
    function dispatch(type, payload) {
        const handler = actions[type];
        if (!handler) {
            console.error(`[ZenoStore] Action ${type} not found`);
            return Promise.reject(new Error(`Action ${type} not found`));
        }
        // Action context
        const context = {
            state,
            commit,
            dispatch,
            getters
        };
        return Promise.resolve(handler(context, payload));
    }

    // The Store Instance
    const store = {
        state,
        getters,
        commit,
        dispatch,
        install(Zeno) {
            Zeno.prototype.$store = store;
        }
    };

    return store;
}
