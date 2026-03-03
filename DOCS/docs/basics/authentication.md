# Authentication (JWT)

ZenoEngine is designed from the ground up to support modern, stateless architectures. Its primary and native method for authentication is **JSON Web Tokens (JWT)**. 

By avoiding file-based or database-reliant sessions, ZenoEngine APIs can scale infinitely and serve multi-platform clients (Web, iOS, Android) seamlessly using Bearer tokens.

## 1. Environment Configuration

Before issuing tokens, ensure your JWT Secret is configured in the `.env` file. This signature key is crucial to prevent token tampering.

```env
# A secure, random 32+ character string
JWT_SECRET="super_secret_zeno_jwt_key_that_is_very_long"
```

## 2. Generating Tokens (Login)

When a user submits their credentials (e.g., email and password), you verify them using ZenoEngine's built-in `hash` slots. If they are correct, you issue a JWT using `jwt.sign`.

```zeno
// src/main.zl

http.post: '/api/login' {
    do: {
        http.json: { as: $credentials }
        
        // 1. Find the user
        db.table: "users"
        db.where: "email" { equals: $credentials.email }
        db.first: { as: $user }
        
        if: $user == null {
            http.response: {
                status: 401
                json: { error: "Invalid credentials" }
            }
            return
        }
        
        // 2. Verify Bcrypt hash
        hash.verify: {
            text: $credentials.password
            hash: $user.password
            as: $isValid
        }
        
        if: $isValid == false {
            http.response: {
                status: 401
                json: { error: "Invalid credentials" }
            }
            return
        }
        
        // 3. Generate JWT Token
        // Define your custom payload (claims). Never put sensitive data like passwords here.
        var: $claims {
            val: {
                sub: $user.id,
                email: $user.email,
                role: "user"
            }
        }
        
        // Issue token expiring in 24 hours (86400 seconds)
        jwt.sign: {
            payload: $claims
            expires_in: 86400
            as: $token
        }
        
        // 4. Return to Client
        http.response: {
            json: {
                message: "Login successful",
                token: $token,
                user: {
                    id: $user.id,
                    name: $user.name
                }
            }
        }
    }
}
```

## 3. Protecting Routes (Middleware)

To secure routes, create an authentication middleware that reads the `Authorization: Bearer <token>` header, verifies the signature, and injects the payload into the request scope.

```zeno
// src/middleware/jwt_auth.zl

http.middleware: 'auth.jwt' {
    do: {
        // 1. Extract Bearer token from header
        http.header: "Authorization" { as: $authHeader }
        
        if: $authHeader == null {
            http.response: { status: 401, json: { error: "Missing token" } }
            return
        }
        
        // (Assuming you strip the "Bearer " prefix logic here)
        var: $token { val: str.replace($authHeader, "Bearer ", "") }
        
        // 2. Verify the Token
        jwt.verify: {
            token: $token
            as: $payload
            error: $jwtError
        }
        
        if: $jwtError != null {
            http.response: { status: 401, json: { error: "Invalid or expired token" } }
            return
        }
        
        // 3. Store the verified payload globally for downstream slots
        var: $authUser { val: $payload }
        
        http.next: true
    }
}
```

Apply the middleware to your private endpoints:

```zeno
// src/main.zl
include: 'src/middleware/jwt_auth.zl'

http.get: '/api/me' {
    middleware: ['auth.jwt']
    do: {
        // $authUser is securely injected by our middleware
        http.response: {
            json: $authUser
        }
    }
}
```

## 4. Refreshing Tokens

Tokens eventually expire. Instead of forcing the user to log in again, you can use `jwt.refresh` if the client holds a valid payload and you want to extend its life.

```zeno
http.post: '/api/refresh' {
    middleware: ['auth.jwt'] // Client must provide their current valid (or recently expired) token
    do: {
        jwt.refresh: {
            token: $token
            expires_in: 86400 // another 24 hours
            as: $newToken
            error: $refreshErr
        }
        
        if: $refreshErr != null {
            http.response: { status: 401, json: { error: "Cannot refresh token" } }
            return
        }
        
        http.response: {
            json: { token: $newToken }
        }
    }
}
```

By relying heavily on JWT, your ZenoEngine backend becomes exceptionally fast and globally scalable out of the box.
