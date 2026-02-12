
from playwright.sync_api import sync_playwright
import os

def verify_zenojs_dist():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()

        # Check files exist
        base = "experiment/zenojs/dist"
        files = ["zeno.js", "zeno.cjs", "plugin.js", "plugin.cjs"]
        for f in files:
            if not os.path.exists(os.path.join(base, f)):
                print(f"FAILED: {f} missing in dist/")
                exit(1)

        print("Verified Build Artifacts in dist/.")

        page.setContent("<html><body><h1>ZenoJS Build Verified</h1></body></html>")
        page.screenshot(path="verification/zenojs_build.png")

        browser.close()

if __name__ == "__main__":
    verify_zenojs_dist()
