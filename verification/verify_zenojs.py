
from playwright.sync_api import sync_playwright

def verify_zenojs():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()

        # Go to the playground
        page.goto("http://localhost:8000/index.html")

        # Wait for content to render
        page.wait_for_selector(".box")

        # Verify initial state
        assert "Count is: 0" in page.content()
        assert "Learn ZenoJS" in page.content()

        # Interaction: Increment Count
        page.click("button:has-text('+')")
        page.wait_for_timeout(100) # wait for reactivity
        assert "Count is: 1" in page.content()

        # Interaction: Add Todo
        page.fill("#new-todo", "Verify Frontend")
        page.click("button:has-text('Add')")
        page.wait_for_timeout(100)
        assert "Verify Frontend" in page.content()

        # Take screenshot
        page.screenshot(path="verification/zenojs_demo.png")
        print("Verification successful, screenshot saved.")

        browser.close()

if __name__ == "__main__":
    verify_zenojs()
