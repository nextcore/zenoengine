# Migrating from ASP.NET Core Identity

If you are migrating an existing C# ASP.NET Core application to ZenoEngine, you can reuse your existing database (containing user credentials hashed with ASP.NET Core Identity) without forcing your users to reset their passwords.

ZenoEngine provides a dedicated native high-performance slot group `aspnet.*` that is fully compatible with the ASP.NET Core Identity database schema and password hashing standard (Identity V3 PBKDF2).

---

## The ASP.NET Core Identity Schema

By default, ASP.NET Core Identity stores user data in the `AspNetUsers` table. The `aspnet.login` slot expects this table to contain at least the following columns:

| Column Name | Type | Description |
| :--- | :--- | :--- |
| **`Id`** | `TEXT` / `VARCHAR` | Unique identifier (usually a GUID/UUID string or integer) |
| **`UserName`** | `TEXT` / `VARCHAR` | The display username of the user |
| **`NormalizedUserName`** | `TEXT` / `VARCHAR` | Uppercased/normalized username for case-insensitive lookup |
| **`Email`** | `TEXT` / `VARCHAR` | The email address of the user |
| **`NormalizedEmail`** | `TEXT` / `VARCHAR` | Uppercased/normalized email address for case-insensitive lookup |
| **`PasswordHash`** | `TEXT` / `VARCHAR` | The base64-encoded Identity V3 hash payload |

---

## Using `aspnet.login` in ZenoLang

The `aspnet.login` slot performs the lookup on `AspNetUsers` using a case-insensitive check against `NormalizedUserName` and `NormalizedEmail` to match either username or email. It then validates the password using the ASP.NET Identity V3 format (PBKDF2 with HMAC-SHA256) and issues a JWT token.

### Syntax Reference

```zl
aspnet.login
  username: $input_username_or_email
  password: $input_password
  [fields: ['TenantId', 'FullName']]
  [db: 'default']
  [secret: env("JWT_SECRET")]
  [expires_in: 86400]
  [as: $token]
  [user_as: $user]
```

### Parameter Reference

* **`username`** (string, **Required**): The username or email input from the login request.
* **`password`** (string, **Required**): The plain-text password provided by the user.
* **`fields`** (list of strings, Optional): Custom database columns to retrieve from the `AspNetUsers` table (e.g. `['TenantId', 'FullName']`).
* **`db`** (string, Optional): The name of the database connection to run the query against. Defaults to `'default'`.
* **`secret`** (string, Optional): The JWT secret key used to sign the token. Defaults to the environment variable `JWT_SECRET`.
* **`expires_in`** (int, Optional): Token expiration duration in seconds. Defaults to `86400` (24 hours).
* **`as`** (string, Optional): Variable name to store the generated JWT token string. Defaults to `token` (resolves to `$token`).
* **`user_as`** (string, Optional): Variable name to store the user profile data map. Defaults to `user` (resolves to `$user` with keys: `id`, `username`, `email` plus any custom fields specified in `fields`).

---

## Example Login Implementation

Here is a complete ZenoLang route handler implementing the legacy migration login flow:

```zl
http.post: '/api/auth/login' {
  # 1. Validate inputs
  validate {
    username: 'required'
    password: 'required'
  }

  # 2. Authenticate using C# AspNetUsers schema with a custom 'TenantId' column
  aspnet.login
    username: $request.username
    password: $request.password
    fields: ['TenantId']
    as: $jwt_token
    user_as: $current_user

  # 3. Return JSON response with JWT token
  response.json {
    success: true
    token: $jwt_token
    user: $current_user
    tenant: $current_user.TenantId
  }
}
```

---

## Creating New Users (Password Hashing)

If you need to create/register new users within ZenoEngine and want to save their passwords in the compatible ASP.NET Identity V3 format (so legacy C# microservices or applications can still verify them), use the `aspnet.hash` slot:

```zl
aspnet.hash: $plain_password {
  [iterations: 10000]
  as: $db_hash
}
```

This slot generates a cryptographically secure 16-byte random salt, performs PBKDF2 hashing with HMAC-SHA256, and formats the output as a Base64 binary blob fully compatible with Microsoft Identity.

---

## Verifying Password Manually

If you need to verify a legacy hash without generating a JWT token or loading a database row, use the `aspnet.verify` slot:

```zl
aspnet.verify
  hash: $db_hash
  password: $plain_password
  as: $is_valid
```

---

## Validating Password Policies

To validate passwords against ASP.NET Core Identity password policies before creating or updating users, use the `aspnet.validate_password` slot:

```zl
aspnet.validate_password: $plain_password {
  [required_length: 6]
  [require_digit: true]
  [require_lowercase: true]
  [require_uppercase: true]
  [require_non_alphanumeric: true]
  [required_unique_chars: 1]
  [as: $is_valid]
  [errors_as: $password_errors]
}
```

This slot returns a boolean validation status (`$is_valid`) and a list of detailed error messages (`$password_errors`) that match Microsoft Identity's default error messages (e.g., "Passwords must have at least one uppercase ('A'-'Z').").

### Example Registration Implementation

Here is how you can use the password validation and hashing slots together in a ZenoLang route handler when registering a new user:

```zl
http.post: '/api/auth/register' {
  # 1. Basic validation (required, confirmation check, email format/uniqueness)
  validate {
    rules: {
      username: 'required|unique:AspNetUsers,UserName'
      email: 'required|email|unique:AspNetUsers,Email'
      password: 'required|confirmed'
    }
    as: $validation_errors
  }

  if $validation_errors_any {
    response.json {
      success: false
      errors: $validation_errors
    }
  }

  # 2. Advanced password policy validation matching ASP.NET Identity rules
  aspnet.validate_password: $request.password {
    required_length: 8
    require_uppercase: true
    require_digit: true
    require_non_alphanumeric: true
    as: $pw_ok
    errors_as: $pw_errors
  }

  if $pw_ok == false {
    response.json {
      success: false
      errors: {
        password: $pw_errors
      }
    }
  }

  # 3. Hash the validated password in compatible format
  aspnet.hash: $request.password {
    as: $hashed_password
  }

  # 4. Insert into the AspNetUsers table
  db.query: 'AspNetUsers' {
    insert {
      Id: uuid.v4
      UserName: $request.username
      NormalizedUserName: upper($request.username)
      Email: $request.email
      NormalizedEmail: upper($request.email)
      PasswordHash: $hashed_password
    }
  }

  response.json {
    success: true
    message: 'User registered successfully!'
  }
}
```

---

## Security Verification Details

Under the hood, the password hash verification runs compiled Go native code doing:
1. Decoding the Base64 password hash string.
2. Checking the version byte (`0x01` indicates ASP.NET Core Identity V3).
3. Extracting the Iteration Count (usually `10000` or `100000`), Salt, and Subkey bytes.
4. Re-hashing the input password with `pbkdf2` using HMAC-SHA256.
5. Performing a constant-time comparison (`subtle.ConstantTimeCompare`) of the resulting subkey against the stored one to mitigate timing attacks.
