# ZenoLang for Zed

Enterprise-grade ZenoLang support for the Zed editor.

## Features
- **Tree-sitter Grammar**: Robust parsing for `.zl` and `.blade.zl` files.
- **Syntax Highlighting**: Accurate coloring for slots, variables, strings, and control keywords.
- **Auto-indentation**: Smart indentation handling for nested blocks.
- **Bracket Matching**: Correct pairing for `{}` brackets.

## Installation (Development Mode)

Since this extension is part of the ZenoEngine mono-repo, you can install it as a "Dev Extension".

1. Open **Zed**.
2. Open the Command Palette (`Ctrl+Shift+P` or `Cmd+Shift+P`).
3. Type **"Extensions: Install Dev Extension"**.
4. Select the `zed-zenolang` folder in this repository.

## Structure
- `languages/zenolang/config.toml`: Language configuration (comments, brackets).
- `languages/zenolang/highlights.scm`: Tree-sitter query for syntax highlighting.
- `grammar.js`: The Tree-sitter grammar definition.
- `extension.toml`: Extension manifest.
