# Migrating from ASP.NET Core Identity

If you are migrating an existing C# ASP.NET Core application to ZenoEngine, you can reuse your existing database (containing user credentials hashed with ASP.NET Core Identity) without forcing your users to reset their passwords.

ZenoEngine provides a native high-performance module slot called `auth.aspnet_login` that is fully compatible with the ASP.NET Core Identity database schema and password hashing standard (Identity V3 PBKDF2).

---

## The ASP.NET Core Identity Schema

By default, ASP.NET Core Identity stores user data in the `AspNetUsers` table. The `auth.aspnet_login` slot expects this table to contain at least the following columns:

| Column Name | Type | Description |
| :--- | :--- | :--- |
| **`Id`** | `TEXT` / `VARCHAR` | Unique identifier (usually a GUID/UUID string or integer) |
| **`UserName`** | `TEXT` / `VARCHAR` | The display username of the user |
| **`NormalizedUserName`** | `TEXT` / `VARCHAR` | Uppercased/normalized username for case-insensitive lookup |
| **`Email`** | `TEXT` / `VARCHAR` | The email address of the user |
| **`NormalizedEmail`** | `TEXT` / `VARCHAR` | Uppercased/normalized email address for case-insensitive lookup |
| **`PasswordHash`** | `TEXT` / `VARCHAR` | The base64-encoded Identity V3 hash payload |

---

## Using `auth.aspnet_login` in ZenoLang

The `auth.aspnet_login` slot performs the lookup on `AspNetUsers` using a case-insensitive check against `NormalizedUserName` and `NormalizedEmail` to match either username or email. It then validates the password using the ASP.NET Identity V3 format (PBKDF2 with HMAC-SHA256) and issues a JWT token.

### Syntax Reference

```zl
auth.aspnet_login
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
  auth.aspnet_login
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

## Security Verification Details

Under the hood, the password hash verification runs compiled Go native code doing:
1. Decoding the Base64 password hash string.
2. Checking the version byte (`0x01` indicates ASP.NET Core Identity V3).
3. Extracting the Iteration Count (usually `10000` or `100000`), Salt, and Subkey bytes.
4. Re-hashing the input password with `pbkdf2` using HMAC-SHA256.
5. Performing a constant-time comparison (`subtle.ConstantTimeCompare`) of the resulting subkey against the stored one to mitigate timing attacks.
