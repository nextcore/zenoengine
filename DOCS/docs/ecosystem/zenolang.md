# ZenoLang

## Overview

**ZenoLang** is the expressive scripting language at the heart of ZenoEngine. It is designed to be human-readable, declarative, and deeply integrated with the ZenoEngine runtime.

ZenoLang's syntax can be thought of as a cross between YAML and a modern scripting language — deeply consistent and predictable, with no surprises.

## Basic Syntax

ZenoLang uses a **slot-based** syntax. Every statement is a `slot_name: value` pair or a `slot_name: value { child blocks }` block.

```zeno
// This is a comment
var: $name { val: 'Alice' }
log: $name

// These are equivalent:
var: $age { val: 30 }
log: $age
```

## Variables

All variables are prefixed with `$`. Assignment is done via `set`:

```zeno
var: $name { val: 'Alice' }
var: $age { val: 30 }
var: $active { val: true }
```

## String Interpolation

```zeno
var: $greeting { val: 'Hello, ' + $name + '!' }
log: $greeting
```

## Control Flow

### If / Else

```zeno
if: $age >= 18 {
    then: {
        log: 'Adult'
    }
    else: {
        log: 'Minor'
    }
}
```

### Foreach Loop

```zeno
foreach: $users {
    as: $user
    do: {
        log: $user.name
    }
}
```

### While Loop

```zeno
var: $i { val: 0 }
while: $i < 10 {
    do: {
        log: $i
        var: $i { val: $i + 1 }
    }
}
```

## Functions

```zeno
fn: 'greet' {
    params: ['name']
    do: {
        return: 'Hello, ' + $name
    }
}

call: 'greet' {
    args: { name: 'Alice' }
    as: $result
}
log: $result
```

## Slots & Extensions

ZenoLang's power comes from **slots** — built-in or plugin-registered commands that can do anything from query a database to render a Blade template. Each slot runs in the Go runtime with full concurrency and type safety.

```zeno
// Built-in slots
db.table: 'users'
db.get { as: $users }

view: 'users.index' { users: $users }
```

## Running Scripts

ZenoLang scripts are run directly by the ZenoEngine runtime:

```bash
zeno run my-script.zl
```

Or they are loaded at server startup as part of routes: in your `src/main.zl`.

## Language Specification

The full ZenoLang language specification is available in the [DOCS/LANGUAGE_SPECIFICATION.md](https://github.com/zenoengine/zenoengine/blob/main/DOCS/LANGUAGE_SPECIFICATION.md) file in the repository.
