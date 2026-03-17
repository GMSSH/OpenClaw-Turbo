<template>
  <div class="conv-page">
    <!-- Top bar -->
    <div class="conv-topbar">
      <div class="topbar-left">
        <div class="status-capsule">
          <span class="capsule-dot" :class="connState"></span>
          <span class="topbar-title">Chat</span>
          <span class="capsule-sep"></span>
          <n-switch :value="wsEnabled" @update:value="handleWsToggle" :loading="wsToggling" size="small" />
          <template v-if="wsEnabled">
            <span class="capsule-sep"></span>
            <span class="bind-toggle has-tip" :class="{ 'is-private': wsBindMode === 'private' }" @click="switchBindMode" :data-tip="bindTip">
              <!-- lucide:globe for public -->
              <svg v-if="wsBindMode === 'public'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20a14.5 14.5 0 0 0 0-20M2 12h20"/></svg>
              <!-- tabler:shield-lock for private -->
              <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M12 3a12 12 0 0 0 8.5 3A12 12 0 0 1 12 21A12 12 0 0 1 3.5 6A12 12 0 0 0 12 3"/><path d="M11 11a1 1 0 1 0 2 0a1 1 0 1 0-2 0m1 1v2.5"/></svg>
              <span class="bind-label">{{ wsBindMode === 'public' ? t('console.webUIPublic') : t('console.webUIPrivate') }}</span>
            </span>
            <span class="capsule-sep"></span>
            <span class="ws-port-pill">
              <span class="port-label">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
              </span>
              <n-input-number v-model:value="wsPort" :min="1024" :max="65535" size="tiny" :show-button="false" :disabled="wsToggling" @blur="handlePortChange" @keydown.enter="handlePortChange" />
            </span>
          </template>
          <span v-else class="ws-off-tip">对话已关闭</span>
        </div>
      </div>
      <div class="topbar-actions">
        <button class="tb-icon-btn has-tip" :class="{ active: showDebug }" @click="toggleDebug" data-tip="显示工具和思考过程">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a7 7 0 0 0-7 7c0 2.38 1.19 4.47 3 5.74V17a1 1 0 001 1h6a1 1 0 001-1v-2.26c1.81-1.27 3-3.36 3-5.74a7 7 0 00-7-7z"/><line x1="9" y1="21" x2="15" y2="21"/><line x1="10" y1="24" x2="14" y2="24"/><line x1="12" y1="17" x2="12" y2="12"/><line x1="9.5" y1="14" x2="12" y2="12"/><line x1="14.5" y1="14" x2="12" y2="12"/></svg>
        </button>
        <span class="tb-divider"></span>
        <button class="tb-icon-btn has-tip" @click="loadHistory" data-tip="刷新" :disabled="loadingHistory">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/></svg>
        </button>
      </div>
    </div>
    <!-- 首次访问提示 -->
    <div v-if="showHint" class="chat-hint">
      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/></svg>
      <span>如遇无法连接，请确认服务器是否有公网，点击上方「公网访问 / 内网访问」按钮切换或更改端口尝试。</span>
      <button class="hint-close" @click="dismissHint">
        <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
      </button>
    </div>
    <!-- Messages area -->
    <div class="messages-area" ref="messagesRef">
      <div class="messages-inner">
        <!-- Welcome -->
        <div v-if="messages.length === 0 && !loadingHistory" class="welcome-block">
          <div class="welcome-avatar">
            <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M12 2a7 7 0 0 1 7 7c0 3-2 5.5-4 7l-1 3h-4l-1-3c-2-1.5-4-4-4-7a7 7 0 0 1 7-7z"/>
              <line x1="9" y1="22" x2="15" y2="22"/>
            </svg>
          </div>
          <p class="welcome-text">输入消息开始对话</p>
        </div>

        <!-- Loading history -->
        <div v-if="loadingHistory" class="loading-history">
          <div class="loading-spinner"></div>
          <span>加载消息...</span>
        </div>

        <!-- Messages -->
        <template v-for="(msg, idx) in messages" :key="idx">
          <!-- User message -->
          <div v-if="msg.role === 'user'" class="msg-row user-row">
            <div class="user-bubble">
              <div class="msg-md" v-html="renderMd(msg.text)"></div>
            </div>
            <div class="avatar user-av">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            </div>
          </div>

          <!-- Assistant message (hide if only thinking/tools and debug off) -->
          <div v-else-if="msg.text || showDebug" class="msg-row bot-row">
            <div class="avatar bot-av">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 2a7 7 0 0 1 7 7c0 3-2 5.5-4 7l-1 3h-4l-1-3c-2-1.5-4-4-4-7a7 7 0 0 1 7-7z"/><line x1="9" y1="22" x2="15" y2="22"/></svg>
            </div>
            <div class="bot-content">
              <!-- Thinking -->
              <div v-if="showDebug && msg.thinking" class="think-row">
                <div class="think-toggle" @click="msg.showThinking = !msg.showThinking">
                  <svg :class="{ open: msg.showThinking }" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="9 6 15 12 9 18"/></svg>
                  <span>Thinking</span>
                </div>
                <div v-show="msg.showThinking" class="think-body">{{ msg.thinking }}</div>
              </div>
              <!-- Tool calls -->
              <div v-if="showDebug && msg.tools && msg.tools.length" class="tool-calls">
                <div v-for="(tool, ti) in msg.tools" :key="ti" class="tool-chip-wrap">
                  <div class="tool-chip" @click="tool.expanded = !tool.expanded">
                    <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 000 1.4l1.6 1.6a1 1 0 001.4 0l3.77-3.77a6 6 0 01-7.94 7.94l-6.91 6.91a2.12 2.12 0 01-3-3l6.91-6.91a6 6 0 017.94-7.94l-3.76 3.76z"/></svg>
                    <span class="tool-name">{{ tool.name }}</span>
                    <svg :class="{ open: tool.expanded }" viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="9 6 15 12 9 18"/></svg>
                  </div>
                  <div v-show="tool.expanded" class="tool-detail">
                    <pre>{{ tool.content }}</pre>
                  </div>
                </div>
              </div>
              <!-- Text card -->
              <div v-if="msg.text" class="bot-card">
                <div class="msg-md" v-html="renderMd(msg.text)"></div>
                <span v-if="msg.streaming" class="cursor-blink"></span>
              </div>
            </div>
          </div>
        </template>

        <!-- Typing indicator -->
        <div v-if="agentTyping && !currentAssistantMsg" class="msg-row bot-row">
          <div class="avatar bot-av">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 2a7 7 0 0 1 7 7c0 3-2 5.5-4 7l-1 3h-4l-1-3c-2-1.5-4-4-4-7a7 7 0 0 1 7-7z"/><line x1="9" y1="22" x2="15" y2="22"/></svg>
          </div>
          <div class="bot-card typing-card">
            <div class="dots"><span></span><span></span><span></span></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Input bar -->
    <div class="input-bar">
      <div class="input-inner">
        <div class="conn-pill" :class="connState" :title="connLabel">
          <span class="cd"></span>
        </div>
        <div class="input-box" :class="{ focused: inputFocused }">
          <textarea
            ref="inputRef"
            v-model="inputText"
            :disabled="!wsReady || sending"
            :placeholder="wsReady ? 'Message (Enter to send, Shift+Enter for new line)' : t('console.waitingConnection')"
            rows="1"
            @focus="inputFocused = true"
            @blur="inputFocused = false"
            @keydown="onKeyDown"
            @compositionstart="isComposing = true"
            @compositionend="isComposing = false"
            @input="autoResize"
          ></textarea>
          <button v-if="agentTyping" class="stop-btn has-tip" @click="abortReply" :data-tip="t('console.stopReply')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
          </button>
          <button v-else class="send-btn" :disabled="!canSend" @click="sendMessage">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<!-- Removed module-level script to prevent white screen crashes and memory leaks -->

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { NSwitch, NInputNumber } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import { getClawWsInfo, getWsProxyStatus, toggleWsProxy } from '@/api/conversation'
import gm from '@/utils/gmssh'

marked.setOptions({ breaks: true, gfm: true })

const { t } = useI18n()

// ====== WebSocket internal state ======
let ws = null
let currentRunId = null
let pendingReqs = {}
let wsRetries = 0
let retryTimer = null

const messages = ref([])
const inputText = ref('')
const inputFocused = ref(false)
const showDebug = ref(localStorage.getItem('conv_debug') === '1')
function toggleDebug() { showDebug.value = !showDebug.value; localStorage.setItem('conv_debug', showDebug.value ? '1' : '0') }
let isComposing = false
const sending = ref(false)
const agentTyping = ref(false)
const loadingHistory = ref(false)
const connState = ref('disconnected')
const wsReady = computed(() => connState.value === 'connected')
const canSend = computed(() => wsReady.value && inputText.value.trim() && !sending.value)
const messagesRef = ref(null)
const inputRef = ref(null)

const connLabel = computed(() => ({ disconnected: '未连接', connecting: '连接中...', connected: '已连接' }[connState.value]))
const bindTip = computed(() => {
  if (wsBindMode.value === 'public') return t('console.webUIPublic') + ' ' + (window.$gm?.publicIp || 'localhost')
  return t('console.webUIPrivate') + ' ' + (window.$gm?.privateIp || 'localhost')
})
const currentAssistantMsg = computed(() => {
  const l = messages.value[messages.value.length - 1]
  return l?.role === 'assistant' && l?.streaming ? l : null
})

// WS 代理设置
const wsEnabled = ref(false)
const wsPort = ref(37300)
let lastSavedPort = 37300
const wsBindMode = ref(localStorage.getItem('ws_bind_mode') || 'public')
const wsRunning = ref(false)
const wsToggling = ref(false)
const isHttps = (window.$gm?.baseURL || '').startsWith('https')
const showHint = ref(!localStorage.getItem('chat_hint_dismissed'))
function dismissHint() { showHint.value = false; localStorage.setItem('chat_hint_dismissed', '1') }

async function loadWsStatus() {
  try {
    const res = await getWsProxyStatus()
    wsEnabled.value = !!res?.enabled
    wsPort.value = res?.port || 37300
    lastSavedPort = wsPort.value
    wsRunning.value = !!res?.running
  } catch {}
}

async function handleWsToggle(val) {
  wsToggling.value = true
  try {
    // 关闭开关时先断开 WS
    if (!val) {
      if (ws) { ws.onclose = null; ws.close(); ws = null }
      connState.value = 'disconnected'
      if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
    }
    const res = await toggleWsProxy({ enabled: val, port: wsPort.value, useSSL: isHttps })
    if (res?.success) {
      wsEnabled.value = val
      wsRunning.value = val
      if (res.port) wsPort.value = res.port
      gm.success(res.message)
      if (val) {
        // 先断开旧连接再重连
        if (ws) { ws.onclose = null; ws.close(); ws = null }
        connState.value = 'disconnected'
        wsRetries = 0
        if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
        connectWs()
      }
    } else {
      gm.error(res?.message || '操作失败')
    }
  } catch (e) {
    gm.error('操作失败: ' + (e.message || e))
  }
  wsToggling.value = false
}

async function handlePortChange() {
  if (!wsEnabled.value) return
  if (wsPort.value === lastSavedPort) return
  wsToggling.value = true
  try {
    const res = await toggleWsProxy({ enabled: true, port: wsPort.value, useSSL: isHttps })
    if (res?.success) {
      wsRunning.value = true
      if (res.port) wsPort.value = res.port
      lastSavedPort = wsPort.value
      gm.success(res.message)
      // 端口变了，关闭旧连接并重连
      if (ws) { ws.onclose = null; ws.close(); ws = null }
      connState.value = 'disconnected'
      wsRetries = 0
      connectWs()
    } else {
      gm.error(res?.message || '端口切换失败')
    }
  } catch (e) {
    gm.error('端口切换失败: ' + (e.message || e))
  }
  wsToggling.value = false
}

function switchBindMode() {
  const newMode = wsBindMode.value === 'public' ? 'private' : 'public'
  wsBindMode.value = newMode
  localStorage.setItem('ws_bind_mode', newMode)
  // 清理现有前端 WS 连接
  if (ws) { ws.onclose = null; ws.close(); ws = null }
  connState.value = 'disconnected'
  wsRetries = 0
  if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
  // 后端代理绑定 0.0.0.0，无需重启，前端用新 IP 重连即可
  if (wsEnabled.value) {
    gm.success(newMode === 'private' ? t('console.webUIPrivate') : t('console.webUIPublic'))
    connectWs()
  }
}

function renderMd(text) {
  if (!text) return ''
  try { return marked.parse(text) } catch { return text }
}

function genId() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
    const r = Math.random() * 16 | 0
    return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16)
  })
}

// ====== WebSocket (module-level singleton) ======
async function connectWs() {
  // 如果已连接，不重复连接
  if (ws && ws.readyState === WebSocket.OPEN) return
  // 清理残留连接
  if (ws) { try { ws.onclose = null; ws.close() } catch {} ws = null }
  if (connState.value === 'connecting') return
  connState.value = 'connecting'
  try {
    const info = await getClawWsInfo()
    const host = wsBindMode.value === 'private'
      ? (window.$gm?.privateIp || 'localhost')
      : gm.getPublicIp()
    const b64 = btoa(info.token).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
    const isHttps = (window.$gm?.baseURL || '').startsWith('https')
    const wsProto = isHttps ? 'wss' : 'ws'
    ws = new WebSocket(`${wsProto}://${host}:${info.port}/ws/chat?token=${b64}`)
    ws.onopen = () => {}
    ws.onmessage = ev => { let d; try { d = JSON.parse(ev.data) } catch { return }; handleMsg(d) }
    ws.onerror = () => { connState.value = 'disconnected' }
    ws.onclose = () => {
      ws = null
      connState.value = 'disconnected'
      if (wsRetries < 5) {
        const delay = Math.min(3000 * Math.pow(2, wsRetries), 60000)
        wsRetries++
        if (retryTimer) clearTimeout(retryTimer)
        retryTimer = setTimeout(() => { retryTimer = null; connectWs() }, delay)
      }
    }
  } catch (e) { connState.value = 'disconnected'; gm.error('连接失败: ' + (e.message || '')) }
}

function handleMsg(d) {
  if (d.type === 'event' && d.event === 'proxy.connected') { connState.value = 'connected'; wsRetries = 0; loadHistory(); return }
  if (d.type === 'event' && d.event === 'proxy.error') { gm.error(d.payload?.message || ''); connState.value = 'disconnected'; return }
  if (d.type === 'res') { const h = pendingReqs[d.id]; if (h) { delete pendingReqs[d.id]; d.ok ? h.resolve(d.payload) : h.reject(d.payload) }; return }
  if (d.type === 'event') handleEvent(d)
}

function wsSend(method, params = {}) {
  return new Promise((resolve, reject) => {
    const id = genId(); pendingReqs[id] = { resolve, reject }
    ws.send(JSON.stringify({ type: 'req', id, method, params }))
  })
}

// ====== History ======
async function loadHistory() {
  loadingHistory.value = true
  try {
    const res = await wsSend('chat.history', { sessionKey: 'agent:main:main', limit: 200 })
    if (res?.messages) {
      const parsed = []
      let pendingTools = []
      for (const m of res.messages) {
        // Skip session start
        if (m.role === 'user' && (m.content?.[0]?.text || '').startsWith('A new session was started')) continue
        // Collect tool calls/results
        if (m.role === 'toolUse') {
          const name = m.toolName || m.content?.[0]?.toolName || 'tool'
          pendingTools.push({ name, content: JSON.stringify(m.input || m.content, null, 2), expanded: false })
          continue
        }
        if (m.role === 'toolResult') {
          const name = m.toolName || 'tool'
          let content = ''
          if (m.details) content = JSON.stringify(m.details, null, 2)
          else if (m.content?.[0]?.text) content = m.content[0].text
          else content = JSON.stringify(m.content, null, 2)
          pendingTools.push({ name, content, expanded: false })
          continue
        }
        // Flush pending tools into the next assistant message
        const msg = {
          role: m.role,
          text: (m.content?.filter(c => c.type === 'text') || []).map(t => t.text).join(''),
          thinking: (m.content?.filter(c => c.type === 'thinking') || []).map(t => t.thinking).join(''),
          tools: [], showThinking: false, streaming: false,
        }
        if (m.role === 'assistant' && pendingTools.length) {
          msg.tools = pendingTools
          pendingTools = []
        }
        parsed.push(msg)
      }
      // If there are remaining pending tools, attach to the last assistant msg or create one
      if (pendingTools.length) {
        const lastBot = [...parsed].reverse().find(m => m.role === 'assistant')
        if (lastBot) lastBot.tools = [...(lastBot.tools || []), ...pendingTools]
        else parsed.push({ role: 'assistant', text: '', thinking: '', tools: pendingTools, showThinking: false, streaming: false })
      }
      messages.value = parsed
    }
  } catch {}
  finally { loadingHistory.value = false; await nextTick(); scrollBottom() }
}

// ====== Events ======
function handleEvent(d) {
  const { event, payload } = d
  if (event === 'agent') {
    const { runId, stream, data: ev } = payload
    if (stream === 'lifecycle') {
      if (ev?.phase === 'start') { agentTyping.value = true; currentRunId = runId }
      else if (ev?.phase === 'end') {
        agentTyping.value = false
        const l = messages.value[messages.value.length - 1]
        if (l?.streaming) l.streaming = false
        currentRunId = null
        // Refetch history to backfill tools & thinking
        backfillLastMsg()
      }
      return
    }
    if (stream === 'assistant' && ev) {
      const l = messages.value[messages.value.length - 1]
      if (l?.role === 'assistant' && l?.streaming) { if (ev.delta) l.text = ev.text || (l.text + ev.delta) }
      else { messages.value.push({ role: 'assistant', text: ev.delta || ev.text || '', thinking: '', showThinking: false, streaming: true }) }
      nextTick(scrollBottom); return
    }
    if ((stream === 'thinking' || stream === 'reasoning') && ev) {
      let l = messages.value[messages.value.length - 1]
      if (!(l?.role === 'assistant' && l?.streaming)) {
        messages.value.push({ role: 'assistant', text: '', thinking: '', showThinking: false, streaming: true })
        l = messages.value[messages.value.length - 1]
      }
      l.thinking = (l.thinking || '') + (ev.delta || ev.text || ''); return
    }
  }
  if (event === 'chat' && payload?.state === 'final') {
    const m = payload.message
    if (m?.role === 'assistant') {
      const l = messages.value[messages.value.length - 1]
      if (l?.role === 'assistant') {
        l.text = (m.content?.filter(c => c.type === 'text') || []).map(t => t.text).join('')
        l.streaming = false
        const tk = m.content?.filter(c => c.type === 'thinking') || []
        if (tk.length) l.thinking = tk.map(t => t.thinking).join('')
      }
      nextTick(scrollBottom)
    }
  }
}

// ====== Backfill tools & thinking after run ends ======
async function backfillLastMsg() {
  try {
    const res = await wsSend('chat.history', { sessionKey: 'agent:main:main', limit: 20 })
    if (!res?.messages) return
    const hist = res.messages
    // Find the last assistant message in history
    const lastAssist = [...hist].reverse().find(m => m.role === 'assistant')
    if (!lastAssist) return
    // Collect tool messages before it
    const assistIdx = hist.indexOf(lastAssist)
    const tools = []
    for (let i = 0; i < assistIdx; i++) {
      const m = hist[i]
      if (m.role === 'toolUse') {
        tools.push({ name: m.toolName || m.content?.[0]?.toolName || 'tool', content: JSON.stringify(m.input || m.content, null, 2), expanded: false })
      } else if (m.role === 'toolResult') {
        let content = ''
        if (m.details) content = JSON.stringify(m.details, null, 2)
        else if (m.content?.[0]?.text) content = m.content[0].text
        else content = JSON.stringify(m.content, null, 2)
        tools.push({ name: m.toolName || 'tool', content, expanded: false })
      }
    }
    // Patch the last assistant message in our UI
    const uiLast = messages.value[messages.value.length - 1]
    if (uiLast?.role === 'assistant') {
      const thinking = (lastAssist.content?.filter(c => c.type === 'thinking') || []).map(t => t.thinking).join('')
      if (thinking) uiLast.thinking = thinking
      if (tools.length) uiLast.tools = tools
    }
  } catch {}
}

// ====== Send ======
async function sendMessage() {
  const text = inputText.value.trim(); if (!text || !wsReady.value) return
  sending.value = true; inputText.value = ''; resetTA()
  messages.value.push({ role: 'user', text, thinking: '', showThinking: false, streaming: false })
  await nextTick(); scrollBottom()
  try { await wsSend('chat.send', { sessionKey: 'agent:main:main', message: text, deliver: false, idempotencyKey: genId() }) }
  catch (e) { gm.error('发送失败: ' + (e.message || '')) }
  finally { sending.value = false }
}

function onKeyDown(e) { if (e.key === 'Enter' && !e.shiftKey && !isComposing) { e.preventDefault(); sendMessage() } }
function abortReply() {
  if (!currentRunId || !wsReady.value) return
  wsSend('chat.abort', { sessionKey: 'agent:main:main', runId: currentRunId }).catch(() => {})
}
function autoResize() { const el = inputRef.value; if (!el) return; el.style.height = 'auto'; el.style.height = Math.min(el.scrollHeight, 150) + 'px' }
function resetTA() { nextTick(() => { const el = inputRef.value; if (el) el.style.height = 'auto' }) }
function scrollBottom() { const el = messagesRef.value; if (el) el.scrollTop = el.scrollHeight }

// Component lifecycle — properly manage WS to avoid leaks
onMounted(async () => {
  await loadWsStatus()
  if (!wsEnabled.value) return   // WS 未启用，不连接

  // 如果配置已启用但代理未运行（比如后端重启过），先启动代理
  if (!wsRunning.value) {
    try {
      const res = await toggleWsProxy({ enabled: true, port: wsPort.value, useSSL: isHttps })
      if (res?.success) {
        wsRunning.value = true
        if (res.port) wsPort.value = res.port
      } else {
        gm.error(res?.message || '启动 WS 代理失败')
        return
      }
    } catch (e) {
      gm.error('启动 WS 代理失败: ' + (e.message || e))
      return
    }
  }

  connectWs()
})

onUnmounted(() => {
  // 彻底清理 WS 防止切到其他页面时由于消息到来而触发响应式更新导致崩溃
  if (ws) {
    ws.onclose = null
    ws.onerror = null
    ws.onmessage = null
    ws.close()
    ws = null
  }
  if (retryTimer) {
    clearTimeout(retryTimer)
    retryTimer = null
  }
})
</script>

<style scoped>
.conv-page { width: 100%; height: 100%; display: flex; flex-direction: column; overflow: hidden; }

/* ===== First-visit hint ===== */
.chat-hint {
  flex-shrink: 0; display: flex; align-items: center; gap: 8px;
  padding: 8px 16px; margin: 0;
  background: rgba(59, 130, 246, 0.08);
  border-bottom: 1px solid rgba(59, 130, 246, 0.15);
  font-size: 12px; color: var(--jm-accent-5); line-height: 1.5;
  animation: hintSlideIn 0.3s ease-out;
}
.chat-hint svg { flex-shrink: 0; color: #3b82f6; }
.chat-hint span { flex: 1; }
.hint-close {
  flex-shrink: 0; width: 20px; height: 20px; border-radius: 6px;
  border: none; background: transparent; color: var(--jm-accent-4);
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.2s;
}
.hint-close:hover { background: rgba(var(--jm-accent-2-rgb), 0.3); color: var(--jm-accent-6); }
@keyframes hintSlideIn { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: none; } }

/* ===== Top bar — Floating Glass ===== */
.conv-topbar {
  flex-shrink: 0; padding: 8px 20px;
  display: flex; align-items: center; justify-content: space-between;
  border-bottom: none;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(20px); -webkit-backdrop-filter: blur(20px);
  position: relative; z-index: 10;
  box-shadow:
    var(--jm-glass-inner-glow),
    0 1px 2px rgba(0, 0, 0, 0.06),
    0 4px 16px rgba(0, 0, 0, 0.04);
}
.conv-topbar::after {
  content: '';
  position: absolute; bottom: 0; left: 5%; right: 5%; height: 1px;
  background: linear-gradient(90deg,
    transparent,
    rgba(var(--jm-primary-1-rgb), 0.15) 30%,
    rgba(var(--jm-primary-1-rgb), 0.25) 50%,
    rgba(var(--jm-primary-1-rgb), 0.15) 70%,
    transparent
  );
}
/* WS 代理内联 */
.ws-port-pill {
  display: flex; align-items: center; gap: 0;
  border-radius: 8px; overflow: hidden;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.2);
}
.ws-port-pill .port-label {
  display: flex; align-items: center; justify-content: center;
  padding: 0 5px 0 6px; color: var(--jm-accent-4);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  border-right: 1px solid var(--jm-glass-border);
}
.ws-port-pill :deep(.n-input-number) { width: 64px; font-size: 12px; }
.ws-port-pill :deep(.n-input) {
  --n-height: 22px !important;
  --n-border: none !important; --n-border-hover: none !important;
  --n-border-focus: none !important; --n-box-shadow-focus: none !important;
  --n-color: transparent !important; --n-color-focus: transparent !important;
  font-size: 12px; font-weight: 600;
  font-family: var(--jm-font-mono);
}
.ws-port-pill :deep(.n-input__input-el) { text-align: center; }
.ws-off-tip { font-size: 11px; color: var(--jm-accent-4); }

/* Bind toggle (公网/内网) —— 增强视觉效果 */
.bind-toggle {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 3px 10px; border-radius: 12px;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.25);
  cursor: pointer; transition: all 0.3s ease;
  font-size: 11px; font-weight: 600; color: #22c55e;
  box-shadow: 0 0 8px rgba(34, 197, 94, 0.1);
}
.bind-toggle:hover {
  background: rgba(34, 197, 94, 0.18);
  box-shadow: 0 0 12px rgba(34, 197, 94, 0.2);
  transform: scale(1.03);
}
.bind-toggle.is-private {
  color: #f59e0b;
  border-color: rgba(245, 158, 11, 0.3);
  background: rgba(245, 158, 11, 0.1);
  box-shadow: 0 0 8px rgba(245, 158, 11, 0.1);
}
.bind-toggle.is-private:hover {
  background: rgba(245, 158, 11, 0.18);
  box-shadow: 0 0 12px rgba(245, 158, 11, 0.2);
}
.bind-label { white-space: nowrap; }
.topbar-left { display: flex; align-items: center; }
.status-capsule {
  display: inline-flex; align-items: center; gap: 8px;
  padding: 5px 14px; border-radius: 20px;
  background: rgba(var(--jm-accent-1-rgb), 0.25);
  backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);
  border: 1px solid var(--jm-glass-border);
  box-shadow: var(--jm-glass-inner-glow), 0 1px 4px rgba(0, 0, 0, 0.06);
}
.capsule-dot {
  width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0;
  transition: all 0.3s;
}
.capsule-dot.connected { background: #3cff7a; box-shadow: 0 0 6px rgba(60, 255, 122, 0.5); animation: led-breathe-green 2s ease-in-out infinite; }
.capsule-dot.connecting { background: #fb923c; animation: pulse 1s ease-in-out infinite; }
.capsule-dot.disconnected { background: #f87171; }
@keyframes led-breathe-green {
  0%, 100% { box-shadow: 0 0 4px rgba(60, 255, 122, 0.4); }
  50% { box-shadow: 0 0 10px rgba(60, 255, 122, 0.7), 0 0 20px rgba(60, 255, 122, 0.2); }
}
.capsule-sep {
  width: 1px; height: 14px;
  background: var(--jm-glass-border); margin: 0 2px;
}
.topbar-title { font-size: 14px; font-weight: 600; color: var(--jm-accent-7); letter-spacing: -0.01em; }
.topbar-actions { display: flex; align-items: center; gap: 4px; }
.tb-icon-btn {
  width: 28px; height: 28px; border-radius: 8px;
  border: none;
  background: transparent; color: var(--jm-accent-4);
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.25s;
}
.tb-icon-btn:hover:not(:disabled) { color: var(--jm-accent-6); background: var(--jm-glass-bg-hover); transform: scale(1.08); }
.tb-icon-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.tb-icon-btn.active { color: var(--jm-primary-1); border-color: rgba(var(--jm-primary-1-rgb), 0.3); background: rgba(var(--jm-primary-1-rgb), 0.08); box-shadow: 0 0 8px rgba(var(--jm-primary-1-rgb), 0.1); }
.tb-divider { display: none; }
/* Instant tooltip */
.has-tip { position: relative; }
.has-tip::after {
  content: attr(data-tip);
  position: absolute; bottom: -28px; left: 50%; transform: translateX(-50%);
  padding: 3px 8px; border-radius: 6px;
  background: var(--jm-overlay-bg, rgba(0,0,0,0.85)); color: #fff;
  font-size: 11px; white-space: nowrap;
  pointer-events: none; opacity: 0; transition: opacity 0.15s; z-index: 10;
}
.has-tip:hover::after { opacity: 1; }

/* ===== Messages ===== */
.messages-area { flex: 1; overflow-y: auto; scroll-behavior: smooth; }
.messages-inner { padding: 24px 28px 16px; display: flex; flex-direction: column; gap: 14px; min-height: 100%; }

/* Welcome */
.welcome-block { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: var(--jm-accent-4); }
.welcome-avatar {
  width: 56px; height: 56px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, rgba(var(--jm-primary-1-rgb), 0.15), rgba(var(--jm-primary-1-rgb), 0.05));
  color: var(--jm-primary-1); box-shadow: 0 4px 20px rgba(var(--jm-primary-1-rgb), 0.08);
}
.welcome-text { margin: 0; font-size: 14px; }

/* Loading */
.loading-history { display: flex; align-items: center; justify-content: center; gap: 8px; padding: 20px; font-size: 13px; color: var(--jm-accent-4); }
.loading-spinner { width: 16px; height: 16px; border: 2px solid var(--jm-glass-border); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.6s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== Message rows ===== */
.msg-row { display: flex; gap: 10px; align-items: flex-start; animation: msgPop 0.3s cubic-bezier(0.22, 1, 0.36, 1) both; }
@keyframes msgPop { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: none; } }
.user-row { justify-content: flex-end; }
.bot-row { justify-content: flex-start; }

/* Avatars */
.avatar { width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0; display: flex; align-items: center; justify-content: center; margin-top: 2px; }
.bot-av { background: linear-gradient(135deg, var(--jm-primary-1), var(--jm-primary-2)); color: #fff; box-shadow: 0 2px 8px rgba(var(--jm-primary-1-rgb), 0.2); }
.user-av { background: rgba(var(--jm-accent-5-rgb, 180,180,180), 0.12); color: var(--jm-accent-5); border: 1px solid var(--jm-glass-border); }

/* User bubble — Inner Glow */
.user-bubble {
  max-width: 85%; padding: 10px 14px;
  border-radius: 16px 16px 4px 16px;
  background: var(--jm-primary-1); color: #fff;
  font-size: 13.5px; line-height: 1.6; word-break: break-word;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.08), 0 1px 4px rgba(0,0,0,0.08);
}
.user-bubble .msg-md :deep(p) { margin: 0; }
.user-bubble .msg-md :deep(code) { background: rgba(255,255,255,0.15); color: #fff; padding: 1px 4px; border-radius: 4px; font-size: 0.88em; font-family: var(--jm-font-mono); }
.user-bubble .msg-md :deep(a) { color: #fff; text-decoration: underline; }

/* Bot content wrapper */
.bot-content { display: flex; flex-direction: column; gap: 6px; max-width: 85%; }

/* Bot card — Soft Drop Shadow + Glass */
.bot-card {
  padding: 12px 16px;
  border-radius: 4px 16px 16px 16px;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--jm-glass-border);
  font-size: 13.5px; line-height: 1.7; color: var(--jm-accent-7);
  word-break: break-word;
  box-shadow: 0 1px 6px rgba(0,0,0,0.05);
}

/* ===== Thinking — glass ===== */
.think-row { display: flex; flex-direction: column; gap: 4px; }
.think-toggle {
  display: inline-flex; align-items: center; gap: 5px; width: fit-content;
  font-size: 13px; font-weight: 500; color: var(--jm-accent-5);
  cursor: pointer; padding: 5px 12px; border-radius: 10px;
  background: var(--jm-glass-bg); border: 1px solid var(--jm-glass-border);
  transition: all 0.2s;
}
.think-toggle:hover { background: var(--jm-glass-bg-hover); color: var(--jm-accent-6); }
.think-toggle svg { transition: transform 0.2s; }
.think-toggle svg.open { transform: rotate(90deg); }
.think-body {
  padding: 8px 12px; margin-left: 4px; font-size: 12px; color: var(--jm-accent-4);
  line-height: 1.6; white-space: pre-wrap; border-left: 2px solid var(--jm-glass-border);
  max-height: 300px; overflow-y: auto;
}

/* ===== Tool call chips — Glass ===== */
.tool-calls { display: flex; flex-direction: column; gap: 4px; }
.tool-chip-wrap { display: flex; flex-direction: column; }
.tool-chip {
  display: inline-flex; align-items: center; gap: 6px; width: fit-content;
  font-size: 12.5px; color: var(--jm-accent-5);
  cursor: pointer; padding: 4px 10px; border-radius: 10px;
  background: var(--jm-glass-bg); border: 1px solid var(--jm-glass-border);
  transition: all 0.2s;
}
.tool-chip:hover { background: var(--jm-glass-bg-hover); color: var(--jm-accent-6); border-color: var(--jm-glass-border-hover); }
.tool-chip svg { flex-shrink: 0; }
.tool-chip svg:last-child { transition: transform 0.2s; }
.tool-chip svg.open { transform: rotate(90deg); }
.tool-name { font-weight: 500; font-family: var(--jm-font-mono); }
.tool-detail { margin-top: 4px; margin-left: 4px; border-left: 2px solid var(--jm-glass-border); max-height: 250px; overflow: auto; }
.tool-detail pre { margin: 0; padding: 8px 12px; font-size: 11.5px; line-height: 1.5; color: var(--jm-accent-5); font-family: var(--jm-font-mono); white-space: pre-wrap; word-break: break-all; }

/* ===== Markdown ===== */
.msg-md :deep(p) { margin: 0 0 6px; }
.msg-md :deep(p:last-child) { margin-bottom: 0; }
.msg-md :deep(h1), .msg-md :deep(h2), .msg-md :deep(h3), .msg-md :deep(h4) { margin: 10px 0 4px; font-weight: 600; color: var(--jm-accent-7); }
.msg-md :deep(h1) { font-size: 1.3em; } .msg-md :deep(h2) { font-size: 1.15em; } .msg-md :deep(h3) { font-size: 1.05em; }
.msg-md :deep(strong) { font-weight: 600; }
.msg-md :deep(code) { background: rgba(var(--jm-accent-1-rgb), 0.5); padding: 1px 5px; border-radius: 4px; font-size: 0.88em; font-family: var(--jm-font-mono); color: var(--jm-primary-2); }
.msg-md :deep(pre) { background: rgba(var(--jm-accent-1-rgb), 0.4); border: 1px solid var(--jm-glass-border); border-radius: 10px; padding: 10px 14px; overflow-x: auto; margin: 6px 0; }
.msg-md :deep(pre code) { background: none; padding: 0; font-size: 12px; line-height: 1.5; color: var(--jm-accent-7); }
.msg-md :deep(ul), .msg-md :deep(ol) { margin: 4px 0; padding-left: 18px; }
.msg-md :deep(li) { margin: 2px 0; }
.msg-md :deep(blockquote) { margin: 6px 0; padding: 4px 12px; border-left: 3px solid var(--jm-primary-1); color: var(--jm-accent-5); background: rgba(var(--jm-primary-1-rgb), 0.04); border-radius: 0 8px 8px 0; }
.msg-md :deep(a) { color: var(--jm-primary-1); text-decoration: none; }
.msg-md :deep(a:hover) { text-decoration: underline; }
.msg-md :deep(table) { border-collapse: collapse; margin: 6px 0; width: 100%; }
.msg-md :deep(th), .msg-md :deep(td) { border: 1px solid var(--jm-glass-border); padding: 5px 8px; font-size: 12.5px; }
.msg-md :deep(th) { background: rgba(var(--jm-accent-1-rgb), 0.3); font-weight: 600; }
.msg-md :deep(hr) { border: none; border-top: 1px solid var(--jm-glass-border); margin: 10px 0; }

/* Typing */
.cursor-blink { display: inline-block; width: 2px; height: 14px; background: var(--jm-primary-1); margin-left: 1px; vertical-align: text-bottom; animation: blink .8s step-end infinite; }
@keyframes blink { 50% { opacity: 0; } }
.typing-card { padding: 10px 16px; }
.dots { display: flex; gap: 4px; }
.dots span { width: 6px; height: 6px; border-radius: 50%; background: var(--jm-accent-4); animation: bounce 1.2s ease-in-out infinite; }
.dots span:nth-child(2) { animation-delay: .15s; }
.dots span:nth-child(3) { animation-delay: .3s; }
@keyframes bounce { 0%,60%,100% { transform: translateY(0); opacity: .4; } 30% { transform: translateY(-5px); opacity: 1; } }

/* ===== Input bar — Frosted ===== */
.input-bar { flex-shrink: 0; padding: 12px 28px 18px; }
.input-inner { display: flex; align-items: flex-end; gap: 10px; }

/* Connection pill */
.conn-pill { width: 8px; height: 8px; border-radius: 50%; margin-bottom: 16px; flex-shrink: 0; }
.conn-pill .cd { display: block; width: 100%; height: 100%; border-radius: 50%; background: currentColor; }
.conn-pill.connected { color: #3cff7a; }
.conn-pill.connecting { color: #fb923c; animation: pulse 1s ease-in-out infinite; }
.conn-pill.disconnected { color: #f87171; }
@keyframes pulse { 50% { opacity: .3; } }

/* Input box — glass + neon focus */
.input-box {
  flex: 1; display: flex; align-items: flex-end;
  border: 1px solid var(--jm-glass-border); border-radius: 14px;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  padding: 6px 6px 6px 14px; transition: all 0.3s;
}
.input-box.focused { border-color: var(--jm-glass-border-hover); box-shadow: 0 0 0 2px rgba(var(--jm-primary-1-rgb), 0.12), 0 0 12px rgba(var(--jm-primary-1-rgb), 0.06); }
.input-box textarea { flex: 1; border: none; outline: none; background: transparent; color: var(--jm-accent-7); font-size: 14px; line-height: 1.5; resize: none; max-height: 150px; font-family: inherit; padding: 6px 0; }
.input-box textarea::placeholder { color: var(--jm-accent-4); }

.send-btn {
  width: 34px; height: 34px; flex-shrink: 0; border-radius: 10px; border: none;
  background: var(--jm-primary-1); color: #fff;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.2s ease;
  box-shadow: 0 1px 4px rgba(0,0,0,0.1);
}
.send-btn:hover:not(:disabled) { background: var(--jm-primary-2); transform: scale(1.04); box-shadow: 0 2px 8px rgba(0,0,0,0.12); }
.send-btn:disabled { opacity: 0.2; cursor: not-allowed; }

.stop-btn {
  width: 34px; height: 34px; flex-shrink: 0; border-radius: 10px;
  border: 2px solid rgba(248, 113, 113, 0.4); background: transparent; color: #f87171;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.25s;
  animation: stop-ring-pulse 1.5s ease-in-out infinite;
}
.stop-btn:hover { background: rgba(248,113,113,0.08); border-color: rgba(248,113,113,0.5); transform: scale(1.06); }
@keyframes stop-ring-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(248, 113, 113, 0.2); }
  50% { box-shadow: 0 0 0 4px rgba(248, 113, 113, 0.08); }
}
</style>

