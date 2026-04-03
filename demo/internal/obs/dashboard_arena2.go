package obs

import "net/http"

const arena2DashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>UCP Arena 2</title>
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
.nav-link { color: #E5004C; text-decoration: none; font-weight: 700; font-size: 0.8rem; padding: 0.25rem 0.5rem; border: 1px solid #E5004C; border-radius: 8px; transition: background 0.15s, color 0.15s; }
.nav-link:hover { background: #E5004C; color: #fff; }

.winner-banner { display: none; background: #E5004C; padding: 1.2rem 1.5rem; text-align: center; flex-shrink: 0; animation: banner-in 0.4s ease-out; }
.winner-banner h2 { font-size: 1.5rem; color: #fff; margin-bottom: 0.25rem; font-weight: 800; }
.winner-banner p { font-size: 1rem; color: rgba(255,255,255,0.85); }
@keyframes banner-in { from { opacity: 0; transform: translateY(-20px); } to { opacity: 1; transform: translateY(0); } }

.main-area { flex: 1; display: flex; overflow: hidden; }

/* --- Activity panel (left) --- */
.activity-panel { width: 300px; background: #FFFFFF; border-right: 1px solid #E0E0E0; display: flex; flex-direction: column; flex-shrink: 0; }
.activity-panel .panel-header { font-weight: 700; font-size: 0.75rem; color: #E5004C; text-transform: uppercase; letter-spacing: 0.05em; padding: 0.6rem 0.75rem; border-bottom: 1px solid #E0E0E0; }
.activity-panel .panel-body { flex: 1; overflow-y: auto; padding: 0.5rem 0.75rem; font-size: 0.95rem; line-height: 1.5; white-space: pre-wrap; word-wrap: break-word; color: #666; }
.activity-panel .panel-body .thinking-entry, .activity-panel .panel-body .result-entry, .activity-panel .panel-body .error-entry, .activity-panel .panel-body .arena-registration, .activity-panel .panel-body .arena-sale, .activity-panel .panel-body .arena-config, .activity-panel .panel-body .tool-call-entry { cursor: pointer; transition: opacity 0.15s; }
.activity-panel .panel-body .thinking-entry:hover, .activity-panel .panel-body .result-entry:hover, .activity-panel .panel-body .error-entry:hover, .activity-panel .panel-body .arena-registration:hover, .activity-panel .panel-body .arena-sale:hover, .activity-panel .panel-body .arena-config:hover, .activity-panel .panel-body .tool-call-entry:hover { opacity: 0.75; }
.activity-panel .panel-body .thinking-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #F9FAFB; border-radius: 8px; color: #2D2D2D; }
.activity-panel .panel-body .result-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #DCFCE7; border: 1px solid #16A34A; border-radius: 8px; color: #16A34A; }
.activity-panel .panel-body .error-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #FEF2F2; border: 1px solid #DC2626; border-radius: 8px; color: #DC2626; }
.activity-panel .panel-body .arena-registration { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #FDE8E8; border: 1px solid #E5004C; border-radius: 8px; color: #E5004C; }
.activity-panel .panel-body .arena-sale { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #DCFCE7; border: 1px solid #16A34A; border-radius: 8px; color: #16A34A; }
.activity-panel .panel-body .arena-config { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #FFF7ED; border: 1px solid #F59E0B; border-radius: 8px; color: #D97706; }
.activity-panel .panel-body .tool-call-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #F3F4F6; border: 1px solid #9CA3AF; border-radius: 8px; color: #6B7280; }
.proto-toggle { display: inline-block; font-size: 0.75rem; margin-left: 4px; color: #999; cursor: pointer; vertical-align: middle; user-select: none; }
.proto-toggle:hover { color: #E5004C; }
.proto-payload { display: none; margin-top: 0.35rem; padding: 0.4rem 0.5rem; background: #1A1A2E; color: #A5F3FC; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 0.7rem; line-height: 1.4; white-space: pre-wrap; word-break: break-all; max-height: 200px; overflow-y: auto; }
.proto-payload.visible { display: block; }
.proto-payload .pk { color: #F9A8D4; }
.proto-payload .ps { color: #A5F3FC; }
.proto-payload .pn { color: #FDE68A; }
.proto-payload .pp { color: #D1D5DB; }
.proto-badge { display: inline-block; font-size: 0.7rem; font-weight: 700; padding: 0.1rem 0.35rem; border-radius: 4px; margin-left: 4px; vertical-align: middle; }
.proto-badge.req { background: #DBEAFE; color: #3B82F6; }
.proto-badge.res { background: #DCFCE7; color: #16A34A; }
.proto-badge.err { background: #FEF2F2; color: #DC2626; }
.proto-duration { font-size: 0.7rem; color: #999; margin-left: 4px; }

/* --- Center area (leaderboard + timeline) --- */
.center-area { flex: 1; display: flex; flex-direction: column; padding: 0.75rem; gap: 0.75rem; overflow: hidden; }
.panel { background: #FFFFFF; border: 1px solid #2D2D2D; border-radius: 16px; overflow: hidden; display: flex; flex-direction: column; }
.panel-hdr { padding: 0.5rem 0.75rem; border-bottom: 1px solid #E0E0E0; font-weight: 700; font-size: 0.75rem; color: #E5004C; text-transform: uppercase; letter-spacing: 0.05em; display: flex; align-items: center; gap: 0.5rem; flex-shrink: 0; }
.panel-hdr .ph-right { margin-left: auto; font-weight: 400; color: #999; text-transform: none; letter-spacing: 0; }
.panel-content { flex: 1; overflow: auto; position: relative; }

.leaderboard-panel { flex: 0 0 55%; }
.timeline-panel { flex: 1; }

/* --- Leaderboard --- */
.lb-table { width: 100%; border-collapse: collapse; font-size: 0.95rem; }
.lb-table th { text-align: left; padding: 0.5rem 0.6rem; color: #999; font-weight: 600; font-size: 0.85rem; text-transform: uppercase; border-bottom: 1px solid #E0E0E0; position: sticky; top: 0; background: #fff; z-index: 1; }
.lb-table td { padding: 0.5rem 0.6rem; border-bottom: 1px solid #F3F4F6; font-variant-numeric: tabular-nums; }
.lb-table tr:hover { background: #FDF0EE; }
.lb-rank { font-weight: 800; width: 2.5rem; text-align: center; font-size: 1.1rem; }
.lb-name { font-weight: 700; font-size: 0.9rem; }
.lb-positive { color: #16A34A; font-weight: 700; }
.lb-negative { color: #DC2626; font-weight: 700; }
.lb-spark { width: 80px; height: 24px; vertical-align: middle; }
.lb-no-data { text-align: center; padding: 2rem; color: #999; font-size: 0.9rem; }

/* --- Timeline --- */
.tl-scroll { overflow-x: auto; overflow-y: auto; height: 100%; padding: 0.5rem; }
.tl-lane { display: flex; align-items: center; margin-bottom: 2px; min-height: 28px; }
.tl-label { width: 110px; flex-shrink: 0; font-size: 0.85rem; font-weight: 700; color: #1A1A2E; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; padding-right: 0.5rem; text-align: right; }
.tl-track { flex: 1; position: relative; height: 24px; background: #F9FAFB; border-radius: 4px; overflow: visible; }
.tl-event { position: absolute; height: 22px; top: 1px; border-radius: 4px; font-size: 0.75rem; font-weight: 600; display: flex; align-items: center; justify-content: center; color: #fff; min-width: 18px; cursor: default; transition: opacity 0.15s; z-index: 1; white-space: nowrap; padding: 0 4px; }
.tl-event:hover { opacity: 0.8; z-index: 10; }
.tl-event.search { background: #6B7280; }
.tl-event.lookup { background: #E5004C; }
.tl-event.checkout { background: #3B82F6; }
.tl-event.promo { background: #F59E0B; }
.tl-event.discount { background: #F97316; }
.tl-event.update { background: #8B5CF6; }
.tl-event.summary { background: #0EA5E9; }
.tl-event.complete { background: #16A34A; }
.tl-event.cancel { background: #DC2626; }
.tl-event.thinking { background: #9CA3AF; }
.tl-legend { display: flex; gap: 0.75rem; flex-wrap: wrap; padding: 0.4rem 0.75rem; border-top: 1px solid #E0E0E0; font-size: 0.75rem; flex-shrink: 0; }
.tl-legend-item { display: flex; align-items: center; gap: 3px; }
.tl-legend-dot { width: 10px; height: 10px; border-radius: 3px; flex-shrink: 0; }
.tl-no-data { text-align: center; padding: 2rem; color: #999; font-size: 0.85rem; }

/* --- Agent panel (right) --- */
.agent-panel { width: 320px; background: #FDF0EE; border-left: 1px solid #E0E0E0; display: flex; flex-direction: column; flex-shrink: 0; padding: 1rem; gap: 0.75rem; }
.agent-card { background: #FFFFFF; border: 1px solid #2D2D2D; border-radius: 16px; box-shadow: 6px 6px 0px #1A1A2E; overflow: hidden; display: flex; flex-direction: column; }
.agent-card-dots { padding: 0.4rem 0.75rem; border-bottom: 1px solid #E0E0E0; display: flex; align-items: center; gap: 6px; cursor: pointer; }
.agent-card-dots:hover { background: #F3F4F6; }
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

/* --- Agent expanded overlay --- */
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
    <a href="/arena" class="nav-link">Monitor</a>
    <a href="/insights" class="nav-link">Insights</a>
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

  <div class="center-area">
    <div class="panel leaderboard-panel" style="box-shadow:4px 4px 0px #F59E0B">
      <div class="panel-hdr">Leaderboard <span class="ph-right" id="lb-count"></span></div>
      <div class="panel-content" id="lb-body">
        <div class="lb-no-data">No merchants registered</div>
      </div>
    </div>
    <div class="panel timeline-panel" style="box-shadow:4px 4px 0px #E5004C">
      <div class="panel-hdr">Negotiation Timeline <span class="ph-right" id="tl-elapsed"></span></div>
      <div class="panel-content">
        <div class="tl-scroll" id="tl-scroll">
          <div class="tl-no-data" id="tl-no-data">Waiting for buyer agent...</div>
        </div>
      </div>
      <div class="tl-legend">
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#6B7280"></div>Search</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#E5004C"></div>Details</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#3B82F6"></div>Checkout</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#F59E0B"></div>Promos</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#F97316"></div>Discount</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#8B5CF6"></div>Update</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#0EA5E9"></div>Verify</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#16A34A"></div>Pay</div>
        <div class="tl-legend-item"><div class="tl-legend-dot" style="background:#DC2626"></div>Cancel</div>
      </div>
    </div>
  </div>

  <div class="agent-panel">
    <div class="agent-card">
      <div class="agent-card-dots"><span class="agent-card-title">External Service</span></div>
      <div class="agent-card-body">
        <div class="agent-identity">
          <div class="agent-avatar">&#x1F6D2;</div>
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
      It discovers and negotiates with merchants on its own via the Shopping Graph.
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
  // === DOM refs ===
  var panelBody = document.getElementById('panel-body');
  var panelHeader = document.getElementById('panel-header');
  var bottomDesc = document.getElementById('bottom-desc');
  var bottomBadge = document.getElementById('bottom-badge');
  var merchantCountEl = document.getElementById('merchant-count');
  var winnerBanner = document.getElementById('winner-banner');
  var winnerName = document.getElementById('winner-name');
  var winnerDetail = document.getElementById('winner-detail');
  var productInfo = document.getElementById('product-info');
  var lbBody = document.getElementById('lb-body');
  var lbCountEl = document.getElementById('lb-count');
  var tlScroll = document.getElementById('tl-scroll');
  var tlNoData = document.getElementById('tl-no-data');
  var tlElapsed = document.getElementById('tl-elapsed');

  // === State ===
  var merchants = {};
  var currentRankings = {};
  var profitSnapshots = {};
  var agentStartTime = null;
  var tlLanes = {};
  var tlMaxTime = 0;
  var bannerTimer = null;
  var PX_PER_SEC = 60;

  // === Action config (timeline) ===
  var actionMeta = {
    'search_products':      {cls:'search',   icon:'S',  label:'Search'},
    'get_product_details':  {cls:'lookup',   icon:'D',  label:'Details'},
    'create_checkout':      {cls:'checkout', icon:'C',  label:'Checkout'},
    'list_promotions':      {cls:'promo',    icon:'P',  label:'Promos'},
    'apply_discount_codes': {cls:'discount', icon:'%',  label:'Discount'},
    'update_checkout':      {cls:'update',   icon:'U',  label:'Update'},
    'get_checkout_summary': {cls:'summary',  icon:'$',  label:'Verify'},
    'complete_checkout':    {cls:'complete', icon:'+',  label:'Pay'},
    'cancel_checkout':      {cls:'cancel',   icon:'X',  label:'Cancel'}
  };

  // === Helpers ===
  function escapeHtml(s) { var d=document.createElement('div'); d.textContent=s; return d.innerHTML; }
  function formatPrice(cents) { return '$'+(cents/100).toFixed(2); }
  function extractMid(url) {
    if(!url) return null;
    var parts=url.replace(/\/+$/,'').split('/');
    return parts[parts.length-1];
  }
  function getMerchantName(mid) {
    var m=merchants[mid];
    return m ? (m.emoji?m.emoji+' ':'')+m.name : mid.substring(0,8);
  }
  function getMerchantColor(mid) {
    var m=merchants[mid];
    return m&&m.accent_color ? m.accent_color : '#E5004C';
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

  function isPolledMessage(s) {
    return s && (s.indexOf('Polled') !== -1 || s.indexOf('Poll failed') !== -1);
  }

  function detectOp(ev) {
    var s = ev.summary || '';
    if (ev.type === 'agent_start') return 'Agent started';
    if (ev.type === 'agent_done') return ev.summary;
    if (ev.type === 'agent_thinking') return s;
    var m;
    if ((m = s.match(/^Searching for:\s*(.+)/i))) return 'Search: ' + m[1];
    if (s.match(/^Getting details for/i)) return 'Consulting price';
    if (s.match(/^Creating checkout/i)) return 'Creating checkout';
    if (s.match(/^Asking .+ for promotions/i)) return 'Looking for promotions';
    if ((m = s.match(/^Applying discount\s+(\S+)/i))) return 'Negotiating: code ' + m[1];
    if (s.match(/^Updating checkout/i)) return 'Updating order';
    if (s.match(/^Getting checkout summary/i)) return 'Verifying checkout';
    if (s.match(/^Completing checkout/i)) return 'Payment in progress';
    if (s.match(/^Cancelling checkout/i)) return 'Cancelling order';
    return s;
  }

  // === Fetch config ===
  fetch('/arena/config')
    .then(function(r) { return r.ok ? r.json() : Promise.reject(); })
    .then(function(cfg) {
      productInfo.textContent = cfg.product_name + ' | Cost: $' + (cfg.cost_price / 100).toFixed(2);
    })
    .catch(function() {});

  // === Data fetching ===
  function fetchMerchants() {
    fetch('/arena/merchants')
      .then(function(r) { return r.json(); })
      .then(function(d) {
        var list = d.merchants || [];
        merchants = {};
        for (var i=0; i<list.length; i++) {
          merchants[list[i].id] = list[i];
          if (!profitSnapshots[list[i].id]) profitSnapshots[list[i].id] = [];
          profitSnapshots[list[i].id].push(list[i].net_profit);
          if (profitSnapshots[list[i].id].length > 30) profitSnapshots[list[i].id].shift();
        }
        merchantCountEl.textContent = list.length + ' merchant' + (list.length !== 1 ? 's' : '');
        renderLeaderboard(list);
      })
      .catch(function() {});
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

  fetchRankings();
  setInterval(fetchRankings, 3000);

  // ============================================================
  // LEADERBOARD
  // ============================================================
  function renderLeaderboard(list) {
    if (!list || list.length === 0) {
      lbBody.innerHTML = '<div class="lb-no-data">No merchants registered</div>';
      lbCountEl.textContent = '';
      return;
    }
    lbCountEl.textContent = list.length + ' merchants';
    // Sort by shopping graph rank (lower = better), unranked at the end
    var sorted = list.slice().sort(function(a,b) {
      var rdA = currentRankings[a.id];
      var ra = (rdA && typeof rdA === 'object' && typeof rdA.rank === 'number') ? rdA.rank : 9999;
      var rdB = currentRankings[b.id];
      var rb = (rdB && typeof rdB === 'object' && typeof rdB.rank === 'number') ? rdB.rank : 9999;
      if (ra !== rb) return ra - rb;
      return b.net_profit - a.net_profit;
    });
    var html = '<table class="lb-table"><thead><tr>';
    html += '<th>#</th><th>Merchant</th><th>Price</th><th>Bid</th><th>Sales</th><th>Ad Spend</th><th>Profit</th><th></th>';
    html += '</tr></thead><tbody>';
    var medals = {1:'&#129351;', 2:'&#129352;', 3:'&#129353;'};
    for (var i=0; i<sorted.length; i++) {
      var m = sorted[i];
      var profitCls = m.net_profit >= 0 ? 'lb-positive' : 'lb-negative';
      var rdM = currentRankings[m.id];
      var rankNum = (rdM && typeof rdM === 'object' && typeof rdM.rank === 'number') ? rdM.rank : 0;
      var medal = rankNum > 0 ? (medals[rankNum] || rankNum) : '-';
      var emoji = m.emoji ? m.emoji+' ' : '';
      html += '<tr>';
      html += '<td class="lb-rank">' + medal + '</td>';
      html += '<td class="lb-name" style="color:' + (m.accent_color||'#1A1A2E') + '">' + escapeHtml(emoji+m.name) + '</td>';
      html += '<td>' + formatPrice(m.price) + '</td>';
      html += '<td>' + formatPrice(m.max_cpc_bid) + '</td>';
      html += '<td>' + m.sales_count + '</td>';
      html += '<td>' + formatPrice(m.total_ad_spend) + '</td>';
      html += '<td class="' + profitCls + '">' + formatPrice(m.net_profit) + '</td>';
      html += '<td><canvas class="lb-spark" data-mid="' + m.id + '"></canvas></td>';
      html += '</tr>';
    }
    html += '</tbody></table>';
    lbBody.innerHTML = html;

    var sparks = lbBody.querySelectorAll('.lb-spark');
    for (var i=0; i<sparks.length; i++) {
      drawSparkline(sparks[i], sparks[i].getAttribute('data-mid'));
    }
  }

  function drawSparkline(canvas, mid) {
    var data = profitSnapshots[mid];
    if (!data || data.length < 2) return;
    var dpr = window.devicePixelRatio || 1;
    var w = canvas.offsetWidth || 80;
    var h = canvas.offsetHeight || 24;
    canvas.width = w * dpr;
    canvas.height = h * dpr;
    var c = canvas.getContext('2d');
    c.setTransform(dpr, 0, 0, dpr, 0, 0);
    var min = Infinity, max = -Infinity;
    for (var i=0; i<data.length; i++) { if(data[i]<min) min=data[i]; if(data[i]>max) max=data[i]; }
    var range = max - min || 1;
    c.beginPath();
    for (var i=0; i<data.length; i++) {
      var x = i / (data.length-1) * w;
      var y = h - ((data[i]-min)/range) * (h-2) - 1;
      if (i===0) c.moveTo(x,y); else c.lineTo(x,y);
    }
    var last = data[data.length-1];
    c.strokeStyle = last >= 0 ? '#16A34A' : '#DC2626';
    c.lineWidth = 1.5;
    c.stroke();
  }

  // ============================================================
  // TIMELINE
  // ============================================================
  function resetTimeline() {
    tlLanes = {};
    tlMaxTime = 0;
    tlScroll.innerHTML = '';
    tlNoData = document.createElement('div');
    tlNoData.className = 'tl-no-data';
    tlNoData.textContent = 'Agent started...';
    tlScroll.appendChild(tlNoData);
  }

  function ensureLane(mid) {
    if (tlLanes[mid]) return tlLanes[mid];
    if (tlNoData && tlNoData.parentNode) tlNoData.parentNode.removeChild(tlNoData);
    tlNoData = null;
    var lane = document.createElement('div');
    lane.className = 'tl-lane';
    var label = document.createElement('div');
    label.className = 'tl-label';
    label.textContent = getMerchantName(mid);
    label.style.color = getMerchantColor(mid);
    var track = document.createElement('div');
    track.className = 'tl-track';
    lane.appendChild(label);
    lane.appendChild(track);
    tlScroll.appendChild(lane);
    tlLanes[mid] = {lane:lane, track:track, label:label};
    return tlLanes[mid];
  }

  // Override for __graph__ lane
  var origEnsureLane = ensureLane;
  ensureLane = function(mid) {
    var result = origEnsureLane(mid);
    if (mid === '__graph__') {
      result.label.textContent = 'Shopping Graph';
      result.label.style.color = '#1A1A2E';
    }
    return result;
  };

  function addTimelineEvent(mid, action, elapsed, durationMs) {
    var meta = actionMeta[action] || {cls:'thinking', icon:'?', label:action};
    var l = ensureLane(mid);
    var left = elapsed * PX_PER_SEC;
    var width = Math.max(18, (durationMs||200)/1000 * PX_PER_SEC);
    var block = document.createElement('div');
    block.className = 'tl-event ' + meta.cls;
    block.style.left = left + 'px';
    block.style.width = width + 'px';
    block.title = meta.label + (durationMs ? ' (' + durationMs + 'ms)' : '');
    block.textContent = meta.icon;
    l.track.appendChild(block);
    var endPx = left + width;
    if (endPx > tlMaxTime) {
      tlMaxTime = endPx;
      var tracks = tlScroll.querySelectorAll('.tl-track');
      for (var i=0; i<tracks.length; i++) {
        tracks[i].style.minWidth = (tlMaxTime+40) + 'px';
      }
    }
    tlScroll.scrollLeft = tlScroll.scrollWidth;
    return block;
  }

  function updateTimelineBlock(mid, action, durationMs) {
    if (!tlLanes[mid]) return;
    var track = tlLanes[mid].track;
    var blocks = track.querySelectorAll('.tl-event');
    var meta = actionMeta[action];
    if (!meta) return;
    for (var i=blocks.length-1; i>=0; i--) {
      if (blocks[i].classList.contains(meta.cls)) {
        var width = Math.max(18, durationMs/1000 * PX_PER_SEC);
        blocks[i].style.width = width + 'px';
        blocks[i].title = meta.label + ' (' + durationMs + 'ms)';
        break;
      }
    }
  }

  // ============================================================
  // ACTIVITY PANEL
  // ============================================================
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

  // ============================================================
  // AUDIO
  // ============================================================
  var audioCtx;
  function initAudio() { if(!audioCtx) try { audioCtx=new(window.AudioContext||window.webkitAudioContext)(); } catch(e) {} }
  document.addEventListener('click', initAudio, {once:true});
  function playKaChing() {
    if(!audioCtx) return;
    var now=audioCtx.currentTime;
    var g=audioCtx.createGain(); g.gain.setValueAtTime(0.25,now); g.gain.exponentialRampToValueAtTime(0.01,now+0.5); g.connect(audioCtx.destination);
    var o1=audioCtx.createOscillator(); o1.type='sine'; o1.frequency.value=523.25; o1.connect(g); o1.start(now); o1.stop(now+0.15);
    var o2=audioCtx.createOscillator(); o2.type='sine'; o2.frequency.value=659.25; o2.connect(g); o2.start(now+0.12); o2.stop(now+0.35);
    var o3=audioCtx.createOscillator(); o3.type='sine'; o3.frequency.value=783.99; o3.connect(g); o3.start(now+0.25); o3.stop(now+0.5);
  }

  // ============================================================
  // SSE WITH AUTO-RECONNECTION
  // ============================================================
  var connDot = document.getElementById('conn-dot');
  var sseRetryDelay = 1000;
  var es = null;
  function sseConnect() {
    if (es) return;
    es = new EventSource('/events');
    es.onopen = function() { connDot.className='conn-dot connected'; connDot.title='Connected'; sseRetryDelay=1000; };
    es.onerror = function() {
      connDot.className='conn-dot disconnected'; connDot.title='Disconnected';
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
      var data = ev.data || {};
      var summary = ev.summary || '';

      if (isPolledMessage(summary)) return;

      var displayText = detectOp(ev);

      // --- Arena-specific events (activity panel) ---
      if (ev.source === 'arena') {
        if (ev.type === 'merchant_registered') {
          appendToPanel('arena-registration', summary);
          fetchRankings();
        } else if (ev.type === 'sale_completed') {
          appendToPanel('arena-sale', summary);
          showWinnerBanner('SOLD!', summary);
          playKaChing();
          fetchRankings();
        } else if (ev.type === 'config_update' || ev.type === 'merchant_left') {
          appendToPanel('arena-config', summary);
          fetchRankings();
        }
      }

      // --- Agent lifecycle ---
      if (ev.type === 'agent_start') {
        agentStartTime = new Date(ev.timestamp);
        panelBody.innerHTML = '';
        panelHeader.textContent = 'Activity Log';
        appendToPanel('thinking-entry', displayText, summary);
        resetTimeline();
      }

      if (ev.type === 'agent_thinking' && summary) {
        appendToPanel('thinking-entry', displayText, summary);
      }

      if (ev.type === 'tool_call' && summary) {
        var pd = ev.data ? Object.assign({_type:'req'}, ev.data) : null;
        appendToPanel('tool-call-entry', displayText, summary, pd);

        // Timeline: add event block
        if (data.action) {
          var mid = extractMid(data.merchant_url);
          var elapsed = agentStartTime ? (new Date(ev.timestamp) - agentStartTime)/1000 : 0;
          if (data.action === 'search_products') {
            addTimelineEvent('__graph__', 'search_products', elapsed, 0);
          } else if (mid) {
            addTimelineEvent(mid, data.action, elapsed, 0);
          }
          if (agentStartTime) {
            tlElapsed.textContent = elapsed.toFixed(1) + 's';
          }
        }
      }

      if (ev.type === 'tool_result' && summary) {
        var pdr = ev.data ? Object.assign({_type:'res'}, ev.data) : null;
        appendToPanel('result-entry', displayText, summary, pdr);

        // Timeline: update block duration
        if (data.action) {
          var mid = data.params ? extractMid(data.params.merchant_url) : null;
          if (mid && data.duration_ms) {
            updateTimelineBlock(mid, data.action, data.duration_ms);
          }
          if (data.action === 'search_products' && data.duration_ms) {
            updateTimelineBlock('__graph__', 'search_products', data.duration_ms);
          }
        }
      }

      if (ev.type === 'tool_error' && summary) {
        var pde = ev.data ? Object.assign({_type:'err'}, ev.data) : null;
        appendToPanel('error-entry', displayText, summary, pde);
        if (data.action) {
          var mid = data.params ? extractMid(data.params.merchant_url) : null;
          if (mid) {
            var elapsed = agentStartTime ? (new Date(ev.timestamp) - agentStartTime)/1000 : 0;
            addTimelineEvent(mid, data.action, elapsed, 100);
          }
        }
      }

      if (ev.type === 'agent_error' && summary) appendToPanel('error-entry', displayText, summary);

      if (ev.type === 'agent_done' && summary) {
        panelHeader.textContent = 'Agent Result';
        appendToPanel('result-entry', displayText, summary);
        tlElapsed.textContent = '';
        fetchRankings();
        showAgentModal(summary, 5000);
      }

      // --- Bottom bar ---
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

      // Refresh on completions
      if (ev.type === 'tool_result' && summary && summary.indexOf('omplete') !== -1) {
        fetchRankings();
      }

    } catch(ex) { console.error('SSE handler error:', ex); }
  }
  sseConnect();

  // ============================================================
  // AGENT PANEL
  // ============================================================
  var cmdInput = document.getElementById('command-input');
  var btnSend = document.getElementById('btn-send');
  var sendStatus = document.getElementById('send-status');
  var merchantCountInput = document.getElementById('merchant-count-input');
  var merchantCountValue = document.getElementById('merchant-count-value');

  merchantCountInput.addEventListener('input', function() { merchantCountValue.textContent = merchantCountInput.value; });

  var agentPanel = document.querySelector('.agent-panel');
  var agentDots = document.querySelector('.agent-card-dots');
  agentDots.addEventListener('click', function() {
    agentPanel.classList.toggle('expanded');
    if (agentPanel.classList.contains('expanded')) cmdInput.focus();
  });
  agentPanel.addEventListener('click', function(e) {
    if (e.target === agentPanel && agentPanel.classList.contains('expanded')) agentPanel.classList.remove('expanded');
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
      .then(function(d) {
        if (d.connected) {
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
  cmdInput.addEventListener('keydown', function(e) { if (e.key === 'Enter') submitCommand(); });

  // Agent status polling
  var agentStatusEl = document.getElementById('agent-status');
  function pollAgentStatus() {
    fetch('/status')
      .then(function(r) { return r.json(); })
      .then(function(d) {
        if (d.agent_connected) {
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

  // ============================================================
  // AGENT MODAL
  // ============================================================
  var modalOverlay = document.getElementById('agent-modal-overlay');
  var modalBody = document.getElementById('agent-modal-body');
  var modalProgress = document.getElementById('agent-modal-progress');
  var modalDismissTimer = null;
  var modalOpenedAt = 0;

  if (typeof marked !== 'undefined') { marked.setOptions({ breaks: true }); }

  function showAgentModal(text, autoDismissMs, protoData) {
    autoDismissMs = autoDismissMs || 12000;
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
    modalProgress.offsetWidth;
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

  // Pause modal auto-dismiss on hover
  var modalRemaining = 0;
  modalOverlay.querySelector('.agent-modal').addEventListener('mouseenter', function() {
    if (modalDismissTimer) {
      modalRemaining = Math.max(0, (modalOpenedAt + 12000) - Date.now());
      clearTimeout(modalDismissTimer);
      modalDismissTimer = null;
      modalProgress.style.transition = 'none';
    }
  });
  modalOverlay.querySelector('.agent-modal').addEventListener('mouseleave', function() {
    if (modalRemaining > 0 && modalOverlay.classList.contains('visible')) {
      modalProgress.style.transition = 'width ' + modalRemaining + 'ms linear';
      modalProgress.style.width = '0%';
      modalDismissTimer = setTimeout(hideAgentModal, modalRemaining);
    }
  });

  document.getElementById('agent-modal-close').addEventListener('click', hideAgentModal);
  modalOverlay.addEventListener('click', function(e) {
    if (e.target === modalOverlay) hideAgentModal();
  });
})();
</script>
</body>
</html>`

func (h *Handler) handleArena2Dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Write([]byte(arena2DashboardHTML))
}
