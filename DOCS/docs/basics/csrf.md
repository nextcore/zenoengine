# CSRF Protection

## Introduction

Cross-site request forgeries are a type of malicious exploit whereby unauthorized commands are performed on behalf of an authenticated user. ZenoEngine makes it easy to protect your application from CSRF attacks.

ZenoEngine automatically generates a CSRF "token" for each active user session. This token is used to verify that the authenticated user is the person actually making the requests.

## Preventing CSRF Requests

ZenoEngine's CSRF middleware verifies the token in the form submission. To include the token in your forms, use the `@csrf` Blade directive:

```blade
<form method="POST" action="/profile">
    @csrf
    ...
</form>
```

This generates a hidden input field:

```html
<input type="hidden" name="_token" value="xGF...">
```

## Excluding URIs

If you are building a pure API and handling authentication via JWT tokens, you may want to exclude certain routes from CSRF verification. API routes (`/api/...`) are excluded from CSRF protection by default.
