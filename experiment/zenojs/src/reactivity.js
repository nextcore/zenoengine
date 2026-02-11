
// Reactivity System (Vue 3/Solid Style)
let activeEffect = null;
const targetMap = new WeakMap();

/**
 * Tracks dependency. Call this inside getter.
 * @param {object} target - The object being accessed
 * @param {string} key - The property key being accessed
 */
export function track(target, key) {
    if (activeEffect) {
        let depsMap = targetMap.get(target);
        if (!depsMap) {
            depsMap = new Map();
            targetMap.set(target, depsMap);
        }
        let dep = depsMap.get(key);
        if (!dep) {
            dep = new Set();
            depsMap.set(key, dep);
        }
        dep.add(activeEffect);
    }
}

/**
 * Triggers effects. Call this inside setter.
 * @param {object} target - The object being modified
 * @param {string} key - The property key being modified
 */
export function trigger(target, key) {
    const depsMap = targetMap.get(target);
    if (!depsMap) return;

    const dep = depsMap.get(key);
    if (dep) {
        dep.forEach(effect => effect());
    }
}

/**
 * Creates a reactive object using Proxy.
 * @param {object} target - The object to make reactive
 */
export function reactive(target) {
    // If primitive, return as is (use ref() for primitives - not implemented here yet)
    if (typeof target !== 'object' || target === null) {
        return target;
    }

    return new Proxy(target, {
        get(target, key, receiver) {
            const res = Reflect.get(target, key, receiver);
            track(target, key);

            // Deep reactivity
            if (typeof res === 'object' && res !== null) {
                return reactive(res);
            }
            return res;
        },
        set(target, key, value, receiver) {
            const oldValue = target[key];
            const res = Reflect.set(target, key, value, receiver);
            if (oldValue !== value) {
                trigger(target, key);
            }
            return res;
        }
    });
}

/**
 * Registers a side effect that re-runs when dependencies change.
 * @param {function} fn - The function to run
 */
export function effect(fn) {
    const effectFn = () => {
        try {
            activeEffect = effectFn;
            fn();
        } finally {
            activeEffect = null;
        }
    };
    effectFn(); // Run immediately
    return effectFn;
}
