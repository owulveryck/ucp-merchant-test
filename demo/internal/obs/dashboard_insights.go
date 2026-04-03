package obs

import "net/http"

const insightsDashboardHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>UCP Arena - Insights</title>
<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;600;700;800&display=swap" rel="stylesheet">
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:'Outfit',system-ui,sans-serif;background:#FDF0EE;color:#1A1A2E;height:100vh;display:flex;flex-direction:column;overflow:hidden}

.topbar{background:#FFFFFF;padding:0.5rem 1.5rem;display:flex;align-items:center;gap:1rem;border-bottom:1px solid #E0E0E0;flex-shrink:0}
.topbar h1{font-size:1.1rem;font-weight:800;letter-spacing:0.02em}
.topbar h1 span{color:#E5004C}
.topbar .right{margin-left:auto;display:flex;align-items:center;gap:0.75rem}
.nav-link{color:#E5004C;text-decoration:none;font-weight:700;font-size:0.85rem;padding:0.3rem 0.6rem;border:1px solid #E5004C;border-radius:8px;transition:background 0.15s,color 0.15s}
.nav-link:hover{background:#E5004C;color:#fff}
.live-dot{width:8px;height:8px;border-radius:50%;background:#E5004C;display:inline-block;margin-right:4px;animation:pulse-dot 1.5s ease-in-out infinite}
@keyframes pulse-dot{0%,100%{opacity:1}50%{opacity:0.3}}
.conn-dot{width:8px;height:8px;border-radius:50%;display:inline-block;margin-left:6px;vertical-align:middle}
.conn-dot.connected{background:#16A34A}
.conn-dot.disconnected{background:#DC2626;animation:pulse-dot 1s ease-in-out infinite}

.main-area{flex:1;display:flex;flex-direction:column;padding:0.75rem;gap:0.75rem;overflow:hidden}

.panel{background:#FFFFFF;border:1px solid #2D2D2D;border-radius:16px;overflow:hidden;display:flex;flex-direction:column}
.panel-header{padding:0.45rem 0.75rem;border-bottom:1px solid #E0E0E0;font-weight:700;font-size:0.7rem;color:#E5004C;text-transform:uppercase;letter-spacing:0.05em;display:flex;align-items:center;gap:0.5rem;flex-shrink:0}
.panel-header .ph-right{margin-left:auto;font-weight:400;color:#999;text-transform:none;letter-spacing:0}
.panel-body{flex:1;overflow:auto;position:relative}

.top-row{flex:0 0 42%;display:flex;gap:0.75rem;min-height:0}
.bottom-row{flex:1;display:flex;gap:0.75rem;min-height:0}
.graph-panel{flex:0 0 28%}
.timeline-panel{flex:1}
.leaderboard-panel{flex:0 0 40%}
.decision-panel{flex:1}

/* --- Graph --- */
#graph-canvas{width:100%;height:100%;display:block}

/* --- Timeline --- */
.tl-scroll{overflow-x:auto;overflow-y:auto;height:100%;padding:0.5rem}
.tl-lane{display:flex;align-items:center;margin-bottom:2px;min-height:28px}
.tl-label{width:100px;flex-shrink:0;font-size:0.85rem;font-weight:700;color:#1A1A2E;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;padding-right:0.5rem;text-align:right}
.tl-track{flex:1;position:relative;height:24px;background:#F9FAFB;border-radius:4px;overflow:visible}
.tl-event{position:absolute;height:22px;top:1px;border-radius:4px;font-size:0.75rem;font-weight:600;display:flex;align-items:center;justify-content:center;color:#fff;min-width:18px;cursor:default;transition:opacity 0.15s;z-index:1;white-space:nowrap;padding:0 4px}
.tl-event:hover{opacity:0.8;z-index:10}
.tl-event.search{background:#6B7280}
.tl-event.lookup{background:#E5004C}
.tl-event.checkout{background:#3B82F6}
.tl-event.promo{background:#F59E0B}
.tl-event.discount{background:#F97316}
.tl-event.update{background:#8B5CF6}
.tl-event.summary{background:#0EA5E9}
.tl-event.complete{background:#16A34A}
.tl-event.cancel{background:#DC2626}
.tl-event.thinking{background:#9CA3AF}
.tl-legend{display:flex;gap:0.75rem;flex-wrap:wrap;padding:0.4rem 0.5rem;border-top:1px solid #E0E0E0;font-size:0.75rem;flex-shrink:0}
.tl-legend-item{display:flex;align-items:center;gap:3px}
.tl-legend-dot{width:10px;height:10px;border-radius:3px;flex-shrink:0}
.tl-no-data{text-align:center;padding:2rem;color:#999;font-size:0.85rem}
.tl-time-axis{position:absolute;top:0;left:100px;right:0;height:16px;border-bottom:1px solid #E0E0E0;font-size:0.55rem;color:#999;overflow:hidden;flex-shrink:0}
.tl-time-mark{position:absolute;top:0;height:100%;border-left:1px dashed #E0E0E0;padding-left:3px;line-height:16px}

/* --- Leaderboard --- */
.lb-table{width:100%;border-collapse:collapse;font-size:0.95rem}
.lb-table th{text-align:left;padding:0.4rem 0.5rem;color:#999;font-weight:600;font-size:0.85rem;text-transform:uppercase;border-bottom:1px solid #E0E0E0;position:sticky;top:0;background:#fff;z-index:1}
.lb-table td{padding:0.4rem 0.5rem;border-bottom:1px solid #F3F4F6;font-variant-numeric:tabular-nums}
.lb-table tr:hover{background:#FDF0EE}
.lb-rank{font-weight:800;width:2rem;text-align:center}
.lb-name{font-weight:700}
.lb-positive{color:#16A34A;font-weight:600}
.lb-negative{color:#DC2626;font-weight:600}
.lb-spark{width:60px;height:20px;vertical-align:middle}
.lb-no-data{text-align:center;padding:2rem;color:#999;font-size:0.85rem}

/* --- Decision --- */
.dc-table{width:100%;border-collapse:collapse;font-size:0.8rem}
.dc-table th{text-align:left;padding:0.4rem 0.5rem;color:#999;font-weight:600;font-size:0.7rem;text-transform:uppercase;border-bottom:1px solid #E0E0E0;position:sticky;top:0;background:#fff;z-index:1}
.dc-table td{padding:0.4rem 0.5rem;border-bottom:1px solid #F3F4F6;font-variant-numeric:tabular-nums}
.dc-winner{background:#F0FDF4 !important}
.dc-winner td{color:#16A34A;font-weight:700}
.dc-canceled{opacity:0.5}
.dc-canceled td{text-decoration:line-through}
.dc-status{font-weight:700;font-size:0.75rem;padding:0.15rem 0.4rem;border-radius:20px;display:inline-block}
.dc-status.won{background:#DCFCE7;color:#16A34A}
.dc-status.lost{background:#FEF2F2;color:#DC2626}
.dc-status.pending{background:#F3F4F6;color:#9CA3AF}
.dc-no-data{text-align:center;padding:2rem;color:#999;font-size:0.85rem}
</style>
</head>
<body>

<div class="topbar">
  <h1>UCP <span>Arena</span> &middot; Insights</h1>
  <div class="right">
    <a href="/arena" class="nav-link">Arena Monitor</a>
    <span><span class="live-dot"></span>LIVE<span class="conn-dot disconnected" id="conn-dot"></span></span>
  </div>
</div>

<div class="main-area">
  <div class="top-row">
    <div class="panel graph-panel" style="box-shadow:4px 4px 0px #1A1A2E">
      <div class="panel-header">Shopping Graph</div>
      <div class="panel-body"><canvas id="graph-canvas"></canvas></div>
    </div>
    <div class="panel timeline-panel" style="box-shadow:4px 4px 0px #E5004C">
      <div class="panel-header">Negotiation Timeline <span class="ph-right" id="tl-elapsed"></span></div>
      <div class="panel-body">
        <div class="tl-scroll" id="tl-scroll">
          <div class="tl-no-data" id="tl-no-data">En attente de l'agent acheteur...</div>
        </div>
      </div>
      <div class="tl-legend">
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#6B7280"></div>Recherche</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#E5004C"></div>Consultation</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#3B82F6"></div>Panier</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#F59E0B"></div>Promos</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#F97316"></div>Code promo</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#8B5CF6"></div>Mise a jour</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#0EA5E9"></div>Verification</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#16A34A"></div>Paiement</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#DC2626"></div>Annulation</div>
      </div>
    </div>
  </div>
  <div class="bottom-row">
    <div class="panel leaderboard-panel" style="box-shadow:4px 4px 0px #F59E0B">
      <div class="panel-header">Leaderboard</div>
      <div class="panel-body" id="lb-body">
        <div class="lb-no-data">Aucun marchand enregistre</div>
      </div>
    </div>
    <div class="panel decision-panel" style="box-shadow:4px 4px 0px #16A34A">
      <div class="panel-header">Decision Recap <span class="ph-right" id="dc-status-text"></span></div>
      <div class="panel-body" id="dc-body">
        <div class="dc-no-data" id="dc-no-data">En attente d'une comparaison...</div>
      </div>
    </div>
  </div>
</div>

<script>
(function() {
  // === State ===
  var merchants = {};
  var rankings = {};
  var agentStartTime = null;
  var tlEvents = [];
  var tlLanes = {};
  var checkoutMerchants = {};
  var decisionData = {};
  var profitSnapshots = {};
  var PX_PER_SEC = 60;

  // === Action config ===
  var actionMeta = {
    'search_products':      {cls:'search',   icon:'S',  label:'Recherche'},
    'get_product_details':  {cls:'lookup',   icon:'D',  label:'Consultation'},
    'create_checkout':      {cls:'checkout', icon:'C',  label:'Panier'},
    'list_promotions':      {cls:'promo',    icon:'P',  label:'Promos'},
    'apply_discount_codes': {cls:'discount', icon:'%',  label:'Code promo'},
    'update_checkout':      {cls:'update',   icon:'U',  label:'Mise a jour'},
    'get_checkout_summary': {cls:'summary',  icon:'$',  label:'Verification'},
    'complete_checkout':    {cls:'complete', icon:'+',  label:'Paiement'},
    'cancel_checkout':      {cls:'cancel',   icon:'X',  label:'Annulation'}
  };

  // === Helpers ===
  function esc(s){var d=document.createElement('div');d.textContent=s;return d.innerHTML}
  function fmt(cents){return '$'+(cents/100).toFixed(2)}
  function extractMid(url){
    if(!url)return null;
    var parts=url.replace(/\/+$/,'').split('/');
    return parts[parts.length-1];
  }
  function getMerchantName(mid){
    var m=merchants[mid];
    return m?(m.emoji?m.emoji+' ':'')+m.name:mid.substring(0,8);
  }
  function getMerchantColor(mid){
    var m=merchants[mid];
    return m&&m.accent_color?m.accent_color:'#E5004C';
  }

  // === Data fetching ===
  function fetchMerchants(){
    fetch('/arena/merchants').then(function(r){return r.json()}).then(function(d){
      var list=d.merchants||[];
      merchants={};
      for(var i=0;i<list.length;i++){
        merchants[list[i].id]=list[i];
        if(!profitSnapshots[list[i].id])profitSnapshots[list[i].id]=[];
        profitSnapshots[list[i].id].push(list[i].net_profit);
        if(profitSnapshots[list[i].id].length>30)profitSnapshots[list[i].id].shift();
      }
      renderLeaderboard(list);
      layoutGraph();
    }).catch(function(){});
  }
  function fetchRankings(){
    fetch('/arena/rankings').then(function(r){return r.json()}).then(function(d){
      rankings=d.rankings||{};
    }).catch(function(){});
  }
  fetchRankings();
  fetchMerchants();
  setInterval(function(){fetchRankings();fetchMerchants()},3000);

  // ============================================================
  // GRAPH VISUALIZATION
  // ============================================================
  var canvas=document.getElementById('graph-canvas');
  var ctx=canvas.getContext('2d');
  var graphNodes=[];
  var centerNode={id:'graph',name:'Shopping\nGraph',x:0,y:0,color:'#1A1A2E',radius:28,pulse:0,emoji:''};
  var edgeGlows={};

  function layoutGraph(){
    var ids=Object.keys(merchants);
    var n=ids.length;
    graphNodes=[centerNode];
    for(var i=0;i<n;i++){
      var m=merchants[ids[i]];
      var angle=(i/n)*Math.PI*2-Math.PI/2;
      var r=Math.min(canvas.offsetWidth,canvas.offsetHeight)*0.32;
      if(r<60)r=60;
      var existing=null;
      for(var j=1;j<graphNodes.length;j++){
        if(graphNodes[j].id===ids[i]){existing=graphNodes[j];break}
      }
      graphNodes.push({
        id:ids[i],
        name:m.name,
        x:existing?existing.x:Math.cos(angle)*r,
        y:existing?existing.y:Math.sin(angle)*r,
        tx:Math.cos(angle)*r,
        ty:Math.sin(angle)*r,
        color:m.accent_color||'#E5004C',
        emoji:m.emoji||'',
        radius:18,
        pulse:existing?existing.pulse:0
      });
    }
  }

  function pulseNode(mid){
    for(var i=0;i<graphNodes.length;i++){
      if(graphNodes[i].id===mid)graphNodes[i].pulse=1;
    }
  }
  function pulseCenter(){centerNode.pulse=1}
  function glowEdge(mid){edgeGlows[mid]=1}

  function renderGraph(){
    var dpr=window.devicePixelRatio||1;
    var w=canvas.offsetWidth;
    var h=canvas.offsetHeight;
    if(w===0||h===0){requestAnimationFrame(renderGraph);return}
    canvas.width=w*dpr;
    canvas.height=h*dpr;
    ctx.setTransform(dpr,0,0,dpr,0,0);
    var cx=w/2,cy=h/2;
    ctx.clearRect(0,0,w,h);

    // Animate positions
    for(var i=1;i<graphNodes.length;i++){
      var nd=graphNodes[i];
      if(nd.tx!==undefined){nd.x+=(nd.tx-nd.x)*0.08;nd.y+=(nd.ty-nd.y)*0.08}
    }

    // Draw edges
    for(var i=1;i<graphNodes.length;i++){
      var nd=graphNodes[i];
      var glow=edgeGlows[nd.id]||0;
      ctx.beginPath();
      ctx.moveTo(cx+centerNode.x,cy+centerNode.y);
      ctx.lineTo(cx+nd.x,cy+nd.y);
      if(glow>0){
        ctx.strokeStyle=nd.color;
        ctx.lineWidth=2+glow*4;
        ctx.globalAlpha=0.4+glow*0.6;
        edgeGlows[nd.id]=Math.max(0,glow-0.015);
      } else {
        ctx.strokeStyle='#D1D5DB';
        ctx.lineWidth=1;
        ctx.globalAlpha=0.4;
      }
      ctx.stroke();
      ctx.globalAlpha=1;
    }

    // Draw nodes
    for(var i=0;i<graphNodes.length;i++){
      var nd=graphNodes[i];
      var x=cx+nd.x,y=cy+nd.y;
      var r=nd.radius;

      // Pulse ring
      if(nd.pulse>0){
        ctx.beginPath();
        ctx.arc(x,y,r+nd.pulse*20,0,Math.PI*2);
        ctx.strokeStyle=nd.color;
        ctx.lineWidth=2;
        ctx.globalAlpha=nd.pulse*0.4;
        ctx.stroke();
        ctx.globalAlpha=1;
        nd.pulse=Math.max(0,nd.pulse-0.02);
      }

      // Circle
      ctx.beginPath();
      ctx.arc(x,y,r,0,Math.PI*2);
      ctx.fillStyle=nd.color;
      ctx.fill();

      // Center node text
      if(i===0){
        ctx.fillStyle='#fff';
        ctx.font='700 9px Outfit';
        ctx.textAlign='center';
        ctx.textBaseline='middle';
        ctx.fillText('Shopping',x,y-5);
        ctx.fillText('Graph',x,y+7);
      }

      // Merchant label below
      if(i>0){
        ctx.fillStyle='#1A1A2E';
        ctx.font='600 10px Outfit';
        ctx.textAlign='center';
        ctx.textBaseline='top';
        var label=nd.emoji?(nd.emoji+' '+nd.name):nd.name;
        if(label.length>14)label=label.substring(0,12)+'..';
        ctx.fillText(label,x,y+r+4);
      }
    }

    requestAnimationFrame(renderGraph);
  }
  requestAnimationFrame(renderGraph);

  // ============================================================
  // TIMELINE
  // ============================================================
  var tlScroll=document.getElementById('tl-scroll');
  var tlNoData=document.getElementById('tl-no-data');
  var tlElapsed=document.getElementById('tl-elapsed');
  var tlMaxTime=0;

  function resetTimeline(){
    tlEvents=[];
    tlLanes={};
    tlMaxTime=0;
    tlScroll.innerHTML='';
    tlNoData=document.createElement('div');
    tlNoData.className='tl-no-data';
    tlNoData.textContent='Agent demarre...';
    tlScroll.appendChild(tlNoData);
  }

  function ensureLane(mid){
    if(tlLanes[mid])return tlLanes[mid];
    if(tlNoData&&tlNoData.parentNode)tlNoData.parentNode.removeChild(tlNoData);
    tlNoData=null;
    var lane=document.createElement('div');
    lane.className='tl-lane';
    var label=document.createElement('div');
    label.className='tl-label';
    label.textContent=getMerchantName(mid);
    label.style.color=getMerchantColor(mid);
    var track=document.createElement('div');
    track.className='tl-track';
    lane.appendChild(label);
    lane.appendChild(track);
    tlScroll.appendChild(lane);
    tlLanes[mid]={lane:lane,track:track,label:label};
    return tlLanes[mid];
  }

  function addTimelineEvent(mid,action,elapsed,durationMs){
    var meta=actionMeta[action]||{cls:'thinking',icon:'?',label:action};
    var l=ensureLane(mid);
    var left=elapsed*PX_PER_SEC;
    var width=Math.max(18,(durationMs||200)/1000*PX_PER_SEC);
    var block=document.createElement('div');
    block.className='tl-event '+meta.cls;
    block.style.left=left+'px';
    block.style.width=width+'px';
    block.title=meta.label+(durationMs?' ('+durationMs+'ms)':'');
    block.textContent=meta.icon;
    l.track.appendChild(block);
    var endPx=left+width;
    if(endPx>tlMaxTime){
      tlMaxTime=endPx;
      var tracks=tlScroll.querySelectorAll('.tl-track');
      for(var i=0;i<tracks.length;i++){
        tracks[i].style.minWidth=(tlMaxTime+40)+'px';
      }
    }
    tlScroll.scrollLeft=tlScroll.scrollWidth;
    return block;
  }

  function updateTimelineBlock(mid,action,durationMs){
    if(!tlLanes[mid])return;
    var track=tlLanes[mid].track;
    var blocks=track.querySelectorAll('.tl-event');
    var meta=actionMeta[action];
    if(!meta)return;
    for(var i=blocks.length-1;i>=0;i--){
      if(blocks[i].classList.contains(meta.cls)){
        var width=Math.max(18,durationMs/1000*PX_PER_SEC);
        blocks[i].style.width=width+'px';
        blocks[i].title=meta.label+' ('+durationMs+'ms)';
        break;
      }
    }
  }

  // ============================================================
  // LEADERBOARD
  // ============================================================
  var lbBody=document.getElementById('lb-body');

  function renderLeaderboard(list){
    if(!list||list.length===0){
      lbBody.innerHTML='<div class="lb-no-data">Aucun marchand enregistre</div>';
      return;
    }
    var sorted=list.slice().sort(function(a,b){return b.net_profit-a.net_profit});
    var html='<table class="lb-table"><thead><tr>';
    html+='<th>#</th><th>Marchand</th><th>Ventes</th><th>Revenu</th><th>Pub</th><th>Profit</th><th></th>';
    html+='</tr></thead><tbody>';
    var medals={0:'&#129351;',1:'&#129352;',2:'&#129353;'};
    for(var i=0;i<sorted.length;i++){
      var m=sorted[i];
      var profitCls=m.net_profit>=0?'lb-positive':'lb-negative';
      var medal=medals[i]||(i+1);
      var emoji=m.emoji?m.emoji+' ':'';
      html+='<tr>';
      html+='<td class="lb-rank">'+medal+'</td>';
      html+='<td class="lb-name" style="color:'+(m.accent_color||'#1A1A2E')+'">'+esc(emoji+m.name)+'</td>';
      html+='<td>'+m.sales_count+'</td>';
      html+='<td>'+fmt(m.price*m.sales_count)+'</td>';
      html+='<td>'+fmt(m.total_ad_spend)+'</td>';
      html+='<td class="'+profitCls+'">'+fmt(m.net_profit)+'</td>';
      html+='<td><canvas class="lb-spark" data-mid="'+m.id+'"></canvas></td>';
      html+='</tr>';
    }
    html+='</tbody></table>';
    lbBody.innerHTML=html;

    // Draw sparklines
    var sparks=lbBody.querySelectorAll('.lb-spark');
    for(var i=0;i<sparks.length;i++){
      drawSparkline(sparks[i],sparks[i].getAttribute('data-mid'));
    }
  }

  function drawSparkline(canvas,mid){
    var data=profitSnapshots[mid];
    if(!data||data.length<2)return;
    var dpr=window.devicePixelRatio||1;
    var w=canvas.offsetWidth||60;
    var h=canvas.offsetHeight||20;
    canvas.width=w*dpr;
    canvas.height=h*dpr;
    var c=canvas.getContext('2d');
    c.setTransform(dpr,0,0,dpr,0,0);
    var min=Infinity,max=-Infinity;
    for(var i=0;i<data.length;i++){if(data[i]<min)min=data[i];if(data[i]>max)max=data[i]}
    var range=max-min||1;
    c.beginPath();
    for(var i=0;i<data.length;i++){
      var x=i/(data.length-1)*w;
      var y=h-((data[i]-min)/range)*(h-2)-1;
      if(i===0)c.moveTo(x,y);else c.lineTo(x,y);
    }
    var last=data[data.length-1];
    c.strokeStyle=last>=0?'#16A34A':'#DC2626';
    c.lineWidth=1.5;
    c.stroke();
  }

  // ============================================================
  // DECISION RECAP
  // ============================================================
  var dcBody=document.getElementById('dc-body');
  var dcNoData=document.getElementById('dc-no-data');
  var dcStatusText=document.getElementById('dc-status-text');

  function updateDecision(mid,totals,status){
    if(!decisionData[mid])decisionData[mid]={subtotal:0,discount:0,shipping:0,total:0,status:'pending'};
    if(totals){
      for(var i=0;i<totals.length;i++){
        var t=totals[i];
        if(t.type==='subtotal')decisionData[mid].subtotal=t.amount||0;
        if(t.type==='discount'||t.type==='items_discount')decisionData[mid].discount+=(t.amount||0);
        if(t.type==='fulfillment')decisionData[mid].shipping=t.amount||0;
        if(t.type==='total')decisionData[mid].total=t.amount||0;
      }
    }
    if(status)decisionData[mid].status=status;
    renderDecision();
  }

  function renderDecision(){
    var ids=Object.keys(decisionData);
    if(ids.length===0)return;
    if(dcNoData)dcNoData.style.display='none';

    var hasWinner=false;
    for(var i=0;i<ids.length;i++){if(decisionData[ids[i]].status==='winner')hasWinner=true}
    dcStatusText.textContent=hasWinner?'Termine':'En cours...';

    var sorted=ids.slice().sort(function(a,b){
      var da=decisionData[a],db=decisionData[b];
      if(da.status==='winner')return -1;
      if(db.status==='winner')return 1;
      return (da.total||Infinity)-(db.total||Infinity);
    });

    var html='<table class="dc-table"><thead><tr>';
    html+='<th>Marchand</th><th>Sous-total</th><th>Remise</th><th>Livraison</th><th>Total</th><th>Statut</th>';
    html+='</tr></thead><tbody>';
    for(var i=0;i<sorted.length;i++){
      var mid=sorted[i];
      var d=decisionData[mid];
      var rowCls='';
      var statusHtml='<span class="dc-status pending">...</span>';
      if(d.status==='winner'){
        rowCls='dc-winner';
        statusHtml='<span class="dc-status won">GAGNANT</span>';
      } else if(d.status==='canceled'){
        rowCls='dc-canceled';
        statusHtml='<span class="dc-status lost">Annule</span>';
      }
      html+='<tr class="'+rowCls+'">';
      html+='<td style="font-weight:700;color:'+getMerchantColor(mid)+'">'+esc(getMerchantName(mid))+'</td>';
      html+='<td>'+(d.subtotal?fmt(d.subtotal):'-')+'</td>';
      html+='<td>'+(d.discount?'-'+fmt(d.discount):'-')+'</td>';
      html+='<td>'+(d.shipping?fmt(d.shipping):'-')+'</td>';
      html+='<td style="font-weight:800">'+(d.total?fmt(d.total):'-')+'</td>';
      html+='<td>'+statusHtml+'</td>';
      html+='</tr>';
    }
    html+='</tbody></table>';
    dcBody.innerHTML=html;
  }

  // ============================================================
  // SSE
  // ============================================================
  var connDot=document.getElementById('conn-dot');
  var sseRetryDelay=1000;
  var es=null;
  function sseConnect(){
    if(es)return;
    es=new EventSource('/events');
    es.onopen=function(){connDot.className='conn-dot connected';sseRetryDelay=1000};
    es.onerror=function(){connDot.className='conn-dot disconnected';es.close();es=null;setTimeout(sseConnect,sseRetryDelay);sseRetryDelay=Math.min(sseRetryDelay*2,8000)};
    es.onmessage=handleSSE;
  }
  document.addEventListener('visibilitychange',function(){
    if(document.hidden){if(es){es.close();es=null;connDot.className='conn-dot disconnected'}}
    else{sseConnect();fetchRankings();fetchMerchants()}
  });

  function handleSSE(msg){
    try{
      var ev=JSON.parse(msg.data);
      var data=ev.data||{};
      var summary=ev.summary||'';

      // Agent lifecycle
      if(ev.type==='agent_start'){
        agentStartTime=new Date(ev.timestamp);
        decisionData={};
        checkoutMerchants={};
        resetTimeline();
        renderDecision();
        dcStatusText.textContent='Demarre...';
        if(dcNoData)dcNoData.style.display='';
        pulseCenter();
      }

      if(ev.type==='agent_done'){
        dcStatusText.textContent='Termine';
        tlElapsed.textContent='';
      }

      // Tool calls -> timeline + graph
      if(ev.type==='tool_call'&&data.action){
        var mid=extractMid(data.merchant_url);
        var elapsed=agentStartTime?(new Date(ev.timestamp)-agentStartTime)/1000:0;
        if(data.action==='search_products'){
          pulseCenter();
          // Add search to a "Graph" lane
          addTimelineEvent('__graph__','search_products',elapsed,0);
        } else if(mid){
          addTimelineEvent(mid,data.action,elapsed,0);
          pulseNode(mid);
          glowEdge(mid);
        }
        if(agentStartTime){
          tlElapsed.textContent=elapsed.toFixed(1)+'s';
        }

        // Track checkout merchant mapping
        if(data.action==='create_checkout'&&mid){
          // We will map checkout_id -> mid when we get the result
        }
        if(data.action==='complete_checkout'&&mid){
          updateDecision(mid,null,'winner');
        }
        if(data.action==='cancel_checkout'&&mid){
          updateDecision(mid,null,'canceled');
        }
      }

      // Tool results -> update timeline duration + decision data
      if(ev.type==='tool_result'&&data.action){
        var mid=data.params?extractMid(data.params.merchant_url):null;
        if(mid&&data.duration_ms){
          updateTimelineBlock(mid,data.action,data.duration_ms);
        }
        if(data.action==='search_products'&&data.duration_ms){
          updateTimelineBlock('__graph__','search_products',data.duration_ms);
        }

        // Extract checkout totals for decision
        if(mid&&data.response){
          var resp=data.response;
          // Response could be the checkout object (with totals) or a result wrapper
          var totals=resp.totals||(resp.result&&resp.result.totals);
          if(totals){
            updateDecision(mid,totals,null);
          }
        }
      }

      // Tool errors
      if(ev.type==='tool_error'&&data.action){
        var mid=data.params?extractMid(data.params.merchant_url):null;
        if(mid){
          var elapsed=agentStartTime?(new Date(ev.timestamp)-agentStartTime)/1000:0;
          addTimelineEvent(mid,data.action,elapsed,100);
        }
      }

      // Arena events -> graph effects
      if(ev.source==='arena'){
        if(ev.type==='sale_completed'){
          fetchMerchants();
        }
        if(ev.type==='merchant_registered'||ev.type==='config_update'||ev.type==='merchant_left'){
          fetchMerchants();
        }
      }

      // Merchant activity -> graph pulse
      if(ev.type==='checkout_created'||ev.type==='checkout_updated'||ev.type==='checkout_canceled'||ev.type==='product_details'){
        var cards=Object.keys(merchants);
        for(var i=0;i<cards.length;i++){
          if(merchants[cards[i]].name===ev.source){
            pulseNode(cards[i]);
            glowEdge(cards[i]);
            break;
          }
        }
      }

    }catch(ex){console.error('SSE handler error:',ex)}
  }
  sseConnect();

  // Fix graph lane label
  var origEnsureLane=ensureLane;
  ensureLane=function(mid){
    var result=origEnsureLane(mid);
    if(mid==='__graph__'){
      result.label.textContent='Shopping Graph';
      result.label.style.color='#1A1A2E';
    }
    return result;
  };
})();
</script>
</body>
</html>`

func (h *Handler) handleInsightsDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Write([]byte(insightsDashboardHTML))
}
