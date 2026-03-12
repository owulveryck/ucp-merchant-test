package obs

import "net/http"

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>UCP Shopping Demo</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: system-ui, -apple-system, sans-serif; background: #0E2356; color: #FFFFFF; overflow: hidden; height: 100vh; display: flex; flex-direction: column; }
.topbar { background: #263967; padding: 0.6rem 1.5rem; display: flex; align-items: center; gap: 1rem; border-bottom: 1px solid #3E4F78; flex-shrink: 0; }
.topbar h1 { font-size: 1.1rem; font-weight: 600; letter-spacing: 0.02em; }
.topbar .controls { margin-left: auto; display: flex; align-items: center; gap: 0.5rem; }
.topbar .controls button { background: #00D2DD; color: #0E2356; border: none; border-radius: 6px; padding: 0.35rem 0.75rem; font-size: 0.8rem; font-weight: 600; cursor: pointer; transition: opacity 0.2s; }
.topbar .controls button:hover { opacity: 0.85; }
.topbar .controls button.active { background: #FFFFFF; }
.topbar .step-info { font-size: 0.8rem; color: #B7BDCC; font-variant-numeric: tabular-nums; }
.topbar .live-dot { width: 8px; height: 8px; border-radius: 50%; background: #00D2DD; display: inline-block; margin-right: 4px; }
.topbar .live-dot.recording { animation: pulse-dot 1.5s ease-in-out infinite; }
@keyframes pulse-dot { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
.main-area { flex: 1; display: flex; overflow: hidden; }
.canvas-wrap { flex: 1; position: relative; overflow: hidden; }
.canvas-wrap svg { width: 100%; height: 100%; }
.activity-panel { width: 280px; background: #1B2F5E; border-right: 1px solid #3E4F78; display: flex; flex-direction: column; flex-shrink: 0; }
.activity-panel .panel-header { font-weight: 600; font-size: 0.75rem; color: #00D2DD; text-transform: uppercase; letter-spacing: 0.05em; padding: 0.6rem 0.75rem; border-bottom: 1px solid #3E4F78; }
.activity-panel .panel-body { flex: 1; overflow-y: auto; padding: 0.5rem 0.75rem; font-size: 0.8rem; line-height: 1.5; white-space: pre-wrap; word-wrap: break-word; color: #B7BDCC; }
.activity-panel .panel-body .thinking-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #263967; border-radius: 6px; color: #DDE0E8; }
.activity-panel .panel-body .result-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #1A3A2A; border: 1px solid #2ECC71; border-radius: 6px; color: #FFFFFF; }
.activity-panel .panel-body .error-entry { margin-bottom: 0.5rem; padding: 0.4rem 0.5rem; background: #3A1A1A; border: 1px solid #FF6B6B; border-radius: 6px; color: #FF6B6B; }
.bottombar { background: #263967; padding: 0.5rem 1.5rem; border-top: 1px solid #3E4F78; display: flex; align-items: center; gap: 1rem; flex-shrink: 0; min-height: 48px; }
.bottombar .desc { flex: 1; font-size: 0.85rem; color: #B7BDCC; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.bottombar .badge { background: #00D2DD; color: #0E2356; border-radius: 4px; padding: 0.2rem 0.6rem; font-size: 0.75rem; font-weight: 600; white-space: nowrap; flex-shrink: 0; }
.node-rect { fill: #3E4F78; stroke: #4A5A85; stroke-width: 1.5; rx: 12; ry: 12; transition: stroke 0.3s, stroke-width 0.3s, filter 0.3s; }
.node-rect.active { stroke: #00D2DD; stroke-width: 3; filter: url(#glow); }
.node-label { fill: #FFFFFF; font-size: 15px; font-weight: 600; text-anchor: middle; dominant-baseline: central; pointer-events: none; }
.node-sublabel { fill: #B7BDCC; font-size: 11px; text-anchor: middle; dominant-baseline: central; pointer-events: none; }
.conn-path { fill: none; stroke: #2A3D6B; stroke-width: 1.5; }
@keyframes dash-travel { to { stroke-dashoffset: 0; } }
@keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }
#node-client { cursor: pointer; }
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
  <h1>UCP Shopping Demo</h1>
  <div class="controls">
    <button id="btn-replay" title="Replay">&#x27F2; Replay</button>
    <button id="btn-step" title="Step forward">&#x25B6; Step</button>
    <button id="btn-play" title="Play / Pause">&#x23EF; Play</button>
    <span class="step-info" id="step-info">0 / 0</span>
    <span><span class="live-dot recording" id="live-dot"></span><span id="mode-label">LIVE</span></span>
  </div>
</div>

<div class="main-area">
<div class="activity-panel">
  <div class="panel-header" id="panel-header">Activity Log</div>
  <div class="panel-body" id="panel-body"></div>
</div>
<div class="canvas-wrap">
<svg id="svg-canvas" viewBox="0 0 1200 600" preserveAspectRatio="xMidYMid meet">
  <defs>
    <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto" markerUnits="strokeWidth">
      <polygon points="0 0, 10 3.5, 0 7" fill="#00D2DD" />
    </marker>
    <filter id="glow" x="-30%" y="-30%" width="160%" height="160%">
      <feGaussianBlur stdDeviation="4" result="blur" />
      <feMerge><feMergeNode in="blur" /><feMergeNode in="SourceGraphic" /></feMerge>
    </filter>
  </defs>

  <!-- connection paths (behind nodes) -->
  <path id="path-graph" class="conn-path" d="M240,300 C400,300 400,75 460,75" />
  <path id="path-super" class="conn-path" d="M240,300 C500,300 700,120 860,120" />
  <path id="path-mega"  class="conn-path" d="M240,300 C500,300 700,300 860,300" />
  <path id="path-budget" class="conn-path" d="M240,300 C500,300 700,480 860,480" />
  <path id="path-graph-super"  class="conn-path" d="M640,75 C750,75 800,120 860,120" style="stroke-dasharray: 6 4; stroke: #2A3D6B;" />
  <path id="path-graph-mega"   class="conn-path" d="M640,75 C750,75 800,300 860,300" style="stroke-dasharray: 6 4; stroke: #2A3D6B;" />
  <path id="path-graph-budget" class="conn-path" d="M640,75 C750,175 800,480 860,480" style="stroke-dasharray: 6 4; stroke: #2A3D6B;" />

  <!-- animation layer -->
  <g id="anim-layer"></g>

  <!-- nodes -->
  <g id="node-client" transform="translate(60,265)">
    <rect class="node-rect" width="180" height="70" />
    <text class="node-label" x="90" y="30">Client Agent</text>
    <text class="node-sublabel" x="90" y="50">Gemini</text>
  </g>
  <g id="node-graph" transform="translate(460,40)">
    <rect class="node-rect" width="180" height="70" />
    <text class="node-label" x="90" y="30">Shopping Graph</text>
    <text class="node-sublabel" x="90" y="50">Product Index</text>
  </g>
  <g id="node-super" transform="translate(860,85)">
    <rect class="node-rect" width="180" height="70" />
    <text class="node-label" x="90" y="30">SuperShop</text>
    <text class="node-sublabel" x="90" y="50">:8182</text>
  </g>
  <g id="node-mega" transform="translate(860,265)">
    <rect class="node-rect" width="180" height="70" />
    <text class="node-label" x="90" y="30">MegaMart</text>
    <text class="node-sublabel" x="90" y="50">:8183</text>
  </g>
  <g id="node-budget" transform="translate(860,445)">
    <rect class="node-rect" width="180" height="70" />
    <text class="node-label" x="90" y="30">BudgetBuy</text>
    <text class="node-sublabel" x="90" y="50">:8184</text>
  </g>
</svg>
</div>
</div>

<div class="bottombar">
  <div class="desc" id="bottom-desc">Waiting for events...</div>
  <div class="badge" id="bottom-badge" style="display:none"></div>
</div>

<div class="modal-overlay" id="command-modal">
  <div class="modal">
    <h2>Send instruction to Client Agent</h2>
    <input type="text" id="command-input" placeholder="e.g. find roses" autocomplete="off" />
    <div class="modal-buttons">
      <button class="btn-cancel" id="modal-cancel">Cancel</button>
      <button class="btn-search" id="modal-submit">Search</button>
    </div>
  </div>
</div>

<script>
(function() {
  var NODES = {
    client: 'node-client',
    graph:  'node-graph',
    super:  'node-super',
    mega:   'node-mega',
    budget: 'node-budget'
  };
  var PATHS = {
    graph:  'path-graph',
    super:  'path-super',
    mega:   'path-mega',
    budget: 'path-budget',
    'graph-super':  'path-graph-super',
    'graph-mega':   'path-graph-mega',
    'graph-budget': 'path-graph-budget'
  };

  var state = {
    events: [],
    currentIndex: -1,
    mode: 'live',
    playing: false,
    speed: 1500,
    timerId: null,
    clearTimerId: null
  };

  var animLayer = document.getElementById('anim-layer');
  var stepInfo = document.getElementById('step-info');
  var modeLabel = document.getElementById('mode-label');
  var liveDot = document.getElementById('live-dot');
  var bottomDesc = document.getElementById('bottom-desc');
  var bottomBadge = document.getElementById('bottom-badge');
  var panelHeader = document.getElementById('panel-header');
  var panelBody = document.getElementById('panel-body');

  function detectTarget(ev) {
    var s = ev.summary || '';
    if (s.indexOf('Searching for:') !== -1) return 'graph';
    if (s.indexOf('localhost:8182') !== -1) return 'super';
    if (s.indexOf('localhost:8183') !== -1) return 'mega';
    if (s.indexOf('localhost:8184') !== -1) return 'budget';
    return null;
  }

  function detectOp(ev) {
    var s = ev.summary || '';
    var t = ev.type || '';
    if (t === 'agent_start') return 'Agent Start';
    if (t === 'agent_done') return 'Agent Done';
    if (t === 'agent_thinking') return 'Thinking';
    if (t === 'tool_result') return 'Tool Result';
    if (t === 'tool_error') return 'Tool Error';
    if (t === 'agent_error') return 'Error';
    if (t === 'agent_step') { var ms = s.match(/Step (\d+)/); return ms ? 'Step ' + ms[1] : 'Step'; }
    if (s.indexOf('Polled') !== -1 || s.indexOf('Poll failed') !== -1) {
      var m = s.match(/(\d+) products/);
      return m ? 'Poll: ' + m[1] + ' products' : 'Catalog Poll';
    }
    if (s.indexOf('Searching for:') !== -1) return 'Catalog Search';
    if (s.indexOf('Getting details') !== -1) return 'Product Details';
    if (s.indexOf('Creating checkout') !== -1) return 'Create Checkout';
    if (s.indexOf('Applying discount') !== -1) {
      var m2 = s.match(/discount (.+?) to/);
      return m2 ? 'Discount: ' + m2[1] : 'Apply Discounts';
    }
    if (s.indexOf('Updating checkout') !== -1) {
      var m3 = s.match(/: (.+?) at /);
      return m3 ? 'Update: ' + m3[1] : 'Update Checkout';
    }
    if (s.indexOf('Getting checkout') !== -1) return 'Get Checkout';
    if (s.indexOf('Completing checkout') !== -1) return 'Complete + Pay';
    if (s.indexOf('Cancelling checkout') !== -1) return 'Cancel Checkout';
    return 'Tool Call';
  }

  function clearActive() {
    var rects = document.querySelectorAll('.node-rect');
    for (var i = 0; i < rects.length; i++) {
      rects[i].classList.remove('active');
    }
  }

  function setActive(nodeKey) {
    var el = document.getElementById(NODES[nodeKey]);
    if (el) el.querySelector('.node-rect').classList.add('active');
  }

  function clearAnim() {
    while (animLayer.firstChild) animLayer.removeChild(animLayer.firstChild);
  }

  function animateArrow(pathId, label) {
    clearAnim();
    var refPath = document.getElementById(pathId);
    if (!refPath) return;
    var totalLen = refPath.getTotalLength();

    // animated arrow path
    var ns = 'http://www.w3.org/2000/svg';
    var arrow = document.createElementNS(ns, 'path');
    arrow.setAttribute('d', refPath.getAttribute('d'));
    arrow.setAttribute('fill', 'none');
    arrow.setAttribute('stroke', '#00D2DD');
    arrow.setAttribute('stroke-width', '2.5');
    arrow.setAttribute('marker-end', 'url(#arrowhead)');
    arrow.setAttribute('stroke-dasharray', String(totalLen));
    arrow.setAttribute('stroke-dashoffset', String(totalLen));
    arrow.style.animation = 'dash-travel 700ms ease-out forwards';
    animLayer.appendChild(arrow);

    // traveling dot
    var dot = document.createElementNS(ns, 'circle');
    dot.setAttribute('r', '4');
    dot.setAttribute('fill', '#00D2DD');
    var motionAnim = document.createElementNS(ns, 'animateMotion');
    motionAnim.setAttribute('dur', '700ms');
    motionAnim.setAttribute('fill', 'freeze');
    var mpath = document.createElementNS(ns, 'mpath');
    mpath.setAttributeNS('http://www.w3.org/1999/xlink', 'xlink:href', '#' + pathId);
    motionAnim.appendChild(mpath);
    dot.appendChild(motionAnim);
    animLayer.appendChild(dot);

    // label pill at midpoint
    if (label) {
      var mid = refPath.getPointAtLength(totalLen / 2);
      var g = document.createElementNS(ns, 'g');
      g.style.animation = 'fade-in 300ms ease-out 350ms both';

      var textEl = document.createElementNS(ns, 'text');
      textEl.setAttribute('x', String(mid.x));
      textEl.setAttribute('y', String(mid.y - 14));
      textEl.setAttribute('text-anchor', 'middle');
      textEl.setAttribute('fill', '#0E2356');
      textEl.setAttribute('font-size', '11');
      textEl.setAttribute('font-weight', '600');
      textEl.textContent = label;
      // measure text for pill background - approximate width
      var approxW = label.length * 7 + 16;

      var pill = document.createElementNS(ns, 'rect');
      pill.setAttribute('x', String(mid.x - approxW / 2));
      pill.setAttribute('y', String(mid.y - 26));
      pill.setAttribute('width', String(approxW));
      pill.setAttribute('height', '20');
      pill.setAttribute('rx', '10');
      pill.setAttribute('fill', '#00D2DD');

      g.appendChild(pill);
      g.appendChild(textEl);
      animLayer.appendChild(g);
    }
  }

  function escapeHtml(s) {
    var d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }

  var META_TYPES = { agent_step: 1, agent_thinking: 1, tool_result: 1, tool_error: 1, agent_error: 1 };

  function appendToPanel(className, text) {
    var div = document.createElement('div');
    div.className = className;
    div.textContent = text;
    panelBody.appendChild(div);
    panelBody.scrollTop = panelBody.scrollHeight;
  }

  function showEvent(idx) {
    if (idx < 0 || idx >= state.events.length) return;
    state.currentIndex = idx;
    var ev = state.events[idx];
    var op = detectOp(ev);
    var isMeta = META_TYPES[ev.type] === 1;

    // --- Activity panel updates (always) ---
    if (ev.type === 'agent_start') {
      panelBody.innerHTML = '';
      panelHeader.textContent = 'Activity Log';
    }
    if (ev.type === 'agent_thinking' && ev.summary) {
      appendToPanel('thinking-entry', ev.summary);
    }
    if (ev.type === 'tool_error' && ev.summary) {
      appendToPanel('error-entry', ev.summary);
    }
    if (ev.type === 'agent_error' && ev.summary) {
      appendToPanel('error-entry', ev.summary);
    }
    if (ev.type === 'agent_done' && ev.summary) {
      panelHeader.textContent = 'Agent Result';
      appendToPanel('result-entry', ev.summary);
    }

    // --- Bottom bar (always) ---
    bottomDesc.textContent = ev.summary || ev.type || '';
    if (op) {
      bottomBadge.textContent = op;
      bottomBadge.style.display = '';
      if (ev.type === 'tool_error' || ev.type === 'agent_error') {
        bottomBadge.style.background = '#FF6B6B';
        bottomBadge.style.color = '#FFFFFF';
      } else {
        bottomBadge.style.background = '#00D2DD';
        bottomBadge.style.color = '#0E2356';
      }
    } else {
      bottomBadge.style.display = 'none';
    }

    updateStepInfo();

    // --- Node/animation updates (skip for meta events) ---
    if (isMeta) return;

    var target = detectTarget(ev);

    clearActive();

    var isGraphSource = ev.source === 'shopping-graph';
    if (isGraphSource && target && (target === 'super' || target === 'mega' || target === 'budget')) {
      setActive('graph');
      setActive(target);
      var graphPathId = PATHS['graph-' + target];
      if (graphPathId) animateArrow(graphPathId, op);
    } else if (target) {
      setActive('client');
      setActive(target);
      var pathId = PATHS[target];
      if (pathId) animateArrow(pathId, op);
    } else {
      setActive('client');
      clearAnim();
      if (ev.type === 'agent_done') {
        var keys = ['graph', 'super', 'mega', 'budget'];
        for (var i = 0; i < keys.length; i++) setActive(keys[i]);
      }
    }

    if (state.clearTimerId) clearTimeout(state.clearTimerId);
    if (ev.type !== 'agent_done') {
      state.clearTimerId = setTimeout(function() {
        clearActive();
        clearAnim();
        bottomDesc.textContent = 'Idle';
        bottomBadge.style.display = 'none';
      }, 1500);
    }
  }

  function updateStepInfo() {
    stepInfo.textContent = String(state.currentIndex + 1) + ' / ' + String(state.events.length);
  }

  function addEventLive(ev) {
    state.events.push(ev);
    if (state.mode === 'live') {
      showEvent(state.events.length - 1);
    }
    updateStepInfo();
  }

  // Controls
  function stopTimer() {
    if (state.timerId) { clearInterval(state.timerId); state.timerId = null; }
    state.playing = false;
  }

  function setMode(m) {
    state.mode = m;
    if (m === 'live') {
      modeLabel.textContent = 'LIVE';
      liveDot.classList.add('recording');
    } else {
      modeLabel.textContent = 'REPLAY';
      liveDot.classList.remove('recording');
    }
  }

  function stepForward() {
    stopTimer();
    setMode('replay');
    var next = state.currentIndex + 1;
    if (next < state.events.length) showEvent(next);
  }

  function playPause() {
    if (state.playing) {
      stopTimer();
      return;
    }
    setMode('replay');
    state.playing = true;
    state.timerId = setInterval(function() {
      var next = state.currentIndex + 1;
      if (next >= state.events.length) {
        stopTimer();
        return;
      }
      showEvent(next);
    }, state.speed);
  }

  function replay() {
    stopTimer();
    setMode('replay');
    state.currentIndex = -1;
    clearActive();
    clearAnim();
    bottomDesc.textContent = 'Replay from start...';
    bottomBadge.style.display = 'none';
    updateStepInfo();
    // auto play after short delay
    setTimeout(function() { playPause(); }, 300);
  }

  document.getElementById('btn-replay').addEventListener('click', replay);
  document.getElementById('btn-step').addEventListener('click', stepForward);

  var btnPlay = document.getElementById('btn-play');
  btnPlay.addEventListener('click', function() {
    playPause();
  });

  // speed toggle on double click
  btnPlay.addEventListener('dblclick', function(e) {
    e.preventDefault();
    if (state.speed === 1500) state.speed = 800;
    else if (state.speed === 800) state.speed = 400;
    else state.speed = 1500;
    var label = state.speed === 1500 ? '1x' : (state.speed === 800 ? '2x' : '4x');
    btnPlay.textContent = '\u23EF ' + label;
    // restart timer with new speed if playing
    if (state.playing) {
      clearInterval(state.timerId);
      state.timerId = setInterval(function() {
        var next = state.currentIndex + 1;
        if (next >= state.events.length) { stopTimer(); return; }
        showEvent(next);
      }, state.speed);
    }
  });

  // SSE connection
  var es = new EventSource('/events');
  es.onmessage = function(msg) {
    try {
      var ev = JSON.parse(msg.data);
      addEventLive(ev);
    } catch(e) { console.error(e); }
  };

  // Command modal
  var modal = document.getElementById('command-modal');
  var cmdInput = document.getElementById('command-input');
  var modalSubmit = document.getElementById('modal-submit');
  var modalCancel = document.getElementById('modal-cancel');

  document.getElementById('node-client').addEventListener('click', function() {
    modal.classList.add('visible');
    cmdInput.value = '';
    cmdInput.focus();
  });

  function hideModal() { modal.classList.remove('visible'); }

  modalCancel.addEventListener('click', hideModal);

  function submitCommand() {
    var val = cmdInput.value.trim();
    if (!val) return;
    hideModal();
    bottomDesc.textContent = 'Sending: ' + val;
    bottomBadge.textContent = 'Sending...';
    bottomBadge.style.display = '';
    var clientRect = document.querySelector('#node-client .node-rect');
    clientRect.classList.add('active');
    setTimeout(function() { clientRect.classList.remove('active'); }, 1500);
    fetch('/command', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({instruction: val}) });
  }

  modalSubmit.addEventListener('click', submitCommand);
  cmdInput.addEventListener('keydown', function(e) {
    if (e.key === 'Enter') submitCommand();
    if (e.key === 'Escape') hideModal();
  });

  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape' && modal.classList.contains('visible')) hideModal();
  });

})();
</script>
</body>
</html>`

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}
