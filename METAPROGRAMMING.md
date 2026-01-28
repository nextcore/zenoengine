# ZenoLang Metaprogramming Guide

> [!NOTE]
> Metaprogramming in ZenoLang allows code to treat other code as data, enabling dynamic code generation, introspection, and runtime modification of behavior.

## Core Concepts

ZenoLang's metaprogramming capabilities are built on four primary slots:

1.  **`meta.eval`**: Executes a string as ZenoLang code.
2.  **`meta.scope`**: Introspects the current execution scope (variables).
3.  **`meta.parse`**: Parses a code string into an AST (Abstract Syntax Tree) Map.
4.  **`meta.run`**: Executes an AST Map as code.
5.  **`meta.template`**: Renders a Blade template to a code string (Code Generation).

---

## 1. Dynamic Execution (`meta.eval`)

The `meta.eval` slot takes a string containing valid ZenoLang code and executes it within the current context. This is the foundation of dynamic behavior.

### Signature
```zenolang
meta.eval: "string_of_code"
```

### Example
```zenolang
$dynamic_path: "/api/status"
meta.eval: "http.get: '" + $dynamic_path + "' { return: 'OK' }"
```

> [!IMPORTANT]
> Variables used inside the evaluated string must be properly concatenated or resolved. `meta.eval` shares the **current scope**, so it can access variables defined before it runs.

---

## 2. Introspection (`meta.scope`)

`meta.scope` returns the current execution scope (all defined variables) as a Map. This is useful for passing context to views or debugging.

### Signature
```zenolang
$vars: meta.scope
```

### Example
```zenolang
$user: "Max"
$role: "Admin"

view.blade: 'debug.blade.zl' {
    vars: meta.scope  // Passes { user: "Max", role: "Admin" } to the view
}
```

---

## 3. Code as Data (`meta.parse` & `meta.run`)

These slots allow you to manipulate code structure directly.

-   **`meta.parse`**: Converts code string -> Map (AST).
-   **`meta.run`**: Converts Map (AST) -> Execution.

### Example: Modifying Code Structure
```zenolang
// 1. Parse code into a Map
$ast: meta.parse: "print: 'Hello World'"

// 2. Modify the AST Map (e.g., change the printed value)
// (Assuming standard map manipulation slots exist)
$ast.value: "Hello Zeno"

// 3. Execute the modified AST
meta.run: $ast  // Output: "Hello Zeno"
```

---

## Real-World Use Case: Dynamic Admin Generator

A powerful application of metaprogramming is generating boilerplate code automatically. Below is the implementation of a **Dynamic Admin Route Generator** that creates CRUD routes for multiple resources without writing manual handlers.

### The Implementation (`admin_generator.zl`)

```zenolang
// 1. Define Resources
$resources: []
array.push: $resources { val: 'products' }
array.push: $resources { val: 'orders' }
array.push: $resources { val: 'customers' }

// 2. Generate Routes Loop
// 2. Generate Routes Loop
foreach: $resources {
    as: $res
    do: {
        log: "Generating route for: " + $res

        // 3. Generate Code using Template
        // We use a Blade template to generate the ZenoLang code
        meta.template: 'codegen/admin_route.blade.zl' {
            resource: $res
            as: $code
        }

        // 4. Eval/Execute the Code
        meta.eval: $code
    }
}
```

### The Template (`views/codegen/admin_route.blade.zl`)

```blade
http.get: '/admin/{{ $resource }}' {
    view.blade: 'admin/list.blade.zl' {
        title: 'Manage {{ $resource }}'
        resource: '{{ $resource }}'
        params: meta.scope
    }
}
```

### How It Works
1.  **Configuration**: We define a list of resources (`products`, `orders`).
2.  **String Construction**: Inside the loop, we concatenate strings to form valid ZenoLang syntax.
    *   *Resulting String*: `http.get: '/admin/products' { view.blade: ... }`
3.  **Evaluation**: `meta.eval: $code` runs that string. The ZenoEngine runtime sees an `http.get` instruction and registers the route exactly as if it were written manually in the file.
4.  **Scope Sharing**: The `view.blade` call inside uses `meta.scope` to pass the generated variables (like `$res`) into the template.

---

## Best Practices & Safety

> [!WARNING]
> **Performance**: `meta.eval` involves parsing strings at runtime. While ZenoLang's parser is fast, excessive use (e.g., inside hot loops of a request handler) can impact performance. Use it for **initialization** (like generating routes at startup) rather than per-request logic.

> [!CAUTION]
> **Security**: Never pass untrusted user input directly to `meta.eval`. This allows arbitrary code execution.

*   **Debugging**: Use `log` to print the generated code string *before* evaluating it. This helps verify the syntax is correct.
*   **Editor Support**: Dynamically generated code is invisible to static analysis tools and IDEs. Use comments to document what magic is happening.
