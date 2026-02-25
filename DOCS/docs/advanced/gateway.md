# Web Server Gateway & Virtual Hosts

ZenoEngine is designed not just as an application framework but as a **full Web Server Gateway**. You don't necessarily need to put it behind Nginx or Caddy. ZenoEngine provides built-in slots to act as a reverse proxy, host static single-page applications (SPAs), and manage multi-tenant domains.

## Virtual Hosting

You can route traffic based on the incoming domain name using the `http.host` slot. This is the cornerstone of ZenoEngine's "Multi-App Architecture", allowing one single Go binary to host dozens of distinct applications.

```zeno
// main.zl

// App 1: API Server
http.host: "api.mycompany.com" {
    do: {
        http.get: '/v1/users' { ... }
    }
}

// App 2: Landing Page
http.host: "www.mycompany.com" {
    do: {
        http.get: '/' { ... }
    }
}
```

## Static Asset & SPA Hosting

Instead of using Nginx to serve your Vue, React, or Svelte applications, you can use ZenoEngine's `http.static` slot. This slot securely serves files from a given directory and protects against path traversal attacks.

If you are hosting a Single Page Application (SPA) where the frontend router needs to take over, add the `spa: true` flag. This ensures that any 404s will automatically return `index.html` instead.

```zeno
// Host a React App from the /dist folder
http.host: "app.mydomain.com" {
    do: {
        // Serve everything under / prefix from the ./frontend/dist directory
        http.static: "./frontend/dist" {
            path: "/"
            spa: true
        }
    }
}

// Host a regular folder of images
http.static: "./storage/images" {
    path: "/images/"
}
```

## Reverse Proxying (Caddy-style)

Sometimes you have a legacy API in Node.js, Python, or even PHP that you want to host on the same domain as your new ZenoEngine built features. You can seamlessly proxy traffic to those services using `http.proxy`.

```zeno
http.host: "api.mydomain.com" {
    do: {
        // ZenoEngine handles these new endpoints directly
        http.get: '/v2/fast-search' { ... }
        
        // EVERYTHING else gets securely proxied to a legacy Node.js app
        http.proxy: "http://localhost:4000" {
            path: "/"
        }
        
        // Proxy a specific path to another internal service
        http.proxy: "http://10.0.0.5:8080" {
            path: "/ai-models/"
        }
    }
}
```

The `http.proxy` automatically preserves HTTP headers, handles streaming, and propagates the correct `X-Forwarded-*` headers, just like Nginx or Caddy.
