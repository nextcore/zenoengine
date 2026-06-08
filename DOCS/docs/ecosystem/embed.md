# Embedding ZenoLang

ZenoLang core engine dirancang agar bisa di-*embed* (ditanam) ke dalam projek Go lain dengan sangat mudah.

### Instalasi

Untuk mengimpor ZenoLang di dalam projek Go, unduh modul `zeno-go` terlebih dahulu:

```bash
go get github.com/nextcore/zeno-go
```

Kemudian impor package engine:

```go
import (
    "github.com/nextcore/zeno-go/pkg/engine"
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
	"github.com/nextcore/zeno-go/pkg/engine"
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
### Integrasi Modular ZenoEngine (Option Pattern)

Jika Anda menggunakan framework **`zenoengine`** penuh di dalam proyek Go Anda, Anda bisa menggunakan **Option Pattern** untuk mendaftarkan modul-modul bawaan secara selektif. Ini sangat berguna jika Anda tidak ingin membebani aplikasi dengan seluruh fitur monolitik (misalnya, membangun microservice tanpa database, atau sekadar melakukan rendering template Blade).

Untuk menggunakan pendaftaran modular, impor package `app` dari `zenoengine`:

```go
import (
    "github.com/nextcore/zeno-go/pkg/engine"
    "github.com/nextcore/zenoengine/internal/app"
)
```

#### Opsi Pendaftaran Kategori (Broad Options)
- **`app.WithCore()`**: Mendaftarkan seluruh slot inti (Math, Time, Logic, FileSystem, Metadata, dll).
- **`app.WithWeb(r *chi.Mux)`**: Mendaftarkan slot web server (Router, Blade, Inertia, HTTP Client/Server, Session, Captcha).
- **`app.WithData(dbMgr *dbmanager.DBManager)`**: Mendaftarkan slot database, ORM, Schema, Validator, serta Auth.
- **`app.WithExtra(queue worker.JobQueue, setConfig func([]string))`**: Mendaftarkan modul pengiriman Email, in-memory caching, serta background jobs.

#### Opsi Pendaftaran Granular
Untuk kontrol yang lebih detail, Anda bisa memilih modul tertentu secara individual:
- **`app.WithBlade()`**: Mengaktifkan mesin templating Blade.
- **`app.WithRouter(r *chi.Mux)`**: Mengaktifkan routing HTTP.
- **`app.WithDB(dbMgr)`**: Mengaktifkan Database, ORM, Schema, dan DB Hooks.
- **`app.WithMail()`**: Mengaktifkan modul Email.
- **`app.WithCache()`**: Mengaktifkan in-memory Cache.
- **`app.WithJob(queue, setConfig)`**: Mengaktifkan background worker queue.

#### Contoh Integrasi Selektif

Berikut adalah contoh inisialisasi engine yang hanya menggunakan fitur **Core** dan **Blade Templating** (tanpa database, worker, ataupun router HTTP):

```go
package main

import (
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zenoengine/internal/app"
)

func main() {
	// 1. Inisialisasi Core Engine
	eng := engine.NewEngine()

	// 2. Registrasi Fitur secara Selektif (hanya Core & Blade Template)
	app.RegisterSlots(eng,
		app.WithCore(),
		app.WithBlade(),
	)

	// Sekarang Anda dapat mengeksekusi template Blade atau skrip ZenoLang secara langsung!
}
```
