# Inertia.js Support for ZenoEngine

## Overview

Inertia.js is now supported in ZenoEngine! Build modern single-page applications (SPAs) using server-side routing with your favorite frontend framework (React, Vue, or Svelte).

## Features

- ✅ Server-side routing with ZenoLang
- ✅ Automatic JSON/HTML response handling
- ✅ Shared data (auth, flash, errors)
- ✅ Compatible with existing auth system
- ✅ Multi-tenant ready
- ✅ Coexists with Blade templates

## Quick Start

### 1. Backend (ZenoLang)

**routes.zl**:
```zenolang
http.get: /dashboard {
    auth.middleware: {
        secret: env("JWT_SECRET")
        do: {
            // Get data
            db.select: {
                table: "products"
                as: $products
            }
            
            // Render with Inertia
            inertia.render: {
                component: "Dashboard"
                props: {
                    products: $products
                    user: $auth
                }
            }
        }
    }
}
```

### 2. Frontend (React + Vite)

**Setup**:
```bash
cd public/inertia
npm create vite@latest . -- --template react
npm install @inertiajs/react
npm install
npm run build
```

**app.jsx**:
```jsx
import { createInertiaApp } from '@inertiajs/react'
import { createRoot } from 'react-dom/client'

createInertiaApp({
  resolve: name => {
    const pages = import.meta.glob('./Pages/**/*.jsx', { eager: true })
    return pages[`./Pages/${name}.jsx`]
  },
  setup({ el, App, props }) {
    createRoot(el).render(<App {...props} />)
  },
})
```

**Pages/Dashboard.jsx**:
```jsx
import { Head, Link } from '@inertiajs/react'

export default function Dashboard({ products, user }) {
  return (
    <>
      <Head title="Dashboard" />
      <h1>Welcome, {user.email}</h1>
      
      <div>
        {products.map(product => (
          <div key={product.id}>
            {product.name} - ${product.price}
          </div>
        ))}
      </div>
      
      <Link href="/products">View All Products</Link>
    </>
  )
}
```

## Available Slots

### `inertia.render`

Render an Inertia component with props.

```zenolang
inertia.render: {
    component: "Dashboard"
    props: {
        user: $auth
        data: $data
    }
}
```

**Parameters**:
- `component` (required): Component name (e.g., "Dashboard", "Products/Index")
- `props` (optional): Data to pass to component

**Behavior**:
- Initial request: Returns HTML with embedded page data
- Subsequent requests (X-Inertia: true): Returns JSON

### `inertia.share`

Share data across all Inertia requests.

```zenolang
inertia.share: {
    auth: $auth
    flash: $flash
    errors: $errors
}
```

### `inertia.location`

Force a full page reload to a URL.

```zenolang
inertia.location: {
    url: "/login"
}
```

## Complete Example

### Backend

**routes.zl**:
```zenolang
include: "src/tutorial/inertia_demo/controller.zl"

// Products List
http.get: /products {
    auth.middleware: {
        secret: env("JWT_SECRET")
        do: {
            call: get_products
        }
    }
}

// Product Create
http.post: /products {
    auth.middleware: {
        secret: env("JWT_SECRET")
        do: {
            call: create_product
        }
    }
}
```

**controller.zl**:
```zenolang
fn: get_products {
    db.select: {
        table: "products"
        as: $products
    }
    
    inertia.render: {
        component: "Products/Index"
        props: {
            products: $products
        }
    }
}

fn: create_product {
    http.form: { as: $form }
    
    // Validate
    if: !$form.name {
        then: {
            http.redirect: {
                url: "/products/create"
                with: {
                    errors: { name: "Name is required" }
                }
            }
            stop
        }
    }
    
    // Insert
    db.insert: {
        table: "products"
        data: $form
        as: $product
    }
    
    // Redirect with success
    http.redirect: {
        url: "/products"
        with: {
            flash: { success: "Product created!" }
        }
    }
}
```

### Frontend

**Pages/Products/Index.jsx**:
```jsx
import { Head, Link, router } from '@inertiajs/react'

export default function Index({ products, flash }) {
  const deleteProduct = (id) => {
    if (confirm('Are you sure?')) {
      router.delete(`/products/${id}`)
    }
  }

  return (
    <>
      <Head title="Products" />
      
      {flash?.success && (
        <div className="alert">{flash.success}</div>
      )}
      
      <h1>Products</h1>
      <Link href="/products/create">Create New</Link>
      
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Price</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {products.map(product => (
            <tr key={product.id}>
              <td>{product.name}</td>
              <td>${product.price}</td>
              <td>
                <Link href={`/products/${product.id}/edit`}>Edit</Link>
                <button onClick={() => deleteProduct(product.id)}>
                  Delete
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  )
}
```

## Shared Data

Automatically shared with all components:
- `auth` - Current user from auth.middleware
- `flash` - Flash messages
- `errors` - Validation errors

Access in components:
```jsx
export default function MyComponent({ auth, flash, errors }) {
  return (
    <div>
      {auth && <p>Logged in as: {auth.email}</p>}
      {flash?.success && <div>{flash.success}</div>}
      {errors?.name && <span>{errors.name}</span>}
    </div>
  )
}
```

## Forms

**Backend**:
```zenolang
http.post: /products {
    auth.middleware: {
        secret: env("JWT_SECRET")
        do: {
            http.form: { as: $form }
            
            db.insert: {
                table: "products"
                data: $form
            }
            
            http.redirect: { url: "/products" }
        }
    }
}
```

**Frontend**:
```jsx
import { useForm } from '@inertiajs/react'

export default function Create() {
  const { data, setData, post, processing, errors } = useForm({
    name: '',
    price: '',
  })

  const submit = (e) => {
    e.preventDefault()
    post('/products')
  }

  return (
    <form onSubmit={submit}>
      <input
        value={data.name}
        onChange={e => setData('name', e.target.value)}
      />
      {errors.name && <span>{errors.name}</span>}
      
      <button disabled={processing}>Create</button>
    </form>
  )
}
```

## Best Practices

1. **Always use auth.middleware** for protected routes
2. **Share common data** using inertia.share
3. **Use flash messages** for user feedback
4. **Validate on server** and return errors
5. **Use Link component** for navigation (no full reload)

## Demo

Demo routes available at:
- `/inertia/dashboard` - Dashboard example
- `/inertia/products` - Products list
- `/inertia/products/create` - Create form

## Production

1. Build frontend: `cd public/inertia && npm run build`
2. Assets served from `/inertia/dist/assets/`
3. Configure asset versioning for cache busting

## Summary

Inertia.js support enables modern SPA development while keeping server-side routing in ZenoLang. Perfect for:
- ✅ Admin panels
- ✅ Dashboards
- ✅ CRUD applications
- ✅ Multi-tenant apps

**Status**: Production Ready 🚀
