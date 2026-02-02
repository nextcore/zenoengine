# Zeno-X Framework - Laravel-Like Features

## 🎉 New Features Added

The Zeno-X framework has been enhanced with Laravel-like features to improve developer experience and productivity.

### ✅ Implemented Features

#### 1. **Request Validation System**
Laravel-style validation with fluent API:

```zenolang
call: 'validate' {
    args: [
        $data,
        {
            name: 'required|min:3|max:50'
            email: 'required|email'
            age: 'numeric|min:18'
        },
        {
            'name.required': 'Name is required'
            'email.email': 'Invalid email format'
        }
    ]
    as: $validator
}

if: $validator.fails {
    then: {
        return: response.validationError($validator.errors)
    }
}
```

**Supported Rules:**
- `required` - Field must be present and not empty
- `email` - Valid email format
- `min:X` - Minimum length/value
- `max:X` - Maximum length/value
- `numeric` - Must be a number

#### 2. **Response Helpers**
Convenient response functions for common HTTP responses:

```zenolang
// JSON responses
return: response.json({ message: 'Success' }, 200)
return: response.success('User created', $user)
return: response.error('Not found', 404)

// Specialized responses
return: response.created('Resource created', $resource)
return: response.notFound('User not found')
return: response.unauthorized('Invalid credentials')
return: response.forbidden('Access denied')
return: response.validationError($errors)

// Redirects
return: response.redirect('/dashboard')
return: response.redirect('/login', 'Please login first')
return: response.back('Changes saved')

// File downloads
return: response.download('/path/to/file.pdf', 'document.pdf')
```

#### 3. **Global Helper Functions**
Laravel-like helper functions available globally:

```zenolang
// Environment variables
call: 'env' { args: ['APP_NAME', 'Zeno-X'] as: $appName }

// Debugging
call: 'dd' { args: [$user, $data] }  // Dump and die

// HTTP abort
call: 'abort' { args: [404, 'Not found'] }

// Old input (after validation)
call: 'old' { args: ['email'] as: $oldEmail }

// Asset URLs
call: 'asset' { args: ['css/app.css'] as: $cssUrl }

// CSRF token
call: 'csrfToken' { args: [] as: $token }

// Date/time
call: 'now' { args: [] as: $timestamp }
call: 'today' { args: [] as: $date }

// String helpers
call: 'strRandom' { args: [16] as: $random }
call: 'strSlug' { args: ['Hello World'] as: $slug }  // 'hello-world'

// Value checks
call: 'blank' { args: [$value] as: $isBlank }
call: 'filled' { args: [$value] as: $isFilled }
```

#### 4. **Password Hashing**
Secure password hashing with bcrypt:

```zenolang
// Hash password
call: 'hashMake' {
    args: ['password123']
    as: $hashed
}

// Verify password
call: 'hashCheck' {
    args: ['password123', $hashedPassword]
    as: $isValid
}
```

#### 5. **Authentication System**
Complete authentication with JWT tokens:

```zenolang
// Attempt login
call: 'authAttempt' {
    args: [{ email: $email, password: $password }]
    as: $success
}

// Get authenticated user
call: 'authUser' { args: [] as: $user }

// Check if authenticated
call: 'authCheck' { args: [] as: $isAuth }

// Get user ID
call: 'authId' { args: [] as: $userId }

// Logout
call: 'authLogout' { args: [] }

// Manual login
call: 'authLogin' { args: [$user] as: $token }

// Get token
call: 'authToken' { args: [] as: $token }
```

#### 6. **Authentication Middleware**
Protect routes requiring authentication:

```zenolang
http.get: '/dashboard' {
    middleware: ['Authenticate']
    do: {
        // Only authenticated users can access
        call: 'authUser' { args: [] as: $user }
        return: view('dashboard', { user: $user })
    }
}
```

#### 7. **Authentication Controller**
Pre-built auth endpoints:

- `GET /register` - Registration form
- `POST /register` - Process registration
- `GET /login` - Login form
- `POST /login` - Process login
- `POST /logout` - Logout user
- `GET /profile` - Get user profile (protected)
- `PUT /profile` - Update profile (protected)

#### 8. **Modern Auth Views**
Beautiful, responsive authentication pages included.

#### 9. **Enhanced Controller Stubs**
Generated controllers now include:
- Validation
- Response helpers
- Error handling
- Proper HTTP status codes

---

## 🚀 Quick Start

### Creating a New Project

**Recommended: Cross-Platform Method**

```bash
# Works on Linux, macOS, and Windows
zeno run new-project.zl my-blog
```

This is the **recommended method** - fully portable and works everywhere ZenoEngine runs!

**Alternative: Platform-Specific Scripts**

```bash
# Linux / macOS
./zeno-new my-blog

# Windows
zeno-new.bat my-blog
```

This creates a complete project with:
- ✅ All framework modules pre-installed
- ✅ Authentication system ready
- ✅ .env file auto-generated
- ✅ Directory structure organized
- ✅ README and documentation included

**See [CREATING_NEW_PROJECTS.md](file:///home/max/Documents/PROJ/ZenoEngine%20-%20Public/CREATING_NEW_PROJECTS.md) for complete guide.**

### Working with Existing zeno-x

The framework is automatically loaded via `bootstrap.zl` in your `src/main.zl`:

```zenolang
// Load Framework
include: "bootstrap.zl"

// Load Controllers
include: "app/controllers/AuthController.zl"

// Load Routes
include: "routes/web.zl"
```

### 2. Create a Controller with Validation

```bash
../../zeno run artisan.zl make:crud Product
```

This generates a controller with built-in validation and response helpers.

### 3. Test the Framework

```bash
# Test validation and helpers
../../zeno run test_framework.zl

# Start the server
../../zeno
```

### 4. Test Authentication

```bash
# Register a new user
curl -X POST http://localhost:3000/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"secret123"}'

# Login
curl -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'

# Access protected route
curl -X GET http://localhost:3000/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 📁 New File Structure

```
examples/zeno-x/
├── bootstrap.zl                    # Framework loader
├── app/
│   ├── Support/                    # NEW: Framework modules
│   │   ├── Validator.zl           # Validation system
│   │   ├── Response.zl            # Response helpers
│   │   ├── Helpers.zl             # Global helpers
│   │   ├── Hash.zl                # Password hashing
│   │   └── Auth.zl                # Authentication
│   ├── controllers/
│   │   └── AuthController.zl      # NEW: Auth endpoints
│   └── middleware/
│       └── Authenticate.zl        # NEW: Auth middleware
├── views/
│   └── auth/                       # NEW: Auth views
│       ├── login.blade.zl
│       └── register.blade.zl
└── test_framework.zl              # NEW: Framework tests
```

---

## 📖 Usage Examples

### Example 1: Controller with Validation

```zenolang
// app/controllers/ProductController.zl
include: "app/Support/Validator.zl"
include: "app/Support/Response.zl"

http.post: '/products' {
    do: {
        http.form: { as: $data }
        
        // Validate
        call: 'validate' {
            args: [
                $data,
                {
                    name: 'required|min:3',
                    price: 'required|numeric|min:0',
                    stock: 'numeric'
                },
                {
                    'name.required': 'Product name is required',
                    'price.numeric': 'Price must be a number'
                }
            ]
            as: $validator
        }
        
        if: $validator.fails {
            then: {
                return: response.validationError($validator.errors)
            }
        }
        
        // Save product
        orm.model: 'products'
        orm.save: $data
        orm.last: { as: $product }
        
        return: response.created('Product created', { product: $product })
    }
}
```

### Example 2: Protected Route

```zenolang
// routes/web.zl
include: "app/middleware/Authenticate.zl"

http.get: '/admin/dashboard' {
    middleware: ['Authenticate']
    do: {
        call: 'authUser' { args: [] as: $user }
        
        // Only authenticated users reach here
        return: view('admin/dashboard', { user: $user })
    }
}
```

### Example 3: Custom Validation

```zenolang
http.post: '/contact' {
    do: {
        http.form: { as: $data }
        
        call: 'validate' {
            args: [
                $data,
                {
                    name: 'required|min:3|max:100',
                    email: 'required|email',
                    message: 'required|min:10'
                },
                {
                    'name.required': 'Please provide your name',
                    'name.min': 'Name must be at least 3 characters',
                    'email.email': 'Please provide a valid email address',
                    'message.min': 'Message must be at least 10 characters'
                }
            ]
            as: $validator
        }
        
        if: $validator.fails {
            then: {
                return: response.validationError($validator.errors)
            }
        }
        
        // Process contact form...
        return: response.success('Message sent successfully')
    }
}
```

---

## 🎯 Next Steps

### Planned Features (Future Phases)

- **Advanced Routing**: Route groups, named routes, resource routes
- **ORM Enhancements**: Soft deletes, scopes, accessors/mutators
- **File Upload**: File validation and storage
- **Session Management**: Flash messages, session data
- **More Artisan Commands**: `migrate:rollback`, `route:list`, etc.
- **Error Pages**: Custom 404, 500 pages
- **Example Blog App**: Complete showcase application

---

## 🐛 Testing

Run the framework tests:

```bash
cd examples/zeno-x
../../zeno run test_framework.zl
```

Expected output:
- ✅ Validation tests pass
- ✅ Helper functions work
- ✅ Password hashing works
- ✅ All features load without errors

---

## 📝 Migration from Old Code

### Before (Old Style)
```zenolang
http.post: '/products' {
    do: {
        http.form: { as: $data }
        
        // Manual validation
        if: $data.name == nil {
            then: {
                return: { error: 'Name required' }
            }
        }
        
        orm.model: 'products'
        orm.save: $data
        
        return: { status: 'success' }
    }
}
```

### After (New Style)
```zenolang
http.post: '/products' {
    do: {
        http.form: { as: $data }
        
        // Automatic validation
        call: 'validate' {
            args: [$data, { name: 'required|min:3' }, {}]
            as: $validator
        }
        
        if: $validator.fails {
            then: { return: response.validationError($validator.errors) }
        }
        
        orm.model: 'products'
        orm.save: $data
        orm.last: { as: $product }
        
        return: response.created('Product created', { product: $product })
    }
}
```

---

## 🤝 Contributing

To add new features to the framework:

1. Create module in `app/Support/`
2. Add to `bootstrap.zl`
3. Update documentation
4. Add tests in `test_framework.zl`

---

## 📄 License

Same as ZenoEngine - Public

---

**Happy coding with Zeno-X! 🚀**
