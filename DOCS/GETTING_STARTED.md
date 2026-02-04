# Getting Started with ZenoEngine

Welcome to **ZenoEngine**, the high-performance engine for **ZenoLang**. This guide will help you set up your environment and start building applications.

## 🚀 Two Ways to Use ZenoEngine

Depending on your goals, there are two ways to work with ZenoEngine:

### 1. Application Development (Recommended)
**Use this if you want to build apps using ZenoLang.**
- **How:** Download the latest [Zeno binary](https://github.com/user/ZenoEngine/releases).
- **Why:** It's faster, easier to manage, and isolates your application logic from the engine's core. You don't need Go installed.
- **Setup:**
    1. Place the `zeno` binary in your path or project folder.
    2. Create your `.zl` files in a `src/` directory.
    3. Run your app: `./zeno src/main.zl`

### 2. Core/Engine Development (Advanced)
**Use this if you want to contribute to ZenoEngine itself or modify its internals.**
- **How:** Clone this repository and run using the Go source.
- **Why:** Full access to the engine's Go code, internal slots, and performance tuning.
- **Setup:**
    1. Install Go (1.21+).
    2. Run the engine: `go run cmd/zeno/zeno.go`

---

## 🛠️ Installation & Setup

Regardless of your choice, follow these steps to initialize your project:

### 1. Initialize Configuration
Copy the example environment file and generate your security key:
```bash
cp .env.example .env
# If using binary:
zeno key:generate
# If using source:
go run cmd/zeno/zeno.go key:generate
```

### 2. Configure Database
By default, ZenoEngine uses **SQLite**. You can change this in your `.env`:
```env
DB_DRIVER=sqlite
DB_NAME=zeno_app.db
```

### 3. Run the Demo
The project comes with a built-in **MAX Demo** (Full Feature Showcase).
```bash
# Start the engine
zeno
# Or via source
go run cmd/zeno/zeno.go
```
Once started, visit `http://localhost:3000` to see the demo.

---

## 📖 Next Steps
- Learn the syntax: Check [LANGUAGE_SPECIFICATION.md](file:///home/max/Documents/PROJ/ZenoEngine%20-%20Public/LANGUAGE_SPECIFICATION.md)
- Follow the style guide: [ZENOLANG_STYLE_GUIDE.md](file:///home/max/Documents/PROJ/ZenoEngine%20-%20Public/ZENOLANG_STYLE_GUIDE.md)
- Explore the tutorials: Check the folder `src/tutorial/`
