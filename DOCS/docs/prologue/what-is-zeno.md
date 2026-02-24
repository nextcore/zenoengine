# What is ZenoEngine?

ZenoEngine is a web application framework built with Go under the hood, but designed with a familiar, expressive syntax called **ZenoLang** that Laravel developers will immediately feel comfortable with.

## Philosophy

Laravel made PHP web development elegant. ZenoEngine takes that same philosophy but applies it to a modern, compiled, high-performance runtime.

> "The best of Laravel, reborn in Go."

## Why ZenoEngine?

- 🚀 **Blazing Fast**: Native Go performance. No PHP overhead.
- 🦅 **Familiar Syntax**: ZenoLang reads like a mix of Laravel and YAML. No new mental models.
- 📦 **Single Binary**: Your entire app ships as one file. No runtime, no config tuning.
- 🛡️ **Type Safe**: Strong typing baked in.
- 🌿 **Full Stack**: Built-in Blade templating, ORM, routing, validation, and more.

## How it Works

```zeno
// routes/web.zl
http.get: '/' {
    view: 'welcome' {
        name: 'World'
    }
}

http.post: '/users' {
    validate {
        name: 'required'
        email: 'required,email'
    }
    orm.model: 'users'
    orm.save: $form
    redirect: '/users'
}
```

## Community

ZenoEngine is fully open source. Contributions are welcome on GitHub.
