# ZenoLang Language Specification v1.0

## 1. Introduction

ZenoLang is a domain-specific language for web development designed for:
- **Simplicity** - Easy to understand syntax
- **Productivity** - Zero boilerplate, focus on logic
- **Safety** - Built-in validation and error handling
- **Performance** - Optimized for web workloads

## 2. Lexical Structure

### 2.1 Character Set
- **Encoding:** UTF-8
- **Line Ending:** LF (`\n`) atau CRLF (`\r\n`)
- **Whitespace:** Space, Tab, Newline

### 2.2 Comments
```zenolang
// Single-line comment

/* Multi-line comment
   Can span multiple lines */
```

### 2.3 Identifiers
```ebnf
identifier ::= letter (letter | digit | '_')*
letter     ::= 'a'..'z' | 'A'..'Z'
digit      ::= '0'..'9'
```

**Examples:**
```zenolang
http
http.get
db.query
user_name
```

### 2.4 Keywords
```
Reserved Keywords:
- if, else, try, catch
- while, loop, for, foreach, forelse
- break, continue
- true, false, null
```

### 2.5 Operators
```
Arithmetic: +, -, *, /, %
Comparison: ==, !=, <, >, <=, >=
Logical:    &&, ||, !
Assignment: =
Increment:  ++, --
```

### 2.6 Literals

**String:**
```zenolang
"double quoted"
'single quoted'
```

**Number:**
```zenolang
123          // Integer
123.45       // Float
-42          // Negative
```

**Boolean:**
```zenolang
true
false
```

**Null:**
```zenolang
null
```

## 3. Syntax

### 3.1 Basic Structure

```ebnf
program    ::= statement*
statement  ::= slot | block | assignment
slot       ::= identifier ':' value? block?
block      ::= '{' statement* '}'
assignment ::= variable ':' value
variable   ::= '$' identifier
value      ::= string | number | boolean | null | variable | expression
```

### 3.2 Slot Syntax

**Format:**
```zenolang
slot_name: value {
    child_slot: value
}
```

**Examples:**
```zenolang
// Slot with value only
http.get: "/api/users"

// Slot with block only
http.get {
    path: "/api/users"
}

// Slot with value and block
http.get: "/api/users" {
    middleware: "auth"
}
```

### 3.3 Variable Declaration

**Syntax:**
```zenolang
$variable_name: value
```

**Examples:**
```zenolang
$name: "John Doe"
$age: 25
$is_active: true
$data: null
```

### 3.4 Variable Shorthand

**Object:**
```zenolang
$user: {
    name: "John"
    age: 25
    email: "john@example.com"
}
```

**Array:**
```zenolang
$items: ["apple", "banana", "orange"]
```

### 3.5 Variable Reference

```zenolang
$name: "John"
$greeting: "Hello, $name"  // String interpolation
$copy: $name               // Variable reference
```

### 3.6 Conditionals

**If-Else:**
```zenolang
if: "$age >= 18" {
    then: {
        // Adult logic
    }
    else: {
        // Minor logic
    }
}
```

**Unless:**
```zenolang
unless: "$is_logged_in" {
    do: {
        // Redirect to login
    }
}
```

### 3.7 Loops

**While Loop:**
```zenolang
$i: 0
while: "$i < 10" {
    do: {
        // Loop body
        $i: "$i + 1"
    }
}
```

**For Loop (C-style):**
```zenolang
for: "$i = 0; $i < 10; $i++" {
    do: {
        // Loop body
    }
}
```

**Foreach Loop:**
```zenolang
for: $items {
    as: $item
    do: {
        // Process $item
    }
}
```

**Forelse:**
```zenolang
forelse: $items {
    as: $item
    do: {
        // Process $item
    }
    empty: {
        // No items
    }
}
```

### 3.8 Error Handling

**Try-Catch:**
```zenolang
try {
    do: {
        // Code that might fail
    }
    catch: {
        // Error handling
        // $error contains error message
    }
}
```

### 3.9 Loop Control

**Break:**
```zenolang
while: true {
    do: {
        break: "$i == 5"  // Conditional break
    }
}
```

**Continue:**
```zenolang
for: $items {
    as: $item
    do: {
        continue: "$item == null"  // Skip null items
    }
}
```

## 4. Type System

### 4.1 Dynamic Typing

ZenoLang menggunakan dynamic typing dengan automatic type coercion.

**Types:**
- `string` - Text data
- `number` - Integer or Float
- `boolean` - true or false
- `null` - Absence of value
- `array` - Ordered collection
- `object` - Key-value pairs

### 4.2 Type Coercion Rules

**To String:**
```
number  → string:  123 → "123"
boolean → string:  true → "true"
null    → string:  null → ""
```

**To Number:**
```
string  → number:  "123" → 123
boolean → number:  true → 1, false → 0
null    → number:  null → 0
```

**To Boolean:**
```
string  → boolean: "" → false, non-empty → true
number  → boolean: 0 → false, non-zero → true
null    → boolean: null → false
```

### 4.3 Type Checking

**Runtime Type Checking:**
```zenolang
// Type is determined at runtime
$value: "123"
$number: "$value + 1"  // Coerced to number: 124
```

## 5. Standard Library

### 5.1 HTTP Slots

**Routing:**
```zenolang
http.get: "/path"
http.post: "/path"
http.put: "/path"
http.delete: "/path"
http.patch: "/path"
```

**Responses:**
```zenolang
http.ok: { data: $result }
http.created: { data: $result }
http.error: { message: "Error" }
http.not_found: { message: "Not found" }
http.unauthorized: { message: "Unauthorized" }
```

### 5.2 Database Slots

**Query:**
```zenolang
db.query: "SELECT * FROM users"
db.exec: "INSERT INTO users ..."
db.transaction: { /* queries */ }
```

**Query Builder:**
```zenolang
db.table: "users" {
    select: ["id", "name", "email"]
    where: { status: "active" }
    orderBy: "created_at"
    limit: 10
}
```

### 5.3 Validation Slots

```zenolang
validate: $request.body {
    rules: {
        name: "required|string|min:3"
        email: "required|email"
        age: "required|integer|min:18"
    }
    as: $validated
}
```

### 5.4 Collection Slots

```zenolang
collections.map: $items {
    callback: { /* transform */ }
    as: $result
}

collections.filter: $items {
    callback: { /* condition */ }
    as: $result
}

collections.reduce: $items {
    callback: { /* accumulate */ }
    initial: 0
    as: $result
}
```

## 6. Scoping Rules

### 6.1 Scope Hierarchy

```
Global Scope
  ├─ Route Scope (per HTTP request)
  │   ├─ Block Scope (if, while, for, etc)
  │   └─ Block Scope
  └─ Route Scope
```

### 6.2 Variable Visibility

**Parent-to-Child:**
```zenolang
$global: "visible"

http.get: "/test" {
    // Can access $global
    $local: "only here"
    
    if: true {
        then: {
            // Can access both $global and $local
        }
    }
}
```

**Child-to-Parent:**
```zenolang
http.get: "/test" {
    if: true {
        then: {
            $inner: "not visible outside"
        }
    }
    // Cannot access $inner here
}
```

## 7. Error Handling

### 7.1 Error Types

**Syntax Error:**
```
[file.zl:10:5] lexical error: unexpected character '&'
```

**Validation Error:**
```
[file.zl:15:3] validation error: unknown attribute 'invalid' for slot 'http.get'
```

**Runtime Error:**
```
[file.zl:20:7] execution error in 'db.query': database connection failed
```

### 7.2 Error Format

```
[filename:line:column] error_type: message
```

**Components:**
- `filename` - Source file
- `line` - Line number (1-indexed)
- `column` - Column number (1-indexed)
- `error_type` - Category of error
- `message` - Detailed description

## 8. Execution Model

### 8.1 Execution Flow

```
1. Parse Source → AST
2. Validate AST
3. Execute AST
   ├─ Resolve Variables
   ├─ Execute Slots
   └─ Handle Errors
```

### 8.2 Slot Execution

**Synchronous:**
```zenolang
db.query: "SELECT * FROM users"
// Waits for query to complete
```

**Asynchronous (Background Jobs):**
```zenolang
job.dispatch: "send_email" {
    data: { email: $user.email }
}
// Returns immediately
```

### 8.3 Context Management

**Request Context:**
- Timeout (default 30s)
- Cancellation
- Request data
- Response writer

**Execution Context:**
- Variables
- Scope chain
- Error state

## 9. Performance Considerations

### 9.1 Optimization Strategies

**Caching:**
- AST caching
- Handler caching
- Template caching

**Pooling:**
- Scope pooling
- Map pooling
- Buffer pooling

**Fast Paths:**
- Common operations optimized
- Bypass validation for cached handlers

### 9.2 Resource Limits

**Defaults:**
- Max loop iterations: 10,000
- Request timeout: 30s
- Max file size: 32MB
- Max request body: 10MB

## 10. Security

### 10.1 Built-in Protection

**Automatic:**
- CSRF protection
- XSS prevention (Blade)
- SQL injection prevention (parameterized queries)
- Rate limiting

**Configurable:**
- Request timeout
- Max iterations
- File size limits

### 10.2 Best Practices

**DO:**
- Use parameterized queries
- Validate all inputs
- Use HTTPS in production
- Enable CSRF protection

**DON'T:**
- Concatenate SQL queries
- Trust user input
- Disable validation
- Skip authentication

## 11. Compatibility

### 11.1 Cross-Platform Guarantees

**MUST be identical:**
- Syntax
- Semantics
- Error messages
- Standard library API

**MAY differ:**
- Performance characteristics
- Internal implementation
- Platform-specific features

### 11.2 Version Compatibility

**Semantic Versioning:**
```
MAJOR.MINOR.PATCH

MAJOR: Breaking changes
MINOR: New features (backward compatible)
PATCH: Bug fixes
```

**Compatibility Promise:**
- Same MAJOR version = Compatible
- Different MAJOR version = May break

## 12. Future Considerations

### 12.1 Planned Features

- **Modules:** Import/export system
- **Generics:** Type parameters
- **Async/Await:** Explicit async syntax
- **Pattern Matching:** Advanced conditionals
- **Macros:** Code generation

### 12.2 Deprecation Policy

**Process:**
1. Mark as deprecated (1 major version)
2. Show warnings (1 major version)
3. Remove feature (next major version)

**Example:**
```
v1.0: Feature introduced
v2.0: Feature deprecated (warnings)
v3.0: Feature removed
```

---

**Specification Version:** 1.0  
**Last Updated:** 2025-01-01  
**Status:** Draft  
**Implementations:** Go (1.0), .NET (Planned), Rust (Planned)
