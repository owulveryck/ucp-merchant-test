package main

import (
	"fmt"
	"net/http"
)

func serveDashboard(w http.ResponseWriter, r *http.Request, tenantID, merchantName string, costPrice int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, dashboardHTML, merchantName, tenantID, float64(costPrice)/100, costPrice, tenantID, costPrice)
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s - Dashboard</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'Segoe UI',system-ui,sans-serif;background:#0a0a1a;color:#e0e0e0;min-height:100vh;padding:1rem}
.header{text-align:center;margin-bottom:1.5rem}
.header h1{font-size:1.5rem;background:linear-gradient(135deg,#00d4ff,#7b2ff7);-webkit-background-clip:text;-webkit-text-fill-color:transparent}
.header .id{font-size:.8rem;color:#555;margin-top:.3rem}
.card{background:#111;border:2px solid #222;border-radius:12px;padding:1.2rem;margin-bottom:1rem}
.card h2{font-size:1rem;color:#888;margin-bottom:.8rem;text-transform:uppercase;letter-spacing:.05em}
.field{margin-bottom:1rem}
.field label{display:block;font-size:.85rem;color:#aaa;margin-bottom:.3rem}
.field input[type=range]{width:100%%;accent-color:#00d4ff}
.field .value{font-size:1.8rem;font-weight:700;color:#00d4ff;text-align:center}
.field .subvalue{font-size:.8rem;color:#666;text-align:center}
.cost-info{font-size:.8rem;color:#ff8800;text-align:center;margin-top:.2rem}
.discount-section{margin-top:.5rem}
.discount-row{display:flex;gap:.5rem;align-items:center;margin-bottom:.5rem;flex-wrap:wrap}
.discount-row input,.discount-row select{padding:.4rem;border:1px solid #333;border-radius:6px;background:#1a1a2e;color:#fff;font-size:.85rem}
.discount-row input[type=text]{flex:1;min-width:80px}
.discount-row input[type=number]{width:60px}
.discount-row select{width:90px}
.discount-row label.cb{display:flex;align-items:center;gap:.3rem;font-size:.75rem;color:#aaa;white-space:nowrap}
.btn-sm{padding:.3rem .6rem;border:none;border-radius:6px;cursor:pointer;font-size:.8rem}
.btn-add{background:#1a3a1a;color:#4caf50;border:1px solid #2e7d32}
.btn-rm{background:#3a1a1a;color:#ef5350;border:1px solid #c62828}
.save-status{text-align:center;font-size:.8rem;color:#4caf50;margin-top:.5rem;opacity:0;transition:opacity .3s}
.activity-log{max-height:250px;overflow-y:auto;font-size:.8rem}
.activity-entry{padding:.4rem .6rem;margin-bottom:.3rem;border-radius:6px;display:flex;align-items:baseline;gap:.5rem}
.activity-entry .act-time{color:#555;font-size:.7rem;flex-shrink:0}
.activity-entry .act-text{flex:1}
.activity-entry.act-catalog{background:#0d1b2a;border-left:3px solid #00d4ff;color:#00d4ff}
.activity-entry.act-checkout{background:#0d1b2a;border-left:3px solid #00e5cc;color:#00e5cc}
.activity-entry.act-cart{background:#0d1b2a;border-left:3px solid #7b86a2;color:#b0b8cc}
.activity-entry.act-cancel{background:#1a0d0d;border-left:3px solid #ff4444;color:#ff4444}
.activity-entry.act-sale{background:#0d1a0d;border-left:3px solid #4caf50;color:#4caf50}
.sale-overlay{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,212,255,.95);display:none;align-items:center;justify-content:center;flex-direction:column;z-index:9999;animation:pulse .5s ease-in-out infinite alternate}
.sale-overlay h1{font-size:4rem;color:#fff;font-weight:900;text-shadow:0 4px 20px rgba(0,0,0,.3)}
.sale-overlay p{font-size:1.5rem;color:#fff;margin-top:1rem}
@keyframes pulse{from{background:rgba(0,212,255,.95)}to{background:rgba(123,47,247,.95)}}
</style>
</head>
<body>

<div class="header">
<h1>Mon Commerce</h1>
<div class="id">ID: %s</div>
</div>

<div class="card">
<h2>Prix de vente</h2>
<div class="field">
<div class="value" id="price-display">--</div>
<div class="cost-info">Prix d'achat: $%.2f (minimum)</div>
<input type="range" id="price-slider" min="%d" max="30000" step="100" value="6000">
</div>
</div>

<div class="card">
<h2>Stock</h2>
<div class="field">
<div class="value" id="stock-display">--</div>
<input type="range" id="stock-slider" min="0" max="50" step="1" value="10">
</div>
</div>

<div class="card">
<h2>Boost (visibilite)</h2>
<div class="field">
<div class="value" id="boost-display">--</div>
<div class="subvalue">Plus le boost est eleve, plus vous etes visible</div>
<input type="range" id="boost-slider" min="0" max="100" step="5" value="50">
<div class="cost-info" id="boost-cost-info"></div>
<div class="cost-info" id="boost-margin-info" style="color:#4caf50"></div>
</div>
</div>

<div class="card">
<h2>Rentabilite</h2>
<div class="field">
<div class="value" id="profit-display">$0.00</div>
<div class="subvalue" id="sales-count-display">0 ventes</div>
</div>
</div>

<div class="card">
<h2>Activite</h2>
<div class="activity-log" id="activity-log"></div>
</div>

<div class="card">
<h2>Codes promo</h2>
<div class="discount-section" id="discounts"></div>
<button class="btn-sm btn-add" onclick="addDiscount()">+ Ajouter un code promo</button>
</div>

<div class="save-status" id="save-status">Sauvegarde...</div>

<div class="sale-overlay" id="sale-overlay">
<h1>VENDU !</h1>
<p id="sale-detail"></p>
</div>

<script>
const TID='%s';
const API='/'+TID+'/api';
const COST_PRICE=%d;

let config={selling_price:6000,stock:10,discount_codes:[],boost_score:50};
let saveTimer=null;

async function loadConfig(){
  try{
    const r=await fetch(API+'/config');
    config=await r.json();
    document.getElementById('price-slider').value=config.selling_price;
    document.getElementById('stock-slider').value=config.stock;
    document.getElementById('boost-slider').value=config.boost_score;
    updateDisplays();
    renderDiscounts();
  }catch(e){}
}

function updateDisplays(){
  document.getElementById('price-display').textContent='$'+(config.selling_price/100).toFixed(2);
  document.getElementById('stock-display').textContent=config.stock;
  document.getElementById('boost-display').textContent=config.boost_score;
  const margin=config.selling_price-COST_PRICE;
  const boostCost=config.boost_score*margin/100;
  const netMargin=margin-boostCost;
  const bcEl=document.getElementById('boost-cost-info');
  bcEl.textContent='Cout du boost par vente: $'+(boostCost/100).toFixed(2);
  const bmEl=document.getElementById('boost-margin-info');
  bmEl.textContent='Marge nette par vente: $'+(netMargin/100).toFixed(2);
  if(netMargin<=margin*0.2){bmEl.style.color='#ff4444'}else{bmEl.style.color='#4caf50'}
  if(config.total_profit!==undefined){
    document.getElementById('profit-display').textContent='$'+(config.total_profit/100).toFixed(2);
    document.getElementById('sales-count-display').textContent=config.sales_count+' vente(s)';
  }
}

function schedSave(){
  clearTimeout(saveTimer);
  const el=document.getElementById('save-status');
  el.textContent='Sauvegarde...';el.style.opacity='1';
  saveTimer=setTimeout(async()=>{
    try{
      await fetch(API+'/config',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(config)});
      el.textContent='Sauvegarde !';
      setTimeout(()=>el.style.opacity='0',1000);
    }catch(e){el.textContent='Erreur';el.style.color='#ff4444'}
  },300);
}

document.getElementById('price-slider').oninput=function(){config.selling_price=parseInt(this.value);updateDisplays();schedSave()};
document.getElementById('stock-slider').oninput=function(){config.stock=parseInt(this.value);updateDisplays();schedSave()};
document.getElementById('boost-slider').oninput=function(){config.boost_score=parseInt(this.value);updateDisplays();schedSave()};

function renderDiscounts(){
  const el=document.getElementById('discounts');
  el.innerHTML='';
  (config.discount_codes||[]).forEach((dc,i)=>{
    const row=document.createElement('div');
    row.className='discount-row';
    row.innerHTML=
      '<input type="text" value="'+dc.code+'" placeholder="CODE" onchange="updDiscount('+i+',\'code\',this.value)">'+
      '<select onchange="updDiscount('+i+',\'type\',this.value)"><option value="percentage"'+(dc.type==='percentage'?' selected':'')+'>%%</option><option value="fixed"'+(dc.type==='fixed'?' selected':'')+'>Fixe</option></select>'+
      '<input type="number" value="'+dc.value+'" min="1" onchange="updDiscount('+i+',\'value\',parseInt(this.value))">'+
      '<label class="cb"><input type="checkbox"'+(dc.new_customer_only?' checked':'')+' onchange="updDiscount('+i+',\'new_customer_only\',this.checked)">Nouveau client</label>'+
      '<button class="btn-sm btn-rm" onclick="rmDiscount('+i+')">X</button>';
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

// SSE for sale + activity notifications
const actLog=document.getElementById('activity-log');
let actCount=0;
function addActivity(cls,text){
  const now=new Date();
  const ts=now.getHours().toString().padStart(2,'0')+':'+now.getMinutes().toString().padStart(2,'0')+':'+now.getSeconds().toString().padStart(2,'0');
  const div=document.createElement('div');
  div.className='activity-entry '+cls;
  div.innerHTML='<span class="act-time">'+ts+'</span><span class="act-text">'+text+'</span>';
  actLog.appendChild(div);
  actLog.scrollTop=actLog.scrollHeight;
  actCount++;
  if(actCount>50){actLog.removeChild(actLog.firstChild);actCount--}
}
const evtSrc=new EventSource('/'+TID+'/api/notifications');
evtSrc.addEventListener('message',function(e){
  try{
    const d=JSON.parse(e.data);
    if(d.type==='sale'){
      const ov=document.getElementById('sale-overlay');
      document.getElementById('sale-detail').textContent='Commande '+d.order_id+' - $'+(d.total/100).toFixed(2);
      ov.style.display='flex';
      if(navigator.vibrate)navigator.vibrate([200,100,200,100,200]);
      setTimeout(()=>{ov.style.display='none';loadConfig()},5000);
      addActivity('act-sale',d.summary||('Vente: '+d.order_id));
    } else if(d.type==='catalog_browse'||d.type==='product_details'){
      addActivity('act-catalog',d.summary||d.type);
    } else if(d.type==='checkout_created'||d.type==='checkout_updated'){
      addActivity('act-checkout',d.summary||d.type);
    } else if(d.type==='cart_created'){
      addActivity('act-cart',d.summary||d.type);
    } else if(d.type==='checkout_canceled'){
      addActivity('act-cancel',d.summary||d.type);
    }
  }catch(e){}
});

loadConfig();
</script>
</body>
</html>`
