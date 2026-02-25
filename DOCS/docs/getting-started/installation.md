# Installation

## Meet ZenoEngine

ZenoEngine is a web application framework with expressive, elegant syntax. We've already laid the foundation — freeing you to create without sweating the small things.

If you are coming from Laravel, you will feel right at home. ZenoEngine brings the elegant syntax, routing, ORM, and Blade templating you love, but executes them at the speed of compiled Go.

### Step 1: Download & Install

ZenoEngine is distributed as a single executable binary.
1. Go to the [Releases](https://github.com/zenolang/zenoengine/releases) page.
2. Download the appropriate binary file for your OS (`zeno-linux-amd64`, `zeno-darwin-arm64`, etc.).
3. Rename the file to `zeno` (or `zeno.exe` on Windows).
4. Make it executable: `chmod +x zeno`
5. Move it to your `/usr/local/bin/` so you can use it globally.

### Step 2: Create a New Project

ZenoEngine comes with two powerful officially supported templates built-in: a classic Laravel-style **MVC** template, and a modern Domain-Driven **Modular** template. Both come fully configured with Authentication, User Management, Roles, and an SQLite database!

To scaffold a new project, run:

```bash
zeno new my_first_app
```

You will be presented with an interactive prompt asking you to choose your preferred template:
```text
🚀 Welcome to ZenoEngine!
Choose a starting boilerplate for your project:
  1) MVC (Classic Laravel-style architecture)
  2) Modular (Domain-Driven, feature-based architecture)

Enter choice (1 or 2) [default: 1]: 
```

*(Alternatively, you can skip the prompt by passing `--template=mvc` or `--template=modular`)*

### Step 3: Run the Server

Once your project is created, navigate into it and start the development server:

```bash
cd my_first_app
zeno run src/main.zl
```

Or just run the binary by itself if it's in the project root:
```bash
./zeno
```
