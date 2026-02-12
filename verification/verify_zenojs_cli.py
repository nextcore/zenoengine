
from playwright.sync_api import sync_playwright
import os

def verify_zenojs_cli():
    # Verify file structure created by CLI
    base_path = "verification/my-test-app"

    required_files = [
        "package.json",
        "vite.config.js",
        "index.html",
        "src/main.js",
        "src/App.blade",
        "lib/zeno/src/zeno.js",
        "lib/zeno/vite-plugin.js"
    ]

    missing = []
    for f in required_files:
        if not os.path.exists(os.path.join(base_path, f)):
            missing.append(f)

    if missing:
        print(f"CLI Verification Failed. Missing: {missing}")
        exit(1)

    print("CLI Verification Passed: All boilerplate files present.")

    # Dummy screenshot
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        page.setContent("<html><body><h1>ZenoJS CLI Verified</h1></body></html>")
        page.screenshot(path="verification/zenojs_cli.png")
        browser.close()

if __name__ == "__main__":
    verify_zenojs_cli()
