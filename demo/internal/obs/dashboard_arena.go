package obs

import "net/http"

const arenaDashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>UCP Arena Monitor</title>
<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;600;700;800&display=swap" rel="stylesheet">
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: 'Outfit', system-ui, sans-serif; background: #FDF0EE; color: #1A1A2E; overflow: hidden; height: 100vh; display: flex; flex-direction: column; }

.topbar { background: #FFFFFF; padding: 0.6rem 1.5rem; display: flex; align-items: center; gap: 1rem; border-bottom: 1px solid #E0E0E0; flex-shrink: 0; }
.topbar h1 { font-size: 1.1rem; font-weight: 800; letter-spacing: 0.02em; color: #1A1A2E; }
.topbar h1 span { color: #E5004C; }
.topbar .product-info { font-size: 0.85rem; color: #666; margin-left: 0.5rem; }
.topbar .right { margin-left: auto; display: flex; align-items: center; gap: 0.75rem; }
.topbar .live-dot { width: 8px; height: 8px; border-radius: 50%; background: #E5004C; display: inline-block; margin-right: 4px; animation: pulse-dot 1.5s ease-in-out infinite; }
@keyframes pulse-dot { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
.conn-dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; margin-left: 6px; vertical-align: middle; }
.conn-dot.connected { background: #16A34A; }
.conn-dot.disconnected { background: #DC2626; animation: pulse-conn 1s ease-in-out infinite; }
@keyframes pulse-conn { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
.topbar .merchant-count { font-size: 0.8rem; color: #666; }

.winner-banner { display: none; background: #E5004C; padding: 1.2rem 1.5rem; text-align: center; flex-shrink: 0; animation: banner-in 0.4s ease-out; }
.winner-banner h2 { font-size: 1.5rem; color: #fff; margin-bottom: 0.25rem; font-weight: 800; }
.winner-banner p { font-size: 1rem; color: rgba(255,255,255,0.85); }
@keyframes banner-in { from { opacity: 0; transform: translateY(-20px); } to { opacity: 1; transform: translateY(0); } }

.main-area { flex: 1; display: flex; overflow: hidden; }

.activity-panel { width: 300px; background: #FFFFFF; border-right: 1px solid #E0E0E0; display: flex; flex-direction: column; flex-shrink: 0; }
.activity-panel .panel-header { font-weight: 700; font-size: 0.75rem; color: #E5004C; text-transform: uppercase; letter-spacing: 0.05em; padding: 0.6rem 0.75rem; border-bottom: 1px solid #E0E0E0; }
.activity-panel .panel-body { flex: 1; overflow-y: auto; padding: 0.5rem 0.75rem; font-size: 0.8rem; line-height: 1.5; white-space: pre-wrap; word-wrap: break-word; color: #666; }
.activity-panel .panel-body .thinking-entry, .activity-panel .panel-body .result-entry, .activity-panel .panel-body .error-entry, .activity-panel .panel-body .arena-registration, .activity-panel .panel-body .arena-sale, .activity-panel .panel-body .arena-config, .activity-panel .panel-body .tool-call-entry { cursor: pointer; transition: opacity 0.15s; }
.activity-panel .panel-body .thinking-entry:hover, .activity-panel .panel-body .result-entry:hover, .activity-panel .panel-body .error-entry:hover, .activity-panel .panel-body .arena-registration:hover, .activity-panel .panel-body .arena-sale:hover, .activity-panel .panel-body .arena-config:hover, .activity-panel .panel-body .tool-call-entry:hover { opacity: 0.75; }
.activity-panel .panel-body .thinking-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #F9FAFB; border-radius: 8px; color: #2D2D2D; }
.activity-panel .panel-body .result-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #DCFCE7; border: 1px solid #16A34A; border-radius: 8px; color: #16A34A; }
.activity-panel .panel-body .error-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #FEF2F2; border: 1px solid #DC2626; border-radius: 8px; color: #DC2626; }
.activity-panel .panel-body .arena-registration { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #FDE8E8; border: 1px solid #E5004C; border-radius: 8px; color: #E5004C; }
.activity-panel .panel-body .arena-sale { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #DCFCE7; border: 1px solid #16A34A; border-radius: 8px; color: #16A34A; }
.activity-panel .panel-body .arena-config { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #FFF7ED; border: 1px solid #F59E0B; border-radius: 8px; color: #D97706; }
.activity-panel .panel-body .tool-call-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #F3F4F6; border: 1px solid #9CA3AF; border-radius: 8px; color: #6B7280; }
.proto-toggle { display: inline-block; font-size: 0.65rem; margin-left: 4px; color: #999; cursor: pointer; vertical-align: middle; user-select: none; }
.proto-toggle:hover { color: #E5004C; }
.proto-payload { display: none; margin-top: 0.35rem; padding: 0.4rem 0.5rem; background: #1A1A2E; color: #A5F3FC; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 0.7rem; line-height: 1.4; white-space: pre-wrap; word-break: break-all; max-height: 200px; overflow-y: auto; }
.proto-payload.visible { display: block; }
.proto-payload .pk { color: #F9A8D4; }
.proto-payload .ps { color: #A5F3FC; }
.proto-payload .pn { color: #FDE68A; }
.proto-payload .pp { color: #D1D5DB; }
.proto-badge { display: inline-block; font-size: 0.6rem; font-weight: 700; padding: 0.1rem 0.35rem; border-radius: 4px; margin-left: 4px; vertical-align: middle; }
.proto-badge.req { background: #DBEAFE; color: #3B82F6; }
.proto-badge.res { background: #DCFCE7; color: #16A34A; }
.proto-badge.err { background: #FEF2F2; color: #DC2626; }
.proto-duration { font-size: 0.6rem; color: #999; margin-left: 4px; }

.merchants-area { flex: 1; overflow-y: auto; padding: 1.5rem; }
.merchants-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1rem; }
.merchant-card { background: #FFFFFF; border: 1px solid #2D2D2D; border-radius: 16px; padding: 1.2rem; transition: border-color 0.3s, box-shadow 0.3s; box-shadow: 4px 4px 0px #E5004C; }
.merchant-card.active { border-color: #E5004C; box-shadow: 6px 6px 0px #E5004C; }
.merchant-card.state-checkout { border-color: #3B82F6; background: #EFF6FF; box-shadow: 4px 4px 0px #3B82F6; }
.merchant-card.state-negotiate { border-color: #F59E0B; background: #FFFBEB; box-shadow: 4px 4px 0px #F59E0B; }
.merchant-card.state-sale { border-color: #16A34A; background: #F0FDF4; box-shadow: 6px 6px 0px #16A34A; }
.merchant-card.state-lookup { border-color: #E5004C; background: #FDE8E8; box-shadow: 4px 4px 0px #E5004C; }
.mc-status { display: none; margin-top: 0.6rem; padding: 0.3rem 0.6rem; border-radius: 20px; font-size: 0.75rem; font-weight: 600; text-align: center; animation: status-in 0.25s ease-out; }
@keyframes status-in { from { opacity: 0; transform: translateY(4px); } to { opacity: 1; transform: translateY(0); } }
.mc-status.visible { display: block; }
.mc-status.st-lookup { background: #FDE8E8; color: #E5004C; border: 1px solid #E5004C; }
.mc-status.st-checkout { background: #DBEAFE; color: #3B82F6; border: 1px solid #3B82F6; }
.mc-status.st-negotiate { background: #FFF7ED; color: #D97706; border: 1px solid #F59E0B; }
.mc-status.st-sale { background: #DCFCE7; color: #16A34A; border: 1px solid #16A34A; }
.mc-status.st-cancel { background: #FEF2F2; color: #DC2626; border: 1px solid #DC2626; }
.merchant-card .mc-name { font-size: 1rem; font-weight: 700; margin-bottom: 0.75rem; display: flex; justify-content: space-between; align-items: center; color: #1A1A2E; }
.merchant-card .mc-name .sales-badge { background: #FDE8E8; color: #E5004C; font-size: 0.7rem; font-weight: 700; padding: 0.15rem 0.5rem; border-radius: 20px; }
.mc-rank { font-size: 0.7rem; font-weight: 700; padding: 0.15rem 0.5rem; border-radius: 20px; background: #FDE8E8; color: #E5004C; }
.mc-rank.top-1 { background: #FFD700; color: #1A1A2E; }
.mc-rank.top-2 { background: #C0C0C0; color: #1A1A2E; }
.mc-rank.top-3 { background: #CD7F32; color: #FFF; }
.merchant-card .mc-row { display: flex; justify-content: space-between; align-items: center; padding: 0.25rem 0; font-size: 0.85rem; color: #666; border-bottom: 1px solid #E0E0E0; }
.merchant-card .mc-row:last-child { border-bottom: none; }
.merchant-card .mc-row .mc-label { color: #999; }
.merchant-card .mc-row .mc-value { font-weight: 600; color: #1A1A2E; font-variant-numeric: tabular-nums; }
.merchant-card .mc-row .mc-value.positive { color: #16A34A; }
.merchant-card .mc-row .mc-value.negative { color: #DC2626; }
.merchant-card .mc-row .mc-value.zero-stock { color: #DC2626; }

.no-merchants { text-align: center; padding: 4rem 2rem; color: #999; }
.no-merchants h2 { font-size: 1.2rem; margin-bottom: 0.5rem; color: #666; font-weight: 700; }
.no-merchants p { font-size: 0.9rem; }

/* --- Agent Acheteur panel (right side, visually separate) --- */
.agent-panel { width: 320px; background: #FDF0EE; border-left: 1px solid #E0E0E0; display: flex; flex-direction: column; flex-shrink: 0; padding: 1rem; gap: 0.75rem; }
.agent-card { background: #FFFFFF; border: 1px solid #2D2D2D; border-radius: 16px; box-shadow: 6px 6px 0px #1A1A2E; overflow: hidden; display: flex; flex-direction: column; }
.agent-card-dots { padding: 0.4rem 0.75rem; border-bottom: 1px solid #E0E0E0; display: flex; align-items: center; gap: 6px; }
.agent-card-dots::before { content: ''; width: 10px; height: 10px; border-radius: 50%; background: #1A1A2E; display: inline-block; }
.agent-card-dots::after { content: ''; width: 10px; height: 10px; border-radius: 50%; background: #CCC; display: inline-block; }
.agent-card-dots .agent-card-title { margin-left: 0.5rem; font-size: 0.7rem; font-weight: 700; color: #999; text-transform: uppercase; letter-spacing: 0.05em; }
.agent-card-body { padding: 1rem; }
.agent-card-body .agent-identity { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.75rem; }
.agent-avatar { width: 40px; height: 40px; border-radius: 50%; background: #1A1A2E; display: flex; align-items: center; justify-content: center; font-size: 1.2rem; flex-shrink: 0; }
.agent-info .agent-name { font-size: 1rem; font-weight: 800; color: #1A1A2E; }
.agent-info .agent-role { font-size: 0.75rem; color: #999; }
.agent-status-line { display: flex; align-items: center; gap: 0.35rem; font-size: 0.8rem; font-weight: 600; margin-bottom: 0.75rem; }
.agent-status-line .status-dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; }
.agent-status-line.connected { color: #16A34A; }
.agent-status-line.connected .status-dot { background: #16A34A; }
.agent-status-line.disconnected { color: #999; }
.agent-status-line.disconnected .status-dot { background: #999; }
.agent-separator { border: none; border-top: 1px solid #E0E0E0; margin: 0.25rem 0 0.75rem; }
.agent-input-wrap { display: flex; gap: 0.5rem; }
.agent-input-wrap input { flex: 1; padding: 0.5rem 0.75rem; border: 1px solid #CCC; border-radius: 8px; background: #FFFFFF; color: #1A1A2E; font-size: 0.85rem; outline: none; font-family: 'Outfit', system-ui, sans-serif; }
.agent-input-wrap input:focus { border-color: #1A1A2E; }
.agent-input-wrap input:disabled { opacity: 0.4; }
.agent-input-wrap button { background: #1A1A2E; color: #FFFFFF; border: none; border-radius: 8px; padding: 0.5rem 0.75rem; font-size: 0.85rem; font-weight: 700; cursor: pointer; transition: opacity 0.2s; white-space: nowrap; }
.agent-input-wrap button:hover { opacity: 0.85; }
.agent-input-wrap button:disabled { opacity: 0.4; cursor: not-allowed; }
.agent-merchant-count { display: flex; align-items: center; gap: 0.5rem; font-size: 0.75rem; color: #999; margin-top: 0.5rem; }
.agent-merchant-count input[type=range] { flex: 1; accent-color: #1A1A2E; }
.agent-merchant-count .count-val { font-weight: 700; color: #1A1A2E; min-width: 1.2rem; text-align: center; }
.agent-send-status { font-size: 0.75rem; margin-top: 0.4rem; min-height: 1.1rem; }
.agent-send-status.success { color: #16A34A; }
.agent-send-status.error { color: #DC2626; }
.agent-send-status.warning { color: #D97706; }
.agent-send-status.sending { color: #999; }
.agent-note { font-size: 0.7rem; color: #999; line-height: 1.4; margin-top: auto; padding-top: 0.5rem; }

/* --- Agent expanded (overlay) state --- */
.agent-panel.expanded { position: fixed; inset: 0; width: 100%; z-index: 200; background: rgba(0,0,0,0.35); display: flex; align-items: center; justify-content: center; padding: 2rem; border: none; }
.agent-panel.expanded .agent-card { width: 680px; max-height: 85vh; box-shadow: 10px 10px 0px #1A1A2E; animation: agent-pop 0.25s ease-out; }
.agent-panel.expanded .agent-card-body { padding: 2rem; }
.agent-panel.expanded .agent-avatar { width: 72px; height: 72px; font-size: 2rem; }
.agent-panel.expanded .agent-name { font-size: 1.8rem; }
.agent-panel.expanded .agent-role { font-size: 1rem; }
.agent-panel.expanded .agent-status-line { font-size: 1.1rem; margin-bottom: 1rem; }
.agent-panel.expanded .agent-separator { margin: 0.5rem 0 1rem; }
.agent-panel.expanded .agent-input-wrap input { font-size: 1.15rem; padding: 0.8rem 1.2rem; }
.agent-panel.expanded .agent-input-wrap button { font-size: 1.15rem; padding: 0.8rem 1.2rem; }
.agent-panel.expanded .agent-merchant-count { font-size: 1rem; margin-top: 0.75rem; }
.agent-panel.expanded .agent-send-status { font-size: 0.95rem; }
.agent-panel.expanded .agent-note { font-size: 0.9rem; }
@keyframes agent-pop { from { opacity: 0; transform: scale(0.92); } to { opacity: 1; transform: scale(1); } }

/* --- Agent Modal Overlay --- */
.agent-modal-overlay { display: none; position: fixed; inset: 0; z-index: 300; background: rgba(26,26,46,0.6); align-items: center; justify-content: center; padding: 2rem; }
.agent-modal-overlay.visible { display: flex; }
.agent-modal { background: #FFFFFF; border: 2px solid #2D2D2D; border-radius: 16px; box-shadow: 8px 8px 0px #E5004C; width: 720px; max-width: 90vw; max-height: 80vh; display: flex; flex-direction: column; animation: agent-pop 0.3s ease-out; overflow: hidden; }
.agent-modal-dots { padding: 0.6rem 1rem; border-bottom: 1px solid #E0E0E0; display: flex; align-items: center; }
.agent-modal-dots::before { content: ''; width: 10px; height: 10px; border-radius: 50%; background: #E5004C; display: inline-block; margin-right: 8px; }
.agent-modal-title { font-size: 0.75rem; font-weight: 700; color: #E5004C; text-transform: uppercase; letter-spacing: 0.05em; flex: 1; }
.agent-modal-close { background: none; border: none; font-size: 1.5rem; color: #999; cursor: pointer; padding: 0 0.25rem; line-height: 1; }
.agent-modal-close:hover { color: #E5004C; }
.agent-modal-body { padding: 2rem 2.5rem; overflow-y: auto; font-size: 1.15rem; line-height: 1.7; color: #1A1A2E; }
.agent-modal-body h1, .agent-modal-body h2, .agent-modal-body h3 { font-weight: 800; color: #1A1A2E; margin: 1rem 0 0.5rem; }
.agent-modal-body strong { color: #E5004C; }
.agent-modal-body ul, .agent-modal-body ol { margin: 0.75rem 0; padding-left: 1.5rem; }
.agent-modal-body li { margin-bottom: 0.4rem; }
.agent-modal-body a { color: #E5004C; text-decoration: underline; }
.agent-modal-body p { margin-bottom: 0.75rem; }
.agent-modal-body code { background: #F3F4F6; padding: 0.15rem 0.4rem; border-radius: 4px; font-size: 0.95em; }
.agent-modal-progress { height: 3px; background: #E5004C; width: 100%; transition: width linear; }

.agent-card-dots { cursor: pointer; }
.agent-card-dots:hover { background: #F3F4F6; }

.bottombar { background: #FFFFFF; padding: 0.5rem 1.5rem; border-top: 1px solid #E0E0E0; display: flex; align-items: center; gap: 1rem; flex-shrink: 0; min-height: 48px; }
.bottombar .desc { flex: 1; font-size: 0.85rem; color: #666; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.bottombar .badge { background: #FDE8E8; color: #E5004C; border-radius: 20px; padding: 0.2rem 0.6rem; font-size: 0.75rem; font-weight: 600; white-space: nowrap; flex-shrink: 0; }
</style>
<script src="https://cdn.jsdelivr.net/npm/marked@12.0.0/marked.min.js"></script>
</head>
<body>

<div class="topbar">
  <h1>UCP <span>Arena</span></h1>
  <span class="product-info" id="product-info"></span>
  <div class="right">
    <a href="/insights" style="color:#E5004C;text-decoration:none;font-weight:700;font-size:0.8rem;padding:0.25rem 0.5rem;border:1px solid #E5004C;border-radius:8px">Insights</a>
    <span class="merchant-count" id="merchant-count">0 merchants</span>
    <span><span class="live-dot"></span>LIVE<span class="conn-dot disconnected" id="conn-dot" title="SSE"></span></span>
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
  <div class="agent-panel">
    <div class="agent-card">
      <div class="agent-card-dots"><span class="agent-card-title">External Service</span></div>
      <div class="agent-card-body">
        <div class="agent-identity">
          <div class="agent-avatar">🛒</div>
          <div class="agent-info">
            <div class="agent-name">Agent Acheteur</div>
            <div class="agent-role">Gemini &middot; independent buyer</div>
          </div>
        </div>
        <div id="agent-status" class="agent-status-line disconnected">
          <span class="status-dot"></span> Disconnected
        </div>
        <hr class="agent-separator">
        <div class="agent-input-wrap">
          <input type="text" id="command-input" placeholder="What should I buy?" autocomplete="off" disabled />
          <button id="btn-send" disabled>Send</button>
        </div>
        <div class="agent-merchant-count">
          <span>Merchants:</span>
          <input type="range" id="merchant-count-input" min="1" max="100" value="3" />
          <span class="count-val" id="merchant-count-value">3</span>
        </div>
        <div class="agent-send-status" id="send-status"></div>
      </div>
    </div>
    <div class="agent-note">
      The buying agent is an independent process (Gemini).
      It is not part of the arena — it discovers and negotiates with merchants on its own via the Shopping Graph.
    </div>
  </div>
</div>

<div class="agent-modal-overlay" id="agent-modal-overlay">
  <div class="agent-modal">
    <div class="agent-modal-dots">
      <span class="agent-modal-title">Agent Acheteur</span>
      <button class="agent-modal-close" id="agent-modal-close">&times;</button>
    </div>
    <div class="agent-modal-body" id="agent-modal-body"></div>
    <div class="agent-modal-progress" id="agent-modal-progress"></div>
  </div>
</div>

<div class="bottombar">
  <div class="desc" id="bottom-desc">Waiting for events...</div>
  <div class="badge" id="bottom-badge" style="display:none"></div>
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
  var currentRankings = {};

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

  function syntaxHighlight(json) {
    var s = JSON.stringify(json, null, 2);
    return s.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function(m) {
      if (/^"/.test(m)) {
        if (/:$/.test(m)) return '<span class="pk">' + m + '</span>';
        return '<span class="ps">' + m + '</span>';
      }
      if (/true|false/.test(m)) return '<span class="pn">' + m + '</span>';
      if (/null/.test(m)) return '<span class="pp">' + m + '</span>';
      return '<span class="pn">' + m + '</span>';
    });
  }

  function appendToPanel(className, text, rawText, protoData) {
    var div = document.createElement('div');
    div.className = className;
    var textSpan = document.createElement('span');
    textSpan.textContent = text;
    div.appendChild(textSpan);
    div.setAttribute('data-raw', rawText || text);

    if (protoData) {
      var display = Object.assign({}, protoData);
      delete display._type;
      div.setAttribute('data-proto', JSON.stringify(display));
      var badge = document.createElement('span');
      badge.className = 'proto-badge ' + (protoData._type || 'req');
      badge.textContent = protoData._type === 'res' ? 'RES' : protoData._type === 'err' ? 'ERR' : 'REQ';
      div.appendChild(badge);
      if (protoData.duration_ms) {
        var dur = document.createElement('span');
        dur.className = 'proto-duration';
        dur.textContent = protoData.duration_ms + 'ms';
        div.appendChild(dur);
      }
      var toggle = document.createElement('span');
      toggle.className = 'proto-toggle';
      toggle.textContent = '{ }';
      toggle.title = 'Show A2A payload';
      div.appendChild(toggle);
      var payload = document.createElement('div');
      payload.className = 'proto-payload';
      payload.innerHTML = syntaxHighlight(display);
      div.appendChild(payload);
      toggle.addEventListener('click', function(e) {
        e.stopPropagation();
        payload.classList.toggle('visible');
        toggle.textContent = payload.classList.contains('visible') ? 'Hide' : '{ }';
      });
    }

    div.addEventListener('click', function(e) {
      if (e.target.classList.contains('proto-toggle')) return;
      var proto = div.getAttribute('data-proto');
      showAgentModal(div.getAttribute('data-raw'), 5000, proto ? JSON.parse(proto) : null);
    });
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

  function fetchRankings() {
    fetch('/arena/rankings')
      .then(function(r) { return r.ok ? r.json() : Promise.reject(); })
      .then(function(data) {
        currentRankings = data.rankings || {};
      })
      .catch(function() {})
      .then(function() { fetchMerchants(); });
  }

  // --- Merchant cards ---
  function renderMerchants(data) {
    var merchants = data.merchants || [];
    merchantCount.textContent = merchants.length + ' merchant' + (merchants.length !== 1 ? 's' : '');

    if (merchants.length === 0) {
      merchantsGrid.innerHTML = '<div class="no-merchants"><h2>No merchants registered</h2><p>Waiting for merchants to join the arena...</p></div>';
      return;
    }

    merchants.sort(function(a, b) {
      var rdA = currentRankings[a.id];
      var ra = (rdA && typeof rdA === 'object' && typeof rdA.rank === 'number') ? rdA.rank : 9999;
      var rdB = currentRankings[b.id];
      var rb = (rdB && typeof rdB === 'object' && typeof rdB.rank === 'number') ? rdB.rank : 9999;
      if (ra !== rb) return ra - rb;
      return (a.id < b.id) ? -1 : 1;
    });

    var html = '';
    for (var i = 0; i < merchants.length; i++) {
      var m = merchants[i];
      var isActive = activeMerchant && activeMerchant === m.name;
      var activeClass = isActive ? ' active' : '';
      var profitClass = m.net_profit > 0 ? 'positive' : (m.net_profit < 0 ? 'negative' : '');
      var stockClass = m.stock <= 0 ? 'zero-stock' : '';

      var rData = currentRankings[m.id];
      var rankNum = (rData && typeof rData === 'object' && typeof rData.rank === 'number') ? rData.rank : 0;
      var rankBadge = '';
      if (rankNum > 0) {
        var rankCls = rankNum <= 3 ? ' top-' + rankNum : '';
        rankBadge = '<span class="mc-rank' + rankCls + '">#' + rankNum + '</span>';
      }

      var cardStyle = m.accent_color ? ' style="border-left:4px solid ' + m.accent_color + '"' : '';
      var emojiPrefix = m.emoji ? m.emoji + ' ' : '';
      html += '<div class="merchant-card' + activeClass + '"' + cardStyle + ' data-name="' + escapeHtml(m.name) + '" data-id="' + escapeHtml(m.id || m.name) + '">' +
        '<div class="mc-name"><span>' + rankBadge + ' ' + emojiPrefix + escapeHtml(m.name) + '</span>' +
        (m.sales_count > 0 ? '<span class="sales-badge">' + m.sales_count + ' sale' + (m.sales_count !== 1 ? 's' : '') + '</span>' : '') +
        '</div>' +
        '<div class="mc-row"><span class="mc-label">Price</span><span class="mc-value">' + formatPrice(m.price) + '</span></div>' +
        '<div class="mc-row"><span class="mc-label">Stock</span><span class="mc-value ' + stockClass + '">' + m.stock + '</span></div>' +
        '<div class="mc-row"><span class="mc-label">Max Bid</span><span class="mc-value">' + formatPrice(m.max_cpc_bid) + '/visit</span></div>' +
        '<div class="mc-row"><span class="mc-label">Actual CPC</span><span class="mc-value">' + formatPrice(m.actual_cpc) + '/visit</span></div>' +
        '<div class="mc-row"><span class="mc-label">Ad Spend</span><span class="mc-value">' + formatPrice(m.total_ad_spend) + ' (' + (m.consultation_count || 0) + ' visits)</span></div>' +
        '<div class="mc-row"><span class="mc-label">Profit</span><span class="mc-value ' + profitClass + '">' + formatPrice(m.net_profit) + '</span></div>' +
        '<div class="mc-status" data-status-for="' + escapeHtml(m.name) + '"></div>' +
        '</div>';
    }
    // Capture active states before replacing DOM
    var prevStates = {};
    var oldCards = merchantsGrid.querySelectorAll('.merchant-card');
    for (var j = 0; j < oldCards.length; j++) {
      var cn = oldCards[j].getAttribute('data-name');
      var se = oldCards[j].querySelector('.mc-status');
      if (cn && se && se.classList.contains('visible')) {
        prevStates[cn] = { cls: se.className, text: se.textContent, cardCls: oldCards[j].className };
      }
    }
    merchantsGrid.innerHTML = html;
    // Restore active states on new cards
    var newCards = merchantsGrid.querySelectorAll('.merchant-card');
    for (var k = 0; k < newCards.length; k++) {
      var nm = newCards[k].getAttribute('data-name');
      if (nm && prevStates[nm]) {
        var ps = prevStates[nm];
        // Restore card-level state classes
        ['state-checkout','state-negotiate','state-sale','state-lookup'].forEach(function(sc) {
          if (ps.cardCls.indexOf(sc) !== -1) newCards[k].classList.add(sc);
        });
        if (ps.cardCls.indexOf('active') !== -1) newCards[k].classList.add('active');
        var nse = newCards[k].querySelector('.mc-status');
        if (nse) { nse.className = ps.cls; nse.textContent = ps.text; }
      }
    }
  }

  function fetchMerchants() {
    fetch('/arena/merchants')
      .then(function(r) { return r.ok ? r.json() : Promise.reject(new Error('unreachable')); })
      .then(renderMerchants)
      .catch(function() {});
  }

  fetchRankings();
  setInterval(fetchRankings, 3000);

  // --- SSE events ---
  function detectMerchantName(ev) {
    var s = (ev.summary || '') + ' ' + (ev.source || '');
    var cards = document.querySelectorAll('.merchant-card');
    for (var i = 0; i < cards.length; i++) {
      var name = cards[i].getAttribute('data-name');
      var id = cards[i].getAttribute('data-id');
      if (name && s.indexOf(name) !== -1) return name;
      if (id && s.indexOf(id) !== -1) return name;
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

  // Translate raw agent summaries into human-readable descriptions
  function detectOp(ev) {
    var s = ev.summary || '';
    if (ev.type === 'agent_start') return 'Agent demarre';
    if (ev.type === 'agent_done') return ev.summary;
    if (ev.type === 'agent_thinking') return s;
    var m;
    if ((m = s.match(/^Searching for:\s*(.+)/i))) return 'Recherche: ' + m[1];
    if (s.match(/^Getting details for/i)) return 'Consultation prix';
    if (s.match(/^Creating checkout/i)) return 'Creation du panier';
    if (s.match(/^Asking .+ for promotions/i)) return 'Recherche de promotions';
    if ((m = s.match(/^Applying discount\s+(\S+)/i))) return 'Negociation: code ' + m[1];
    if (s.match(/^Updating checkout/i)) return 'Mise a jour commande';
    if (s.match(/^Getting checkout summary/i)) return 'Verification du panier';
    if (s.match(/^Completing checkout/i)) return 'Paiement en cours';
    if (s.match(/^Cancelling checkout/i)) return 'Annulation de commande';
    return s;
  }

  function isPolledMessage(s) {
    return s && (s.indexOf('Polled') !== -1 || s.indexOf('Poll failed') !== -1);
  }

  var stateLabels = {
    'state-lookup': { css: 'st-lookup', text: 'Consultation prix' },
    'state-checkout': { css: 'st-checkout', text: 'Creation panier' },
    'state-negotiate': { css: 'st-negotiate', text: 'Negociation' },
    'state-sale': { css: 'st-sale', text: 'Paiement en cours' },
    'state-cancel': { css: 'st-cancel', text: 'Annulation' }
  };
  var stateTimers = {};

  // Set a visual state class on a merchant card with a status label, auto-remove after duration ms
  function setMerchantState(name, cls, duration, label) {
    var cards = document.querySelectorAll('.merchant-card');
    for (var i = 0; i < cards.length; i++) {
      if (cards[i].getAttribute('data-name') !== name) continue;
      var card = cards[i];
      card.classList.remove('state-checkout', 'state-negotiate', 'state-sale', 'state-lookup');
      var statusEl = card.querySelector('.mc-status');
      if (statusEl) {
        statusEl.className = 'mc-status';
        statusEl.textContent = '';
      }
      if (stateTimers[name]) { clearTimeout(stateTimers[name]); delete stateTimers[name]; }
      if (cls) {
        card.classList.add(cls);
        var info = stateLabels[cls];
        if (statusEl && info) {
          statusEl.className = 'mc-status visible ' + info.css;
          statusEl.textContent = label || info.text;
        }
        stateTimers[name] = (function(c, n) {
          return setTimeout(function() {
            c.classList.remove('state-checkout', 'state-negotiate', 'state-sale', 'state-lookup');
            var se = c.querySelector('.mc-status');
            if (se) { se.className = 'mc-status'; se.textContent = ''; }
            delete stateTimers[n];
          }, duration);
        })(card, name);
      }
    }
  }

  // --- Audio feedback ---
  var audioCtx;
  function initAudio(){if(!audioCtx)try{audioCtx=new(window.AudioContext||window.webkitAudioContext)()}catch(e){}}
  document.addEventListener('click',initAudio,{once:true});
  function playKaChing(){
    if(!audioCtx)return;
    var now=audioCtx.currentTime;
    var g=audioCtx.createGain();g.gain.setValueAtTime(0.25,now);g.gain.exponentialRampToValueAtTime(0.01,now+0.5);g.connect(audioCtx.destination);
    var o1=audioCtx.createOscillator();o1.type='sine';o1.frequency.value=523.25;o1.connect(g);o1.start(now);o1.stop(now+0.15);
    var o2=audioCtx.createOscillator();o2.type='sine';o2.frequency.value=659.25;o2.connect(g);o2.start(now+0.12);o2.stop(now+0.35);
    var o3=audioCtx.createOscillator();o3.type='sine';o3.frequency.value=783.99;o3.connect(g);o3.start(now+0.25);o3.stop(now+0.5);
  }

  // --- SSE with auto-reconnection ---
  var connDot = document.getElementById('conn-dot');
  var sseRetryDelay = 1000;
  var es = null;
  function sseConnect() {
    if (es) return;
    es = new EventSource('/events');
    es.onopen = function() { connDot.className='conn-dot connected'; connDot.title='Connecte'; sseRetryDelay=1000; };
    es.onerror = function() {
      connDot.className='conn-dot disconnected'; connDot.title='Deconnecte';
      es.close(); es=null;
      setTimeout(sseConnect, sseRetryDelay);
      sseRetryDelay = Math.min(sseRetryDelay*2, 8000);
    };
    es.onmessage = handleSSEMessage;
  }
  document.addEventListener('visibilitychange', function() {
    if (document.hidden) { if(es){es.close();es=null;connDot.className='conn-dot disconnected'} }
    else { sseConnect(); fetchRankings(); }
  });
  function handleSSEMessage(msg) {
    try {
      var ev = JSON.parse(msg.data);
      var summary = ev.summary || '';

      // Skip polled messages everywhere
      if (isPolledMessage(summary)) return;

      var displayText = detectOp(ev);

      // Arena-specific events
      if (ev.source === 'arena') {
        if (ev.type === 'merchant_registered') {
          appendToPanel('arena-registration', summary);
          fetchRankings();
        } else if (ev.type === 'sale_completed') {
          appendToPanel('arena-sale', summary);
          showWinnerBanner('SOLD!', summary);
          playKaChing();
          fetchRankings();
        } else if (ev.type === 'config_update') {
          appendToPanel('arena-config', summary);
          fetchRankings();
        }
      }

      // Agent activity panel (only milestones, skip verbose polling)
      if (ev.type === 'agent_start') {
        panelBody.innerHTML = '';
        panelHeader.textContent = 'Activity Log';
        appendToPanel('thinking-entry', displayText, summary);
      }
      if (ev.type === 'agent_thinking' && summary) appendToPanel('thinking-entry', displayText, summary);
      if (ev.type === 'tool_call' && summary) {
        var pd = ev.data ? Object.assign({_type:'req'}, ev.data) : null;
        appendToPanel('tool-call-entry', displayText, summary, pd);
      }
      if (ev.type === 'tool_result' && summary) {
        var pdr = ev.data ? Object.assign({_type:'res'}, ev.data) : null;
        appendToPanel('result-entry', displayText, summary, pdr);
      }
      if (ev.type === 'tool_error' && summary) {
        var pde = ev.data ? Object.assign({_type:'err'}, ev.data) : null;
        appendToPanel('error-entry', displayText, summary, pde);
      }
      if (ev.type === 'agent_error' && summary) appendToPanel('error-entry', displayText, summary);
      if (ev.type === 'agent_done' && summary) {
        panelHeader.textContent = 'Agent Result';
        appendToPanel('result-entry', displayText, summary);
        fetchRankings();
        showAgentModal(summary, 5000);
      }

      // Bottom bar
      bottomDesc.textContent = displayText || ev.type || '';
      if (ev.type === 'tool_error' || ev.type === 'agent_error') {
        bottomBadge.textContent = 'Error';
        bottomBadge.style.display = '';
        bottomBadge.style.background = '#FEF2F2';
        bottomBadge.style.color = '#DC2626';
      } else if (ev.type) {
        bottomBadge.textContent = ev.type.replace(/_/g, ' ');
        bottomBadge.style.display = '';
        bottomBadge.style.background = '#FDE8E8';
        bottomBadge.style.color = '#E5004C';
      }

      // Highlight merchant + state colors + status label
      var mName = detectMerchantName(ev);
      if (mName) {
        highlightMerchant(mName);
        var dm;
        if (summary.match(/^Getting details for/i)) {
          setMerchantState(mName, 'state-lookup', 4000);
        } else if (summary.match(/^Creating checkout/i)) {
          setMerchantState(mName, 'state-checkout', 5000);
        } else if (summary.match(/^Asking .+ for promotions/i)) {
          setMerchantState(mName, 'state-negotiate', 5000, 'Recherche de promotions');
        } else if ((dm = summary.match(/^Applying discount\s+(\S+)/i))) {
          setMerchantState(mName, 'state-negotiate', 5000, 'Code promo: ' + dm[1]);
        } else if (summary.match(/^Updating checkout/i)) {
          setMerchantState(mName, 'state-negotiate', 5000, 'Mise a jour commande');
        } else if (summary.match(/^Getting checkout summary/i)) {
          setMerchantState(mName, 'state-checkout', 4000, 'Verification panier');
        } else if (summary.match(/^Completing checkout/i)) {
          setMerchantState(mName, 'state-sale', 8000);
        } else if (summary.match(/^Cancelling checkout/i)) {
          setMerchantState(mName, 'state-cancel', 5000);
        }

        // Merchant activity events (from merchant callbacks, source = merchant name)
        if (ev.type === 'product_details' || ev.type === 'catalog_browse') {
          setMerchantState(mName, 'state-lookup', 4000);
        } else if (ev.type === 'checkout_created' || ev.type === 'cart_created') {
          setMerchantState(mName, 'state-checkout', 5000);
        } else if (ev.type === 'checkout_updated') {
          var label = 'Mise a jour';
          if (summary.indexOf('promo') !== -1) label = 'Code promo';
          else if (summary.indexOf('livraison') !== -1) label = 'Livraison';
          else if (summary.indexOf('acheteur') !== -1) label = 'Info acheteur';
          setMerchantState(mName, 'state-negotiate', 5000, label);
        } else if (ev.type === 'checkout_canceled') {
          setMerchantState(mName, 'state-cancel', 5000);
        }
      }

      // Refresh merchants on sales or completions
      if (ev.type === 'tool_result' && summary && summary.indexOf('omplete') !== -1) {
        fetchRankings();
      }
    } catch(e) { console.error(e); }
  }
  sseConnect();

  // --- Agent panel (inline, no modal) ---
  var cmdInput = document.getElementById('command-input');
  var btnSend = document.getElementById('btn-send');
  var sendStatus = document.getElementById('send-status');
  var merchantCountInput = document.getElementById('merchant-count-input');
  var merchantCountValue = document.getElementById('merchant-count-value');

  merchantCountInput.addEventListener('input', function() { merchantCountValue.textContent = merchantCountInput.value; });

  // --- Toggle expanded agent panel ---
  var agentPanel = document.querySelector('.agent-panel');
  var agentDots = document.querySelector('.agent-card-dots');
  agentDots.addEventListener('click', function() {
    agentPanel.classList.toggle('expanded');
    if (agentPanel.classList.contains('expanded')) {
      cmdInput.focus();
    }
  });
  // Click on backdrop (outside card) to collapse
  agentPanel.addEventListener('click', function(e) {
    if (e.target === agentPanel && agentPanel.classList.contains('expanded')) {
      agentPanel.classList.remove('expanded');
    }
  });
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
      if (document.getElementById('agent-modal-overlay').classList.contains('visible')) {
        hideAgentModal();
      } else if (agentPanel.classList.contains('expanded')) {
        agentPanel.classList.remove('expanded');
      }
    }
  });

  function submitCommand() {
    var val = cmdInput.value.trim();
    if (!val) return;
    sendStatus.textContent = 'Sending...';
    sendStatus.className = 'agent-send-status sending';
    btnSend.disabled = true;
    fetch('/command', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({instruction: val, merchant_count: parseInt(merchantCountInput.value, 10)}) })
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.connected) {
          sendStatus.textContent = 'Instruction sent to agent';
          sendStatus.className = 'agent-send-status success';
          cmdInput.value = '';
        } else {
          sendStatus.textContent = 'Agent not connected';
          sendStatus.className = 'agent-send-status warning';
        }
        btnSend.disabled = false;
      })
      .catch(function() {
        sendStatus.textContent = 'Failed to reach agent';
        sendStatus.className = 'agent-send-status error';
        btnSend.disabled = false;
      });
  }

  btnSend.addEventListener('click', submitCommand);
  cmdInput.addEventListener('keydown', function(e) {
    if (e.key === 'Enter') submitCommand();
  });

  // --- Agent status polling ---
  var agentStatusEl = document.getElementById('agent-status');
  function pollAgentStatus() {
    fetch('/status')
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.agent_connected) {
          agentStatusEl.className = 'agent-status-line connected';
          agentStatusEl.innerHTML = '<span class="status-dot"></span> Connected';
          cmdInput.disabled = false;
          btnSend.disabled = false;
        } else {
          agentStatusEl.className = 'agent-status-line disconnected';
          agentStatusEl.innerHTML = '<span class="status-dot"></span> Disconnected';
          cmdInput.disabled = true;
          btnSend.disabled = true;
        }
      })
      .catch(function() {});
  }
  pollAgentStatus();
  setInterval(pollAgentStatus, 3000);

  // --- Agent Modal (full-screen markdown overlay) ---
  var modalOverlay = document.getElementById('agent-modal-overlay');
  var modalBody = document.getElementById('agent-modal-body');
  var modalProgress = document.getElementById('agent-modal-progress');
  var modalDismissTimer = null;
  var modalOpenedAt = 0;

  if (typeof marked !== 'undefined') {
    marked.setOptions({ breaks: true });
  }

  function showAgentModal(text, autoDismissMs, protoData) {
    autoDismissMs = autoDismissMs || 5000;
    if (typeof marked !== 'undefined') {
      modalBody.innerHTML = marked.parse(text);
    } else {
      modalBody.innerHTML = '<p>' + escapeHtml(text) + '</p>';
    }
    if (protoData) {
      var protoDiv = document.createElement('div');
      protoDiv.className = 'proto-payload visible';
      protoDiv.style.marginTop = '1rem';
      protoDiv.innerHTML = syntaxHighlight(protoData);
      modalBody.appendChild(protoDiv);
    }
    modalOverlay.classList.add('visible');
    modalOpenedAt = Date.now();
    if (modalDismissTimer) clearTimeout(modalDismissTimer);
    modalProgress.style.transition = 'none';
    modalProgress.style.width = '100%';
    modalProgress.offsetWidth; // force reflow
    modalProgress.style.transition = 'width ' + autoDismissMs + 'ms linear';
    modalProgress.style.width = '0%';
    modalDismissTimer = setTimeout(hideAgentModal, autoDismissMs);
  }

  function hideAgentModal() {
    modalOverlay.classList.remove('visible');
    if (modalDismissTimer) { clearTimeout(modalDismissTimer); modalDismissTimer = null; }
    modalProgress.style.transition = 'none';
    modalProgress.style.width = '0%';
  }

  document.getElementById('agent-modal-close').addEventListener('click', hideAgentModal);
  modalOverlay.addEventListener('click', function(e) {
    if (e.target === modalOverlay) hideAgentModal();
  });
})();
</script>
</body>
</html>`

func (h *Handler) handleArenaDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Write([]byte(arenaDashboardHTML))
}
