# Sessions

ZenoEngine provides a simple, cookie-based session management system. While JWT is recommended for stateless APIs, traditional web applications often rely on sessions for maintaining state across requests.

## 1. Setting Session Data

Use `session.set` to store data in the session. Currently, ZenoEngine uses encrypted/encoded cookies to persist this data.

```zeno
// Store user ID after login
session.set: "user_id" { val: $user.id }

// Store a complex object
session.set: "preferences" { val: { theme: "dark", lang: "en" } }
```

## 2. Retrieving Session Data

Use `session.get` to retrieve data from the session.

```zeno
session.get: "user_id" { as: $userId }

if: $userId != null {
    then: {
        dump: "User is logged in with ID: " + $userId
    }
}
```

## 3. Flash Data (One-Time Messages)

Flash data is special session data that is deleted automatically after being read once. This is perfect for "success" or "error" messages after form submissions.

### Setting Flash Data
```zeno
session.flash: "success" { val: "Profile updated successfully!" }
```

### Retrieving Flash Data
```zeno
session.get_flash: "success" { as: $message }

if: $message != null {
    then: {
        // This will only be displayed once
        http.view: 'profile' { message: $message }
    }
}
```

## 4. Destroying Sessions

To log out a user or clear all session data, use `session.destroy`.

```zeno
http.post: '/logout' {
    do: {
        session.destroy: true
        http.redirect: '/login'
    }
}
```

## 5. Security: Regenerate

For security reasons (preventing session fixation), it's good practice to regenerate the session ID after login.

```zeno
session.set: "user_id" { val: $user.id }
session.regenerate: true
```

> [!NOTE]
> Currently, `session.regenerate` is a placeholder that clears the path for future server-side session store implementations. In the current cookie-based implementation, it ensures the session is fresh.
