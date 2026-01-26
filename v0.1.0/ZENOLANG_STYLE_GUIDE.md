# ZenoLang Style Guide

## Overview

This guide defines the official coding standards for ZenoLang (.zl) files in ZenoEngine projects.

## 1. Indentation

**Standard**: 4 spaces per indentation level

```zenolang
// ✅ Good
http.get: /endpoint {
    auth.middleware: {
        secret: env("JWT_SECRET")
        do: {
            call: my_function
        }
    }
}

// ❌ Bad (2 spaces)
http.get: /endpoint {
  auth.middleware: {
    do: {
      call: my_function
    }
  }
}
```

## 2. Naming Conventions

### Functions
**snake_case** - lowercase with underscores

```zenolang
// ✅ Good
fn: get_user_profile { }
fn: create_product { }
fn: send_email_notification { }

// ❌ Bad
fn: getUserProfile { }
fn: CreateProduct { }
fn: Send-Email { }
```

### Variables
**snake_case** with `$` prefix

```zenolang
// ✅ Good
$user_id: 123
$product_list: []
$total_amount: 0

// ❌ Bad
$userId: 123
$ProductList: []
$TotalAmount: 0
```

### Routes
**kebab-case** in URLs

```zenolang
// ✅ Good
http.get: /api/v1/user-profile
http.post: /api/v1/create-order

// ❌ Bad
http.get: /api/v1/user_profile
http.post: /api/v1/CreateOrder
```

## 3. Spacing

### After Colons
Always add space after `:`

```zenolang
// ✅ Good
http.get: /endpoint {
    summary: "Get endpoint"
}

// ❌ Bad
http.get:/endpoint{
    summary:"Get endpoint"
}
```

### Blank Lines
- One blank line between major blocks
- No blank lines within small blocks

```zenolang
// ✅ Good
fn: function_one {
    log: "First function"
}

fn: function_two {
    log: "Second function"
}

// ❌ Bad
fn: function_one {
    log: "First function"
}
fn: function_two {
    log: "Second function"
}
```

### No Trailing Spaces
Remove all trailing whitespace

## 4. Comments

### Single-Line Comments
Use `//` with space after

```zenolang
// ✅ Good
// This is a comment
http.get: /endpoint { }

// ❌ Bad
//This is a comment
http.get: /endpoint { }
```

### Section Headers
Use comments to organize code

```zenolang
// ============================================
// Authentication Routes
// ============================================

http.post: /login { }
http.post: /register { }

// ============================================
// Product Routes
// ============================================

http.get: /products { }
```

### Inline Comments
For complex logic

```zenolang
fn: calculate_total {
    // Calculate subtotal
    $subtotal: $price * $qty
    
    // Apply discount if applicable
    if: $discount > 0 {
        then: {
            $subtotal: $subtotal - $discount
        }
    }
}
```

## 5. Block Structure

### Route Definitions

```zenolang
http.METHOD: /path {
    summary: "Description"
    auth.middleware: {
        secret: env("JWT_SECRET")
        do: {
            call: function_name
        }
    }
}
```

### Function Definitions

```zenolang
fn: function_name {
    // Get data
    db.select: {
        table: "table_name"
        as: $result
    }
    
    // Process
    if: $condition {
        then: {
            // Action
        }
    }
    
    // Return
    http.json: {
        data: $result
    }
}
```

### Middleware Usage

```zenolang
// Preferred: Explicit block
auth.middleware: {
    secret: env("JWT_SECRET")
    do: {
        call: protected_function
    }
}

// Alternative: Inline (for simple cases)
auth.middleware: {
    secret: env("JWT_SECRET")
    do: {
        http.json: { message: "OK" }
    }
}
```

## 6. Database Operations

### Consistent Formatting

```zenolang
// ✅ Good
db.select: {
    table: "users"
    where: {
        id: $user_id
        active: true
    }
    as: $user
}

// ❌ Bad
db.select:{table:"users" where:{id:$user_id active:true} as:$user}
```

## 7. Conditionals

### If-Then Structure

```zenolang
// ✅ Good
if: $condition {
    then: {
        log: "Condition is true"
    }
}

// With else
if: $condition {
    then: {
        log: "True"
    }
    else: {
        log: "False"
    }
}

// ❌ Bad
if:$condition{then:{log:"True"}}
```

## 8. Loops

### For Each

```zenolang
// ✅ Good
for: $item in $items {
    do: {
        log: $item.name
    }
}

// ❌ Bad
for:$item in $items{do:{log:$item.name}}
```

## 9. File Organization

### Structure

```zenolang
// 1. Includes (if any)
include: "path/to/file.zl"

// 2. Function definitions
fn: helper_function {
    // Implementation
}

// 3. Route definitions
http.get: /endpoint {
    // Implementation
}
```

### Example

```zenolang
// ============================================
// Product Controller
// ============================================

include: "modules/products/helpers.zl"

// ============================================
// Helper Functions
// ============================================

fn: validate_product {
    if: !$name || !$price {
        then: {
            http.bad_request: {
                error: "Validation failed"
            }
            stop
        }
    }
}

// ============================================
// Main Functions
// ============================================

fn: get_products {
    db.select: {
        table: "products"
        as: $products
    }
    
    http.json: {
        data: $products
    }
}
```

## 10. Best Practices

### Use env() for Secrets
```zenolang
// ✅ Good
secret: env("JWT_SECRET")

// ❌ Bad
secret: "hardcoded_secret"
```

### Descriptive Variable Names
```zenolang
// ✅ Good
$user_email: "user@example.com"
$total_price: 1000

// ❌ Bad
$e: "user@example.com"
$t: 1000
```

### Error Handling
```zenolang
// ✅ Good
try: {
    do: {
        db.insert: { }
    }
    catch: {
        http.internal_error: {
            message: "Database error"
        }
    }
}
```

### Consistent Quotes
Use double quotes for strings

```zenolang
// ✅ Good
log: "This is a message"

// ❌ Bad (inconsistent)
log: 'This is a message'
```

## Summary

Following these standards ensures:
- ✅ Readable code
- ✅ Consistent style across projects
- ✅ Easier maintenance
- ✅ Better collaboration

**Remember**: Consistency is key! 🚀
