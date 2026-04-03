package main

import (
	"fmt"
	"net/http"
)

func serveDashboard(w http.ResponseWriter, r *http.Request, tenantID, merchantName string, costPrice int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, dashboardHTML, merchantName, merchantName, tenantID, costPrice, float64(costPrice)/100, tenantID, costPrice)
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
<title>%s - Dashboard</title>
<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;600;700;800&display=swap" rel="stylesheet">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'Outfit',system-ui,sans-serif;background:#FDF0EE;color:#1A1A2E;overflow:hidden;height:100vh;height:100dvh}

/* === TOP PANEL === */
.top-panel{position:fixed;top:0;left:0;right:0;height:50vh;height:50dvh;background:#FDF0EE;padding:.8rem 1rem;overflow-y:auto;z-index:10;display:flex;flex-direction:column;gap:.6rem;transition:background .4s ease}
.top-panel.flash-checkout{background:#DBEAFE}
.top-panel.flash-negotiate{background:#FFF7ED}
.top-panel.flash-sale{background:#DCFCE7}

/* === BOTTOM PANEL === */
.bottom-panel{position:fixed;bottom:0;left:0;right:0;height:50vh;height:50dvh;background:#FFFFFF;border-top:2px solid #E5004C;z-index:10;display:flex;flex-direction:column;padding-bottom:env(safe-area-inset-bottom)}
.bottom-header{display:flex;align-items:center;justify-content:space-between;padding:.4rem .8rem;border-bottom:1px solid #F0F0F0;flex-shrink:0}
.bottom-header h2{font-size:1.1rem;color:#999;text-transform:uppercase;letter-spacing:.05em;font-weight:700}
.activity-log{flex:1;overflow-y:auto;padding:.3rem .8rem;font-size:1rem}

/* === METRICS ROW === */
.metrics-row{display:flex;align-items:center;gap:.8rem;flex-wrap:wrap}
.rank-badge{display:inline-flex;align-items:center;gap:.3rem;padding:.4rem 1rem;border-radius:14px;font-size:1.1rem;font-weight:700;background:#FDE8E8;color:#E5004C;border:1px solid #E5004C;flex-shrink:0}
.rank-badge .rank-medal{font-size:1.5rem}
.rank-badge.rank-1{background:#FEF9C3;border-color:#CA8A04;color:#854D0E}
.rank-badge.rank-2{background:#F1F5F9;border-color:#94A3B8;color:#475569}
.rank-badge.rank-3{background:#FFF7ED;border-color:#EA580C;color:#9A3412}
.rank-badge.rank-none{background:#F3F4F6;border-color:#D1D5DB;color:#6B7280}
.metric-pill{display:inline-flex;align-items:center;gap:.2rem;padding:.4rem .8rem;border-radius:14px;font-size:1.1rem;font-weight:700;background:#fff;border:1px solid #E0E0E0;flex-shrink:0}
.metric-pill.profit-positive{color:#16A34A;border-color:#16A34A}
.metric-pill.profit-negative{color:#DC2626;border-color:#DC2626}
.metric-pill.sales{color:#3B82F6;border-color:#3B82F6}

/* === STOCK WARNING === */
.stock-warning{display:none;background:#DC2626;color:#fff;text-align:center;padding:.6rem;border-radius:8px;font-weight:700;font-size:1rem;animation:pulse-warn 1s ease-in-out infinite alternate}
@keyframes pulse-warn{from{opacity:1}to{opacity:.7}}

/* === SLIDER ROWS === */
.slider-row{display:flex;align-items:center;gap:.6rem}
.slider-row label{width:80px;font-size:1rem;color:#666;flex-shrink:0;font-weight:600}
.slider-row .value{width:90px;font-size:1.3rem;font-weight:800;color:#E5004C;text-align:right;flex-shrink:0}
.slider-row input[type=range]{flex:1;accent-color:#E5004C;min-width:0}
.slider-row input[type=range].algo-active{accent-color:#999;pointer-events:none}
.slider-sub{display:flex;justify-content:space-between;font-size:.85rem;color:#999;padding:0 0 0 145px}
.cost-info-inline{font-size:.85rem;color:#D97706;text-align:center}

/* === ALGO BAR === */
.algo-bar{display:flex;gap:.5rem;flex-wrap:wrap;justify-content:center}
.algo-btn{position:relative;padding:.4rem .8rem;border:1px solid #E5004C;border-radius:20px;background:#FDE8E8;color:#E5004C;font-size:.85rem;font-weight:600;cursor:pointer;font-family:'Outfit',system-ui,sans-serif;transition:all .2s}
.algo-btn:hover{background:#E5004C;color:#fff}
.algo-btn.active{background:#E5004C;color:#fff;box-shadow:0 2px 8px rgba(229,0,76,.3)}
.algo-btn .algo-tooltip{display:none;position:absolute;bottom:calc(100%% + 8px);left:50%%;transform:translateX(-50%%);background:#1A1A2E;color:#fff;padding:.4rem .6rem;border-radius:8px;font-size:.8rem;font-weight:400;width:240px;text-align:center;z-index:100;line-height:1.3}
.algo-btn .algo-tooltip::after{content:'';position:absolute;top:100%%;left:50%%;transform:translateX(-50%%);border:6px solid transparent;border-top-color:#1A1A2E}
.algo-btn:hover .algo-tooltip{display:block}

/* === FUNNEL === */
.funnel{display:flex;align-items:center;justify-content:center;gap:.5rem;flex-wrap:wrap;font-size:1rem}
.funnel-step{text-align:center;padding:.3rem .7rem;border-radius:10px;font-weight:600}
.funnel-step .funnel-val{font-weight:800;margin-right:.15rem}
.funnel-step.f-visits{background:#FDE8E8;color:#E5004C}
.funnel-step.f-checkouts{background:#DBEAFE;color:#3B82F6}
.funnel-step.f-sales{background:#DCFCE7;color:#16A34A}
.funnel-arrow{color:#999;font-size:1.1rem;font-weight:700}
.funnel-rate{font-size:.85rem;color:#999;font-weight:600}

/* === ACTIVITY ENTRIES === */
.activity-entry{padding:.5rem .8rem;margin-bottom:.3rem;border-radius:8px;display:flex;align-items:baseline;gap:.4rem}
.activity-entry .act-time{color:#999;font-size:.85rem;flex-shrink:0}
.activity-entry .act-text{flex:1}
.activity-entry .act-cost{font-size:.9rem;font-weight:700;flex-shrink:0}
.activity-entry.act-catalog{background:#FDE8E8;border-left:3px solid #E5004C;color:#E5004C}
.activity-entry.act-checkout{background:#DBEAFE;border-left:3px solid #3B82F6;color:#3B82F6}
.activity-entry.act-cart{background:#F3F4F6;border-left:3px solid #9CA3AF;color:#6B7280}
.activity-entry.act-cancel{background:#FEF2F2;border-left:3px solid #DC2626;color:#DC2626}
.activity-entry.act-sale{background:#DCFCE7;border-left:3px solid #16A34A;color:#16A34A}

/* === PROMO SECTION === */
.promo-section-title{font-size:.95rem;color:#999;text-transform:uppercase;letter-spacing:.05em;font-weight:700;margin:.5rem 0 .3rem}

/* === DISCOUNT ROWS === */
.discount-row{display:flex;gap:.4rem;align-items:center;margin-bottom:.4rem;flex-wrap:wrap}
.discount-row input,.discount-row select{padding:.35rem;border:1px solid #CCC;border-radius:8px;background:#FFFFFF;color:#1A1A2E;font-size:.95rem;font-family:'Outfit',system-ui,sans-serif}
.discount-row input[type=text]{flex:1;min-width:70px}
.discount-row input[type=number]{width:55px}
.discount-row select{width:80px}
.discount-row label.cb{display:flex;align-items:center;gap:.2rem;font-size:.85rem;color:#666;white-space:nowrap}
.discount-usage{font-size:.8rem;color:#16A34A;font-weight:600;white-space:nowrap}
.btn-sm{padding:.25rem .5rem;border:none;border-radius:8px;cursor:pointer;font-size:.9rem;font-weight:600}
.btn-add{background:#FDE8E8;color:#E5004C;border:1px solid #E5004C}
.btn-rm{background:#FEF2F2;color:#DC2626;border:1px solid #DC2626}
.algo-auto-tag{display:inline-block;background:#E5004C;color:#fff;font-size:.7rem;font-weight:700;padding:.1rem .3rem;border-radius:6px;margin-left:.2rem;vertical-align:middle}

/* === SAVE STATUS === */
.save-status{text-align:center;font-size:.85rem;color:#16A34A;opacity:0;transition:opacity .3s}

/* === TREND === */
.trend{font-size:.8rem;font-weight:600;margin-left:.2rem;opacity:.8}
.trend.trend-up{color:#16A34A}
.trend.trend-down{color:#DC2626}

/* === CONNECTION DOT === */
.conn-dot{width:8px;height:8px;border-radius:50%%;display:inline-block;margin-left:.4rem;vertical-align:middle}
.conn-dot.connected{background:#16A34A}
.conn-dot.disconnected{background:#DC2626;animation:pulse-conn 1s ease-in-out infinite}
@keyframes pulse-conn{0%%,100%%{opacity:1}50%%{opacity:.3}}

/* === CONFETTI CANVAS === */
#confetti-canvas{position:absolute;top:0;left:0;width:100%%;height:100%%;pointer-events:none}

/* === SALE OVERLAY === */
.sale-overlay{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(229,0,76,.95);display:none;align-items:center;justify-content:center;flex-direction:column;z-index:9999;animation:pulse .5s ease-in-out infinite alternate;transition:transform .25s ease,opacity .25s ease;touch-action:none}
.sale-overlay.dismissing{transform:translateY(-100%%);opacity:0}
.sale-overlay h1{font-size:3rem;color:#fff;font-weight:900;text-shadow:0 4px 20px rgba(0,0,0,.3)}
.sale-overlay p{font-size:1.2rem;color:#fff;margin-top:.8rem}
.sale-overlay .sale-breakdown{font-size:1rem;color:rgba(255,255,255,.9);margin-top:.6rem;line-height:1.8}
.sale-overlay .sale-profit{font-size:1.8rem;font-weight:800;margin-top:.4rem}
.sale-overlay .sale-profit.positive{color:#BBF7D0}
.sale-overlay .sale-profit.negative{color:#FCA5A5}
.sale-overlay .swipe-hint{font-size:.8rem;color:rgba(255,255,255,.6);margin-top:1.2rem;animation:bounce-hint 1.5s ease-in-out infinite}
@keyframes pulse{from{background:rgba(229,0,76,.95)}to{background:rgba(253,232,232,.95)}}
@keyframes bounce-hint{0%%,100%%{transform:translateY(0)}50%%{transform:translateY(-6px)}}

/* === TABLET / DESKTOP === */
@media(min-width:768px){
  .top-panel{height:100vh;height:100dvh;width:50vw;right:auto}
  .bottom-panel{height:100vh;height:100dvh;width:50vw;left:auto;border-top:none;border-left:2px solid #E5004C}
  .top-panel{padding:1.2rem 1.5rem;gap:.8rem}
  .slider-row .value{font-size:1.4rem}
  .algo-btn{font-size:.95rem;padding:.4rem .9rem}
}
@media(min-width:1200px){
  .rank-badge{font-size:1.4rem;padding:.5rem 1.2rem}
  .rank-badge .rank-medal{font-size:2rem}
  .metric-pill{font-size:1.4rem;padding:.5rem 1rem}
  .slider-row label{font-size:1.1rem;width:100px}
  .slider-row .value{font-size:1.6rem;width:110px}
  .funnel{font-size:1.2rem;gap:.8rem}
  .funnel-step{padding:.4rem 1rem;border-radius:12px}
  .funnel-arrow{font-size:1.4rem}
  .funnel-rate{font-size:1rem}
  .algo-btn{font-size:1rem;padding:.5rem 1rem}
  .activity-log{font-size:1.1rem}
  .activity-entry{padding:.6rem 1rem}
  .activity-entry .act-time{font-size:1rem}
  .activity-entry .act-cost{font-size:1rem}
  .bottom-header h2{font-size:1.3rem}
}
</style>
</head>
<body>

<!-- TOP PANEL: Controls & Metrics -->
<div class="top-panel" id="top-panel">

  <!-- Merchant name -->
  <div style="text-align:center;font-size:1.3rem;font-weight:800;color:#1A1A2E">%s <span style="font-size:.8rem;color:#999;font-weight:400">%s</span><span class="conn-dot disconnected" id="conn-dot" title="SSE"></span></div>

  <!-- Metrics row -->
  <div class="metrics-row">
    <div class="rank-badge rank-none" id="rank-badge">
      <span class="rank-medal" id="rank-medal">-</span>
      <span id="rank-text">...</span>
    </div>
    <div class="metric-pill" id="profit-pill">
      <span id="profit-display">$0.00</span><span class="trend" id="profit-trend"></span>
    </div>
    <div class="metric-pill sales" id="sales-pill">
      <span id="sales-count-display">0 vt</span>
    </div>
  </div>

  <!-- Stock warning -->
  <div class="stock-warning" id="stock-warning">Stock epuise — invisible dans le Graph !</div>

  <!-- Price slider -->
  <div class="slider-row">
    <label>Prix</label>
    <div class="value" id="price-display">--</div>
    <input type="range" id="price-slider" min="%d" max="30000" step="100" value="6000">
  </div>
  <div class="cost-info-inline">Achat: $%.2f (min)</div>

  <!-- Stock slider -->
  <div class="slider-row">
    <label>Stock</label>
    <div class="value" id="stock-display">--</div>
    <input type="range" id="stock-slider" min="0" max="50" step="1" value="10">
  </div>

  <!-- Bid slider -->
  <div class="slider-row">
    <label>Enchere</label>
    <div class="value" id="bid-display">--</div>
    <input type="range" id="bid-slider" min="0" max="200" step="5" value="50">
  </div>
  <div class="slider-sub">
    <span id="bid-cpc-info"></span>
    <span id="bid-spend-info" style="color:#16A34A"></span>
  </div>

  <!-- Algo bar -->
  <div class="algo-bar" id="algo-bar">
    <button class="algo-btn active" data-algo="manual" onclick="setAlgo('manual')">Manuel<span class="algo-tooltip">Controle libre du prix, encheres et promos.</span></button>
    <button class="algo-btn" data-algo="markup" onclick="setAlgo('markup')">Markup<span class="algo-tooltip">Stock eleve = prix bas + encheres hautes + promo auto. Stock bas = prix haut.</span></button>
    <button class="algo-btn" data-algo="threshold" onclick="setAlgo('threshold')">Seuils<span class="algo-tooltip">5 paliers selon le stock. Promos et encheres auto.</span></button>
    <button class="algo-btn" data-algo="markdown" onclick="setAlgo('markdown')">Demarque<span class="algo-tooltip">Faible conversion = prix baisse + encheres montent + promo auto.</span></button>
    <button class="algo-btn" data-algo="surge" onclick="setAlgo('surge')">Surge<span class="algo-tooltip">Forte demande = prix monte, encheres baissent. Aucune promo.</span></button>
  </div>

  <!-- Funnel -->
  <div class="funnel" id="funnel">
    <div class="funnel-step f-visits"><span class="funnel-val" id="f-visits">0</span>vis</div>
    <div class="funnel-arrow">&rarr;</div>
    <div class="funnel-step f-checkouts"><span class="funnel-val" id="f-checkouts">0</span>chk</div>
    <div class="funnel-rate" id="f-rate1"></div>
    <div class="funnel-arrow">&rarr;</div>
    <div class="funnel-step f-sales"><span class="funnel-val" id="f-sales">0</span>vt</div>
    <div class="funnel-rate" id="f-rate2"></div>
  </div>

  <!-- Promos -->
  <div class="promo-section-title">Codes promo</div>
  <div id="discounts"></div>
  <button class="btn-sm btn-add" onclick="addDiscount()" style="margin-top:.3rem">+ Ajouter</button>

  <!-- Shipping -->
  <div class="promo-section-title">Livraison</div>
  <div id="shipping-options"></div>

  <div class="save-status" id="save-status">Sauvegarde...</div>
  <button class="btn-sm" id="leave-btn" onclick="leaveArena()" style="margin-top:.5rem;width:100%%;padding:.6rem;background:#DC2626;color:#fff;border:none;border-radius:8px;font-size:1rem;font-weight:700;cursor:pointer">Quitter l'arene</button>
</div>

<!-- BOTTOM PANEL: Activity Log -->
<div class="bottom-panel">
  <div class="bottom-header">
    <h2>Activite</h2>
  </div>
  <div class="activity-log" id="activity-log"></div>
</div>

<!-- SALE OVERLAY -->
<div class="sale-overlay" id="sale-overlay">
<canvas id="confetti-canvas"></canvas>
<h1>VENDU !</h1>
<p id="sale-detail"></p>
<div class="sale-breakdown" id="sale-breakdown"></div>
<div class="sale-profit" id="sale-profit"></div>
<div class="swipe-hint">swipe ou tap pour fermer</div>
</div>

<script>
const TID='%s';
const API='/'+TID+'/api';
const COST_PRICE=%d;

let config={selling_price:6000,stock:10,discount_codes:[],max_cpc_bid:50,pricing_algo:'manual'};
let saveTimer=null;
let checkoutCount=0;
let discountUsage={};
let currentAlgo='manual';
const MAX_STOCK_REF=20;

// Trend tracking
let prevProfit=null;

// Haptic patterns per event type
function haptic(type){
  if(!navigator.vibrate)return;
  const patterns={
    catalog:[30],
    cart:[40,30,40],
    checkout:[60,40,60],
    cancel:[100],
    sale:[200,100,200,100,200]
  };
  navigator.vibrate(patterns[type]||[30]);
}

// --- Audio feedback (Web Audio API) ---
let audioCtx;
function initAudio(){if(!audioCtx)try{audioCtx=new(window.AudioContext||window.webkitAudioContext)()}catch(e){}}
document.addEventListener('touchstart',initAudio,{once:true});
document.addEventListener('click',initAudio,{once:true});
function playKaChing(){
  if(!audioCtx)return;
  const now=audioCtx.currentTime;
  const g=audioCtx.createGain();g.gain.setValueAtTime(0.3,now);g.gain.exponentialRampToValueAtTime(0.01,now+0.4);g.connect(audioCtx.destination);
  const o1=audioCtx.createOscillator();o1.type='sine';o1.frequency.value=523.25;o1.connect(g);o1.start(now);o1.stop(now+0.12);
  const o2=audioCtx.createOscillator();o2.type='sine';o2.frequency.value=659.25;o2.connect(g);o2.start(now+0.1);o2.stop(now+0.35);
  const o3=audioCtx.createOscillator();o3.type='sine';o3.frequency.value=783.99;o3.connect(g);o3.start(now+0.2);o3.stop(now+0.4);
}
function playTick(){
  if(!audioCtx)return;
  const now=audioCtx.currentTime;
  const g=audioCtx.createGain();g.gain.setValueAtTime(0.1,now);g.gain.exponentialRampToValueAtTime(0.01,now+0.08);g.connect(audioCtx.destination);
  const o=audioCtx.createOscillator();o.type='sine';o.frequency.value=880;o.connect(g);o.start(now);o.stop(now+0.06);
}

// --- Confetti ---
function launchConfetti(){
  const canvas=document.getElementById('confetti-canvas');
  if(!canvas)return;
  const ov=document.getElementById('sale-overlay');
  canvas.width=ov.offsetWidth;canvas.height=ov.offsetHeight;
  const ctx=canvas.getContext('2d');
  const colors=['#FFD700','#FFFFFF','#16A34A','#3B82F6','#F97316','#EC4899'];
  const particles=[];
  for(let i=0;i<60;i++){
    particles.push({x:canvas.width/2,y:canvas.height/2,vx:(Math.random()-0.5)*12,vy:(Math.random()-0.5)*12-4,size:Math.random()*6+3,color:colors[Math.floor(Math.random()*colors.length)],rot:Math.random()*Math.PI*2,rotV:(Math.random()-0.5)*0.3,alpha:1});
  }
  const start=performance.now();
  function frame(now){
    const elapsed=(now-start)/1000;
    if(elapsed>2.5){ctx.clearRect(0,0,canvas.width,canvas.height);return}
    ctx.clearRect(0,0,canvas.width,canvas.height);
    for(const p of particles){
      p.x+=p.vx;p.y+=p.vy;p.vy+=0.25;p.vx*=0.99;p.rot+=p.rotV;
      p.alpha=Math.max(0,1-elapsed/2.5);
      ctx.save();ctx.translate(p.x,p.y);ctx.rotate(p.rot);ctx.globalAlpha=p.alpha;
      ctx.fillStyle=p.color;ctx.fillRect(-p.size/2,-p.size/2,p.size,p.size*0.6);
      ctx.restore();
    }
    requestAnimationFrame(frame);
  }
  requestAnimationFrame(frame);
}

// --- Animated counters ---
function animateValue(el,from,to,dur,fmt){
  if(from===to)return;
  const start=performance.now();
  function step(now){
    const t=Math.min((now-start)/dur,1);
    const eased=t<0.5?2*t*t:1-Math.pow(-2*t+2,2)/2;
    el.textContent=fmt(Math.round(from+(to-from)*eased));
    if(t<1)requestAnimationFrame(step);
  }
  requestAnimationFrame(step);
}
const fmtDollar=v=>(v>=0?'+':'')+' $'+(v/100).toFixed(2);
const fmtInt=v=>''+v;

// Swipe-to-dismiss overlay
(function(){
  const ov=document.getElementById('sale-overlay');
  let startY=0;
  ov.addEventListener('touchstart',function(e){startY=e.touches[0].clientY},{passive:true});
  ov.addEventListener('touchmove',function(e){
    const dy=startY-e.touches[0].clientY;
    if(dy>60)dismissSale();
  },{passive:true});
  ov.addEventListener('click',function(){dismissSale()});
})();
let saleAutoTimer=null;
function dismissSale(){
  clearTimeout(saleAutoTimer);
  const ov=document.getElementById('sale-overlay');
  ov.classList.add('dismissing');
  setTimeout(()=>{ov.style.display='none';ov.classList.remove('dismissing');loadConfig()},250);
}

async function loadConfig(){
  try{
    const r=await fetch(API+'/config');
    config=await r.json();
    document.getElementById('price-slider').value=config.selling_price;
    document.getElementById('stock-slider').value=config.stock;
    document.getElementById('bid-slider').value=config.max_cpc_bid;
    updateDisplays();
    renderDiscounts();
    renderShipping();
    currentAlgo=config.pricing_algo||'manual';
    document.querySelectorAll('.algo-btn').forEach(b=>{b.classList.toggle('active',b.dataset.algo===currentAlgo)});
    if(config.accent_color){document.documentElement.style.setProperty('--accent',config.accent_color);document.getElementById('price-slider').style.accentColor=config.accent_color}
    const sl=document.getElementById('price-slider');
    const bl=document.getElementById('bid-slider');
    if(currentAlgo!=='manual'){sl.classList.add('algo-active');bl.classList.add('algo-active');recalcAlgoPrice()}
    else{sl.classList.remove('algo-active');bl.classList.remove('algo-active')}
  }catch(e){console.error('loadConfig:',e)}
}

function updateDisplays(){
  document.getElementById('price-display').textContent='$'+(config.selling_price/100).toFixed(2);
  document.getElementById('stock-display').textContent=config.stock;
  document.getElementById('bid-display').textContent='$'+(config.max_cpc_bid/100).toFixed(2);
  const cpcEl=document.getElementById('bid-cpc-info');
  const actualCPC=config.actual_cpc||0;
  const maxBid=config.max_cpc_bid||0;
  if(maxBid===0){
    cpcEl.textContent='Organique (gratuit)';
    cpcEl.style.color='#6B7280';
  } else {
    cpcEl.textContent='CPC: $'+(actualCPC/100).toFixed(2)+' / max $'+(maxBid/100).toFixed(2);
    cpcEl.style.color='#D97706';
  }
  const spendEl=document.getElementById('bid-spend-info');
  const totalAdSpend=config.total_ad_spend||0;
  spendEl.textContent='Pub: $'+(totalAdSpend/100).toFixed(2);
  if(totalAdSpend>0){spendEl.style.color='#D97706'}else{spendEl.style.color='#16A34A'}
  // Profit pill + trend (animated)
  if(config.net_profit!==undefined){
    const np=config.net_profit;
    const profitEl=document.getElementById('profit-display');
    const pill=document.getElementById('profit-pill');
    const trendEl=document.getElementById('profit-trend');
    const prevNP=profitEl._prevValue||0;
    if(prevNP!==np)animateValue(profitEl,prevNP,np,400,fmtDollar);
    else profitEl.textContent=fmtDollar(np);
    profitEl._prevValue=np;
    pill.className='metric-pill '+(np>=0?'profit-positive':'profit-negative');
    if(prevProfit!==null&&np!==prevProfit){
      const delta=np-prevProfit;
      const arrow=delta>0?'\u2191':'\u2193';
      trendEl.textContent=arrow+(delta>0?'+':'')+' $'+(delta/100).toFixed(2);
      trendEl.className='trend '+(delta>0?'trend-up':'trend-down');
      setTimeout(()=>{trendEl.textContent='';trendEl.className='trend'},8000);
    }
    prevProfit=np;
    const cc=config.consultation_count||0;
    const sc=config.sales_count||0;
    document.getElementById('sales-count-display').textContent=sc+' vt / '+cc+' vis';
    // Funnel (animated)
    const fv=document.getElementById('f-visits');
    const fc=document.getElementById('f-checkouts');
    const fs=document.getElementById('f-sales');
    const pvV=fv._prevValue||0,pvC=fc._prevValue||0,pvS=fs._prevValue||0;
    if(pvV!==cc)animateValue(fv,pvV,cc,300,fmtInt);fv._prevValue=cc;
    if(pvC!==checkoutCount)animateValue(fc,pvC,checkoutCount,300,fmtInt);fc._prevValue=checkoutCount;
    if(pvS!==sc)animateValue(fs,pvS,sc,300,fmtInt);fs._prevValue=sc;
    document.getElementById('f-rate1').textContent=cc>0?Math.round(checkoutCount/cc*100)+'%%':'';
    document.getElementById('f-rate2').textContent=checkoutCount>0?Math.round(sc/checkoutCount*100)+'%%':'';
  }
  // Stock warning
  const warn=document.getElementById('stock-warning');
  if(config.stock<=0){warn.style.display='block'}else{warn.style.display='none'}
}

function schedSave(){
  clearTimeout(saveTimer);
  const el=document.getElementById('save-status');
  el.textContent='Sauvegarde...';el.style.opacity='1';
  saveTimer=setTimeout(async()=>{
    try{
      const r=await fetch(API+'/config',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(config)});
      if(r.ok){el.textContent='OK';const d=await r.json();config=d;updateDisplays();setTimeout(()=>el.style.opacity='0',800);}
      else{el.textContent='Erreur';el.style.color='#DC2626';loadConfig();}
    }catch(e){el.textContent='Erreur';el.style.color='#DC2626'}
  },300);
}

function setAlgo(algo){
  currentAlgo=algo;
  config.pricing_algo=algo;
  document.querySelectorAll('.algo-btn').forEach(b=>{b.classList.toggle('active',b.dataset.algo===algo)});
  const ps=document.getElementById('price-slider');
  const bs=document.getElementById('bid-slider');
  if(algo==='manual'){
    ps.classList.remove('algo-active');
    bs.classList.remove('algo-active');
    removeAutoDiscount();
  }else{
    ps.classList.add('algo-active');
    bs.classList.add('algo-active');
    recalcAlgoPrice();
  }
  schedSave();
}

function removeAutoDiscount(){
  config.discount_codes=(config.discount_codes||[]).filter(dc=>!dc.code.startsWith('AUTO'));
  renderDiscounts();
}

function setAutoDiscount(pct){
  config.discount_codes=config.discount_codes||[];
  const idx=config.discount_codes.findIndex(dc=>dc.code.startsWith('AUTO'));
  if(pct<=0){
    if(idx>=0)config.discount_codes.splice(idx,1);
  }else{
    const dc={code:'AUTO'+pct,type:'percentage',value:pct,new_customer_only:false};
    if(idx>=0)config.discount_codes[idx]=dc;
    else config.discount_codes.push(dc);
  }
  renderDiscounts();
}

function recalcAlgoPrice(){
  if(currentAlgo==='manual')return;
  const stock=config.stock||0;
  const consult=config.consultation_count||0;
  const sales=config.sales_count||0;
  let newPrice=config.selling_price;
  let newBid=config.max_cpc_bid;
  let autoPct=0;

  if(currentAlgo==='markup'){
    const maxS=Math.max(MAX_STOCK_REF,stock);
    const ratio=maxS>0?stock/maxS:0;
    const adj=(1-ratio)*0.4-ratio*0.15;
    newPrice=Math.round(COST_PRICE*(1+0.30+adj));
    newBid=Math.round(ratio*150);
    if(ratio>0.6)autoPct=Math.round((ratio-0.6)*25);

  }else if(currentAlgo==='threshold'){
    const maxS=Math.max(MAX_STOCK_REF,stock);
    const pct=maxS>0?stock/maxS:0;
    if(pct>0.8){newPrice=Math.round(COST_PRICE*1.10);newBid=175;autoPct=15}
    else if(pct>0.6){newPrice=Math.round(COST_PRICE*1.20);newBid=125;autoPct=10}
    else if(pct>0.4){newPrice=Math.round(COST_PRICE*1.30);newBid=75;autoPct=5}
    else if(pct>0.2){newPrice=Math.round(COST_PRICE*1.45);newBid=25;autoPct=0}
    else{newPrice=Math.round(COST_PRICE*1.65);newBid=0;autoPct=0}

  }else if(currentAlgo==='markdown'){
    const ratio=consult>0?sales/consult:1;
    const disc=Math.min(0.25,(1-ratio)*0.3);
    newPrice=Math.round(COST_PRICE*(1.30-disc));
    newPrice=Math.max(newPrice,Math.round(COST_PRICE*1.05));
    newBid=Math.round((1-ratio)*175);
    if(ratio<0.5)autoPct=Math.round((0.5-ratio)*30);

  }else if(currentAlgo==='surge'){
    const sf=Math.min(2.0,1.0+consult*0.05);
    newPrice=Math.round(COST_PRICE*1.20*sf);
    newBid=Math.round(Math.max(0,150-consult*10));
    autoPct=0;
  }

  newPrice=Math.max(COST_PRICE,Math.min(30000,newPrice));
  newPrice=Math.round(newPrice/100)*100;
  if(newPrice<COST_PRICE)newPrice=COST_PRICE;
  newBid=Math.max(0,Math.min(200,Math.round(newBid/5)*5));

  config.selling_price=newPrice;
  config.max_cpc_bid=newBid;
  document.getElementById('price-slider').value=newPrice;
  document.getElementById('bid-slider').value=newBid;
  setAutoDiscount(autoPct);
  updateDisplays();
  schedSave();
}

document.getElementById('price-slider').oninput=function(){if(currentAlgo!=='manual')return;config.selling_price=parseInt(this.value);updateDisplays();schedSave()};
document.getElementById('stock-slider').oninput=function(){config.stock=parseInt(this.value);updateDisplays();if(currentAlgo!=='manual'){recalcAlgoPrice()}else{schedSave()}};
document.getElementById('bid-slider').oninput=function(){if(currentAlgo!=='manual')return;config.max_cpc_bid=parseInt(this.value);updateDisplays();schedSave()};

function renderDiscounts(){
  const el=document.getElementById('discounts');
  el.innerHTML='';
  (config.discount_codes||[]).forEach((dc,i)=>{
    const row=document.createElement('div');
    row.className='discount-row';
    const isAuto=dc.code.startsWith('AUTO');
    const usage=discountUsage[dc.code.toUpperCase()]||0;
    const usageBadge=usage>0?'<span class="discount-usage">('+usage+'x)</span>':'';
    if(isAuto){
      row.innerHTML=
        '<span style="color:#E5004C;font-weight:700;font-size:.8rem">'+dc.code+'</span>'+
        '<span class="algo-auto-tag">ALGO</span>'+
        '<span style="color:#666;font-size:.8rem;margin-left:auto">-'+dc.value+'%%</span>'+
        usageBadge;
    }else{
      row.innerHTML=
        '<input type="text" value="'+dc.code+'" placeholder="CODE" onchange="updDiscount('+i+',\'code\',this.value)">'+
        '<select onchange="updDiscount('+i+',\'type\',this.value)"><option value="percentage"'+(dc.type==='percentage'?' selected':'')+'>%%</option><option value="fixed"'+(dc.type==='fixed'?' selected':'')+'>Fixe</option></select>'+
        '<input type="number" value="'+dc.value+'" min="1" onchange="updDiscount('+i+',\'value\',parseInt(this.value))">'+
        usageBadge+
        '<label class="cb"><input type="checkbox"'+(dc.new_customer_only?' checked':'')+' onchange="updDiscount('+i+',\'new_customer_only\',this.checked)">Nouveau</label>'+
        '<button class="btn-sm btn-rm" onclick="rmDiscount('+i+')">X</button>';
    }
    el.appendChild(row);
  });
}

function addDiscount(){
  config.discount_codes=config.discount_codes||[];
  config.discount_codes.push({code:'PROMO',type:'percentage',value:10,new_customer_only:false});
  renderDiscounts();schedSave();
}
function updDiscount(i,k,v){config.discount_codes[i][k]=v;schedSave()}
function rmDiscount(i){config.discount_codes.splice(i,1);renderDiscounts();schedSave()}

function renderShipping(){
  const el=document.getElementById('shipping-options');
  el.innerHTML='';
  (config.shipping_options||[]).forEach((so)=>{
    const row=document.createElement('div');
    row.className='discount-row';
    row.innerHTML='<span style="color:#666;font-size:.8rem">'+so.title+'</span><span style="color:#E5004C;font-weight:700;margin-left:auto">$'+(so.cost/100).toFixed(2)+'</span>';
    el.appendChild(row);
  });
  if(!config.shipping_options||config.shipping_options.length===0){
    el.innerHTML='<div style="color:#555;font-size:.8rem">Aucune option configuree</div>';
  }
}

// Rank badge
async function fetchRank(){
  try{
    const r=await fetch('/rankings');
    const d=await r.json();
    const rankings=d.rankings||{};
    const myRank=rankings[TID];
    const badge=document.getElementById('rank-badge');
    const medal=document.getElementById('rank-medal');
    const text=document.getElementById('rank-text');
    const total=Object.keys(rankings).length;
    if(myRank&&typeof myRank==='object'&&myRank.rank>0){
      const rk=myRank.rank;
      const medals={1:'\uD83E\uDD47',2:'\uD83E\uDD48',3:'\uD83E\uDD49'};
      medal.textContent=medals[rk]||'#'+rk;
      text.textContent=rk+(rk===1?'er':'e')+'/'+total;
      badge.className='rank-badge'+(rk<=3?' rank-'+rk:'');
    } else {
      medal.textContent='-';
      text.textContent='--/'+total;
      badge.className='rank-badge rank-none';
    }
  }catch(e){}
}

// Background flash on events
let flashTimer=null;
function flashBg(cls){
  clearTimeout(flashTimer);
  const panel=document.getElementById('top-panel');
  panel.className='top-panel '+cls;
  flashTimer=setTimeout(()=>{panel.className='top-panel'},3000);
}

// SSE for sale + activity notifications
const actLog=document.getElementById('activity-log');
let actCount=0;
function addActivity(cls,text,costAnnotation){
  const now=new Date();
  const ts=now.getHours().toString().padStart(2,'0')+':'+now.getMinutes().toString().padStart(2,'0')+':'+now.getSeconds().toString().padStart(2,'0');
  const div=document.createElement('div');
  div.className='activity-entry '+cls;
  var timeEl=document.createElement('span');timeEl.className='act-time';timeEl.textContent=ts;
  var textEl=document.createElement('span');textEl.className='act-text';textEl.textContent=text;
  div.appendChild(timeEl);div.appendChild(textEl);
  if(costAnnotation){var costEl=document.createElement('span');costEl.className='act-cost';costEl.innerHTML=costAnnotation;div.appendChild(costEl)}
  actLog.appendChild(div);
  actLog.scrollTop=actLog.scrollHeight;
  actCount++;
  if(actCount>50){actLog.removeChild(actLog.firstChild);actCount--}
}
// SSE lifecycle with auto-reconnection + connection indicator
let evtSrc=null;
let sseRetryDelay=1000;
const connDot=document.getElementById('conn-dot');
function setConnected(ok){
  connDot.className='conn-dot '+(ok?'connected':'disconnected');
  connDot.title=ok?'Connecte':'Deconnecte';
}
// --- Sale celebration ---
let lastCelebratedSales=0;
function showSaleCelebration(saleTotal,orderID,summary){
  const profit=config.selling_price-COST_PRICE-(config.actual_cpc||0);
  const ov=document.getElementById('sale-overlay');
  document.getElementById('sale-detail').textContent='Commande '+(orderID||'?')+' - $'+((saleTotal||0)/100).toFixed(2);
  const bd=document.getElementById('sale-breakdown');
  bd.innerHTML='Prix: $'+(config.selling_price/100).toFixed(2)+'<br>Achat: -$'+(COST_PRICE/100).toFixed(2)+'<br>CPC: -$'+((config.actual_cpc||0)/100).toFixed(2);
  const sp=document.getElementById('sale-profit');
  sp.textContent=(profit>=0?'+':'')+' $'+(profit/100).toFixed(2);
  sp.className='sale-profit '+(profit>=0?'positive':'negative');
  ov.style.display='flex';ov.classList.remove('dismissing');
  haptic('sale');playKaChing();launchConfetti();
  clearTimeout(saleAutoTimer);
  saleAutoTimer=setTimeout(()=>{dismissSale()},8000);
  const profitStr=(profit>=0?'+':'-')+' $'+Math.abs(profit/100).toFixed(2);
  addActivity('act-sale',summary||('Vente: '+(orderID||'')),'<span style="color:#16A34A">'+profitStr+'</span>');
  flashBg('flash-sale');
}

function handleSSEMessage(e){
  try{
    const d=JSON.parse(e.data);
    if(d.type==='sale'){
      showSaleCelebration(d.total,d.order_id,d.summary);
      lastCelebratedSales=(config.sales_count||0)+1;
      loadConfig();
      if(d.summary){
        (config.discount_codes||[]).forEach(dc=>{
          if(d.summary.toUpperCase().indexOf(dc.code.toUpperCase())>=0){
            discountUsage[dc.code.toUpperCase()]=(discountUsage[dc.code.toUpperCase()]||0)+1;
            renderDiscounts();
          }
        });
      }
    } else if(d.type==='catalog_browse'||d.type==='product_details'){
      const cpcCost=(config.actual_cpc||0);
      const costStr=cpcCost>0?'-$'+(cpcCost/100).toFixed(2):'';
      addActivity('act-catalog',d.summary||d.type,costStr?'<span style="color:#D97706">'+costStr+'</span>':'');
      haptic('catalog');
    } else if(d.type==='checkout_created'){
      checkoutCount++;
      addActivity('act-checkout',d.summary||d.type);
      flashBg('flash-checkout');
      haptic('checkout');playTick();
      updateDisplays();
      if(d.summary){
        (config.discount_codes||[]).forEach(dc=>{
          if(d.summary.toUpperCase().indexOf(dc.code.toUpperCase())>=0){
            discountUsage[dc.code.toUpperCase()]=(discountUsage[dc.code.toUpperCase()]||0)+1;
            renderDiscounts();
          }
        });
      }
    } else if(d.type==='checkout_updated'){
      addActivity('act-checkout',d.summary||d.type);
      flashBg('flash-negotiate');
      haptic('checkout');playTick();
      if(d.summary){
        (config.discount_codes||[]).forEach(dc=>{
          if(d.summary.toUpperCase().indexOf(dc.code.toUpperCase())>=0){
            discountUsage[dc.code.toUpperCase()]=(discountUsage[dc.code.toUpperCase()]||0)+1;
            renderDiscounts();
          }
        });
      }
    } else if(d.type==='cart_created'){
      addActivity('act-cart',d.summary||d.type);
      haptic('cart');playTick();
    } else if(d.type==='checkout_canceled'){
      addActivity('act-cancel',d.summary||d.type);
      haptic('cancel');
    } else if(d.type==='promotions_listed'){
      addActivity('act-catalog',d.summary||'Promotions consultees');
      haptic('catalog');
    }
  }catch(ex){console.error('SSE handler error:',ex)}
}
function sseConnect(){
  if(evtSrc)return;
  evtSrc=new EventSource('/'+TID+'/api/notifications');
  evtSrc.onopen=function(){setConnected(true);sseRetryDelay=1000};
  evtSrc.onerror=function(){
    setConnected(false);
    evtSrc.close();evtSrc=null;
    setTimeout(sseConnect,sseRetryDelay);
    sseRetryDelay=Math.min(sseRetryDelay*2,8000);
  };
  evtSrc.addEventListener('message',handleSSEMessage);
}
function sseDisconnect(){
  if(evtSrc){evtSrc.close();evtSrc=null;setConnected(false)}
}
document.addEventListener('visibilitychange',()=>{
  if(document.hidden){sseDisconnect()}else{sseConnect();loadConfig();fetchRank()}
});
if(!document.hidden){sseConnect()}

loadConfig().then(()=>{lastCelebratedSales=config.sales_count||0});
fetchRank();
setInterval(()=>{
  fetchRank();
  loadConfig().then(()=>{
    const sc=config.sales_count||0;
    if(sc>lastCelebratedSales&&lastCelebratedSales>=0){
      showSaleCelebration(0,'','Vente detectee !');
      lastCelebratedSales=sc;
    }
  });
},3000);

// --- Leave arena ---
let hasLeft=false;
async function leaveArena(){
  if(!confirm("Quitter l'arene ? Votre marchand disparaitra."))return;
  hasLeft=true;
  try{await fetch('/'+TID+'/leave',{method:'POST'})}catch(e){}
  sseDisconnect();
  document.body.innerHTML='<div style="display:flex;align-items:center;justify-content:center;height:100vh;font-family:Outfit,system-ui,sans-serif;font-size:1.5rem;color:#666">Vous avez quitte l\'arene. <a href="/" style="margin-left:.5rem;color:#E5004C">Retour</a></div>';
}
window.addEventListener('beforeunload',function(e){
  if(hasLeft)return;
  e.preventDefault();
});
window.addEventListener('pagehide',function(){
  if(hasLeft)return;
  navigator.sendBeacon('/'+TID+'/leave');
});
</script>
</body>
</html>`
