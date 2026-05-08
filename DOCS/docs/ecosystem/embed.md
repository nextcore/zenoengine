# Embedding ZenoLang

ZenoLang core engine dirancang agar bisa di-*embed* (ditanam) ke dalam projek Go murni dengan sangat mudah. Ini memungkinkan Anda menggunakan ZenoLang sebagai bahasa skrip di dalam aplikasi Anda sendiri.

## Instalasi

Pastikan Anda telah menginisialisasi modul Go di projek Anda. Untuk saat ini, karena projek ini masih dalam pengembangan lokal, Anda dapat mereferensikan folder `zenoengine` di `go.mod` Anda atau mengimportnya jika berada di dalam workspace yang sama.

```go
import (
    "zeno/pkg/engine"
)
```

## Dasar Penggunaan

Untuk menjalankan skrip ZenoLang di dalam aplikasi Go Anda, Anda hanya perlu menginisialisasi `engine.Engine` dan mengeksekusi AST (Abstract Syntax Tree) yang dihasilkan oleh parser.

### Contoh Sederhana

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

## Mendaftarkan Slot Kustom

Kekuatan utama ZenoLang adalah kemampuannya untuk diperluas dengan **Slots**. Anda bisa mendaftarkan fungsi Go Anda sebagai slot yang bisa dipanggil dari skrip ZenoLang.

```go
eng.Register("my.slot", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
    // Ambil nilai utama
    val := node.Value
    fmt.Println("Nilai utama:", val)

    // Ambil anak-anak (children) jika ada
    for _, child := range node.Children {
        fmt.Printf("Atribut: %s = %v\n", child.Name, child.Value)
    }

    // Anda juga bisa memanipulasi scope
    scope.Set("my_result", "Sukses!")

    return nil
}, engine.SlotMeta{
    Description: "Slot kustom saya",
    Example:     "my.slot: 'nilai' { attr: 'v' }",
})
```

## Menggunakan Zeno Blade

Jika Anda juga ingin menggunakan sistem template Blade di projek Go Anda, Anda bisa mengimport `zeno/pkg/blade` dan mendaftarkannya:

```go
import "zeno/pkg/blade"

// Di dalam main()
blade.RegisterBladeSlots(eng)
```

Sekarang Anda bisa menggunakan slot `view.blade` di dalam skrip Anda untuk me-render template Blade!
