# ZenoLang

## Overview

**ZenoLang** is the expressive scripting language at the heart of ZenoEngine. It is designed to be human-readable, declarative, and deeply integrated with the ZenoEngine runtime.

ZenoLang's syntax can be thought of as a cross between YAML and a modern scripting language — deeply consistent and predictable, with no surprises.

## Basic Syntax

ZenoLang uses a **slot-based** syntax. Every statement is a `slot_name: value` pair or a `slot_name: value { child blocks }` block.

```zeno
// This is a comment
set: $name { val: 'Alice' }
print: $name

// These are equivalent:
set: $age { val: 30 }
print: $age
```

## Variables

All variables are prefixed with `$`. Assignment is done via `set`:

```zeno
set: $name { val: 'Alice' }
set: $age { val: 30 }
set: $active { val: true }
```

## String Interpolation

```zeno
set: $greeting { val: 'Hello, ' + $name + '!' }
print: $greeting
```

## Control Flow

### If / Else

```zeno
if: $age >= 18 {
    print: 'Adult'
} else: {
    print: 'Minor'
}
```

### Foreach Loop

```zeno
foreach: $users as $user {
    print: $user.name
}
```

### While Loop

```zeno
set: $i { val: 0 }
while: $i < 10 {
    print: $i
    set: $i { val: $i + 1 }
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
print: $result
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
