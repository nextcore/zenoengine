var S = Object.defineProperty;
var F = (r, t, n) => t in r ? S(r, t, { enumerable: !0, configurable: !0, writable: !0, value: n }) : r[t] = n;
var m = (r, t, n) => F(r, typeof t != "symbol" ? t + "" : t, n);
import { c as L } from "./compiler-Dkyzkoo9.js";
let g = null;
const $ = /* @__PURE__ */ new WeakMap();
function A(r, t) {
  if (g) {
    let n = $.get(r);
    n || (n = /* @__PURE__ */ new Map(), $.set(r, n));
    let s = n.get(t);
    s || (s = /* @__PURE__ */ new Set(), n.set(t, s)), s.add(g);
  }
}
function T(r, t) {
  const n = $.get(r);
  if (!n) return;
  const s = n.get(t);
  s && s.forEach((e) => e());
}
function b(r) {
  return typeof r != "object" || r === null ? r : new Proxy(r, {
    get(t, n, s) {
      const e = Reflect.get(t, n, s);
      return A(t, n), typeof e == "object" && e !== null ? b(e) : e;
    },
    set(t, n, s, e) {
      const o = t[n], i = Reflect.set(t, n, s, e);
      return o !== s && T(t, n), i;
    }
  });
}
function j(r) {
  const t = () => {
    try {
      g = t, r();
    } finally {
      g = null;
    }
  };
  return t(), t;
}
const D = {
  required: (r) => r != null && r !== "",
  email: (r) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(r),
  numeric: (r) => !isNaN(parseFloat(r)) && isFinite(r),
  min: (r, t) => (typeof r == "string" ? r.length : r) >= parseFloat(t),
  max: (r, t) => (typeof r == "string" ? r.length : r) <= parseFloat(t)
}, I = {
  required: "This field is required.",
  email: "Please enter a valid email.",
  numeric: "This field must be a number.",
  min: "Minimum value is :arg.",
  max: "Maximum value is :arg."
};
function q(r, t) {
  const n = {};
  let s = !0;
  for (const e in t) {
    const o = t[e].split("|"), i = M(r, e);
    for (const a of o) {
      const [l, c] = a.split(":"), d = D[l];
      if (d) {
        if (l !== "required" && (i == null || i === ""))
          continue;
        if (!d(i, c)) {
          const h = I[l] || "Invalid field.";
          n[e] = h.replace(":arg", c), s = !1;
          break;
        }
      } else
        console.warn(`[Validator] Unknown rule: ${l}`);
    }
  }
  return { isValid: s, errors: n };
}
function M(r, t) {
  return t.split(".").reduce((n, s) => n && n[s] !== void 0 ? n[s] : void 0, r);
}
const w = "zeno_auth_token", N = {
  // Configuration
  baseURL: "",
  // Set Base URL
  setBaseURL(r) {
    this.baseURL = r;
  },
  // Token Management
  setToken(r) {
    localStorage.setItem(w, r);
  },
  getToken() {
    return localStorage.getItem(w);
  },
  removeToken() {
    localStorage.removeItem(w);
  },
  // Request Helper
  async request(r, t, n = null, s = {}) {
    const e = this.getToken(), o = {
      method: r,
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
        ...s
      }
    };
    e && (o.headers.Authorization = `Bearer ${e}`), n && (o.body = JSON.stringify(n));
    const i = t.startsWith("http") ? t : this.baseURL + t;
    try {
      const a = await fetch(i, o);
      if (a.status === 401)
        throw this.removeToken(), window.dispatchEvent(new CustomEvent("zeno:unauthorized")), new Error("Unauthorized");
      if (!a.ok) {
        const l = await a.json().catch(() => ({}));
        throw new Error(l.message || `HTTP Error ${a.status}`);
      }
      return await a.json();
    } catch (a) {
      throw console.error("[ZenoHTTP]", a), a;
    }
  },
  get(r, t) {
    return this.request("GET", r, null, t);
  },
  post(r, t, n) {
    return this.request("POST", r, t, n);
  },
  put(r, t, n) {
    return this.request("PUT", r, t, n);
  },
  delete(r, t) {
    return this.request("DELETE", r, null, t);
  }
};
let p = {}, f = null;
function U(r) {
  f && f.hooks.mounted.push(r);
}
function G(r) {
  f && f.hooks.unmounted.push(r);
}
const u = class u {
  static use(t) {
    this._plugins.push(t), t.install && t.install(u);
  }
  static component(t, n) {
    this._components[t] = n;
  }
  static layout(t, n) {
    this._layouts[t] = n;
  }
  static service(t, n) {
    this._services[t] = n;
  }
  static view(t, n) {
    this._views[t] = n;
  }
  static create(t) {
    return new u(t);
  }
  constructor(t) {
    this.hooks = { mounted: [], unmounted: [] };
    const n = f;
    f = this;
    try {
      const e = t.data || (() => ({})), o = typeof e == "function" ? e() : e;
      o.$errors = {}, t.mounted && this.hooks.mounted.push(t.mounted.bind(null)), t.unmounted && this.hooks.unmounted.push(t.unmounted), this.data = b(o);
    } finally {
      f = n;
    }
    this.methods = t.methods || {}, this.template = t.template || "", this.props = t.props || [], this.slots = t.slots || {}, this.sections = t.sections || {};
    const s = t.attributes || {};
    this.data.$attributes = {
      ...s,
      merge: (e) => {
        const o = { ...e, ...s };
        return e.class && s.class && (o.class = `${e.class} ${s.class}`), new E(o);
      },
      toString: () => new E(s).toString()
    }, this.el = null, this.data.$helpers = {
      classNames: (e) => {
        let o = [];
        if (typeof e == "string") o.push(e);
        else if (Array.isArray(e))
          e.forEach((i) => {
            if (typeof i == "string") o.push(i);
            else if (typeof i == "object")
              for (const a in i) i[a] && o.push(a);
          });
        else if (typeof e == "object")
          for (const i in e) e[i] && o.push(i);
        return o.join(" ");
      },
      styleNames: (e) => {
        let o = [];
        if (typeof e == "object")
          for (const i in e) e[i] && o.push(`${i}: ${e[i]}`);
        return o.join("; ");
      }
    }, this.data.$slots = {};
    for (const e in this.slots) this.data.$slots[e] = !0;
    if (this.data.$slot = (e = "default") => this.slots[e] ? this.slots[e]() : "", this.data.$sections = this.sections, this.data.$push = (e, o) => {
      p[e] || (p[e] = []), p[e].push(o);
    }, this.data.$stack = (e) => p[e] ? p[e].join("") : "", this.data.$services = u._services, this.data.$http = N, u.prototype.$router && (this.data.$router = u.prototype.$router), u.prototype.$store && (this.data.$store = u.prototype.$store, this.data.user === void 0 && this.data.$store.state.user && (this.data.auth = {
      user: () => this.data.$store.state.user,
      check: () => {
        const e = this.data.$store.state.user;
        return e && e.name !== "Guest" && e.name !== null;
      },
      guest: () => {
        const e = this.data.$store.state.user;
        return !e || e.name === "Guest" || e.name === null;
      }
    })), this.data.$validate = (e) => {
      const o = q(this.data, e);
      return this.data.$errors = o.errors, o.isValid;
    }, t.render)
      this.renderFn = t.render;
    else {
      const e = L(this.template);
      try {
        this.renderFn = new Function(e);
      } catch (o) {
        console.error("Compilation Error:", o), this.renderFn = () => "Error compiling template";
      }
    }
    for (const e in this.methods)
      this.methods[e] = this.methods[e].bind(this.data);
    this.hooks.mounted = this.hooks.mounted.map((e) => e.bind(this.data)), this.hooks.unmounted = this.hooks.unmounted.map((e) => e.bind(this.data)), this.renderComponent = this.renderComponent.bind(this), this.renderLayout = this.renderLayout.bind(this), this.renderInclude = this.renderInclude.bind(this), this.renderDynamic = this.renderDynamic.bind(this);
  }
  // ... [Rest of class same as before] ...
  renderComponent(t, n, s) {
    const e = u._components[t];
    return e ? this.renderDynamic(e, n, s) : `[Component ${t} not found]`;
  }
  renderDynamic(t, n = {}, s = {}) {
    const o = (t.data || (() => ({})))(), i = t.props || [], a = {}, l = {};
    for (const d in n)
      i.includes(d) ? a[d] = n[d] : l[d] = n[d];
    Object.assign(o, a);
    const c = new u({
      data: () => o,
      methods: t.methods,
      template: t.template,
      render: t.render,
      props: t.props,
      attributes: l,
      slots: s,
      mounted: t.mounted,
      unmounted: t.unmounted
    });
    try {
      return c.renderFn.call(c.data);
    } catch (d) {
      return `Error: ${d.message}`;
    }
  }
  renderLayout(t, n) {
    const s = u._layouts[t];
    if (!s) return `[Layout ${t} not found]`;
    const o = (s.data || (() => ({})))(), i = new u({
      data: () => o,
      methods: s.methods,
      template: s.template,
      render: s.render,
      sections: n
    });
    try {
      return i.renderFn.call(i.data);
    } catch (a) {
      return a.message;
    }
  }
  renderInclude(t, n) {
    const s = u._views[t];
    if (!s) return `[View ${t} not found]`;
    const e = { ...this.data, ...n }, o = new u({ data: () => e, template: s });
    try {
      return o.renderFn.call(o.data);
    } catch (i) {
      return i.message;
    }
  }
  mount(t) {
    if (this.el = document.querySelector(t), !this.el) {
      console.error(`Element ${t} not found`);
      return;
    }
    let n = !1;
    j(() => {
      this.render(), n || (n = !0, this.hooks.mounted.forEach((s) => s()));
    });
  }
  render() {
    p = {}, this.data.renderComponent = this.renderComponent, this.data.renderLayout = this.renderLayout, this.data.renderInclude = this.renderInclude, this.data.renderDynamic = this.renderDynamic;
    let t = "";
    try {
      t = this.renderFn.call(this.data);
    } catch (n) {
      console.error("Render Error:", n), t = `<div style="color:red">Render Error: ${n.message}</div>`;
    }
    return this.el && (this.el.innerHTML = t, this.bindEvents()), t;
  }
  bindEvents() {
    if (!this.el) return;
    this.el.querySelectorAll("[data-z-click]").forEach((s) => {
      const e = s.getAttribute("data-z-click");
      this.methods[e] && s.addEventListener("click", (o) => this.methods[e](o)), s.removeAttribute("data-z-click");
    }), this.el.querySelectorAll("[data-z-model]").forEach((s) => {
      const e = s.getAttribute("data-z-model");
      s.addEventListener("input", (o) => {
        const i = o.target.type === "checkbox" ? o.target.checked : o.target.value;
        if (e.includes(".")) {
          const a = e.split(".");
          let l = this.data;
          for (let c = 0; c < a.length - 1; c++)
            l = l[a[c]];
          l[a[a.length - 1]] = i;
        } else
          this.data[e] = i;
      }), s.removeAttribute("data-z-model");
    });
  }
};
m(u, "_components", {}), m(u, "_layouts", {}), m(u, "_services", {}), m(u, "_views", {}), m(u, "_plugins", []);
let v = u;
class E {
  constructor(t) {
    this.attrs = t;
  }
  toString() {
    return Object.entries(this.attrs).map(([t, n]) => n === !0 ? t : n === !1 || n === null || n === void 0 ? "" : `${t}="${n}"`).filter((t) => t).join(" ");
  }
}
class x {
  constructor(t) {
    this.routes = t, this.current = b({
      path: window.location.pathname,
      component: null,
      params: {}
    }), this.globalGuards = [], this._resolve(), window.addEventListener("popstate", () => {
      this._navigate(window.location.pathname);
    }), this._setupLinkInterception();
  }
  beforeEach(t) {
    this.globalGuards.push(t);
  }
  push(t) {
    this._navigate(t, !0);
  }
  async _navigate(t, n = !1) {
    const s = this.routes.find((i) => i.path === t), e = { path: t, meta: s ? s.meta || {} : {} }, o = { path: this.current.path };
    for (const i of this.globalGuards) {
      const a = await i(e, o);
      if (a === !1) return;
      if (typeof a == "string")
        return this.push(a);
    }
    if (s && s.beforeEnter) {
      const i = await s.beforeEnter(e, o);
      if (i === !1) return;
      if (typeof i == "string")
        return this.push(i);
    }
    n && window.history.pushState({}, "", t), this.current.path = t, s ? this.current.component = s.component : this.current.component = {
      template: "<h1>404 Not Found</h1>",
      render: function() {
        return "<h1>404 Not Found</h1>";
      }
    };
  }
  _resolve() {
    this._navigate(window.location.pathname);
  }
  _setupLinkInterception() {
    document.addEventListener("click", (t) => {
      const n = t.target.closest("a");
      if (!n) return;
      const s = n.getAttribute("href");
      s && s.startsWith("/") && n.origin === window.location.origin && (t.preventDefault(), this.push(s));
    });
  }
  install(t) {
    t.prototype.$router = this, t.component("router-view", {
      template: "<!-- Router View -->",
      render: function() {
        const n = this.$router.current.component;
        return n ? this.renderDynamic(n, {}, {}) : "";
      }
    });
  }
}
function _(r = {}) {
  const t = b(r.state || {}), n = {}, s = r.getters || {};
  for (const c in s)
    Object.defineProperty(n, c, {
      get: () => s[c](t),
      enumerable: !0
    });
  const e = r.mutations || {};
  function o(c, d) {
    const h = e[c];
    if (!h) {
      console.error(`[ZenoStore] Mutation ${c} not found`);
      return;
    }
    h(t, d);
  }
  const i = r.actions || {};
  function a(c, d) {
    const h = i[c];
    if (!h)
      return console.error(`[ZenoStore] Action ${c} not found`), Promise.reject(new Error(`Action ${c} not found`));
    const k = {
      state: t,
      commit: o,
      dispatch: a,
      getters: n
    };
    return Promise.resolve(h(k, d));
  }
  const l = {
    state: t,
    getters: n,
    commit: o,
    dispatch: a,
    install(c) {
      c.prototype.$store = l;
    }
  };
  return l;
}
let y = null;
function R(r) {
  let t = r;
  const n = /* @__PURE__ */ new Set();
  return {
    get value() {
      return y && n.add(y), t;
    },
    set value(s) {
      t !== s && (t = s, n.forEach((e) => e()));
    },
    // Helpers
    peek() {
      return t;
    },
    subscribe(s) {
      return n.add(s), () => n.delete(s);
    }
  };
}
function O(r) {
  const t = R(void 0);
  return z(() => {
    t.value = r();
  }), {
    get value() {
      return t.value;
    }
  };
}
function z(r) {
  const t = () => {
    const n = y;
    y = t;
    try {
      r();
    } finally {
      y = n;
    }
  };
  return t(), t;
}
export {
  v as Zeno,
  x as ZenoRouter,
  O as computed,
  _ as createStore,
  z as effect,
  N as http,
  U as onMounted,
  G as onUnmounted,
  R as signal,
  q as validate
};
