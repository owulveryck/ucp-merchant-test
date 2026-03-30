package main

import (
	"fmt"
	"net/http"
)

func serveDashboard(w http.ResponseWriter, r *http.Request, tenantID, merchantName string, costPrice int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, dashboardHTML, merchantName, merchantName, tenantID, float64(costPrice)/100, costPrice, tenantID, costPrice)
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s - Dashboard</title>
<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;600;700;800&display=swap" rel="stylesheet">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'Outfit',system-ui,sans-serif;background:#FDF0EE;color:#1A1A2E;min-height:100vh;padding:1rem;transition:background .4s ease}
body.flash-checkout{background:#DBEAFE}
body.flash-negotiate{background:#FFF7ED}
body.flash-sale{background:#DCFCE7}
.header{text-align:center;margin-bottom:1.5rem}
.header h1{font-size:1.5rem;font-weight:800;color:#1A1A2E}
.header .id{font-size:.8rem;color:#999;margin-top:.3rem}
.rank-badge{display:inline-flex;align-items:center;gap:.4rem;margin-top:.5rem;padding:.4rem 1rem;border-radius:12px;font-size:1.1rem;font-weight:700;background:#FDE8E8;color:#E5004C;border:1px solid #E5004C}
.rank-badge .rank-medal{font-size:1.4rem}
.rank-badge.rank-1{background:#FEF9C3;border-color:#CA8A04;color:#854D0E}
.rank-badge.rank-2{background:#F1F5F9;border-color:#94A3B8;color:#475569}
.rank-badge.rank-3{background:#FFF7ED;border-color:#EA580C;color:#9A3412}
.rank-badge.rank-none{background:#F3F4F6;border-color:#D1D5DB;color:#6B7280}
.stock-warning{display:none;background:#DC2626;color:#fff;text-align:center;padding:.8rem;border-radius:12px;margin-bottom:1rem;font-weight:700;font-size:.95rem;animation:pulse-warn 1s ease-in-out infinite alternate}
@keyframes pulse-warn{from{opacity:1}to{opacity:.7}}
.card{background:#FFFFFF;border:1px solid #2D2D2D;border-radius:16px;margin-bottom:1rem;box-shadow:6px 6px 0px #E5004C;overflow:hidden}
.card-dots{padding:.5rem 1rem;border-bottom:1px solid #E0E0E0;display:flex;align-items:center;gap:6px}
.card-dots::before{content:'';width:10px;height:10px;border-radius:50%%;background:#E5004C;display:inline-block}
.card-dots::after{content:'';width:10px;height:10px;border-radius:50%%;background:#CCC;display:inline-block}
.card-body{padding:1.2rem}
.card h2{font-size:1rem;color:#999;margin-bottom:.8rem;text-transform:uppercase;letter-spacing:.05em;font-weight:700}
.field{margin-bottom:1rem}
.field label{display:block;font-size:.85rem;color:#666;margin-bottom:.3rem}
.field input[type=range]{width:100%%;accent-color:#E5004C}
.field .value{font-size:1.8rem;font-weight:800;color:#E5004C;text-align:center}
.field .subvalue{font-size:.8rem;color:#999;text-align:center}
.cost-info{font-size:.8rem;color:#D97706;text-align:center;margin-top:.2rem}
.bid-bar{height:6px;border-radius:3px;background:#E5E7EB;margin-top:.4rem;overflow:hidden}
.bid-bar-fill{height:100%%;border-radius:3px;background:#E5004C;transition:width .3s}
.funnel{display:flex;align-items:center;justify-content:center;gap:.3rem;flex-wrap:wrap;margin-bottom:.5rem}
.funnel-step{text-align:center;padding:.3rem .6rem;border-radius:8px;font-size:.85rem;font-weight:600}
.funnel-step .funnel-val{font-size:1.3rem;font-weight:800;display:block}
.funnel-step.f-visits{background:#FDE8E8;color:#E5004C}
.funnel-step.f-checkouts{background:#DBEAFE;color:#3B82F6}
.funnel-step.f-sales{background:#DCFCE7;color:#16A34A}
.funnel-arrow{color:#999;font-size:1rem;font-weight:700}
.funnel-rate{font-size:.7rem;color:#999;font-weight:600}
.discount-section{margin-top:.5rem}
.discount-row{display:flex;gap:.5rem;align-items:center;margin-bottom:.5rem;flex-wrap:wrap}
.discount-row input,.discount-row select{padding:.4rem;border:1px solid #CCC;border-radius:8px;background:#FFFFFF;color:#1A1A2E;font-size:.85rem;font-family:'Outfit',system-ui,sans-serif}
.discount-row input[type=text]{flex:1;min-width:80px}
.discount-row input[type=number]{width:60px}
.discount-row select{width:90px}
.discount-row label.cb{display:flex;align-items:center;gap:.3rem;font-size:.75rem;color:#666;white-space:nowrap}
.discount-usage{font-size:.7rem;color:#16A34A;font-weight:600;white-space:nowrap}
.btn-sm{padding:.3rem .6rem;border:none;border-radius:8px;cursor:pointer;font-size:.8rem;font-weight:600}
.btn-add{background:#FDE8E8;color:#E5004C;border:1px solid #E5004C}
.btn-rm{background:#FEF2F2;color:#DC2626;border:1px solid #DC2626}
.save-status{text-align:center;font-size:.8rem;color:#16A34A;margin-top:.5rem;opacity:0;transition:opacity .3s}
.activity-log{max-height:250px;overflow-y:auto;font-size:.8rem}
.activity-entry{padding:.4rem .6rem;margin-bottom:.3rem;border-radius:8px;display:flex;align-items:baseline;gap:.5rem}
.activity-entry .act-time{color:#999;font-size:.7rem;flex-shrink:0}
.activity-entry .act-text{flex:1}
.activity-entry .act-cost{font-size:.75rem;font-weight:700;flex-shrink:0}
.activity-entry.act-catalog{background:#FDE8E8;border-left:3px solid #E5004C;color:#E5004C}
.activity-entry.act-checkout{background:#DBEAFE;border-left:3px solid #3B82F6;color:#3B82F6}
.activity-entry.act-cart{background:#F3F4F6;border-left:3px solid #9CA3AF;color:#6B7280}
.activity-entry.act-cancel{background:#FEF2F2;border-left:3px solid #DC2626;color:#DC2626}
.activity-entry.act-sale{background:#DCFCE7;border-left:3px solid #16A34A;color:#16A34A}
.sale-overlay{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(229,0,76,.95);display:none;align-items:center;justify-content:center;flex-direction:column;z-index:9999;animation:pulse .5s ease-in-out infinite alternate}
.sale-overlay h1{font-size:4rem;color:#fff;font-weight:900;text-shadow:0 4px 20px rgba(0,0,0,.3)}
.sale-overlay p{font-size:1.5rem;color:#fff;margin-top:1rem}
.sale-overlay .sale-breakdown{font-size:1.1rem;color:rgba(255,255,255,.9);margin-top:.8rem;line-height:1.8}
.sale-overlay .sale-profit{font-size:2rem;font-weight:800;margin-top:.5rem}
.sale-overlay .sale-profit.positive{color:#BBF7D0}
.sale-overlay .sale-profit.negative{color:#FCA5A5}
@keyframes pulse{from{background:rgba(229,0,76,.95)}to{background:rgba(253,232,232,.95)}}
.algo-bar{display:flex;gap:.4rem;flex-wrap:wrap;justify-content:center;margin-top:.8rem}
.algo-btn{position:relative;padding:.4rem .8rem;border:1px solid #E5004C;border-radius:20px;background:#FDE8E8;color:#E5004C;font-size:.75rem;font-weight:600;cursor:pointer;font-family:'Outfit',system-ui,sans-serif;transition:all .2s}
.algo-btn:hover{background:#E5004C;color:#fff}
.algo-btn.active{background:#E5004C;color:#fff;box-shadow:0 2px 8px rgba(229,0,76,.3)}
.algo-btn .algo-tooltip{display:none;position:absolute;bottom:calc(100%% + 8px);left:50%%;transform:translateX(-50%%);background:#1A1A2E;color:#fff;padding:.5rem .7rem;border-radius:8px;font-size:.7rem;font-weight:400;width:220px;text-align:center;z-index:100;line-height:1.3}
.algo-btn .algo-tooltip::after{content:'';position:absolute;top:100%%;left:50%%;transform:translateX(-50%%);border:6px solid transparent;border-top-color:#1A1A2E}
.algo-btn:hover .algo-tooltip{display:block}
#price-slider.algo-active,#bid-slider.algo-active{accent-color:#999;pointer-events:none}
.algo-auto-tag{display:inline-block;background:#E5004C;color:#fff;font-size:.6rem;font-weight:700;padding:.1rem .4rem;border-radius:8px;margin-left:.3rem;vertical-align:middle}
</style>
</head>
<body>

<div class="header">
<h1>%s</h1>
<div class="id">ID: %s</div>
<div class="rank-badge rank-none" id="rank-badge">
<span class="rank-medal" id="rank-medal">-</span>
<span id="rank-text">Classement...</span>
</div>
</div>

<div class="stock-warning" id="stock-warning">Stock epuise — vous etes invisible dans le Shopping Graph !</div>

<div class="card">
<div class="card-dots"></div>
<div class="card-body">
<h2>Prix de vente</h2>
<div class="field">
<div class="value" id="price-display">--</div>
<div class="cost-info">Prix d'achat: $%.2f (minimum)</div>
<input type="range" id="price-slider" min="%d" max="30000" step="100" value="6000">
<div class="algo-bar" id="algo-bar">
<button class="algo-btn active" data-algo="manual" onclick="setAlgo('manual')">Manuel<span class="algo-tooltip">Controle libre du prix, encheres et promos avec les curseurs.</span></button>
<button class="algo-btn" data-algo="markup" onclick="setAlgo('markup')">Markup dyn.<span class="algo-tooltip">Stock eleve = prix bas + encheres hautes + promo auto. Stock bas = prix haut + encheres reduites. Marge dynamique.</span></button>
<button class="algo-btn" data-algo="threshold" onclick="setAlgo('threshold')">Seuils stock<span class="algo-tooltip">5 paliers : >80%% = promo -15%% + encheres max. 60-80%% = promo -10%%. 40-60%% = promo -5%%. 20-40%% = prix +15%%. &lt;20%% = prix premium, 0 enchere.</span></button>
<button class="algo-btn" data-algo="markdown" onclick="setAlgo('markdown')">Demarque<span class="algo-tooltip">Faible conversion = prix baisse + encheres montent + promo auto. Optimise pour ecouler quand les visiteurs ne convertissent pas.</span></button>
<button class="algo-btn" data-algo="surge" onclick="setAlgo('surge')">Surge<span class="algo-tooltip">Forte demande = prix monte + encheres baissent (pas besoin de payer). Aucune promo. Maximise la marge sur la rarete.</span></button>
</div>
</div>
</div>
</div>

<div class="card">
<div class="card-dots"></div>
<div class="card-body">
<h2>Stock</h2>
<div class="field">
<div class="value" id="stock-display">--</div>
<input type="range" id="stock-slider" min="0" max="50" step="1" value="10">
</div>
</div>
</div>

<div class="card">
<div class="card-dots"></div>
<div class="card-body">
<h2>Enchere max par visite</h2>
<div class="field">
<div class="value" id="bid-display">--</div>
<div class="subvalue">0 = organique uniquement (gratuit)</div>
<input type="range" id="bid-slider" min="0" max="200" step="5" value="50">
<div class="bid-bar"><div class="bid-bar-fill" id="bid-bar-fill" style="width:0"></div></div>
<div class="cost-info" id="bid-cpc-info"></div>
<div class="cost-info" id="bid-spend-info" style="color:#16A34A"></div>
</div>
</div>
</div>

<div class="card">
<div class="card-dots"></div>
<div class="card-body">
<h2>Rentabilite</h2>
<div class="field">
<div class="value" id="profit-display">$0.00</div>
<div class="subvalue" id="sales-count-display">0 ventes</div>
</div>
<div class="funnel" id="funnel">
<div class="funnel-step f-visits"><span class="funnel-val" id="f-visits">0</span>visites</div>
<div class="funnel-arrow">&rarr;</div>
<div class="funnel-step f-checkouts"><span class="funnel-val" id="f-checkouts">0</span>checkouts</div>
<div class="funnel-rate" id="f-rate1"></div>
<div class="funnel-arrow">&rarr;</div>
<div class="funnel-step f-sales"><span class="funnel-val" id="f-sales">0</span>ventes</div>
<div class="funnel-rate" id="f-rate2"></div>
</div>
</div>
</div>

<div class="card">
<div class="card-dots"></div>
<div class="card-body">
<h2>Activite</h2>
<div class="activity-log" id="activity-log"></div>
</div>
</div>

<div class="card">
<div class="card-dots"></div>
<div class="card-body">
<h2>Codes promo</h2>
<div class="discount-section" id="discounts"></div>
<button class="btn-sm btn-add" onclick="addDiscount()">+ Ajouter un code promo</button>
</div>
</div>

<div class="card">
<div class="card-dots"></div>
<div class="card-body">
<h2>Options de livraison</h2>
<div class="discount-section" id="shipping-options"></div>
</div>
</div>

<div class="save-status" id="save-status">Sauvegarde...</div>

<div class="sale-overlay" id="sale-overlay">
<h1>VENDU !</h1>
<p id="sale-detail"></p>
<div class="sale-breakdown" id="sale-breakdown"></div>
<div class="sale-profit" id="sale-profit"></div>
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
    const sl=document.getElementById('price-slider');
    const bl=document.getElementById('bid-slider');
    if(currentAlgo!=='manual'){sl.classList.add('algo-active');bl.classList.add('algo-active');recalcAlgoPrice()}
    else{sl.classList.remove('algo-active');bl.classList.remove('algo-active')}
  }catch(e){}
}

function updateDisplays(){
  document.getElementById('price-display').textContent='$'+(config.selling_price/100).toFixed(2);
  document.getElementById('stock-display').textContent=config.stock;
  document.getElementById('bid-display').textContent='$'+(config.max_cpc_bid/100).toFixed(2)+' / visite';
  const cpcEl=document.getElementById('bid-cpc-info');
  const actualCPC=config.actual_cpc||0;
  const maxBid=config.max_cpc_bid||0;
  if(maxBid===0){
    cpcEl.textContent='Organique — gratuit, visibilite reduite';
    cpcEl.style.color='#6B7280';
  } else {
    cpcEl.textContent='CPC reel: $'+(actualCPC/100).toFixed(2)+' (max: $'+(maxBid/100).toFixed(2)+')';
    cpcEl.style.color='#D97706';
  }
  const bidBarPct=maxBid>0?Math.min(100,actualCPC/maxBid*100):0;
  document.getElementById('bid-bar-fill').style.width=bidBarPct+'%%';
  const spendEl=document.getElementById('bid-spend-info');
  const totalAdSpend=config.total_ad_spend||0;
  spendEl.textContent='Depenses pub: $'+(totalAdSpend/100).toFixed(2);
  if(totalAdSpend>0){spendEl.style.color='#D97706'}else{spendEl.style.color='#16A34A'}
  if(config.net_profit!==undefined){
    const np=config.net_profit;
    const profitEl=document.getElementById('profit-display');
    profitEl.textContent='$'+(np/100).toFixed(2);
    if(np<0){profitEl.style.color='#DC2626'}else{profitEl.style.color='#16A34A'}
    const cc=config.consultation_count||0;
    const sc=config.sales_count||0;
    const avgCPC=cc>0?(totalAdSpend/cc/100).toFixed(2):'0.00';
    document.getElementById('sales-count-display').textContent=sc+' vente(s) / '+cc+' consultation(s) a $'+avgCPC+' moy.';
    // Funnel
    document.getElementById('f-visits').textContent=cc;
    document.getElementById('f-checkouts').textContent=checkoutCount;
    document.getElementById('f-sales').textContent=sc;
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
      if(r.ok){el.textContent='Sauvegarde !';const d=await r.json();config=d;updateDisplays();setTimeout(()=>el.style.opacity='0',1000);}
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
    // Prix: marge basse si stock eleve, haute si stock bas
    const adj=(1-ratio)*0.4-ratio*0.15;
    newPrice=Math.round(COST_PRICE*(1+0.30+adj));
    // Enchere: forte si stock eleve (besoin de vendre), faible si stock bas
    newBid=Math.round(ratio*150);
    // Promo: discount si stock >60%% pour ecouler
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
    // Prix: baisse si faible conversion
    const disc=Math.min(0.25,(1-ratio)*0.3);
    newPrice=Math.round(COST_PRICE*(1.30-disc));
    newPrice=Math.max(newPrice,Math.round(COST_PRICE*1.05));
    // Enchere: monte si personne n'achete (besoin de visibilite + conversion)
    newBid=Math.round((1-ratio)*175);
    // Promo: discount croissant si conversion faible
    if(ratio<0.5)autoPct=Math.round((0.5-ratio)*30);

  }else if(currentAlgo==='surge'){
    const sf=Math.min(2.0,1.0+consult*0.05);
    newPrice=Math.round(COST_PRICE*1.20*sf);
    // Enchere: baisse avec la demande (pas besoin de payer, la demande vient seule)
    newBid=Math.round(Math.max(0,150-consult*10));
    // Promo: aucune en surge (la demande est forte)
    autoPct=0;
  }

  // Clamp prix
  newPrice=Math.max(COST_PRICE,Math.min(30000,newPrice));
  newPrice=Math.round(newPrice/100)*100;
  if(newPrice<COST_PRICE)newPrice=COST_PRICE;
  // Clamp enchere
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
        '<span style="color:#E5004C;font-weight:700;font-size:.85rem">'+dc.code+'</span>'+
        '<span class="algo-auto-tag">ALGO</span>'+
        '<span style="color:#666;font-size:.85rem;margin-left:auto">-'+dc.value+'%%</span>'+
        usageBadge;
    }else{
      row.innerHTML=
        '<input type="text" value="'+dc.code+'" placeholder="CODE" onchange="updDiscount('+i+',\'code\',this.value)">'+
        '<select onchange="updDiscount('+i+',\'type\',this.value)"><option value="percentage"'+(dc.type==='percentage'?' selected':'')+'>%%</option><option value="fixed"'+(dc.type==='fixed'?' selected':'')+'>Fixe</option></select>'+
        '<input type="number" value="'+dc.value+'" min="1" onchange="updDiscount('+i+',\'value\',parseInt(this.value))">'+
        usageBadge+
        '<label class="cb"><input type="checkbox"'+(dc.new_customer_only?' checked':'')+' onchange="updDiscount('+i+',\'new_customer_only\',this.checked)">Nouveau client</label>'+
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
    row.innerHTML='<span style="color:#666;font-size:.85rem">'+so.title+'</span><span style="color:#E5004C;font-weight:700;margin-left:auto">$'+(so.cost/100).toFixed(2)+'</span>';
    el.appendChild(row);
  });
  if(!config.shipping_options||config.shipping_options.length===0){
    el.innerHTML='<div style="color:#555;font-size:.85rem">Aucune option configuree</div>';
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
      text.textContent=rk+(rk===1?'er':'e')+' / '+total+' marchands';
      badge.className='rank-badge'+(rk<=3?' rank-'+rk:'');
    } else {
      medal.textContent='-';
      text.textContent='Non classe'+(total>0?' ('+total+' marchands)':'');
      badge.className='rank-badge rank-none';
    }
  }catch(e){}
}

// Background flash on events
let flashTimer=null;
function flashBg(cls){
  clearTimeout(flashTimer);
  document.body.className=cls;
  flashTimer=setTimeout(()=>{document.body.className=''},3000);
}

// SSE for sale + activity notifications
const actLog=document.getElementById('activity-log');
let actCount=0;
function addActivity(cls,text,costAnnotation){
  const now=new Date();
  const ts=now.getHours().toString().padStart(2,'0')+':'+now.getMinutes().toString().padStart(2,'0')+':'+now.getSeconds().toString().padStart(2,'0');
  const div=document.createElement('div');
  div.className='activity-entry '+cls;
  const costHtml=costAnnotation?'<span class="act-cost">'+costAnnotation+'</span>':'';
  div.innerHTML='<span class="act-time">'+ts+'</span><span class="act-text">'+text+'</span>'+costHtml;
  actLog.appendChild(div);
  actLog.scrollTop=actLog.scrollHeight;
  actCount++;
  if(actCount>50){actLog.removeChild(actLog.firstChild);actCount--}
}
// SSE lifecycle: connect only when tab is visible to avoid browser connection limit
let evtSrc=null;
function sseConnect(){
  if(evtSrc)return;
  evtSrc=new EventSource('/'+TID+'/api/notifications');
  evtSrc.addEventListener('message',function(e){
    try{
      const d=JSON.parse(e.data);
      if(d.type==='sale'){
        const saleTotal=d.total||0;
        const profit=config.selling_price-COST_PRICE-(config.actual_cpc||0);
        const ov=document.getElementById('sale-overlay');
        document.getElementById('sale-detail').textContent='Commande '+d.order_id+' - $'+(saleTotal/100).toFixed(2);
        const bd=document.getElementById('sale-breakdown');
        bd.innerHTML='Prix de vente: $'+(config.selling_price/100).toFixed(2)+'<br>Prix d\'achat: -$'+(COST_PRICE/100).toFixed(2)+'<br>CPC: -$'+((config.actual_cpc||0)/100).toFixed(2);
        const sp=document.getElementById('sale-profit');
        sp.textContent=(profit>=0?'+':'')+' $'+(profit/100).toFixed(2);
        sp.className='sale-profit '+(profit>=0?'positive':'negative');
        ov.style.display='flex';
        if(navigator.vibrate)navigator.vibrate([200,100,200,100,200]);
        setTimeout(()=>{ov.style.display='none';loadConfig()},5000);
        const profitStr=(profit>=0?'+':'-')+' $'+Math.abs(profit/100).toFixed(2);
        addActivity('act-sale',d.summary||('Vente: '+d.order_id),'<span style="color:#16A34A">'+profitStr+'</span>');
        flashBg('flash-sale');
        // Track discount usage from summary
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
      } else if(d.type==='checkout_created'){
        checkoutCount++;
        addActivity('act-checkout',d.summary||d.type);
        flashBg('flash-checkout');
        updateDisplays();
        // Track discount usage
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
        // Track discount usage
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
      } else if(d.type==='checkout_canceled'){
        addActivity('act-cancel',d.summary||d.type);
      }
    }catch(e){}
  });
}
function sseDisconnect(){
  if(evtSrc){evtSrc.close();evtSrc=null}
}
document.addEventListener('visibilitychange',()=>{
  if(document.hidden){sseDisconnect()}else{sseConnect();loadConfig();fetchRank()}
});
if(!document.hidden){sseConnect()}

loadConfig();
fetchRank();
setInterval(fetchRank,3000);
</script>
</body>
</html>`
