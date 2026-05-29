package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterFrontendRoutes(r *gin.Engine) {
	r.GET("/", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.String(http.StatusOK, indexHTML)
	})
}

const indexHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Books API</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: system-ui, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; }
  .card { background: #1e293b; border-radius: 16px; padding: 40px; width: 100%; max-width: 440px; box-shadow: 0 25px 50px rgba(0,0,0,0.4); }
  h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 8px; }
  .subtitle { color: #94a3b8; font-size: 0.9rem; margin-bottom: 32px; }
  .status-box { border-radius: 12px; padding: 20px; margin-bottom: 24px; }
  .status-box.logged-in { background: #14532d; border: 1px solid #16a34a; }
  .status-box.logged-out { background: #1c1917; border: 1px solid #44403c; }
  .status-label { font-size: 0.75rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 8px; }
  .logged-in .status-label { color: #4ade80; }
  .logged-out .status-label { color: #78716c; }
  .user-email { font-size: 1.1rem; font-weight: 600; }
  .user-id { font-size: 0.75rem; color: #94a3b8; margin-top: 4px; font-family: monospace; }
  .btn { display: block; width: 100%; padding: 12px; border-radius: 10px; border: none; font-size: 0.95rem; font-weight: 600; cursor: pointer; text-align: center; text-decoration: none; transition: opacity 0.15s; }
  .btn:hover { opacity: 0.85; }
  .btn-yandex { background: #fc0; color: #000; margin-bottom: 12px; }
  .btn-logout { background: #7f1d1d; color: #fca5a5; }
  .loading { color: #64748b; font-size: 0.9rem; }
</style>
</head>
<body>
<div class="card">
  <h1>рџ“љ Books API</h1>
  <p class="subtitle">Р›Р°Р±РѕСЂР°С‚РѕСЂРЅР°СЏ СЂР°Р±РѕС‚Р° в„–3</p>

  <div id="status" class="status-box logged-out">
    <div class="status-label">РЎС‚Р°С‚СѓСЃ</div>
    <div class="loading">Р—Р°РіСЂСѓР·РєР°...</div>
  </div>

  <div id="actions"></div>
</div>

<script>
async function checkAuth() {
  try {
    const r = await fetch('/auth/whoami', { credentials: 'include' });
    if (r.ok) {
      const user = await r.json();
      showLoggedIn(user);
    } else {
      showLoggedOut();
    }
  } catch {
    showLoggedOut();
  }
}

function showLoggedIn(user) {
  const box = document.getElementById('status');
  box.className = 'status-box logged-in';
  box.innerHTML = ` + "`" + `
    <div class="status-label">вњ… Р’С‹ РІРѕС€Р»Рё</div>
    <div class="user-email">${user.email}</div>
    <div class="user-id">ID: ${user.id}</div>
  ` + "`" + `;

  document.getElementById('actions').innerHTML = ` + "`" + `
    <button class="btn btn-logout" onclick="logout()">Р’С‹Р№С‚Рё</button>
  ` + "`" + `;
}

function showLoggedOut() {
  const box = document.getElementById('status');
  box.className = 'status-box logged-out';
  box.innerHTML = ` + "`" + `
    <div class="status-label">РќРµ Р°РІС‚РѕСЂРёР·РѕРІР°РЅ</div>
    <div class="user-email" style="color:#94a3b8">Р’РѕР№РґРёС‚Рµ С‡С‚РѕР±С‹ РїСЂРѕРґРѕР»Р¶РёС‚СЊ</div>
  ` + "`" + `;

  document.getElementById('actions').innerHTML = ` + "`" + `
    <a class="btn btn-yandex" href="/auth/oauth/yandex">Р’РѕР№С‚Рё С‡РµСЂРµР· РЇРЅРґРµРєСЃ</a>
  ` + "`" + `;
}

async function logout() {
  await fetch('/auth/logout', { method: 'POST', credentials: 'include' });
  checkAuth();
}

checkAuth();
</script>
</body>
</html>`

