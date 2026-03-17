<template>
  <div class="dashboard-panel">
    <!-- ===== 顶部：状态胶囊 + 操作按钮 ===== -->
    <div class="top-bar">
      <div class="status-pill" :class="!statusLoaded ? 'st-loading' : (running ? 'st-ok' : 'st-off')">
        <span class="st-dot"><span class="st-ring" v-if="running"></span></span>
        <span class="st-text">{{ !statusLoaded ? t('console.statusFetching') : (running ? t('console.statusRunning') : t('console.statusStopped')) }}</span>
        <span class="st-name font-mono">OpenClaw</span>
        <span class="st-sep" v-if="statusData.uptime && statusData.uptime !== '-'">·</span>
        <span class="st-meta" v-if="statusData.uptime && statusData.uptime !== '-'">{{ statusData.uptime }}</span>
      </div>
      <div class="top-actions">
        <!-- 快捷工具胶囊 -->
        <div class="capsule-bar">
          <n-dropdown trigger="click" :options="webUIOptions" @select="onWebUISelect" :theme-overrides="{ color: 'var(--jm-dropdown-bg)', optionColorHover: 'rgba(var(--jm-primary-1-rgb), 0.1)', borderRadius: '10px', boxShadow: 'var(--jm-shadow-elevation-2)', optionTextColor: 'var(--jm-accent-6)', padding: '4px', dividerColor: 'var(--jm-glass-border)' }">
            <button class="cap-btn primary">
              <svg viewBox="0 0 24 24" width="12" height="12" fill="none"><path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><polyline points="15,3 21,3 21,9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><line x1="10" y1="14" x2="21" y2="3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
              Web UI
            </button>
          </n-dropdown>
          <button class="cap-btn" @click="viewLogs">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none"><rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="1.5"/><path d="M8 8l4 4-4 4M14 16h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
           {{ t('console.btnLogs') }}
          </button>
          <button class="cap-btn" @click="openWorkDir">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>
            {{ t('console.btnWorkDir') }}
          </button>
          <button class="cap-btn" @click="openConfig">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><polyline points="14,2 14,8 20,8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
            {{ t('console.btnConfig') }}
          </button>
          <n-tooltip trigger="hover"><template #trigger>
            <button class="cap-btn cap-icon" @click="refreshStatus">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/></svg>
            </button>
          </template>{{ t('console.btnRefresh') }}</n-tooltip>
        </div>
        <!-- 服务控制胶囊 -->
        <div class="capsule-bar capsule-danger">
          <n-tooltip trigger="hover"><template #trigger>
            <button class="cap-btn ctrl-warn" @click="handleRestart" :disabled="actionLoading">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg>
              {{ t('console.btnRestart') }}
            </button>
          </template>{{ t('console.tooltipRestart') }}</n-tooltip>
          <n-tooltip trigger="hover"><template #trigger>
            <button class="cap-btn" @click="handleStop" :disabled="actionLoading">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><path d="M5 7a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              {{ t('console.btnStop') }}
            </button>
          </template>{{ t('console.tooltipStop') }}</n-tooltip>
          <n-tooltip trigger="hover"><template #trigger>
            <button class="cap-btn ctrl-danger-btn" @click="handleUninstall" :disabled="actionLoading">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><path d="M4 7h16m-10 4v6m4-6v6M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2l1-12M9 7V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              {{ t('console.btnUninstall') }}
            </button>
          </template>{{ t('console.tooltipUninstall') }}</n-tooltip>
        </div>
      </div>
    </div>

    <!-- ===== Bento Grid — 非对称布局 ===== -->
    <div class="bento-grid">

      <!-- ■ A区：模型矩阵（视觉锚点） — 最大区域 -->
      <div class="bento-models">
        <div class="section-hd">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
            <path d="M12 2L2 7l10 5 10-5-10-5z" stroke="var(--jm-primary-1)" stroke-width="1.5" stroke-linejoin="round"/>
            <path d="M2 17l10 5 10-5" stroke="var(--jm-primary-2)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12l10 5 10-5" stroke="var(--jm-primary-1)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" opacity="0.5"/>
          </svg>
          <span>{{ t('console.configuredModels') }}</span>
        </div>
        <div class="models-body">
          <div v-if="!modelsLoaded" class="models-placeholder">{{ t('console.modelsLoading') }}</div>
          <div v-else-if="Object.keys(configuredProviders).length === 0" class="models-placeholder">{{ t('console.modelsEmpty') }}</div>
          <template v-else>
            <div v-for="(prov, pid) in configuredProviders" :key="pid" class="provider-row">
              <span class="prov-badge">{{ pid }}</span>
              <div class="model-pills">
                <span v-for="m in (prov.models || [])" :key="m.id"
                  class="model-pill font-mono"
                  :class="{ 'is-default': m.id === configuredDefault }">
                  {{ m.id }}
                </span>
              </div>
            </div>
          </template>
        </div>
        <!-- 默认模型浮动条 -->
        <div v-if="configuredDefault" class="default-bar">
          <span class="default-tag">DEFAULT</span>
          <span class="default-val font-mono">{{ configuredDefault }}</span>
        </div>
      </div>

      <!-- ■ B区（右上）：数据指标 — 悬浮数字 -->
      <div class="bento-metrics">
        <div class="metric-float" v-for="m in metricsData" :key="m.key">
          <span class="mf-num font-mono" :class="m.colorClass">{{ m.display }}</span>
          <span class="mf-label">{{ m.label }} <span class="mf-icon" :class="m.colorClass" v-html="m.icon"></span></span>
        </div>
      </div>

      <!-- ■ C区（右中）：网关信息 + 拓扑 -->
      <div class="bento-gateway">
        <div class="section-hd">
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none">
            <rect x="2" y="2" width="20" height="8" rx="2" stroke="var(--jm-primary-1)" stroke-width="1.5"/>
            <rect x="2" y="14" width="20" height="8" rx="2" stroke="var(--jm-primary-2)" stroke-width="1.5"/>
            <circle cx="6" cy="6" r="1.5" fill="var(--jm-primary-1)"/>
            <circle cx="6" cy="18" r="1.5" fill="var(--jm-primary-2)"/>
          </svg>
          <span>{{ t('console.gatewayTitle') }}</span>
        </div>
        <div class="gw-list">
          <div class="gw-row" v-for="g in gatewayItems" :key="g.label">
            <span class="gw-k">{{ g.label }}</span>
            <span class="gw-v font-mono">{{ g.value }}</span>
          </div>
        </div>
        <!-- 极简拓扑图 -->
        <div class="topo-mini">
          <svg viewBox="0 0 200 36" width="100%" height="36" fill="none">
            <!-- Client 节点 -->
            <rect x="2" y="8" width="44" height="20" rx="6" stroke="var(--jm-accent-3)" stroke-width="1" fill="rgba(var(--jm-accent-1-rgb),0.15)"/>
            <text x="24" y="22" text-anchor="middle" font-size="8" fill="var(--jm-accent-4)" font-family="var(--jm-font-mono)">Client</text>
            <!-- 连线 1 -->
            <line x1="48" y1="18" x2="72" y2="18" stroke="var(--jm-accent-3)" stroke-width="1" stroke-dasharray="3 2"/>
            <circle cx="72" cy="18" r="2" fill="var(--jm-primary-1)"/>
            <!-- Gateway 节点 -->
            <rect x="76" y="8" width="50" height="20" rx="6" stroke="var(--jm-primary-1)" stroke-width="1.2" fill="rgba(var(--jm-primary-1-rgb),0.06)"/>
            <text x="101" y="22" text-anchor="middle" font-size="8" fill="var(--jm-primary-2)" font-weight="600" font-family="var(--jm-font-mono)">Gateway</text>
            <!-- 连线 2 -->
            <line x1="128" y1="18" x2="152" y2="18" stroke="var(--jm-accent-3)" stroke-width="1" stroke-dasharray="3 2"/>
            <circle cx="152" cy="18" r="2" fill="var(--jm-glow-purple, var(--jm-primary-2))"/>
            <!-- Mode 节点 -->
            <rect x="156" y="8" width="40" height="20" rx="6" stroke="var(--jm-accent-3)" stroke-width="1" fill="rgba(var(--jm-accent-1-rgb),0.15)"/>
            <text x="176" y="22" text-anchor="middle" font-size="8" fill="var(--jm-accent-4)" font-family="var(--jm-font-mono)">{{ statusData.bindMode || 'lan' }}</text>
          </svg>
        </div>
      </div>

      <!-- ■ 日志区（右中下）：时间线风格 -->
      <div class="bento-logs">
        <div class="section-hd">
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none">
            <rect x="3" y="3" width="18" height="18" rx="2" stroke="var(--jm-primary-1)" stroke-width="1.5"/>
            <path d="M8 8l4 4-4 4M14 16h4" stroke="var(--jm-primary-2)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <span>{{ t('console.recentLogs') }}</span>
          <button class="log-refresh" @click="loadLogs" :disabled="logsLoading" title="刷新日志">
            <svg :class="{ spinning: logsLoading }" viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/></svg>
          </button>
        </div>
        <div class="log-timeline" ref="logBody">
          <div v-if="!logsLoaded" class="log-placeholder">{{ t('console.logsLoading') }}</div>
          <div v-else-if="parsedLogs.length === 0" class="log-placeholder">{{ t('console.logsEmpty') }}</div>
          <div v-else class="log-entry" v-for="(entry, idx) in parsedLogs" :key="idx">
            <div class="log-rail">
              <span class="log-dot" :class="entry.status"></span>
              <span v-if="idx < parsedLogs.length - 1" class="log-connector"></span>
            </div>
            <div class="log-card">
              <div class="log-meta">
                <span v-if="entry.time" class="log-time font-mono">{{ entry.time }}</span>
                <span v-if="entry.tag" class="log-tag" :class="entry.status">{{ entry.tag }}</span>
              </div>
              <div class="log-content font-mono">{{ entry.text }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, inject, h, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { getClawStatus, getClawConfig, stopClaw, restartClaw, uninstallClaw, getRecentLogs, getOpenClawConfigDirPath } from '@/api/deploy'
import { getModelsConfig } from '@/api/model'
import { getActiveSkillCount } from '@/api/skill'
import { listCronJobs } from '@/api/cron'
import { useRouter } from 'vue-router'
import { NTooltip, NButton, NDropdown } from 'naive-ui'
import gm from '@/utils/gmssh'

const router = useRouter()
const { t } = useI18n()
const setDeployed = inject('setDeployed', () => {})

const statusData = ref({
  running: false, containerName: 'gmssh-openclaw', status: 'unknown',
  webPort: 0, bridgePort: 0, uptime: '-', image: 'gmssh/openclaw:2026.02.17',
})

const configData = ref({
  provider: '', modelName: '', primaryModel: '', baseUrl: '', apiKeyMasked: '',
  contextWindow: 0, maxTokens: 0, gatewayPort: 0, authMode: '', gatewayBind: '',
  gatewayMode: '', containerCPU: '', containerMem: '',
})

const running = computed(() => statusData.value.running)
const statusLoaded = ref(false)
const skillCount = ref(0)
const skillLoaded = ref(false)
const jobCount = ref(0)
const jobLoaded = ref(false)

const metricsData = computed(() => [
  { key: 'skills', display: !statusLoaded.value ? '—' : !skillLoaded.value ? '...' : String(skillCount.value), label: t('console.metricSkills'), colorClass: 'clr-purple', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
  { key: 'crons', display: !statusLoaded.value ? '—' : !jobLoaded.value ? '...' : String(jobCount.value), label: t('console.metricCron'), colorClass: 'clr-cyan', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/><path d="M12 6v6l4 2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
])

const gatewayItems = computed(() => {
  const web = statusData.value.webPort || '—'
  const bridge = statusData.value.bridgePort || '—'
  return [
    { label: t('console.gwAuth'), value: configData.value.authMode || '-' },
    { label: t('console.gwBind'), value: configData.value.gatewayBind || '-' },
    { label: t('console.gwMode'), value: configData.value.gatewayMode || '-' },
    { label: t('console.gwStatus'), value: statusData.value.status || 'unknown' },
    { label: t('console.gwPort'), value: `${web}(web) / ${bridge}(bridge)` },
  ]
})

const webUIOptions = [
  {
    label: () => h('div', { style: 'display:flex;align-items:center;gap:8px' }, [
      h('svg', { viewBox: '0 0 16 16', width: 14, height: 14, innerHTML: '<circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.2" fill="none"/><path d="M1 8h14M8 1c2 2 3 4.5 3 7s-1 5-3 7M8 1c-2 2-3 4.5-3 7s1 5 3 7" stroke="currentColor" stroke-width="1.2" fill="none"/>' }),
      t('console.webUIPublic')
    ]),
    key: 'public'
  },
  {
    label: () => h('div', { style: 'display:flex;align-items:center;gap:8px' }, [
      h('svg', { viewBox: '0 0 16 16', width: 14, height: 14, innerHTML: '<rect x="2" y="4" width="12" height="8" rx="1.5" stroke="currentColor" stroke-width="1.2" fill="none"/><path d="M5 4V3a3 3 0 016 0v1" stroke="currentColor" stroke-width="1.2" fill="none" stroke-linecap="round"/><circle cx="8" cy="8.5" r="1.2" fill="currentColor"/>' }),
      t('console.webUIPrivate')
    ]),
    key: 'private'
  },
]

function getPrivateIp() { return window.$gm?.privateIp || 'localhost' }
function buildWebUrl(ip) {
  const port = statusData.value.webPort; const token = configData.value.gatewayToken || ''
  if (!port) return ''
  return token ? `http://${ip}:${port}/?token=${token}` : `http://${ip}:${port}`
}
function onWebUISelect(key) {
  const ip = key === 'private' ? getPrivateIp() : gm.getPublicIp()
  const url = buildWebUrl(ip)
  if (url) window.open(url, '_blank'); else gm.warning(t('console.webUIUnavailable'))
}

async function fetchAll() {
  skillLoaded.value = false; jobLoaded.value = false
  try {
    const [status, config] = await Promise.all([getClawStatus(), getClawConfig()])
    statusData.value = { ...statusData.value, ...status }
    configData.value = { ...configData.value, ...config }
    statusLoaded.value = true
    getActiveSkillCount().then(res => { skillCount.value = res?.count || 0 }).catch(console.error).finally(() => { skillLoaded.value = true })
    listCronJobs().then(res => { jobCount.value = res?.jobs?.length || 0 }).catch(console.error).finally(() => { jobLoaded.value = true })
    loadConfiguredModels()
    loadLogs()
  } catch (e) { console.error('获取数据失败:', e) }
}

function viewLogs() {
  const gmApi = gm.getGmApi(); const isLocal = configData.value.deployMode === 'local'
  if (isLocal) { if (gmApi?.openShell) gmApi.openShell({ arr: ['journalctl -u openclaw -f --no-pager\n'] }); else gm.info('请在终端执行: journalctl -u openclaw -f') }
  else { if (gmApi?.openShell) gmApi.openShell({ arr: ['docker logs -f gmssh-openclaw\n'] }); else gm.info('请在终端执行: docker logs -f gmssh-openclaw') }
}
function refreshStatus() { fetchAll(); gm.success(t('console.refreshSuccess')) }
async function openWorkDir() {
  try {
    const res = await getOpenClawConfigDirPath()
    const dir = res?.path
    if (!dir) { gm.warning('获取目录失败'); return }
    const gmApi = gm.getGmApi()
    if (gmApi?.openFolder) gmApi.openFolder(dir)
    else gm.info(t('console.btnWorkDir') + ': ' + dir)
  } catch (e) { gm.error('获取目录失败: ' + (e.message || '')) }
}
function openConfig() {
  const gmApi = gm.getGmApi(); const configFile = configData.value.configPath || '/opt/gmclaw/conf/openclaw.json'
  if (gmApi?.openCodeEditor) gmApi.openCodeEditor(configFile); else gm.info(`${t('console.configFile')}: ${configFile}`)
}

const actionLoading = ref(false)
async function doStop() {
  actionLoading.value = true
  try { await stopClaw(); gm.success(t('console.stopSuccess')); await fetchAll() } catch (e) { gm.error(t('console.stopFailed') + ': ' + e.message) } finally { actionLoading.value = false }
}
function handleStop() { window.$gm?.dialog?.warning({ title: t('console.stopTitle'), content: t('console.stopContent'), positiveText: t('console.stopConfirm'), negativeText: t('console.cancel'), maskClosable: false, onPositiveClick: doStop }) || doStop() }
async function handleRestart() {
  actionLoading.value = true
  try { await restartClaw(); gm.success(t('console.restartSuccess')); await fetchAll() } catch (e) { gm.error(t('console.restartFailed') + ': ' + e.message) } finally { actionLoading.value = false }
}
async function doUninstall() {
  actionLoading.value = true
  try { await uninstallClaw(); gm.success(t('console.uninstallSuccess')); setDeployed(false); sessionStorage.removeItem('deploy_token'); sessionStorage.removeItem('deploy_web_port'); sessionStorage.removeItem('deploy_setup_form'); router.replace('/console') }
  catch (e) { gm.error(t('console.uninstallFailed') + ': ' + e.message) } finally { actionLoading.value = false }
}
function handleUninstall() { window.$gm?.dialog?.warning({ title: t('console.uninstallTitle'), content: t('console.uninstallContent'), positiveText: t('console.uninstallConfirm'), negativeText: t('console.cancel'), maskClosable: false, onPositiveClick: () => doUninstall() }) || doUninstall() }

const configuredProviders = reactive({})
const configuredDefault = ref('')
const modelsLoaded = ref(false)
async function loadConfiguredModels() {
  try { const res = await getModelsConfig(); const providers = res?.models?.providers || {}; Object.keys(configuredProviders).forEach(k => delete configuredProviders[k]); Object.assign(configuredProviders, providers); configuredDefault.value = res?.defaultModel || '' } catch {}
  modelsLoaded.value = true
}
onMounted(fetchAll)

// ====== 日志 ======
const logLines = ref([])
const logsLoaded = ref(false)
const logsLoading = ref(false)
const logBody = ref(null)

async function loadLogs() {
  logsLoading.value = true
  try {
    const res = await getRecentLogs(10, configData.value.deployMode || '')
    logLines.value = res?.logs || []
  } catch (e) { console.error('[loadLogs]', e); logLines.value = [] }
  logsLoaded.value = true
  logsLoading.value = false
  nextTick(() => { if (logBody.value) logBody.value.scrollTop = logBody.value.scrollHeight })
}

// 解析日志行 → { time, tag, status, text }
const parsedLogs = computed(() => {
  return logLines.value.map(line => {
    let time = '', tag = '', text = line, status = 'neutral'
    // 提取时间戳（ISO 格式，支持 UTC Z 和时区偏移）
    const timeMatch = line.match(/^(\d{4}-\d{2}-\d{2}T[\d:.]+(?:Z|[+-]\d{2,4})?)\s*/)
    if (timeMatch) {
      try {
        const d = new Date(timeMatch[1])
        time = d.toLocaleTimeString('zh-CN', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
      } catch { time = timeMatch[1].slice(11, 19) }
      text = line.slice(timeMatch[0].length).replace(/^\S+\s+/, '').trim()
    }
    // 提取 [tag] 标记
    const tagMatch = text.match(/^\[([^\]]+)\]\s*/)
    if (tagMatch) {
      tag = tagMatch[1]
      text = text.slice(tagMatch[0].length)
    }
    // 状态探测
    const lower = line.toLowerCase()
    if (/✅|success|started|\bconnected\b|\bactive\b|完成|已启动|已连接/.test(lower)) status = 'ok'
    else if (/❌|error|fail|\bdisconnect|exception|失败|异常/.test(lower)) status = 'err'
    else if (/warn|警告|⚠/.test(lower)) status = 'warn'
    return { time, tag, status, text }
  })
})
</script>

<style scoped>
.dashboard-panel {
  padding: 28px 36px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  height: 100%;
}

/* ===== 顶部条 ===== */
.top-bar { display: flex; align-items: center; justify-content: space-between; gap: 16px; }

/* 状态胶囊 — Mesh Gradient 亚克力 */
.status-pill {
  display: inline-flex; align-items: center; gap: 8px;
  padding: 7px 16px; border-radius: 24px;
  background:
    radial-gradient(ellipse at 20% 50%, rgba(var(--jm-primary-1-rgb), 0.06) 0%, transparent 60%),
    radial-gradient(ellipse at 80% 50%, rgba(162, 120, 255, 0.04) 0%, transparent 60%),
    var(--jm-glass-bg);
  backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--jm-glass-border);
  box-shadow: var(--jm-glass-inner-glow), 0 2px 12px rgba(0,0,0,0.1);
  font-size: 12px; color: var(--jm-accent-5); white-space: nowrap;
}

.st-dot {
  width: 8px; height: 8px; border-radius: 50%;
  position: relative; display: inline-flex; flex-shrink: 0;
}
.st-ok .st-dot {
  background: #3cff7a;
  box-shadow: 0 0 8px #3cff7a, 0 0 16px rgba(60, 255, 122, 0.3);
}
.st-off .st-dot { background: var(--jm-error-color); box-shadow: 0 0 6px rgba(240, 77, 60, 0.4); }
.st-loading .st-dot { background: var(--jm-accent-4); animation: pulse 1.2s ease-in-out infinite; }
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }

.st-ring {
  position: absolute; inset: -4px; border-radius: 50%;
  border: 1.5px solid rgba(60, 255, 122, 0.5);
  animation: ring-out 2.5s ease-out infinite; opacity: 0;
}
@keyframes ring-out { 0% { transform: scale(0.8); opacity: 0.6; } 100% { transform: scale(2.5); opacity: 0; } }

.st-text { font-weight: 600; color: var(--jm-accent-7); }
.st-name { color: var(--jm-accent-4); font-size: 11px; }
.st-sep { color: var(--jm-accent-3); font-size: 10px; }
.st-meta { color: var(--jm-accent-4); font-size: 11px; }

.top-actions { display: flex; align-items: center; gap: 8px; }

/* 胶囊工具栏 */
.capsule-bar {
  display: flex; align-items: center; gap: 1px;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--jm-glass-border);
  border-radius: 10px;
  padding: 2px;
  box-shadow: var(--jm-glass-inner-glow);
}
.capsule-danger { border-color: rgba(var(--jm-accent-2-rgb), 0.2); }

.cap-btn {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 5px 10px; border-radius: 8px;
  border: none; background: transparent;
  color: var(--jm-accent-5); cursor: pointer;
  font-size: 11px; font-weight: 500;
  transition: all 0.2s; white-space: nowrap;
}
.cap-btn:hover { background: var(--jm-glass-bg-hover); color: var(--jm-accent-7); }
.cap-btn.cap-icon { padding: 5px; }
.cap-btn.primary { color: var(--jm-primary-2); }
.cap-btn.primary:hover { background: rgba(var(--jm-primary-1-rgb), 0.1); }
.cap-btn:disabled { opacity: 0.3; cursor: not-allowed; pointer-events: none; }

/* 重启 — 琥珀警告色 */
.cap-btn.ctrl-warn:hover {
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.08);
}
/* 卸载 — 危险红色 */
.cap-btn.ctrl-danger-btn:hover {
  color: var(--jm-error-color);
  background: rgba(220, 38, 38, 0.06);
}

/* ===== Bento Grid — 非对称 ===== */
.bento-grid {
  flex: 1; min-height: 0;
  display: grid;
  grid-template-columns: 1.618fr 1fr;
  grid-template-rows: auto auto 1fr;
  gap: 14px;
}

/* A区：模型矩阵 — 跨越右侧全部行高 */
.bento-models {
  grid-row: 1 / -1;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(var(--jm-glass-blur)); -webkit-backdrop-filter: blur(var(--jm-glass-blur));
  border: 1px solid var(--jm-glass-border);
  border-radius: 16px;
  padding: 22px 24px;
  display: flex; flex-direction: column;
  box-shadow:
    var(--jm-glass-inner-glow),
    0 4px 24px rgba(0, 0, 0, 0.15),
    0 -1px 0 0 rgba(255, 255, 255, 0.03);
  transition: box-shadow 0.3s;
}
.bento-models:hover {
  box-shadow:
    var(--jm-glass-inner-glow),
    0 8px 40px rgba(0, 0, 0, 0.2),
    0 0 20px rgba(var(--jm-primary-1-rgb), 0.06);
}

/* B区：数据指标 */
.bento-metrics {
  display: flex; gap: 8px;
}

/* C区：网关信息 */
.bento-gateway {
  background: var(--jm-glass-bg);
  backdrop-filter: blur(var(--jm-glass-blur)); -webkit-backdrop-filter: blur(var(--jm-glass-blur));
  border: 1px solid var(--jm-glass-border);
  border-radius: 14px;
  padding: 16px 18px;
  box-shadow:
    var(--jm-glass-inner-glow),
    0 2px 12px rgba(0, 0, 0, 0.1),
    0 -1px 0 0 rgba(255, 255, 255, 0.03);
}

/* D区：端口 */
.bento-ports {
  background: var(--jm-glass-bg);
  backdrop-filter: blur(var(--jm-glass-blur)); -webkit-backdrop-filter: blur(var(--jm-glass-blur));
  border: 1px solid var(--jm-glass-border);
  border-radius: 14px;
  padding: 14px 18px;
  display: flex; align-items: center;
  box-shadow:
    var(--jm-glass-inner-glow),
    0 2px 12px rgba(0, 0, 0, 0.1),
    0 -1px 0 0 rgba(255, 255, 255, 0.03);
}

/* 日志区 */
.bento-logs {
  padding: 14px 16px;
  display: flex; flex-direction: column;
  min-height: 0;
}

/* ===== 区块标题 ===== */
.section-hd {
  display: flex; align-items: center; gap: 6px;
  font-size: 10px; font-weight: 600; color: var(--jm-accent-4);
  margin-bottom: 14px; letter-spacing: 0.1em; text-transform: uppercase;
}

/* ===== 模型区 ===== */
.models-body { flex: 1; display: flex; flex-direction: column; gap: 14px; }
.models-placeholder { font-size: 12px; color: var(--jm-accent-4); }

.provider-row { display: flex; align-items: flex-start; gap: 10px; flex-wrap: wrap; }

.prov-badge {
  font-size: 12px; font-weight: 700; letter-spacing: 0.02em;
  padding: 5px 14px; border-radius: 20px;
  background:
    linear-gradient(135deg, rgba(var(--jm-primary-1-rgb), 0.12), rgba(var(--jm-primary-1-rgb), 0.06));
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.18);
  color: var(--jm-primary-2);
  text-transform: capitalize;
  box-shadow: 0 0 12px rgba(var(--jm-primary-1-rgb), 0.08), inset 0 1px 0 rgba(255,255,255,0.06);
  flex-shrink: 0;
  transition: all 0.2s;
}
.prov-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 16px rgba(var(--jm-primary-1-rgb), 0.15), inset 0 1px 0 rgba(255,255,255,0.08);
}

.model-pills { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }

/* 3D 厚度模型标签 */
.model-pill {
  font-size: 11px; padding: 4px 14px; border-radius: 18px;
  background: rgba(var(--jm-accent-2-rgb), 0.15);
  border: 1px solid var(--jm-glass-border);
  color: var(--jm-accent-6);
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  cursor: default;
  /* Soft UI 双阴影 */
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.15),
    0 -1px 0 rgba(255, 255, 255, 0.04);
}
.model-pill:hover {
  transform: translateY(-2px);
  background: rgba(var(--jm-primary-1-rgb), 0.08);
  border-color: rgba(var(--jm-primary-1-rgb), 0.2);
  color: var(--jm-primary-2);
  box-shadow:
    0 4px 12px rgba(var(--jm-primary-1-rgb), 0.12),
    0 -1px 0 rgba(255, 255, 255, 0.06);
}
/* 当前默认模型高亮 */
.model-pill.is-default {
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  border-color: rgba(var(--jm-primary-1-rgb), 0.25);
  color: var(--jm-primary-2);
  font-weight: 600;
  box-shadow:
    0 2px 8px rgba(var(--jm-primary-1-rgb), 0.15),
    0 -1px 0 rgba(255, 255, 255, 0.05),
    0 0 0 1px rgba(var(--jm-primary-1-rgb), 0.08);
}

/* 默认模型浮动条 */
.default-bar {
  margin-top: auto;
  padding-top: 14px;
  border-top: 1px solid var(--jm-glass-border);
  display: flex; align-items: center; gap: 8px;
}
.default-tag {
  font-size: 9px; font-weight: 700; letter-spacing: 0.12em;
  padding: 2px 8px; border-radius: 6px;
  background: rgba(var(--jm-primary-1-rgb), 0.08);
  color: var(--jm-primary-1);
}
.default-val { font-size: 13px; font-weight: 600; color: var(--jm-accent-7); }

/* ===== 指标 — 悬浮大数字 ===== */
.metric-float {
  flex: 1;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 4px;
  padding: 14px 8px;
  border-radius: 14px;
  /* Soft UI — 凹陷嵌入感 */
  background: rgba(var(--jm-accent-1-rgb), 0.15);
  box-shadow:
    inset 0 2px 4px rgba(0, 0, 0, 0.15),
    inset 0 -1px 0 rgba(255, 255, 255, 0.04),
    0 1px 2px rgba(0, 0, 0, 0.05);
  transition: all 0.2s;
}
.metric-float:hover {
  background: rgba(var(--jm-accent-1-rgb), 0.25);
}

.mf-icon {
  display: inline-flex; align-items: center; justify-content: center;
  vertical-align: middle; opacity: 0.7; margin-left: 2px;
}
.mf-icon.clr-purple { color: #a278ff; }
.mf-icon.clr-cyan { color: #63b4ff; }

.mf-num {
  font-size: 26px; font-weight: 800; line-height: 1;
  letter-spacing: -0.04em;
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  background-clip: text;
}
.mf-num.clr-purple { background-image: linear-gradient(135deg, #a278ff, #d0b4ff); }
.mf-num.clr-cyan   { background-image: linear-gradient(135deg, #63b4ff, #a0d4ff); }

.mf-label { font-size: 10px; color: var(--jm-accent-4); letter-spacing: 0.06em; }

/* ===== 网关列表 ===== */
.gw-list { display: flex; flex-direction: column; }
.gw-row {
  display: flex; justify-content: space-between; align-items: center;
  padding: 7px 0; border-bottom: 1px solid rgba(var(--jm-accent-2-rgb), 0.12);
}
.gw-row:last-child { border-bottom: none; }
.gw-k { font-size: 11px; color: var(--jm-accent-4); }
.gw-v { font-size: 12px; font-weight: 700; color: var(--jm-accent-7); }

/* ===== 拓扑图 ===== */
.topo-mini {
  margin-top: 10px; padding: 6px 4px;
  border-radius: 10px;
  background: rgba(var(--jm-accent-1-rgb), 0.08);
  border: 1px solid rgba(var(--jm-accent-2-rgb), 0.1);
}

/* ===== 端口 — LED 数码管 ===== */
.led-port { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 3px; }
.led-sep { width: 1px; height: 30px; background: var(--jm-glass-border); margin: 0 10px; }

.led-label {
  font-size: 9px; font-weight: 600; letter-spacing: 0.15em;
  color: var(--jm-accent-4); text-transform: uppercase;
}
.led-num {
  font-size: 22px; font-weight: 800; letter-spacing: -0.02em;
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  background-clip: text;
  background-image: linear-gradient(180deg, var(--jm-primary-2), var(--jm-primary-1));
  /* LED 发光 */
  filter: drop-shadow(0 0 6px rgba(var(--jm-primary-1-rgb), 0.3));
}

/* ===== 日志面板 — 时间线风格 ===== */
.bento-logs .section-hd { margin-bottom: 10px; }
.log-refresh {
  margin-left: auto; width: 22px; height: 22px; border-radius: 6px;
  border: none; background: transparent; color: var(--jm-accent-4);
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.2s;
}
.log-refresh:hover { color: var(--jm-accent-6); background: var(--jm-glass-bg-hover); }
.log-refresh:disabled { opacity: 0.3; cursor: not-allowed; }
.log-refresh .spinning { animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.log-timeline {
  overflow-y: auto; flex: 1; min-height: 0;
  padding: 2px 0;
}
.log-timeline::-webkit-scrollbar { width: 4px; }
.log-timeline::-webkit-scrollbar-track { background: transparent; }
.log-timeline::-webkit-scrollbar-thumb { background: var(--jm-accent-3); border-radius: 2px; }

.log-placeholder { font-size: 11px; color: var(--jm-accent-4); padding: 12px 0; }

/* Entry = Rail + Card */
.log-entry {
  display: flex; gap: 10px;
  min-height: 0;
}

/* 左侧时间线轨道 */
.log-rail {
  display: flex; flex-direction: column; align-items: center;
  width: 10px; flex-shrink: 0; padding-top: 6px;
}
.log-dot {
  width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0;
  background: var(--jm-accent-3);
  box-shadow: 0 0 0 2px rgba(var(--jm-accent-2-rgb), 0.2);
  transition: all 0.3s;
}
.log-dot.ok {
  background: #22c55e;
  box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.15), 0 0 6px rgba(34, 197, 94, 0.3);
}
.log-dot.err {
  background: #ef4444;
  box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.15), 0 0 6px rgba(239, 68, 68, 0.3);
}
.log-dot.warn {
  background: #f59e0b;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.15), 0 0 6px rgba(245, 158, 11, 0.3);
}
.log-connector {
  flex: 1; width: 1px; min-height: 8px;
  background: linear-gradient(180deg, var(--jm-accent-3), transparent);
}

/* 右侧卡片 */
.log-card {
  flex: 1; min-width: 0;
  padding: 5px 10px 7px;
  border-radius: 8px;
  background: rgba(var(--jm-accent-1-rgb), 0.06);
  border: 1px solid rgba(var(--jm-accent-2-rgb), 0.08);
  margin-bottom: 4px;
  transition: all 0.2s;
}
.log-card:hover {
  background: rgba(var(--jm-primary-1-rgb), 0.04);
  border-color: rgba(var(--jm-primary-1-rgb), 0.12);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}

.log-meta {
  display: flex; align-items: center; gap: 6px; margin-bottom: 2px;
}
.log-time {
  font-size: 9px; color: var(--jm-accent-4); letter-spacing: 0.03em;
}
.log-tag {
  font-size: 8px; font-weight: 600; letter-spacing: 0.05em;
  padding: 1px 5px; border-radius: 4px;
  background: rgba(var(--jm-accent-1-rgb), 0.15);
  color: var(--jm-accent-4); text-transform: uppercase;
}
.log-tag.ok { color: #22c55e; background: rgba(34, 197, 94, 0.08); }
.log-tag.err { color: #ef4444; background: rgba(239, 68, 68, 0.08); }
.log-tag.warn { color: #f59e0b; background: rgba(245, 158, 11, 0.08); }

.log-content {
  font-size: 10px; line-height: 1.5;
  color: var(--jm-accent-5);
  word-break: break-all; white-space: pre-wrap;
}

/* ===== 响应式 ===== */
@media (max-width: 680px) {
  .bento-grid { grid-template-columns: 1fr; grid-template-rows: auto; }
  .bento-models { grid-row: auto; }
  .bento-metrics { flex-wrap: wrap; }
}
</style>
