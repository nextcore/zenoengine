# ZenoLang VSCode Extension

This extension provides syntax highlighting and language support for ZenoLang (`.zl`).

## Features
- **Syntax Highlighting**: Proper coloring for Slots, Variables (`$`), Strings, and Keywords.
- **Bracket Matching**: Auto-closing and matching for `{ }`, `[ ]`, `( )`.
- **Comments**: Support for `//` comments.

## Installation

### Method 1: Manual Install (Development)
1.  Copy the `vscode-zenolang` folder to your VSCode extensions directory:
    -   **Linux/Mac**: `~/.vscode/extensions/`
    -   **Windows**: `%USERPROFILE%\.vscode\extensions\`
2.  Restart VSCode.

### Method 2: Package & Install
1.  Install `vsce`: `npm install -g vsce`
2.  Run `vsce package` inside this folder to create a `.vsix` file.
3.  Install via VSCode "Extensions: Install from VSIX...".

## Usage
Simply open any file with `.zl` extension, and ZenoLang syntax highlighting will activate automatically.
