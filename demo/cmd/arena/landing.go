package main

import (
	"fmt"
	"net/http"
)

func (s *ArenaServer) handleLanding(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	arenaURL := fmt.Sprintf("%s://%s", scheme, r.Host)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	fmt.Fprintf(w, landingHTML, arenaURL, arenaURL, s.productName, float64(s.costPrice)/100)
}

const landingHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>UCP Arena</title>
<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;600;700;800&display=swap" rel="stylesheet">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'Outfit',system-ui,sans-serif;background:#FDF0EE;color:#1A1A2E;min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:2rem}
.logo{font-size:3rem;font-weight:800;color:#1A1A2E;margin-bottom:.5rem}
.logo span{color:#E5004C}
.subtitle{font-size:1.2rem;color:#666;margin-bottom:2rem}
.win-card{background:#FFFFFF;border:1px solid #2D2D2D;border-radius:16px;overflow:hidden;box-shadow:6px 6px 0px #E5004C;margin-bottom:2rem}
.win-dots{padding:.5rem 1rem;border-bottom:1px solid #E0E0E0;display:flex;align-items:center;gap:6px}
.win-dots::before{content:'';width:10px;height:10px;border-radius:50%%;background:#E5004C;display:inline-block}
.win-dots::after{content:'';width:10px;height:10px;border-radius:50%%;background:#CCC;display:inline-block}
.win-body{padding:2rem}
.qr-section{text-align:center}
.qr-placeholder{width:200px;height:200px;background:#fff;border-radius:8px;margin:0 auto 1rem;display:flex;align-items:center;justify-content:center;overflow:hidden;border:1px solid #E0E0E0}
.qr-placeholder img{width:100%%;height:100%%}
.url{font-size:1.4rem;font-weight:700;color:#E5004C;word-break:break-all;margin:.5rem 0}
.url-label{font-size:.9rem;color:#999;margin-bottom:.5rem}
.join-form{width:100%%;max-width:400px}
.join-form h2{font-size:1.2rem;margin-bottom:1rem;color:#1A1A2E;font-weight:700}
.join-form input{width:100%%;padding:.8rem 1rem;border:1px solid #CCC;border-radius:8px;background:#FFFFFF;color:#1A1A2E;font-size:1rem;margin-bottom:1rem;outline:none;font-family:'Outfit',system-ui,sans-serif}
.join-form input:focus{border-color:#E5004C}
.join-form button{width:100%%;padding:.8rem;border:none;border-radius:8px;background:#E5004C;color:#fff;font-size:1.1rem;font-weight:700;cursor:pointer;transition:transform .1s,opacity .2s}
.join-form button:hover{opacity:.9;transform:scale(1.02)}
.join-form button:active{transform:scale(.98)}
.product-info{font-size:.9rem;color:#666;margin-top:.5rem}
.merchants{width:100%%;max-width:500px}
.merchants h3{color:#999;font-size:.9rem;margin-bottom:.5rem;text-transform:uppercase;letter-spacing:.1em;font-weight:700}
#merchant-list{list-style:none}
#merchant-list li{padding:.5rem .8rem;background:#FFFFFF;border:1px solid #E0E0E0;border-radius:8px;margin-bottom:.3rem;display:flex;justify-content:space-between;align-items:center;font-size:.9rem;transition:background .5s ease,border-color .5s ease}
#merchant-list li .m-rank{color:#999;font-size:.75rem;width:2rem;text-align:center;flex-shrink:0}
#merchant-list li .m-name{flex:1;margin:0 .5rem;color:#1A1A2E}
.merchant-price{color:#E5004C;font-weight:600}
#merchant-list li.state-checkout{background:#EFF6FF;border-color:#3B82F6}
#merchant-list li.state-negotiate{background:#FFFBEB;border-color:#F59E0B}
#merchant-list li.state-sale{background:#F0FDF4;border-color:#16A34A}
.error{color:#DC2626;font-size:.9rem;margin-bottom:.5rem;display:none}
</style>
</head>
<body>
<div class="logo">UCP <span>Arena</span></div>
<div class="subtitle">Devenez marchand, fixez votre prix, vendez !</div>

<div class="win-card">
<div class="win-dots"></div>
<div class="win-body qr-section">
<div class="qr-placeholder">
<img src="https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s/auto" alt="QR Code" onerror="this.parentElement.textContent='QR'">
</div>
<div class="url-label">Scannez ou rendez-vous sur</div>
<div class="url">%s</div>
</div>
</div>

<div class="win-card">
<div class="win-dots"></div>
<div class="win-body join-form">
<h2>Rejoindre l'arene</h2>
<div class="product-info">Produit : %s | Prix d'achat : $%.2f<br>Encherissez pour plus de visibilite !</div>
<br>
<div class="error" id="error"></div>
<input type="text" id="name" placeholder="Votre nom ou pseudo" autocomplete="off">
<button onclick="register()">Rejoindre</button>
</div>
</div>

<div class="merchants">
<h3>Marchands en ligne</h3>
<ul id="merchant-list"></ul>
</div>

<script>
let merchants=[];
let rankings={};
let merchantStates={};
let stateTimers={};

async function register(){
  const name=document.getElementById('name').value.trim();
  if(!name){document.getElementById('error').style.display='block';document.getElementById('error').textContent='Entrez votre nom';return}
  try{
    const r=await fetch('/register',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name})});
    const d=await r.json();
    if(!r.ok){document.getElementById('error').style.display='block';document.getElementById('error').textContent=d.detail||'Erreur';return}
    window.location.href=d.dashboard;
  }catch(e){document.getElementById('error').style.display='block';document.getElementById('error').textContent='Erreur de connexion'}
}
document.getElementById('name').addEventListener('keydown',e=>{if(e.key==='Enter')register()});

function renderMerchants(){
  const sorted=[...merchants].sort((a,b)=>{
    const ra=rankings[a.id]?rankings[a.id].rank:9999;
    const rb=rankings[b.id]?rankings[b.id].rank:9999;
    if(ra!==rb)return ra-rb;
    return a.id<b.id?-1:1;
  });
  const list=document.getElementById('merchant-list');
  // Build a set of current IDs for cleanup
  const currentIds=new Set(sorted.map(m=>m.id));
  // Remove LIs that no longer exist
  list.querySelectorAll('li').forEach(li=>{
    const mid=li.id.replace('m-','');
    if(!currentIds.has(mid))li.remove();
  });
  // Update or create LIs in order
  let prevEl=null;
  sorted.forEach(m=>{
    let li=document.getElementById('m-'+m.id);
    if(!li){
      li=document.createElement('li');
      li.id='m-'+m.id;
    }
    const rData=rankings[m.id];
    const rank=(rData&&typeof rData==='object'&&typeof rData.rank==='number')?rData.rank:0;
    const graphPrice=(rData&&typeof rData==='object')?rData.price||0:0;
    const medals={1:'🥇',2:'🥈',3:'🥉'};
    const rankText=rank>0?(medals[rank]||'#'+rank):'-';
    const bidInfo=m.max_cpc_bid>0?' (bid: $'+(m.max_cpc_bid/100).toFixed(2)+')':'';
    let priceLabel='$'+(m.price/100).toFixed(2);
    if(graphPrice>0 && graphPrice!==m.price){
      priceLabel+=' (graph: $'+(graphPrice/100).toFixed(2)+')';
    }
    li.innerHTML='<span class="m-rank">'+rankText+'</span><span class="m-name">'+m.name+'</span><span class="merchant-price">'+priceLabel+bidInfo+'</span>';
    // Preserve state class
    const state=merchantStates[m.id];
    li.className=state||'';
    // Insert in correct order
    if(prevEl){
      if(prevEl.nextSibling!==li)prevEl.after(li);
    } else {
      if(list.firstChild!==li)list.prepend(li);
    }
    prevEl=li;
  });
}

function setMerchantState(mid,cls){
  clearTimeout(stateTimers[mid]);
  merchantStates[mid]=cls;
  const el=document.getElementById('m-'+mid);
  if(el)el.className=cls;
  stateTimers[mid]=setTimeout(()=>{
    delete merchantStates[mid];
    const el2=document.getElementById('m-'+mid);
    if(el2)el2.className='';
  },5000);
}

async function fetchMerchants(){
  try{
    const r=await fetch('/merchants');
    const d=await r.json();
    merchants=d.merchants||[];
    renderMerchants();
  }catch(e){}
}

async function fetchRankings(){
  try{
    const r=await fetch('/rankings');
    const d=await r.json();
    rankings=d.rankings||{};
    renderMerchants();
  }catch(e){}
}

// Initial load
fetchRankings().then(fetchMerchants);

// SSE for live updates (lazy: only when tab visible)
let evtSrc=null;
function sseHandler(e){
  try{
    const d=JSON.parse(e.data);
    if(d.type==='registration'||d.type==='config_update'){
      fetchRankings().then(fetchMerchants);
    } else if(d.type==='sale'){
      setMerchantState(d.merchant_id,'state-sale');
      fetchMerchants();
    } else if(d.type==='checkout_created'){
      setMerchantState(d.merchant_id,'state-checkout');
    } else if(d.type==='checkout_updated'){
      setMerchantState(d.merchant_id,'state-negotiate');
    } else if(d.type==='checkout_canceled'){
      delete merchantStates[d.merchant_id];
      const el=document.getElementById('m-'+d.merchant_id);
      if(el)el.className='';
    }
  }catch(ex){}
}
function sseConnect(){
  if(evtSrc)return;
  evtSrc=new EventSource('/events');
  evtSrc.addEventListener('message',sseHandler);
}
function sseDisconnect(){
  if(evtSrc){evtSrc.close();evtSrc=null}
}
document.addEventListener('visibilitychange',()=>{
  if(document.hidden){sseDisconnect()}else{sseConnect();fetchRankings().then(fetchMerchants)}
});
if(!document.hidden){sseConnect()}

// Periodic refresh
setInterval(()=>fetchRankings().then(fetchMerchants),3000);
</script>
</body>
</html>`
