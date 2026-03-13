package obs

import "net/http"

const arenaDashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>UCP Arena Monitor</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: system-ui, -apple-system, sans-serif; background: #0E2356; color: #FFFFFF; overflow: hidden; height: 100vh; display: flex; flex-direction: column; }

.topbar { background: #263967; padding: 0.6rem 1.5rem; display: flex; align-items: center; gap: 1rem; border-bottom: 1px solid #3E4F78; flex-shrink: 0; }
.topbar h1 { font-size: 1.1rem; font-weight: 600; letter-spacing: 0.02em; }
.topbar .product-info { font-size: 0.85rem; color: #B7BDCC; margin-left: 0.5rem; }
.topbar .right { margin-left: auto; display: flex; align-items: center; gap: 0.75rem; }
.topbar .live-dot { width: 8px; height: 8px; border-radius: 50%; background: #00D2DD; display: inline-block; margin-right: 4px; animation: pulse-dot 1.5s ease-in-out infinite; }
@keyframes pulse-dot { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
.topbar .btn-cmd { background: #00D2DD; color: #0E2356; border: none; border-radius: 6px; padding: 0.35rem 0.75rem; font-size: 0.8rem; font-weight: 600; cursor: pointer; transition: opacity 0.2s; }
.topbar .btn-cmd:hover { opacity: 0.85; }
.topbar .btn-cmd:disabled { opacity: 0.4; cursor: not-allowed; }
.topbar .merchant-count { font-size: 0.8rem; color: #B7BDCC; }
.agent-status { font-size: 0.8rem; font-weight: 600; display: flex; align-items: center; gap: 0.35rem; }
.agent-status .status-dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; }
.agent-status.connected { color: #2ECC71; }
.agent-status.connected .status-dot { background: #2ECC71; }
.agent-status.disconnected { color: #8892A8; }
.agent-status.disconnected .status-dot { background: #8892A8; }

.winner-banner { display: none; background: linear-gradient(135deg, #00D2DD, #7b2ff7); padding: 1.2rem 1.5rem; text-align: center; flex-shrink: 0; animation: banner-in 0.4s ease-out; }
.winner-banner h2 { font-size: 1.5rem; color: #fff; margin-bottom: 0.25rem; }
.winner-banner p { font-size: 1rem; color: rgba(255,255,255,0.85); }
@keyframes banner-in { from { opacity: 0; transform: translateY(-20px); } to { opacity: 1; transform: translateY(0); } }

.main-area { flex: 1; display: flex; overflow: hidden; }

.activity-panel { width: 300px; background: #1B2F5E; border-right: 1px solid #3E4F78; display: flex; flex-direction: column; flex-shrink: 0; }
.activity-panel .panel-header { font-weight: 600; font-size: 0.75rem; color: #00D2DD; text-transform: uppercase; letter-spacing: 0.05em; padding: 0.6rem 0.75rem; border-bottom: 1px solid #3E4F78; }
.activity-panel .panel-body { flex: 1; overflow-y: auto; padding: 0.5rem 0.75rem; font-size: 0.8rem; line-height: 1.5; white-space: pre-wrap; word-wrap: break-word; color: #B7BDCC; }
.activity-panel .panel-body .thinking-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #263967; border-radius: 6px; color: #DDE0E8; }
.activity-panel .panel-body .result-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #1A3A2A; border: 1px solid #2ECC71; border-radius: 6px; color: #FFFFFF; }
.activity-panel .panel-body .error-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #3A1A1A; border: 1px solid #FF6B6B; border-radius: 6px; color: #FF6B6B; }
.activity-panel .panel-body .arena-registration { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #1A2A4A; border: 1px solid #00D2DD; border-radius: 6px; color: #00D2DD; }
.activity-panel .panel-body .arena-sale { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #1A3A2A; border: 1px solid #2ECC71; border-radius: 6px; color: #2ECC71; }
.activity-panel .panel-body .arena-config { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #2A2A1A; border: 1px solid #FF9800; border-radius: 6px; color: #FF9800; }
.activity-panel .panel-body .tool-call-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #1E2A4A; border: 1px solid #7b86a2; border-radius: 6px; color: #B0B8CC; }

.merchants-area { flex: 1; overflow-y: auto; padding: 1.5rem; }
.merchants-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1rem; }
.merchant-card { background: #1B2F5E; border: 1px solid #3E4F78; border-radius: 12px; padding: 1.2rem; transition: border-color 0.3s, box-shadow 0.3s; }
.merchant-card.active { border-color: #00D2DD; box-shadow: 0 0 16px rgba(0, 210, 221, 0.25); }
.merchant-card .mc-name { font-size: 1rem; font-weight: 600; margin-bottom: 0.75rem; display: flex; justify-content: space-between; align-items: center; }
.merchant-card .mc-name .sales-badge { background: #2ECC71; color: #0E2356; font-size: 0.7rem; font-weight: 700; padding: 0.15rem 0.5rem; border-radius: 10px; }
.merchant-card .mc-row { display: flex; justify-content: space-between; align-items: center; padding: 0.25rem 0; font-size: 0.85rem; color: #B7BDCC; border-bottom: 1px solid #2A3D6B; }
.merchant-card .mc-row:last-child { border-bottom: none; }
.merchant-card .mc-row .mc-label { color: #8892A8; }
.merchant-card .mc-row .mc-value { font-weight: 600; color: #DDE0E8; font-variant-numeric: tabular-nums; }
.merchant-card .mc-row .mc-value.positive { color: #2ECC71; }
.merchant-card .mc-row .mc-value.negative { color: #FF6B6B; }
.merchant-card .mc-row .mc-value.zero-stock { color: #FF6B6B; }

.no-merchants { text-align: center; padding: 4rem 2rem; color: #5A6A8A; }
.no-merchants h2 { font-size: 1.2rem; margin-bottom: 0.5rem; color: #8892A8; }
.no-merchants p { font-size: 0.9rem; }

.bottombar { background: #263967; padding: 0.5rem 1.5rem; border-top: 1px solid #3E4F78; display: flex; align-items: center; gap: 1rem; flex-shrink: 0; min-height: 48px; }
.bottombar .desc { flex: 1; font-size: 0.85rem; color: #B7BDCC; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.bottombar .badge { background: #00D2DD; color: #0E2356; border-radius: 4px; padding: 0.2rem 0.6rem; font-size: 0.75rem; font-weight: 600; white-space: nowrap; flex-shrink: 0; }

.modal-overlay { display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.55); z-index: 100; justify-content: center; align-items: center; }
.modal-overlay.visible { display: flex; }
.modal { background: #1B2F5E; border: 1px solid #3E4F78; border-radius: 12px; padding: 1.5rem; width: 380px; box-shadow: 0 8px 32px rgba(0,0,0,0.4); }
.modal h2 { font-size: 1rem; margin-bottom: 1rem; }
.modal input { width: 100%; padding: 0.5rem 0.75rem; border: 1px solid #3E4F78; border-radius: 6px; background: #0E2356; color: #FFF; font-size: 0.9rem; outline: none; }
.modal input:focus { border-color: #00D2DD; }
.modal .modal-buttons { display: flex; gap: 0.5rem; margin-top: 1rem; justify-content: flex-end; }
.modal .modal-buttons button { border: none; border-radius: 6px; padding: 0.4rem 1rem; font-size: 0.85rem; font-weight: 600; cursor: pointer; }
.modal .btn-search { background: #00D2DD; color: #0E2356; }
.modal .btn-cancel { background: #3E4F78; color: #B7BDCC; }
</style>
</head>
<body>

<div class="topbar">
  <h1>UCP Arena Monitor</h1>
  <span class="product-info" id="product-info"></span>
  <div class="right">
    <span class="merchant-count" id="merchant-count">0 merchants</span>
    <span id="agent-status" class="agent-status disconnected" title="No agent connected"><span class="status-dot"></span>Agent: offline</span>
    <button class="btn-cmd" id="btn-command" disabled title="No agent connected">Send Command</button>
    <span><span class="live-dot"></span>LIVE</span>
  </div>
</div>

<div class="winner-banner" id="winner-banner">
  <h2 id="winner-name"></h2>
  <p id="winner-detail"></p>
</div>

<div class="main-area">
  <div class="activity-panel">
    <div class="panel-header" id="panel-header">Activity Log</div>
    <div class="panel-body" id="panel-body"></div>
  </div>
  <div class="merchants-area">
    <div class="merchants-grid" id="merchants-grid">
      <div class="no-merchants">
        <h2>No merchants registered</h2>
        <p>Waiting for merchants to join the arena...</p>
      </div>
    </div>
  </div>
</div>

<div class="bottombar">
  <div class="desc" id="bottom-desc">Waiting for events...</div>
  <div class="badge" id="bottom-badge" style="display:none"></div>
</div>

<div class="modal-overlay" id="command-modal">
  <div class="modal">
    <h2>Send instruction to Agent</h2>
    <input type="text" id="command-input" placeholder="e.g. find the best headphones" autocomplete="off" />
    <div class="modal-buttons">
      <button class="btn-cancel" id="modal-cancel">Cancel</button>
      <button class="btn-search" id="modal-submit">Send</button>
    </div>
  </div>
</div>

<script>
(function() {
  var panelBody = document.getElementById('panel-body');
  var panelHeader = document.getElementById('panel-header');
  var bottomDesc = document.getElementById('bottom-desc');
  var bottomBadge = document.getElementById('bottom-badge');
  var merchantsGrid = document.getElementById('merchants-grid');
  var merchantCount = document.getElementById('merchant-count');
  var winnerBanner = document.getElementById('winner-banner');
  var winnerName = document.getElementById('winner-name');
  var winnerDetail = document.getElementById('winner-detail');
  var productInfo = document.getElementById('product-info');

  var activeMerchant = null;
  var activeTimer = null;
  var bannerTimer = null;

  // Fetch arena config for product info
  fetch('/arena/config')
    .then(function(r) { return r.ok ? r.json() : Promise.reject(); })
    .then(function(cfg) {
      productInfo.textContent = cfg.product_name + ' | Cost: $' + (cfg.cost_price / 100).toFixed(2);
    })
    .catch(function() {});

  function escapeHtml(s) {
    var d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }

  function formatPrice(cents) {
    return '$' + (cents / 100).toFixed(2);
  }

  function appendToPanel(className, text) {
    var div = document.createElement('div');
    div.className = className;
    div.textContent = text;
    panelBody.appendChild(div);
    panelBody.scrollTop = panelBody.scrollHeight;
  }

  function showWinnerBanner(title, detail) {
    winnerName.textContent = title;
    winnerDetail.textContent = detail;
    winnerBanner.style.display = 'block';
    if (bannerTimer) clearTimeout(bannerTimer);
    bannerTimer = setTimeout(function() { winnerBanner.style.display = 'none'; }, 8000);
  }

  // --- Merchant cards ---
  function renderMerchants(data) {
    var merchants = data.merchants || [];
    merchantCount.textContent = merchants.length + ' merchant' + (merchants.length !== 1 ? 's' : '');

    if (merchants.length === 0) {
      merchantsGrid.innerHTML = '<div class="no-merchants"><h2>No merchants registered</h2><p>Waiting for merchants to join the arena...</p></div>';
      return;
    }

    var html = '';
    for (var i = 0; i < merchants.length; i++) {
      var m = merchants[i];
      var isActive = activeMerchant && activeMerchant === m.name;
      var activeClass = isActive ? ' active' : '';
      var profitClass = m.total_profit > 0 ? 'positive' : (m.total_profit < 0 ? 'negative' : '');
      var stockClass = m.stock <= 0 ? 'zero-stock' : '';

      html += '<div class="merchant-card' + activeClass + '" data-name="' + escapeHtml(m.name) + '">' +
        '<div class="mc-name"><span>' + escapeHtml(m.name) + '</span>' +
        (m.sales_count > 0 ? '<span class="sales-badge">' + m.sales_count + ' sale' + (m.sales_count !== 1 ? 's' : '') + '</span>' : '') +
        '</div>' +
        '<div class="mc-row"><span class="mc-label">Price</span><span class="mc-value">' + formatPrice(m.price) + '</span></div>' +
        '<div class="mc-row"><span class="mc-label">Stock</span><span class="mc-value ' + stockClass + '">' + m.stock + '</span></div>' +
        '<div class="mc-row"><span class="mc-label">Boost</span><span class="mc-value">' + m.boost + '%</span></div>' +
        '<div class="mc-row"><span class="mc-label">Margin</span><span class="mc-value">' + formatPrice(m.margin) + '</span></div>' +
        '<div class="mc-row"><span class="mc-label">Net Margin</span><span class="mc-value">' + formatPrice(m.net_margin) + '</span></div>' +
        '<div class="mc-row"><span class="mc-label">Profit</span><span class="mc-value ' + profitClass + '">' + formatPrice(m.total_profit) + '</span></div>' +
        '</div>';
    }
    merchantsGrid.innerHTML = html;
  }

  function fetchMerchants() {
    fetch('/arena/merchants')
      .then(function(r) { return r.ok ? r.json() : Promise.reject(new Error('unreachable')); })
      .then(renderMerchants)
      .catch(function() {});
  }

  fetchMerchants();
  setInterval(fetchMerchants, 5000);

  // --- SSE events ---
  function detectMerchantName(ev) {
    var s = (ev.summary || '') + ' ' + (ev.source || '');
    var cards = document.querySelectorAll('.merchant-card');
    for (var i = 0; i < cards.length; i++) {
      var name = cards[i].getAttribute('data-name');
      if (name && s.indexOf(name) !== -1) return name;
    }
    return null;
  }

  function highlightMerchant(name) {
    if (activeTimer) clearTimeout(activeTimer);
    activeMerchant = name;

    var cards = document.querySelectorAll('.merchant-card');
    for (var i = 0; i < cards.length; i++) {
      if (cards[i].getAttribute('data-name') === name) {
        cards[i].classList.add('active');
      } else {
        cards[i].classList.remove('active');
      }
    }

    activeTimer = setTimeout(function() {
      activeMerchant = null;
      var all = document.querySelectorAll('.merchant-card');
      for (var j = 0; j < all.length; j++) all[j].classList.remove('active');
    }, 3000);
  }

  var es = new EventSource('/events');
  es.onmessage = function(msg) {
    try {
      var ev = JSON.parse(msg.data);

      // Arena-specific events
      if (ev.source === 'arena') {
        if (ev.type === 'merchant_registered') {
          appendToPanel('arena-registration', ev.summary);
          fetchMerchants();
        } else if (ev.type === 'sale_completed') {
          appendToPanel('arena-sale', ev.summary);
          showWinnerBanner('SOLD!', ev.summary);
          fetchMerchants();
        } else if (ev.type === 'config_update') {
          appendToPanel('arena-config', ev.summary);
          fetchMerchants();
        }
      }

      // Agent activity panel (only milestones, skip verbose polling)
      if (ev.type === 'agent_start') {
        panelBody.innerHTML = '';
        panelHeader.textContent = 'Activity Log';
        appendToPanel('thinking-entry', 'Agent started');
      }
      if (ev.type === 'agent_thinking' && ev.summary) appendToPanel('thinking-entry', ev.summary);
      if (ev.type === 'tool_call' && ev.summary) appendToPanel('tool-call-entry', ev.summary);
      if (ev.type === 'tool_result' && ev.summary) appendToPanel('result-entry', ev.summary);
      if (ev.type === 'tool_error' && ev.summary) appendToPanel('error-entry', ev.summary);
      if (ev.type === 'agent_error' && ev.summary) appendToPanel('error-entry', ev.summary);
      if (ev.type === 'agent_done' && ev.summary) {
        panelHeader.textContent = 'Agent Result';
        appendToPanel('result-entry', ev.summary);
        fetchMerchants();
      }

      // Bottom bar
      bottomDesc.textContent = ev.summary || ev.type || '';
      if (ev.type === 'tool_error' || ev.type === 'agent_error') {
        bottomBadge.textContent = 'Error';
        bottomBadge.style.display = '';
        bottomBadge.style.background = '#FF6B6B';
        bottomBadge.style.color = '#FFFFFF';
      } else if (ev.type) {
        bottomBadge.textContent = ev.type.replace(/_/g, ' ');
        bottomBadge.style.display = '';
        bottomBadge.style.background = '#00D2DD';
        bottomBadge.style.color = '#0E2356';
      }

      // Highlight merchant
      var mName = detectMerchantName(ev);
      if (mName) highlightMerchant(mName);

      // Refresh merchants on sales or completions
      if (ev.type === 'tool_result' && ev.summary && ev.summary.indexOf('omplete') !== -1) {
        fetchMerchants();
      }
    } catch(e) { console.error(e); }
  };

  // --- Command modal ---
  var modal = document.getElementById('command-modal');
  var cmdInput = document.getElementById('command-input');

  document.getElementById('btn-command').addEventListener('click', function() {
    modal.classList.add('visible');
    cmdInput.value = '';
    cmdInput.focus();
  });

  function hideModal() { modal.classList.remove('visible'); }
  document.getElementById('modal-cancel').addEventListener('click', hideModal);

  function submitCommand() {
    var val = cmdInput.value.trim();
    if (!val) return;
    hideModal();
    bottomDesc.textContent = 'Sending: ' + val;
    bottomBadge.textContent = 'Sending...';
    bottomBadge.style.display = '';
    bottomBadge.style.background = '#00D2DD';
    bottomBadge.style.color = '#0E2356';
    fetch('/command', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({instruction: val}) })
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.connected) {
          bottomBadge.textContent = 'Command sent';
          bottomBadge.style.background = '#2ECC71';
          bottomBadge.style.color = '#0E2356';
        } else {
          bottomBadge.textContent = 'No agent connected';
          bottomBadge.style.background = '#FF9800';
          bottomBadge.style.color = '#0E2356';
        }
      })
      .catch(function() {
        bottomBadge.textContent = 'Send failed';
        bottomBadge.style.background = '#FF6B6B';
        bottomBadge.style.color = '#FFFFFF';
      });
  }

  document.getElementById('modal-submit').addEventListener('click', submitCommand);
  cmdInput.addEventListener('keydown', function(e) {
    if (e.key === 'Enter') submitCommand();
    if (e.key === 'Escape') hideModal();
  });

  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape' && modal.classList.contains('visible')) hideModal();
  });

  // --- Agent status polling ---
  var agentStatus = document.getElementById('agent-status');
  var btnCommand = document.getElementById('btn-command');
  function pollAgentStatus() {
    fetch('/status')
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.agent_connected) {
          agentStatus.className = 'agent-status connected';
          agentStatus.title = 'Agent connected';
          agentStatus.innerHTML = '<span class="status-dot"></span>Agent: online';
          btnCommand.disabled = false;
          btnCommand.title = '';
        } else {
          agentStatus.className = 'agent-status disconnected';
          agentStatus.title = 'No agent connected';
          agentStatus.innerHTML = '<span class="status-dot"></span>Agent: offline';
          btnCommand.disabled = true;
          btnCommand.title = 'No agent connected';
        }
      })
      .catch(function() {});
  }
  pollAgentStatus();
  setInterval(pollAgentStatus, 3000);
})();
</script>
</body>
</html>`

func (h *Handler) handleArenaDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(arenaDashboardHTML))
}
