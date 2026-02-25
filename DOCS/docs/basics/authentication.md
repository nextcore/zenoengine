# Authentication

Building a secure login and registration system is one of the most common tasks when developing web applications. ZenoEngine provides built-in mechanisms for password hashing, session management, and middleware to make building authentication straightforward.

In this guide, we'll walk through building a complete, basic Authentication flow from scratch.

## 1. Database Setup

First, you need a place to store your users. Create a table using the Query Builder `db.query` or the Schema Builder (if available). The minimum fields required are an identifier (like `email`) and a robust `password` field.

```zeno
// In your database setup script or migration
db.query: "
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
"
```

## 2. The Login View (ZenoBlade)

Let's create the HTML form for users to enter their credentials. Create a file at `resources/views/auth/login.blade.zl`.

Notice how we include a <code v-pre>{{ csrf_field() }}</code>. ZenoEngine requires all `POST` requests to have a CSRF token for security. We also use the `$errors` variable to display any validation or login failures.

<div v-pre>

```html
<!-- resources/views/auth/login.blade.zl -->
<!DOCTYPE html>
<html>
<head>
    <title>Login</title>
</head>
<body>
    <main>
        <h2>Sign In</h2>

        <!-- Display authentication errors -->
        @if(isset($errors['auth']))
            <div style="color: red; padding: 10px;">
                {{ $errors['auth'] }}
            </div>
        @endif

        <form method="POST" action="/login">
            {{ csrf_field() }}

            <div>
                <label>Email</label>
                <input type="email" name="email" required>
            </div>

            <div>
                <label>Password</label>
                <input type="password" name="password" required>
            </div>

            <button type="submit">Log In</button>
        </form>
    </main>
</body>
</html>
```
</div>

## 3. Handling the Routes (ZenoLang)

Now, we need two routes: one `GET` route to display the form, and a `POST` route to process the submission.

```zeno
// src/main.zl

// 1. Show the Login Form
http.get: '/login' {
    do: {
        http.view: 'auth/login'
    }
}

// 2. Process the Login
http.post: '/login' {
    do: {
        // Read the submitted form data
        http.form: { as: $credentials }
        
        // Find the user in the database
        db.table: "users"
        db.where: "email" { equals: $credentials.email }
        db.first: { as: $user }
        
        // If user doesn't exist, redirect back with an error
        if: $user == null {
            then: {
                http.redirect: '/login' {
                    flash: { error: "Invalid credentials." }
                }
                return
            }
        }
        
        // Verify the password using ZenoEngine's built-in Bcrypt hasher
        hash.verify: {
            text: $credentials.password
            hash: $user.password
            as: $isValid
        }
        
        if: $isValid == false {
            then: {
                http.redirect: '/login' {
                    flash: { error: "Invalid credentials." }
                }
                return
            }
        }
        
        // Password is correct! Log the user in by saving their ID to the session.
        session.set: "user_id" { val: $user.id }
        
        // Regenerate the session ID to prevent Session Fixation attacks
        session.regenerate: true
        
        // Redirect to the dashboard
        http.redirect: '/dashboard'
    }
}
```

## 4. Protecting Routes with Middleware

Now that users can log in, we need to protect certain pages (like the `/dashboard`) so that only authenticated users can access them. We achieve this using **Middleware**.

First, define the `auth` middleware logic. If the session does not contain a `user_id`, we redirect them back to the login page.

```zeno
// src/middleware/auth.zl
http.middleware: 'auth' {
    do: {
        session.get: "user_id" { as: $userId }
        
        // If no user_id is found in the session, deny access!
        if: $userId == null {
            then: {
                http.redirect: '/login'
                return
            }
        }
        
        // Optional: Load the full user object and attach it to the request
        db.table: "users"
        db.where: "id" { equals: $userId }
        db.first: { as: $authenticatedUser }
        
        // Store it in the request scope so downstream controllers can use it
        var: $currentUser { val: $authenticatedUser }
        
        // Continue to the intended route
        http.next: true
    }
}
```

Then, apply this middleware to your protected routes:

```zeno
// src/main.zl
include: 'src/middleware/auth.zl'

http.get: '/dashboard' {
    middleware: ['auth']
    do: {
        // Because the 'auth' middleware ran first, we know $currentUser exists here!
        http.view: 'dashboard' {
            user: $currentUser
        }
    }
}
```

## 5. Logging Out

To log a user out, you clear their session data using `session.destroy` or `session.delete`.

```zeno
http.post: '/logout' {
    middleware: ['auth']
    do: {
        // Clear all session data
        session.destroy: true
        
        // Redirect back to the homepage
        http.redirect: '/'
    }
}
```

## Summary

You now have a fully functioning, secure authentication system! 
- You safely store passwords using `hash.make` (when registering) and `hash.verify` (when logging in).
- You prevent Cross-Site Request Forgery (CSRF) on your forms using <code v-pre>{{ csrf_field() }}</code>.
- You protect against Session Fixation using `session.regenerate`.
- You secure private routes using `http.middleware`.
