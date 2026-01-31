# Zeno-X Framework Example

![Zeno Artisan Banner](assets/artisan_banner.png)

This is a working example of the Zeno-X Framework structure.

## 🚀 Critical Lessons Learned (Pengalaman Development)

### 1. Route Syntax is Strict
Routes **MUST** wrap their logic in a `do { ... }` block. Without this, the router matches the URL but does not execute the handler (returns empty 200 OK).

**✅ Correct:**
```zenolang
http.get: '/contact' {
    do: {
        return: view('contact')
    }
}
```

**❌ Incorrect (Silent Failure):**
```zenolang
http.get: '/contact' {
    return: view('contact')
}
```

### 2. View Rendering
Use `view.blade: "filename"` or `view('basename')`.
Ensure the view file exists in `views/` directory.

### 3. AI Artisan (`ai_dev.zl`)
The `ai_dev.zl` script is a neural-scaffolding tool that generates features using Gemini.
- **Security**: API Keys are loaded from `.env` via `system.env: 'GEMINI_API_KEY'`.
- Jangan hardcode API Key di script untuk keamanan repository.

### 4. Security Configuration
The framework middleware (`securecookie`) requires 32-byte keys in `.env`:
- `APP_KEY`, `HASH_KEY`, `BLOCK_KEY`

---
 
 ## 🛠️ Zeno Artisan CLI (The "Laravel Killer")
 
 Zeno Artisan adalah tool baris perintah untuk mempercepat development, sekarang sudah didukung oleh **Zeno Metaprogramming**.
 
 ### Perintah Utama
 ```bash
 # Membuat Controller & Migration
 .\zeno.exe run artisan.zl make:controller ProfileController
 .\zeno.exe run artisan.zl make:migration create_users_table
 
 # Scaffolding Auth (Stateless JWT)
 .\zeno.exe run artisan.zl make:auth
 ```
 
 ---
 
 ## 🗄️ ZenoORM: Fluent & Relational
 
 ZenoORM bukan cuma DDL, tapi sudah mendukung **Advanced Relationships** dan **Eager Loading**.
 
 ### 1. Schema Builder (DDL)
 ```zenolang
 schema.create: 'posts' {
     column.id: 'id'
     column.integer: 'user_id'
     column.string: 'title'
     column.timestamps
 }
 ```
 
 ### 2. Relationships & Eager Loading
 Mendukung relasi antar tabel secara native di engine:
 ```zenolang
 // Define in script or model
 orm.model: 'users' { orm.hasMany: 'posts' { as: 'posts' } }
 
 // Eager load relationships
 orm.model: 'users'
 orm.with: 'posts' {
     orm.find: 1 { as: $user }
 }
 // $user.posts sekarang terisi otomatis!
 ```
 
 ### 3. Data Seeding
 Gunakan slot `db.seed` untuk mengisi data awal:
 ```zenolang
 db.seed: {
     orm.model: 'users'
     orm.save: { name: 'Budi' }
 }
 ```
 
 ## 📂 Project Structure
 ```text
 examples/zeno-x/
 ├── artisan.zl (Command Router)
 ├── ai_dev.zl (AI Feature Generator)
 ├── app/controllers/ (Application Logic)
 ├── database/migrations/ (DB Version Control)
 ├── views/stubs/ (Blade Templates for Scaffolding)
 └── .env (Environment Config)
 ```
 
 ## 🏃 Running the Project
 ```bash
 # Start Server
 .\zeno.exe
 
 # Test Endpoint
 curl http://localhost:3000/contact
 ```
