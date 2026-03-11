package obs

import "net/http"

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Multi-Agent Shopping Demo</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: system-ui, -apple-system, sans-serif; background: #0f172a; color: #e2e8f0; }
.header { background: #1e293b; padding: 1rem 2rem; border-bottom: 1px solid #334155; display: flex; align-items: center; gap: 1rem; }
.header h1 { font-size: 1.25rem; }
.header .status { margin-left: auto; display: flex; gap: 0.5rem; align-items: center; }
.dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; }
.dot.green { background: #22c55e; }
.dot.red { background: #ef4444; }
.dot.yellow { background: #eab308; }
.container { display: grid; grid-template-columns: 1fr 360px; height: calc(100vh - 56px); }
.timeline { overflow-y: auto; padding: 1rem; }
.sidebar { background: #1e293b; border-left: 1px solid #334155; overflow-y: auto; padding: 1rem; }
.sidebar h2 { font-size: 0.875rem; text-transform: uppercase; color: #94a3b8; margin-bottom: 0.75rem; letter-spacing: 0.05em; }
.event-card { background: #1e293b; border: 1px solid #334155; border-radius: 8px; padding: 0.75rem; margin-bottom: 0.5rem; border-left: 3px solid #6366f1; }
.event-card.source-client { border-left-color: #8b5cf6; }
.event-card.source-graph { border-left-color: #06b6d4; }
.event-card.source-merchant { border-left-color: #22c55e; }
.event-card.source-obs { border-left-color: #f59e0b; }
.event-card .meta { display: flex; justify-content: space-between; font-size: 0.75rem; color: #94a3b8; margin-bottom: 0.25rem; }
.event-card .summary { font-size: 0.875rem; }
.event-card .data { font-size: 0.75rem; color: #94a3b8; margin-top: 0.5rem; max-height: 100px; overflow-y: auto; white-space: pre-wrap; font-family: monospace; }
.stat-card { background: #0f172a; border: 1px solid #334155; border-radius: 8px; padding: 0.75rem; margin-bottom: 0.5rem; }
.stat-card .label { font-size: 0.75rem; color: #94a3b8; }
.stat-card .value { font-size: 1.5rem; font-weight: 600; }
.empty { text-align: center; color: #64748b; padding: 3rem; }
#count { font-variant-numeric: tabular-nums; }
</style>
</head>
<body>
<div class="header">
  <h1>Multi-Agent Shopping Demo</h1>
  <div class="status"><span class="dot green"></span> <span id="count">0</span> events</div>
</div>
<div class="container">
  <div class="timeline" id="timeline">
    <div class="empty">Waiting for events...</div>
  </div>
  <div class="sidebar">
    <h2>Stats</h2>
    <div id="stats">
      <div class="stat-card"><div class="label">Events</div><div class="value" id="stat-total">0</div></div>
      <div class="stat-card"><div class="label">Sources</div><div class="value" id="stat-sources">0</div></div>
    </div>
    <h2 style="margin-top:1rem">Recent</h2>
    <div id="recent"></div>
  </div>
</div>
<script>
const timeline = document.getElementById('timeline');
const countEl = document.getElementById('count');
const statTotal = document.getElementById('stat-total');
const statSources = document.getElementById('stat-sources');
const recentEl = document.getElementById('recent');
let count = 0;
const sources = new Set();
const recent = [];

function sourceClass(src) {
  if (src.includes('client')) return 'source-client';
  if (src.includes('graph')) return 'source-graph';
  if (src.includes('merchant') || src.includes('shop') || src.includes('mart') || src.includes('budget')) return 'source-merchant';
  return 'source-obs';
}

function addEvent(e) {
  if (count === 0) timeline.innerHTML = '';
  count++;
  sources.add(e.source);
  countEl.textContent = count;
  statTotal.textContent = count;
  statSources.textContent = sources.size;

  const card = document.createElement('div');
  card.className = 'event-card ' + sourceClass(e.source);
  const ts = new Date(e.timestamp).toLocaleTimeString();
  let dur = '';
  if (e.duration_ms) dur = ' (' + e.duration_ms + 'ms)';
  card.innerHTML = '<div class="meta"><span>' + e.source + '</span><span>' + ts + dur + '</span></div>' +
    '<div class="summary">' + escapeHtml(e.summary) + '</div>' +
    (e.data ? '<div class="data">' + escapeHtml(JSON.stringify(e.data, null, 2)) + '</div>' : '');
  timeline.prepend(card);

  recent.unshift(e);
  if (recent.length > 5) recent.pop();
  renderRecent();
}

function renderRecent() {
  recentEl.innerHTML = recent.map(e => {
    const ts = new Date(e.timestamp).toLocaleTimeString();
    return '<div class="stat-card"><div class="label">' + ts + ' - ' + e.source + '</div><div style="font-size:0.8rem">' + escapeHtml(e.summary) + '</div></div>';
  }).join('');
}

function escapeHtml(s) {
  const d = document.createElement('div');
  d.textContent = s;
  return d.innerHTML;
}

const es = new EventSource('/events');
es.onmessage = function(ev) {
  try { addEvent(JSON.parse(ev.data)); } catch(e) { console.error(e); }
};
</script>
</body>
</html>`

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}
