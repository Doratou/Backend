package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const oauthTestHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Doratou - Google OAuth2 Test Bench</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
        body { background: #f0f2f5; color: #1a1a1a; display: flex; justify-content: center; align-items: center; min-height: 100vh; padding: 20px; }
        .card { background: #fff; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); width: 100%; max-width: 640px; padding: 32px; }
        h1 { font-size: 24px; font-weight: 700; color: #111; margin-bottom: 8px; }
        p.subtitle { color: #666; font-size: 14px; margin-bottom: 24px; line-height: 1.5; }
        .btn-google {
            display: inline-flex; align-items: center; justify-content: center; gap: 12px;
            width: 100%; padding: 14px 20px; border-radius: 8px; border: 1px solid #dadce0;
            background: #fff; color: #3c4043; font-size: 16px; font-weight: 500; cursor: pointer;
            text-decoration: none; transition: background 0.2s, box-shadow 0.2s;
        }
        .btn-google:hover { background: #f8f9fa; box-shadow: 0 1px 3px rgba(60,64,67,0.3); }
        .btn-google svg { width: 20px; height: 20px; }
        .divider { display: flex; align-items: center; margin: 28px 0; color: #999; font-size: 13px; }
        .divider::before, .divider::after { content: ''; flex: 1; height: 1px; background: #e0e0e0; }
        .divider span { padding: 0 12px; }
        .section-title { font-size: 15px; font-weight: 600; margin-bottom: 12px; color: #333; }
        .input-group { display: flex; gap: 8px; margin-bottom: 16px; }
        input[type="text"] {
            flex: 1; padding: 10px 14px; border: 1px solid #d0d7de; border-radius: 6px; font-size: 14px;
            outline: none; font-family: monospace;
        }
        input[type="text"]:focus { border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.2); }
        button.btn-action {
            padding: 10px 18px; border-radius: 6px; border: none; background: #1a73e8; color: #fff;
            font-size: 14px; font-weight: 500; cursor: pointer; transition: background 0.2s;
        }
        button.btn-action:hover { background: #1557b0; }
        pre {
            background: #1e1e2e; color: #cdd6f4; padding: 16px; border-radius: 8px; font-size: 12px;
            overflow-x: auto; max-height: 240px; white-space: pre-wrap; word-break: break-all;
        }
        .status-badge {
            display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 600;
            margin-bottom: 12px;
        }
        .status-badge.info { background: #e8f0fe; color: #1a73e8; }
        .guide { background: #fafafa; border: 1px solid #eee; border-radius: 8px; padding: 14px; font-size: 13px; color: #555; line-height: 1.6; }
        .guide ol { padding-left: 20px; }
    </style>
</head>
<body>
    <div class="card">
        <span class="status-badge info">Dev Testbench</span>
        <h1>Google OAuth 2.0 Test</h1>
        <p class="subtitle">Click the button below to initiate the Google OAuth2 / OIDC flow. The backend handles PKCE/nonce verification, auto-registration, and account linking.</p>

        <a class="btn-google" href="/api/v1/auth/google/login">
            <svg viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/>
            </svg>
            Sign in with Google
        </a>

        <div class="divider"><span>OR TEST PROTECTED API WITH TOKEN</span></div>

        <div class="section-title">Test Token with GET /api/v1/users/me</div>
        <div class="input-group">
            <input type="text" id="tokenInput" placeholder="Paste access_token here..." />
            <button class="btn-action" onclick="testToken()">Verify</button>
        </div>
        <pre id="output">Output will appear here...</pre>

        <div class="divider"><span>HOW IT WORKS</span></div>
        <div class="guide">
            <ol>
                <li>Click <strong>Sign in with Google</strong> to redirect to Google consent screen.</li>
                <li>Google redirects back to <code>/api/v1/auth/google/callback</code> with code and state.</li>
                <li>Backend checks state in Redis, exchanges token, and issues JWT <code>AuthResponse</code>.</li>
                <li>Copy <code>access_token</code> from response JSON and paste above to test.</li>
            </ol>
        </div>
    </div>

    <script>
        async function testToken() {
            const token = document.getElementById('tokenInput').value.trim();
            const output = document.getElementById('output');
            if (!token) {
                output.textContent = 'Please enter an access token.';
                return;
            }
            output.textContent = 'Calling GET /api/v1/users/me ...';
            try {
                const res = await fetch('/api/v1/users/me', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                const data = await res.json();
                output.textContent = JSON.stringify(data, null, 2);
            } catch (err) {
                output.textContent = 'Error: ' + err.message;
            }
        }
    </script>
</body>
</html>`

// OAuthTestUI serves a developer-friendly testbench for Google OAuth at GET /oauth-test.
func OAuthTestUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, oauthTestHTML)
}
