
import re
import random
import string
from playwright.sync_api import Playwright, sync_playwright, expect
from datetime import datetime, timedelta

def get_random_string(length):
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(length))

def run(playwright: Playwright) -> None:
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    random_string = get_random_string(8)
    username = f"testuser_{random_string}"
    email = f"test_{random_string}@example.com"
    password = "password"

    page.goto("http://localhost:8080/")

    # Register
    page.get_by_role("link", name="新規登録").click()
    page.get_by_placeholder("ユーザー名").fill(username)
    page.locator("#register-email").fill(email)
    page.locator("#register-password").fill(password)
    page.get_by_role("button", name="登録").click()

    # Log in
    page.locator("#login-email").fill(email)
    page.locator("#login-password").fill(password)
    page.get_by_role("button", name="ログイン").click()
    expect(page.get_by_role("button", name="ログアウト")).to_be_visible()

    # Create a schedule for today
    now = datetime.now()
    start_time = now.replace(hour=9, minute=0, second=0, microsecond=0)
    end_time = start_time + timedelta(hours=2)

    start_time_str = start_time.strftime('%Y-%m-%dT%H:%M')
    end_time_str = end_time.strftime('%Y-%m-%dT%H:%M')

    page.locator("#schedule-title").fill("Test Event")
    page.locator("#schedule-start-time").fill(start_time_str)
    page.locator("#schedule-end-time").fill(end_time_str)
    page.get_by_role("button", name="追加").click()

    # Wait for the event to appear and take a screenshot
    expect(page.locator(".schedule-item-calendar")).to_be_visible()
    page.screenshot(path="jules-scratch/verification/verification.png")

    context.close()
    browser.close()

with sync_playwright() as playwright:
    run(playwright)
