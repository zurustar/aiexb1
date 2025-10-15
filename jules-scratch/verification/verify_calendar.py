
import re
import random
import string
from playwright.sync_api import Playwright, sync_playwright, expect

def get_random_string(length):
    # choose from all lowercase letter
    letters = string.ascii_lowercase
    result_str = ''.join(random.choice(letters) for i in range(length))
    return result_str

def run(playwright: Playwright) -> None:
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    random_string = get_random_string(8)
    username = f"testuser_{random_string}"
    email = f"test_{random_string}@example.com"
    password = "password"

    page.goto("http://localhost:8080/")

    # Register a new user
    page.get_by_role("link", name="新規登録").click()
    page.get_by_placeholder("ユーザー名").click()
    page.get_by_placeholder("ユーザー名").fill(username)
    page.locator("#register-email").click()
    page.locator("#register-email").fill(email)
    page.locator("#register-password").click()
    page.locator("#register-password").fill(password)
    page.get_by_role("button", name="登録").click()

    # Log in with the new user
    page.locator("#login-email").click()
    page.locator("#login-email").fill(email)
    page.locator("#login-password").click()
    page.locator("#login-password").fill(password)
    page.get_by_role("button", name="ログイン").click()

    # Verify login and calendar visibility
    expect(page.get_by_role("button", name="ログアウト")).to_be_visible()
    expect(page.locator("#calendar-container")).to_be_visible()
    page.screenshot(path="jules-scratch/verification/verification.png")

    context.close()
    browser.close()


with sync_playwright() as playwright:
    run(playwright)
