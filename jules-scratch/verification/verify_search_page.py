from playwright.sync_api import sync_playwright, Page, expect
import time

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page()

    # Wait for the server to start
    time.sleep(15)

    # Navigate to the search page
    page.goto("http://localhost:8081/search")

    # Enter a search query
    page.get_by_placeholder("Search...").fill("test")

    # Click the search button
    page.get_by_role("button", name="Search").click()

    # Wait for the results to load
    expect(page.locator("#results-container")).not_to_be_empty()

    # Take a screenshot
    page.screenshot(path="jules-scratch/verification/verification.png")

    browser.close()

with sync_playwright() as playwright:
    run(playwright)