# Creating New Zeno-X Projects

## 🚀 Quick Start

The easiest way to create a new Zeno-X project:

```bash
# Simply pass the project name as an argument
zeno run new-project.zl my-blog
```

This command will:
- Create a new directory `my-blog/`
- Copy the Zeno-X template
- Generate secure random keys for `.env`
- Create a custom README
- Set up the project structure

**Alternative methods** (if you prefer):

```bash
# Using environment variable
PROJECT_NAME=my-blog zeno run new-project.zl

# Interactive mode (will prompt for name)
zeno run new-project.zl
```

This is the **recommended method** as it's fully portable and works on Linux, macOS, and Windows.

### Platform-Specific Scripts

Alternatively, you can use platform-specific scripts:

#### Linux / macOS

```bash
# Make the script executable (first time only)
chmod +x zeno-new

# Create a new project
./zeno-new my-blog
```

### Windows

```bash
# Create a new project
zeno-new.bat my-blog
```

## 📋 What Gets Created

When you run `zeno-new project-name`, it creates:

```
my-blog/
├── app/
│   ├── controllers/
│   │   └── AuthController.zl       # Pre-built authentication
│   ├── models/
│   │   └── Product.zl              # Example model
│   ├── middleware/
│   │   ├── Authenticate.zl         # Auth middleware
│   │   └── AdminCheck.zl           # Example middleware
│   └── Support/                    # Framework modules
│       ├── Validator.zl            # Validation system
│       ├── Response.zl             # Response helpers
│       ├── Helpers.zl              # Global helpers
│       ├── Hash.zl                 # Password hashing
│       └── Auth.zl                 # Authentication
├── database/
│   ├── migrations/
│   │   └── create_users_table.zl   # User migration
│   └── zeno.sqlite                 # SQLite database
├── routes/
│   └── web.zl                      # Route definitions
├── views/
│   ├── auth/
│   │   ├── login.blade.zl          # Login page
│   │   └── register.blade.zl       # Register page
│   ├── welcome.blade.zl            # Home page
│   └── stubs/                      # Code generation templates
├── public/
│   └── assets/                     # CSS, JS, images
├── storage/
│   ├── logs/                       # Application logs
│   └── uploads/                    # File uploads
├── bootstrap.zl                    # Framework loader
├── artisan.zl                      # CLI tool
├── src/
│   └── main.zl                     # Entry point
├── .env                            # Configuration (auto-generated)
├── .gitignore                      # Git ignore rules
├── README.md                       # Project README
└── LARAVEL_FEATURES.md            # Framework documentation
```

## 🔧 After Creating a Project

### 1. Navigate to Project

```bash
cd my-blog
```

### 2. Configure Environment

Edit `.env` file:

```bash
# Linux/Mac
nano .env

# Windows
notepad .env
```

**Important**: Change the security keys in production!

```env
APP_KEY=base64:YOUR_RANDOM_KEY_HERE
HASH_KEY=base64:YOUR_RANDOM_KEY_HERE
BLOCK_KEY=base64:YOUR_RANDOM_KEY_HERE
JWT_SECRET=base64:YOUR_RANDOM_KEY_HERE
```

### 3. Start the Server

```bash
# Linux/Mac
../../zeno

# Windows
..\..\zeno.exe
```

### 4. Visit Your App

Open browser: `http://localhost:3000`

## 🛠️ Common Tasks

### Create Your First Resource

```bash
# Linux/Mac
../../zeno run artisan.zl make:crud Product

# Windows
..\..\zeno.exe run artisan.zl make:crud Product
```

This creates:
- `app/models/Product.zl`
- `app/controllers/ProductController.zl`
- `database/migrations/create_Products_table.zl`

### Create a Controller

```bash
../../zeno run artisan.zl make:controller BlogController
```

### Create a Model

```bash
../../zeno run artisan.zl make:model Post
```

### Create a Migration

```bash
../../zeno run artisan.zl make:migration create_posts_table
```

### Create Middleware

```bash
../../zeno run artisan.zl make:middleware CheckAge
```

## 📚 Pre-Installed Features

Every new Zeno-X project comes with:

✅ **Request Validation**
```zenolang
call: 'validate' {
    args: [$data, { email: 'required|email' }, {}]
    as: $validator
}
```

✅ **Response Helpers**
```zenolang
return: response.success('User created', $user)
return: response.error('Not found', 404)
```

✅ **Authentication**
```zenolang
call: 'authAttempt' { args: [$credentials] as: $success }
call: 'authUser' { args: [] as: $user }
```

✅ **Password Hashing**
```zenolang
call: 'hashMake' { args: ['password'] as: $hashed }
call: 'hashCheck' { args: ['password', $hashed] as: $valid }
```

✅ **Global Helpers**
```zenolang
call: 'env' { args: ['APP_NAME'] as: $name }
call: 'dd' { args: [$data] }  // Dump and die
call: 'abort' { args: [404] }
```

## 🎯 Example: Building a Blog

### 1. Create the Blog Resource

```bash
../../zeno run artisan.zl make:crud Post
```

### 2. Edit the Migration

Edit `database/migrations/create_Posts_table.zl`:

```zenolang
schema.create: 'posts' {
    column.id: 'id'
    column.string: 'title'
    column.text: 'content'
    column.integer: 'user_id'
    column.timestamps
}
```

### 3. Add Validation to Controller

Edit `app/controllers/PostController.zl`:

```zenolang
http.post: '/posts' {
    do: {
        http.form: { as: $data }
        
        call: 'validate' {
            args: [
                $data,
                {
                    title: 'required|min:5',
                    content: 'required|min:10'
                },
                {}
            ]
            as: $validator
        }
        
        if: $validator.fails {
            then: {
                return: response.validationError($validator.errors)
            }
        }
        
        orm.model: 'posts'
        orm.save: $data
        orm.last: { as: $post }
        
        return: response.created('Post created', { post: $post })
    }
}
```

### 4. Protect Routes

```zenolang
http.post: '/posts' {
    middleware: ['Authenticate']
    do: {
        // Only authenticated users can create posts
    }
}
```

## 🔒 Security Best Practices

### 1. Change Default Keys

**Never use default keys in production!**

Generate secure keys:

```bash
# Linux/Mac
openssl rand -base64 32

# Windows (PowerShell)
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
```

### 2. Set Environment

In `.env`:
```env
APP_ENV=production
APP_DEBUG=false
```

### 3. Secure Database

Change database location and permissions:
```env
DB_DATABASE=/secure/path/to/database.sqlite
```

## 📖 Next Steps

1. **Read Documentation**: Check `LARAVEL_FEATURES.md` in your project
2. **Explore Examples**: Look at `AuthController.zl` for patterns
3. **Build Features**: Use Artisan to scaffold resources
4. **Customize**: Modify views, add middleware, create models

## 🆘 Troubleshooting

### Project Creation Fails

**Problem**: Template not found

**Solution**: Make sure you're running `zeno-new` from the ZenoEngine root directory where `examples/zeno-x` exists.

### Permission Denied (Linux/Mac)

**Problem**: Cannot execute `zeno-new`

**Solution**: 
```bash
chmod +x zeno-new
```

### Server Won't Start

**Problem**: Port already in use

**Solution**: Change port in your code or stop other services using port 3000.

## 💡 Tips

1. **Use Version Control**: Initialize git in your project
   ```bash
   cd my-blog
   git init
   git add .
   git commit -m "Initial commit"
   ```

2. **Keep Template Updated**: When zeno-x gets updates, new projects will include them

3. **Customize Stubs**: Edit `views/stubs/` to change code generation templates

4. **Environment Variables**: Use `.env` for all configuration, never hardcode

## 🎓 Learning Resources

- **Framework Documentation**: `LARAVEL_FEATURES.md`
- **Example Code**: `app/controllers/AuthController.zl`
- **Test Suite**: `test_framework.zl` (in template)
- **ZenoLang Docs**: Check main ZenoEngine documentation

---

**Happy Building! 🚀**
