# Embedding ZenoLang

ZenoLang core engine dirancang agar bisa di-*embed* (ditanam) ke dalam projek lain dengan sangat mudah. Bahasa skrip ZenoLang saat ini dapat di-embed di projek **Go** maupun **Rust**.

---

## 1. Embedding in Go

### Instalasi

Untuk mengimpor ZenoLang di dalam projek Go, Anda dapat mengimpor package engine:

```go
import (
    "zeno/pkg/engine"
)
```

### Dasar Penggunaan

Untuk menjalankan skrip ZenoLang di dalam aplikasi Go Anda, Anda hanya perlu menginisialisasi `engine.Engine` dan mengeksekusi AST (Abstract Syntax Tree) yang dihasilkan oleh parser.

#### Contoh Sederhana

Berikut adalah contoh cara mengeksekusi string skrip ZenoLang langsung dari Go:

```go
package main

import (
	"context"
	"fmt"
	"zeno/pkg/engine"
)

func main() {
	// 1. Inisialisasi Engine
	eng := engine.NewEngine()

	// 2. Kode ZenoLang
	code := `
	var: $name { val: 'ZenoLang' }
	log: 'Hello from ' + $name
	`

	// 3. Parse kode menjadi AST
	root, err := engine.ParseString(code, "eval")
	if err != nil {
		panic(err)
	}

	// 4. Buat Scope untuk variabel
	scope := engine.NewScope(nil)

	// 5. Eksekusi
	ctx := context.Background()
	if err := eng.Execute(ctx, root, scope); err != nil {
		panic(err)
	}
}
```

### Mendaftarkan Slot Kustom (Go)

Anda bisa mendaftarkan fungsi Go Anda sebagai slot yang bisa dipanggil dari skrip ZenoLang:

```go
eng.Register("my.slot", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
    val := node.Value
    fmt.Println("Nilai utama:", val)

    for _, child := range node.Children {
        fmt.Printf("Atribut: %s = %v\n", child.Name, child.Value)
    }

    scope.Set("my_result", "Sukses!")
    return nil
}, engine.SlotMeta{
    Description: "Slot kustom saya",
    Example:     "my.slot: 'nilai' { attr: 'v' }",
})
```

---

## 2. Embedding in Rust (`zenoengine`)

Modular workspace Rust `zenoengine` (dan sub-crate pendukungnya) dapat diintegrasikan langsung ke proyek Rust Anda menggunakan Cargo.

### Instalasi via GitHub

Untuk menambahkan `zenoengine` ke proyek Rust Anda langsung dari repository GitHub, tambahkan baris berikut ke `Cargo.toml` proyek Anda:

```toml
[dependencies]
zenoengine = { git = "https://github.com/nextcore/zenoengine.git", package = "zenoengine" }
```

> [!NOTE]
> Meskipun repositori utama `zenoengine` berisi kode Go di root-nya dan kode Rust berada di subfolder `zeno-rs/`, Cargo secara otomatis akan melakukan pemindaian (scan) rekursif ke seluruh direktori repositori Git tersebut saat proses build untuk mencari file `Cargo.toml` yang mendefinisikan crate dengan nama `zenoengine`. Oleh karena itu, Cargo dapat menemukannya dan melakukan import tanpa masalah.

### Dasar Penggunaan

Untuk membuat instance engine yang telah dimuat dengan standard library (Math & Date slots) bawaan:

```rust
use zenoengine::{new_engine, parser::parse_string, executor::Context, scope::Scope};

fn main() {
    // 1. Buat engine baru dengan slot stdlib pre-registered
    let engine = new_engine();
    let mut ctx = Context::new();
    let scope = Scope::new(None);

    // 2. Tulis skrip ZenoLang
    let script = r#"
        var: $a { val: 10 }
        var: $b { val: 20 }
        math.calc: "$a + $b" {
            as: $sum
        }
    "#;

    // 3. Parse dan Eksekusi
    let root = parse_string(script, "example.zl").unwrap();
    engine.execute(&mut ctx, &root, &scope).unwrap();

    // 4. Ambil hasil dari scope
    if let Some(val) = scope.get("sum") {
        println!("Hasil penjumlahan: {:?}", val); // Output: Float(30.0)
    }
}
```

### Mendaftarkan Slot Kustom (Rust)

Anda dapat memperluas engine Rust dengan mendaftarkan closure atau fungsi sebagai slot:

```rust
use std::sync::Arc;
use zenoengine::{Engine, SlotMeta, InputMeta, Diagnostic};

fn register_custom_slot(engine: &mut Engine) {
    engine.register(
        "custom.log",
        Arc::new(|_engine, _ctx, node, scope| {
            if let Some(ref v) = node.value {
                println!("LOG: {}", v);
            }
            Ok(())
        }),
        SlotMeta {
            description: "Mencetak pesan kustom ke stdout".to_string(),
            example: "custom.log: 'Pesan'".to_string(),
            inputs: std::collections::HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        }
    );
}
```

### Menyajikan Dokumentasi API & Swagger UI

Jika Anda menggunakan framework web Rust seperti Axum atau Actix-web, Anda dapat menggunakan modul `apidoc` dari `zenoengine` untuk mendaftarkan skema API dan menyajikan Swagger UI:

```rust
use zenoengine::apidoc::{self, APIRegistry, RouteDoc};

// 1. Daftarkan rute dokumentasi
APIRegistry::global().register("POST", "/execute", RouteDoc {
    method: "POST".to_string(),
    path: "/execute".to_string(),
    summary: "Eksekusi ZenoLang Script".to_string(),
    description: "Mengevaluasi kode skrip yang dikirimkan client".to_string(),
    tags: vec!["Execution".to_string()],
    params: Vec::new(),
    request_body: None,
    responses: std::collections::HashMap::new(),
});

// 2. Generate HTML Swagger UI untuk handler endpoint
let swagger_html = apidoc::swagger_ui_html("/openapi.json");
```
