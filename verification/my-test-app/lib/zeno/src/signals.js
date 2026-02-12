
// Fine-Grained Signals Implementation (Inspired by Solid/Preact)

let activeSubscriber = null;

// 1. Signal (State)
export function signal(initialValue) {
    let value = initialValue;
    const subscribers = new Set();

    return {
        get value() {
            if (activeSubscriber) {
                subscribers.add(activeSubscriber);
            }
            return value;
        },
        set value(newValue) {
            if (value !== newValue) {
                value = newValue;
                subscribers.forEach(sub => sub());
            }
        },
        // Helpers
        peek() { return value; },
        subscribe(fn) {
            subscribers.add(fn);
            return () => subscribers.delete(fn);
        }
    };
}

// 2. Computed (Derived State)
export function computed(fn) {
    const s = signal(undefined);

    // Effect to re-compute when dependencies change
    effect(() => {
        s.value = fn();
    });

    // Return read-only signal-like object
    return {
        get value() {
            return s.value;
        }
    };
}

// 3. Effect (Side Effects)
export function effect(fn) {
    const run = () => {
        const prev = activeSubscriber;
        activeSubscriber = run;
        try {
            fn();
        } finally {
            activeSubscriber = prev;
        }
    };
    run();
    return run; // return runner (can be used to stop?)
}

// Compatibility with existing "reactive" (Proxy) system
// We can re-implement `reactive` using signals under the hood?
// Or keep them separate for now.
// Current ZenoJS uses `reactive.js` (Proxy).
// Signals are an alternative primitive.
