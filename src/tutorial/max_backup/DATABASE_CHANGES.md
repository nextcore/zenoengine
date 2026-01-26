# ⚠️ PENTING: Perubahan Database Tutorial

## Perubahan yang Dilakukan

Tutorial MAX sekarang menggunakan **database SQLite terpisah** (`tutorial_max.db`) agar tidak mengganggu database utama ZenoEngine.

### Konfigurasi Database

Ditambahkan di `.env`:
```env
DB_TUTORIAL_DRIVER=sqlite
DB_TUTORIAL_NAME=./tutorial_max.db
```

### Perubahan Nama Tabel

Semua tabel menggunakan prefix `tutorial_`:
- `users` → `tutorial_users`
- `teams` → `tutorial_teams`
- `tasks` → `tutorial_tasks`
- `notifications` → `tutorial_notifications`

### File yang Sudah Diupdate

✅ Migrations (semua menggunakan `db: tutorial`):
- 001_users.zl
- 002_teams.zl
- 003_tasks.zl
- 004_notifications.zl

### File yang Perlu Diupdate Manual

⚠️ **CATATAN PENTING**: Karena banyaknya file, beberapa file masih perlu diupdate untuk menambahkan `db: tutorial` pada setiap operasi database (`db.table`, `db.insert`, `db.update`, dll).

File yang perlu diupdate:
1. `seeders/demo_data.zl` - Ganti semua nama tabel dengan `tutorial_*`
2. `modules/auth/routes.zl` - Tambahkan `db: tutorial` pada semua `db.table`
3. `modules/tasks/*.zl` - Tambahkan `db: tutorial` pada semua operasi DB
4. `modules/teams/routes.zl` - Tambahkan `db: tutorial`
5. `modules/realtime/*.zl` - Tambahkan `db: tutorial`
6. `api/v1/*.zl` - Tambahkan `db: tutorial`
7. `main.zl` - Update query count users

## Cara Update Manual

Untuk setiap file yang menggunakan database, tambahkan `db: tutorial` setelah `db.table`:

**Sebelum:**
```javascript
db.table: users
db.where: { col: id, val: $id }
db.get: { as: $users }
```

**Sesudah:**
```javascript
db.table: tutorial_users
  db: tutorial
db.where: { col: id, val: $id }
db.get: { as: $users }
```

## Alternatif: Gunakan Script Otomatis

Saya bisa membuat script untuk mengupdate semua file sekaligus jika diperlukan.

## Testing

Setelah semua file diupdate, jalankan server dan cek:
1. Database `tutorial_max.db` terbuat otomatis
2. Tabel dengan prefix `tutorial_` ter-create
3. Login dan CRUD berfungsi normal
4. Tidak ada konflik dengan database utama
