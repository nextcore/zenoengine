# Static Asset & SPA Hosting

Instead of using Nginx to serve your Vue, React, or Svelte applications, you can use ZenoEngine's `http.static` slot. This slot securely serves files from a given directory and protects against path traversal attacks.

If you are hosting a Single Page Application (SPA) where the frontend router needs to take over, add the `spa: true` flag. This ensures that any 404s will automatically return `index.html` instead.

```zeno
// Host a React/Svelte App from the /dist folder
do: {
    // Serve everything under / prefix from the ./frontend/dist directory
    http.static: "./frontend/dist" {
        path: "/"
        spa: true
    }
}

// Host a regular folder of images
http.static: "./storage/images" {
    path: "/images/"
}
```
