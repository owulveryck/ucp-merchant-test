package main

import (
	"fmt"
	"net/http"
)

func (s *ArenaServer) handleLanding(w http.ResponseWriter, r *http.Request) {
	arenaURL := fmt.Sprintf("http://%s", r.Host)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, landingHTML, arenaURL, arenaURL, s.productName, float64(s.costPrice)/100, arenaURL)
}

const landingHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>UCP Arena</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'Segoe UI',system-ui,sans-serif;background:#0a0a1a;color:#e0e0e0;min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:2rem}
.logo{font-size:3rem;font-weight:800;background:linear-gradient(135deg,#00d4ff,#7b2ff7);-webkit-background-clip:text;-webkit-text-fill-color:transparent;margin-bottom:.5rem}
.subtitle{font-size:1.2rem;color:#888;margin-bottom:2rem}
.qr-section{background:#111;border:2px solid #222;border-radius:16px;padding:2rem;text-align:center;margin-bottom:2rem}
.qr-placeholder{width:200px;height:200px;background:#fff;border-radius:8px;margin:0 auto 1rem;display:flex;align-items:center;justify-content:center;overflow:hidden}
.qr-placeholder img{width:100%%;height:100%%}
.url{font-size:1.4rem;font-weight:600;color:#00d4ff;word-break:break-all;margin:.5rem 0}
.url-label{font-size:.9rem;color:#666;margin-bottom:.5rem}
.join-form{background:#111;border:2px solid #222;border-radius:16px;padding:2rem;width:100%%;max-width:400px;margin-bottom:2rem}
.join-form h2{font-size:1.2rem;margin-bottom:1rem;color:#ccc}
.join-form input{width:100%%;padding:.8rem 1rem;border:2px solid #333;border-radius:8px;background:#1a1a2e;color:#fff;font-size:1rem;margin-bottom:1rem;outline:none}
.join-form input:focus{border-color:#00d4ff}
.join-form button{width:100%%;padding:.8rem;border:none;border-radius:8px;background:linear-gradient(135deg,#00d4ff,#7b2ff7);color:#fff;font-size:1.1rem;font-weight:600;cursor:pointer;transition:transform .1s}
.join-form button:hover{transform:scale(1.02)}
.join-form button:active{transform:scale(.98)}
.product-info{font-size:.9rem;color:#666;margin-top:.5rem}
.merchants{width:100%%;max-width:400px}
.merchants h3{color:#888;font-size:.9rem;margin-bottom:.5rem;text-transform:uppercase;letter-spacing:.1em}
#merchant-list{list-style:none}
#merchant-list li{padding:.5rem .8rem;background:#111;border-radius:8px;margin-bottom:.3rem;display:flex;justify-content:space-between;font-size:.9rem}
.merchant-price{color:#00d4ff;font-weight:600}
.error{color:#ff4444;font-size:.9rem;margin-bottom:.5rem;display:none}
</style>
</head>
<body>
<div class="logo">UCP Arena</div>
<div class="subtitle">Devenez marchand, fixez votre prix, vendez !</div>

<div class="qr-section">
<div class="qr-placeholder">
<img src="https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s/register-page" alt="QR Code" onerror="this.parentElement.textContent='QR'">
</div>
<div class="url-label">Scannez ou rendez-vous sur</div>
<div class="url">%s</div>
</div>

<div class="join-form">
<h2>Rejoindre l'arene</h2>
<div class="product-info">Produit : %s | Prix d'achat : $%.2f</div>
<br>
<div class="error" id="error"></div>
<input type="text" id="name" placeholder="Votre nom ou pseudo" autocomplete="off">
<button onclick="register()">Rejoindre</button>
</div>

<div class="merchants">
<h3>Marchands en ligne</h3>
<ul id="merchant-list"></ul>
</div>

<script>
const BASE='%s';
async function register(){
  const name=document.getElementById('name').value.trim();
  if(!name){document.getElementById('error').style.display='block';document.getElementById('error').textContent='Entrez votre nom';return}
  try{
    const r=await fetch(BASE+'/register',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name})});
    const d=await r.json();
    if(!r.ok){document.getElementById('error').style.display='block';document.getElementById('error').textContent=d.detail||'Erreur';return}
    window.location.href=d.dashboard;
  }catch(e){document.getElementById('error').style.display='block';document.getElementById('error').textContent='Erreur de connexion'}
}
document.getElementById('name').addEventListener('keydown',e=>{if(e.key==='Enter')register()});
async function refreshMerchants(){
  try{
    const r=await fetch(BASE+'/merchants');
    const d=await r.json();
    const list=document.getElementById('merchant-list');
    list.innerHTML='';
    (d.merchants||[]).forEach(m=>{
      const li=document.createElement('li');
      li.innerHTML='<span>'+m.name+'</span><span class="merchant-price">$'+(m.price/100).toFixed(2)+'</span>';
      list.appendChild(li);
    });
  }catch(e){}
}
refreshMerchants();
setInterval(refreshMerchants,3000);
</script>
</body>
</html>`
