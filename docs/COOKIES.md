# Cloud Game Authentication

The Gateway needs a `cloud_auth.json` file containing cookies from a logged-in cloud game session to authenticate with `sr.mihoyo.com/cloud/`.

## Obtaining cloud_auth.json

### Method 1: Playwright (recommended)

```bash
pip install playwright
playwright install chromium
```

```python
# save_cookies.py
from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch(headless=False)
    context = browser.new_context(
        viewport={"width": 1280, "height": 720},
        user_agent="Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36",
    )
    page = context.new_page()
    page.goto("https://sr.mihoyo.com/cloud/")

    print("Log in manually in the browser window, then press Enter...")
    input()

    context.storage_state(path="cloud_auth.json")
    print("Saved cloud_auth.json")
    browser.close()
```

### Method 2: Browser DevTools

1. Open Chrome and navigate to `https://sr.mihoyo.com/cloud/`
2. Log in normally
3. Open DevTools (F12) → Application → Cookies
4. Use a Playwright-compatible cookie export tool, or manually construct the JSON:

```json
{
  "cookies": [
    {"name": "ltoken_v2", "value": "...", "domain": ".mihoyo.com", "path": "/", "httpOnly": false, "secure": true, "sameSite": "Lax"},
    {"name": "ltuid_v2", "value": "...", "domain": ".mihoyo.com", "path": "/", "httpOnly": false, "secure": true, "sameSite": "Lax"},
    ...
  ]
}
```

The key cookies are `ltoken_v2` and `ltuid_v2` — these have ~1 year expiry and are sufficient for authentication.

## Refreshing

- Cookies (`ltoken_v2` / `ltuid_v2`) typically last ~1 year
- When cookies expire, the Gateway's health endpoint will show `state: "DEGRADED"`
- To refresh: re-run the save script above and restart the Gateway container with the new `cloud_auth.json`
- In Docker: `docker run -v ./cloud_auth.json:/app/cloud_auth.json src-web-gateway`
- In Docker Compose: update the file and `docker compose restart`

## Security

- `cloud_auth.json` is gitignored — never commit it
- The file is mounted read-only into the Docker container
- Cookie values are redacted in Gateway logs
