# DB Lifecycle Hooks (`db.hook`)

ZenoEngine mendukung **lifecycle hooks** pada operasi database. Dengan `db.hook`, Anda bisa menjalankan kode secara otomatis sebelum atau sesudah operasi INSERT, UPDATE, atau DELETE pada sebuah tabel.

Ini sangat berguna untuk:
- **Auto-generate field** (misal: `slug`, `updated_at`)
- **Cache invalidation** otomatis setelah data berubah
- **Audit Trail** — mencatat semua perubahan data
- **Notifikasi** — kirim email setelah record disimpan

## Mendaftarkan Hook

Panggil `db.hook` di awal aplikasi (`src/main.zl`) atau di awal setiap route file. Hook tersimpan di memori selama server berjalan.

```zeno
db.hook: 'posts' {
  before_insert: {
    log.info: "Akan menyimpan post baru"
  }
  after_insert: {
    cache.forget: "posts_homepage"
  }
  after_save: {
    # Dijalankan setelah INSERT atau UPDATE
    cache.forget: "posts_list"
  }
  after_update: {
    db.table: activity_log {
      db.insert { action: "updated" model: "posts" user_id: $auth.id }
    }
  }
  before_delete: {
    log.warn: "Menghapus post!"
  }
  after_delete: {
    cache.forget: "posts_list"
  }
}
```

## Referensi Event

| Event | Kapan Dijalankan |
| :--- | :--- |
| `before_insert` | Sebelum `db.insert` dieksekusi |
| `after_insert` | Setelah `db.insert` berhasil |
| `before_update` | Sebelum `db.update` dieksekusi |
| `after_update` | Setelah `db.update` berhasil |
| `before_delete` | Sebelum `db.delete` dieksekusi |
| `after_delete` | Setelah `db.delete` berhasil |
| `before_save` | Sebelum INSERT **atau** UPDATE |
| `after_save` | Setelah INSERT **atau** UPDATE berhasil |

## Variabel yang Tersedia di Dalam Hook

| Variabel | Keterangan |
| :--- | :--- |
| `$data` | Map berisi semua field yang akan di-insert/update |
| `$table` | Nama tabel yang sedang dioperasikan |

```zeno
db.hook: 'articles' {
  before_insert: {
    # $data.title tersedia di sini
    log.info: "Menyimpan artikel: " + $data.title
  }
}
```

> [!NOTE]
> Hook bersifat **global** untuk seluruh lifetime server. Cukup daftarkan sekali di `src/main.zl`.

> [!TIP]
> Gunakan `after_save` jika Anda ingin satu hook untuk menangani INSERT dan UPDATE sekaligus — sangat berguna untuk invalidasi cache.

> [!CAUTION]
> Jika hook menghasilkan error, operasi database **tidak akan di-rollback** secara otomatis kecuali Anda membungkusnya dengan `db.transaction`.
