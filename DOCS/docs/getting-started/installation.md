# Installation

## Meet ZenoEngine

ZenoEngine is a web application framework with expressive, elegant syntax. We've already laid the foundation — freeing you to create without sweating the small things.

If you are coming from Laravel, you will feel right at home. ZenoEngine brings the elegant syntax, routing, ORM, and Blade templating you love, but executes them at the speed of compiled Go.

### Step 1: Download & Install

ZenoEngine is distributed as a single executable binary.
1. Go to the [Releases](https://github.com/nextcore/zenoengine/releases) page.
2. Download the appropriate binary file for your OS (`zeno-linux-amd64`, `zeno-darwin-arm64`, etc.).
3. Rename the file to `zeno` (or `zeno.exe` on Windows).
4. Make it executable: `chmod +x zeno`
5. Move it to your `/usr/local/bin/` so you can use it globally.

### Step 2: Set Up a Project
Since ZenoEngine behaves as a universal runtime, you can write your own `src/main.zl` file or clone a starter template to get started.

A standard project structure looks like this:
```text
my-app/
├── src/
│   └── main.zl
├── views/
│   └── welcome.blade.zl
└── .env
```

### Step 3: Run the Server
To start the application, execute `zeno` by passing the path to your entry point script:

```bash
zeno src/main.zl
```
