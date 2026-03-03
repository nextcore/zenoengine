# Kelemahan Arsitektur & Keamanan ZenoEngine

Berdasarkan tinjauan source code mendalam (terutama pada modul `pkg/engine` dan slot `internal/slots`), berikut adalah beberapa celah keamanan dan kelemahan arsitektur tingkat lanjut yang melekat pada implementasi ZenoEngine saat ini:

## 1. Fragmentasi Scope & Ketidakmampuan Berbagi State Lintas Request
Lokasi: `pkg/engine/scope.go` dan `internal/slots/router.go`

**Masalah:**
Setiap ada panggilan HTTP (*request*), ZenoEngine menciptakan `reqScope` baru (lewat `arena.AllocScope()`) dengan metode kloning-bayangan (*Shadowing*). Secara desain, *Mutex Relational Lock* hanya diterapkan pada ruang memori skrip yang berjalan aktif *(current level)*.

**Dampak:**
Skrip ZenoLang (`.zl`) tidak bisa memiliki variabel *global* antar-permintaan (*request*) secara harafiah. Misalnya, Anda mendefinisikan `set: $pengunjung = $pengunjung + 1` secara global, hal ini hanya akan memengaruhi `reqScope` di permintaan HTTP itu saja (Shadowing). Artinya, berbagi memori (*shared-memory*, misalnya untuk membuat struktur *Cache* L1 in-memory secara *native*) antar sesi menggunakan variabel `.zl` adalah mustahil, mengharuskan *developer* selalu menggunakan I/O lambat (basis data / file / *memory cache slot*) untuk setiap status yang persisten.

**Efek Positif & Kenapa Dipertahankan (*Trade-Off*):**
Arsitektur ini lebih dikenal dengan nama ***Shared-Nothing Architecture*** (mirip dengan pilar desain ekosistem PHP modern). Karena sama sekali tidak ada memori yang saling tumpang tindih antar *request HTTP* atau *thread* yang berbeda, ZenoEngine menjamin bahwa **seluruh eksekusi skrip pengguna 100% *Thread-Safe*** tanpa perlu pusing memikirkan bentrokan *Race Conditions* atau me-*manage* kunci sinkronisasi *(Mutex Locks)* secara manual. Hasilnya, pengembangan web dengan konkurensi masif *(Massive Concurrency)* menggunakan ZenoLang menjadi teramat aman bahkan bagi *developer* pemula.

## 3. Beban Evaluasi AST Tunggal (*AST-Walking Interpreter Overhead*)
Lokasi: `pkg/engine/executor.go`

**Masalah:**
ZenoEngine tidak mengompilasi skrip `.zl` menjadi file biner lokal ataupun *Bytecode*. Skrip diproses dengan berjalan menjelajahi Pohon Sintaksis Abstrak *(AST Walking)* dari mula hingga akhir setiap saat `eng.Execute()` dipanggil.

**Dampak:**
Sifat komputasi yang serba me-*resolve* nama node saat *runtime* berbasiskan `String-Map` akan selalu tertinggal dibandingkan bahasa yang telah melalui fase kompilasi/JIT (seperti V8 JavaScript, atau bahkan *bytecode* milik PHP 8). Hal ini membuat ZenoEngine tidak sesuai untuk kalkulasi berat atau *loop* matematika raksasa di level bahasa skrip (*CPU-Bound*). Meskipun di mitigasi dengan fitur *"FastPath"* pada *node* yang sering dieksekusi, pemborosan memori per lokalisasi *(Heap Objects)* dan memacu *Garbage Collector* Go tetap signifikan bila diakses puluhan ribu orang serentak.

**Efek Positif & Kenapa Dipertahankan (*Trade-Off*):**
Ketiadaan fase kompilasi rumit ke eksekusi *Bytecode* seketika memberikan ZenoEngine masa muat/start (*Startup Time*) yang praktis bernilai nihil. Hal ini memungkinkan fitur andalan bernama **Live Reload / Hot Module Replacement**, di mana *developer* dapat mengubah satu parameter kecil di baris logika *script* basis `.zl` pada aplikasi yang sedang tayang, lalu perubahannya merefleksi seketika tanpa membutuhkan waktu *warming-up JIT* sama sekali. Desain sebagai alat interpretasi penelusuran *Tree (AST Walker)* juga menjaga kompleksitas pustaka sumber `pkg/engine` di bahasa (Go) pada derajat sangat minimalis untuk dirawat. Ia sengaja mengorbankan kecepatan ekstrem sekelas *Rust* untuk sebuah ketangkasan pengembangan yang tiada tara seperti *Laravel*.
