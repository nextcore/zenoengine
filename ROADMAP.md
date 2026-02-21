# Roadmap: Zeno Go - Strategy to Realize "Scriptable Systems Framework"

This document outlines the strategic plan to transform **Zeno Go** into a world-class framework, rivaling Laravel (DX) and ASP.NET Core (Performance). The execution is divided into three phases: **Foundation**, **Expansion**, and **Domination**.

## Phase 1: The Foundation (Core Strength & Stability)
*Focus: Developer Experience (DX), Stability, and Automated Documentation.*

### 1. Self-Documenting Engine (Immediate Action)
Currently, slot documentation is manual. We will implement a metadata system directly in the Go source code.
*   **Action:** Refactor `engine.Register()` to accept `SlotMeta` structs containing descriptions, examples, and parameter types.
*   **Outcome:** `zeno docs` command will auto-generate complete API references. No more outdated docs.
*   **Benefit:** AI (and humans) can instantly understand Zeno's capabilities.

### 2. Standardized Language (ZenoLang 1.0)
Ensure absolute consistency in `.zl` syntax.
*   **Action:** Strict parser rules with human-readable error messages (inspired by Rust/Elm).
*   **Feature:** Optional static type-checking for critical variables in `.zl` scripts.

### 3. Unified Testing Harness
Guarantee backward compatibility.
*   **Action:** Create a massive test suite that runs thousands of `.zl` script scenarios against the engine.
*   **Outcome:** Fearless refactoring of the core engine without breaking user apps.

---

## Phase 1.5: The Laravel Bridge (Seamless Adoption) ✅ COMPLETED
*Focus: Making Zeno feel like home for Laravel Developers.*

### 1. Eloquent ORM Parity (Zeno Query Builder) ✅
Replicate the "magic" and ease of Eloquent without the overhead.
*   **Action:** Create a fluent, expressive query builder in `.zl` that mimics Eloquent.
*   **Status:** Done. `db.table` supports nested config (`where`, `limit`, `get`). `db.where` supports shorthand. Route Model Binding (`bind: { user: "users" }`) implemented.

### 2. Blade Component System (ZenoBlade++) ✅
Upgrade the existing template engine to support modern Component architecture.
*   **Action:** Implement `<x-alert type="error" />` syntax support.
*   **Status:** Done. Parser supports `<x-tag>`, `$attributes` bag implemented with `SafeString`.

### 3. Artisan Console Parity (Zeno CLI) ✅
Ensure every essential `artisan` command has a `zeno` equivalent.
*   **Mapping:**
    *   `make:model` / `make:schema`: ✅ Scaffolds migrations.
    *   `migrate`: ✅ Runs migrations.
    *   `route:list`: ✅ Lists routes (with hang fix).
    *   `docs`: ✅ Auto-generates API docs.

### 4. Middleware & Request Lifecycle ✅
Adopt the robust request pipeline model.
*   **Action:** Implement a middleware stack for Authentication, Throttling, and CSRF protection that mirrors Laravel's Kernel.
*   **Status:** Done. Router supports middleware arrays (`middleware: ['auth', 'web']`) and secure JWT handling. Global helpers (`dd`, `dump`, `abort`, `config`) added.

---

## Phase 2: The Expansion (Ecosystem & Tooling)
*Focus: Matching the Giants in Tooling.*

### 1. Zeno Studio (LSP & IDE Support)
Bring the "IntelliSense" experience to Zeno.
*   **Action:** Build a Language Server Protocol (LSP) implementation for VS Code.
*   **Feature:** Smart autocomplete for Slots, Go-to-Definition, and real-time error detection as you type.
*   **Outcome:** Matches the productivity of C# in Visual Studio or PHPStorm.

### 2. Decentralized Package Manager
Simplify dependency management.
*   **Action:** Git-based plugin system. Import Zeno modules directly from GitHub URLs (like Deno/Go).
*   **Benefit:** No complex centralized registry required.

### 3. Native Database Migrations
Define DB schema in code, not SQL.
*   **Action:** Schema definition in `.zl` files.
*   **Feature:** Auto-calculation of diffs and safe schema migrations.

---

## Phase 3: The Domination (Enterprise Performance)
*Focus: Surpassing Limits & Massive Scale.*

### 1. AOT Compiler (The Game Changer)
Develop as fast as PHP, deploy as fast as C++.
*   **Action:** Transpiler that converts `.zl` scripts into pure Go code (`.go`), compiled into a single binary.
*   **Outcome:** Zero overhead in production. Hot Reload in development. Best of both worlds.

### 2. Islands Architecture (Frontend Agnostic)
Support native partial hydration.
*   **Action:** Allow rendering React/Vue components on the server (SSR) and sending static HTML + minimal JS.
*   **Outcome:** Competes with Astro/Fresh.

### 3. Edge-Ready (WASM)
*   **Action:** Compile Zeno Engine to WebAssembly to run on Cloudflare Workers/Deno Deploy.

---

## Next Steps
We are entering **Phase 2: The Expansion**. The immediate priority is **Zeno Studio (LSP)** to provide autocomplete and syntax checking for developers.
