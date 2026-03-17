<template>
  <div class="cron-page">
    <div class="cron-container fade-in-up">
      <!-- 顶部 -->
      <div class="cron-header">
        <div class="header-left">
          <h2 class="page-title">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
              <polyline points="12,6 12,12 16,14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            {{ t('cron.title') }}
          </h2>
          <span class="header-hint">{{ t('cron.subtitle') }}</span>
        </div>
        <div class="header-actions">
          <button class="refresh-btn" @click="fetchJobs()" :disabled="loading" title="刷新">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/></svg>
          </button>
          <button class="add-btn" @click="openAddForm">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
            <line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
          {{ t('cron.addJob') }}
        </button>
        </div>
      </div>

      <!-- 任务列表 -->
      <div v-if="loading" class="loading-state"><div class="loading-spinner"></div></div>
      <div v-else-if="jobs.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" width="40" height="40" fill="none" opacity="0.3">
          <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
          <polyline points="12,6 12,12 16,14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <p>{{ t('cron.noJobs') }}</p>
        <button class="add-btn small" @click="openAddForm">{{ t('cron.createFirst') }}</button>
      </div>
      <div v-else class="job-list">
        <div v-for="job in jobs" :key="job.id" class="job-card" :class="{ disabled: !job.enabled }">
          <!-- 顶行：名称 + 开关 -->
          <div class="job-top-row">
            <div class="job-title-area">
              <span class="job-name">{{ job.name }}</span>
              <span v-if="job.sessionTarget" class="session-badge" :class="job.sessionTarget">{{ job.sessionTarget === 'isolated' ? t('cron.sessionIsolated') : t('cron.sessionMain') }}</span>
            </div>
            <n-switch size="small" :value="job.enabled" :loading="togglingId === job.id" @update:value="v => toggleJob(job, v)" />
          </div>

          <!-- 调度信息 -->
          <div class="job-schedule-line font-mono">
            <span class="schedule-badge">{{ formatSchedule(job.schedule) }}</span>
          </div>
          <div class="job-time-row" v-if="job.state?.nextRunAtMs || job.state?.lastRunAtMs">
            <span v-if="job.state?.nextRunAtMs" class="next-run">
              Next: {{ formatTime(job.state.nextRunAtMs) }}
            </span>
            <span v-if="job.state?.lastRunAtMs" class="last-run">Last: {{ formatTime(job.state.lastRunAtMs) }}</span>
          </div>

          <!-- 消息/描述 -->
          <div class="job-meta" v-if="job.payload?.message || job.payload?.text || job.description">
            <span v-if="job.payload?.message" class="job-message">{{ truncate(job.payload.message, 50) }}</span>
            <span v-if="job.payload?.text" class="job-message">{{ truncate(job.payload.text, 50) }}</span>
            <span v-if="job.description" class="job-desc">{{ job.description }}</span>
          </div>

          <!-- 操作按钮 — 悬浮显现 -->
          <div class="job-actions-capsule">
            <n-tooltip trigger="hover"><template #trigger><button class="icon-btn" @click="doRunJob(job.id)" :disabled="runningId === job.id">
              <svg v-if="runningId !== job.id" viewBox="0 0 24 24" width="13" height="13" fill="none"><polygon points="5,3 19,12 5,21" fill="currentColor"/></svg>
              <div v-else class="mini-spinner"></div>
            </button></template>{{ t('cron.tooltipRun') }}</n-tooltip>
            <n-tooltip trigger="hover"><template #trigger><button class="icon-btn" @click="openEditForm(job)">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button></template>{{ t('cron.tooltipEdit') }}</n-tooltip>
            <n-tooltip trigger="hover"><template #trigger><button class="icon-btn" @click="viewRuns(job)">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><path d="M12 8v4l3 3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/></svg>
            </button></template>{{ t('cron.tooltipHistory') }}</n-tooltip>
            <n-tooltip trigger="hover"><template #trigger><button class="icon-btn danger" @click="doRemoveJob(job)" :disabled="deletingId === job.id">
              <div v-if="deletingId === job.id" class="mini-spinner"></div>
              <svg v-else viewBox="0 0 24 24" width="13" height="13" fill="none"><polyline points="3 6 5 6 21 6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button></template>{{ t('cron.tooltipDelete') }}</n-tooltip>
          </div>
        </div>
      </div>

      <!-- 新增/编辑弹窗 -->
      <n-modal v-model:show="showForm" preset="card" :title="editingJob ? t('cron.editJobTitle') : t('cron.addJobTitle')" style="max-width: 560px;" :bordered="false" :mask-closable="false" :segmented="{ footer: 'soft' }">
        <div class="form-body">
          <!-- 任务名称 -->
          <div class="form-group">
            <label class="form-label">{{ t('cron.fieldName') }} <span class="required">*</span></label>
            <n-input v-model:value="form.name" :placeholder="t('cron.fieldNamePlaceholder')" size="small" />
          </div>

          <!-- 描述 -->
          <div class="form-group">
            <label class="form-label">{{ t('cron.fieldDesc') }}</label>
            <n-input v-model:value="form.desc" :placeholder="t('cron.fieldDescPlaceholder')" size="small" />
          </div>

          <!-- 请求方式 -->
          <div class="form-group">
            <label class="form-label">{{ t('cron.fieldPayload') }}</label>
            <n-radio-group v-model:value="form.payloadKind" size="small" class="schedule-tabs" :theme-overrides="segmentTheme">
              <n-radio-button value="text">{{ t('cron.payloadText') }}</n-radio-button>
              <n-radio-button value="event">Agent</n-radio-button>
            </n-radio-group>
          </div>

          <!-- 对话内容（两种模式均显示） -->
          <div class="form-group">
            <n-input
              v-model:value="form.message"
              type="textarea"
              :rows="3"
              :placeholder="contentPlaceholder"
              size="small"
            />
          </div>

          <!-- 调度方式 -->
          <div class="form-group">
            <label class="form-label">{{ t('cron.fieldSchedule') }}</label>
            <n-radio-group v-model:value="form.scheduleKind" size="small" class="schedule-tabs" :theme-overrides="segmentTheme">
              <n-radio-button value="cron">{{ t('cron.schedCron') }}</n-radio-button>
              <n-radio-button value="every">{{ t('cron.schedEvery') }}</n-radio-button>
              <n-radio-button value="at">{{ t('cron.schedAt') }}</n-radio-button>
            </n-radio-group>
          </div>

          <!-- Cron 表达式 -->
          <div v-if="form.scheduleKind === 'cron'" class="form-group schedule-detail">
            <n-input v-model:value="form.cron" placeholder="0 7 * * *  (分 时 日 月 周)" size="small" />
            <div class="cron-hints">
              <span class="hint-chip" @click="form.cron = '0 * * * *'">{{ t('cron.hintEveryHour') }}</span>
              <span class="hint-chip" @click="form.cron = '0 7 * * *'">{{ t('cron.hintEveryDay7') }}</span>
              <span class="hint-chip" @click="form.cron = '0 9 * * 1'">{{ t('cron.hintEveryMon9') }}</span>
              <span class="hint-chip" @click="form.cron = '0 0 1 * *'">{{ t('cron.hintEveryMonth1') }}</span>
            </div>
          </div>

          <!-- 固定间隔 -->
          <div v-if="form.scheduleKind === 'every'" class="form-group schedule-detail">
            <div class="interval-row">
              <span class="interval-label">{{ t('cron.everyPrefix') }}</span>
              <n-input-number v-model:value="form.everyValue" :min="1" placeholder="10" size="small" style="width: 100px" />
              <n-select v-model:value="form.everyUnit" :options="intervalUnits" size="small" style="width: 90px" />
              <span v-if="t('cron.everySuffix')" class="interval-label">{{ t('cron.everySuffix') }}</span>
            </div>
          </div>

          <!-- 一次性 -->
          <div v-if="form.scheduleKind === 'at'" class="form-group schedule-detail">
            <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
              <n-date-picker v-model:value="form.atTimestamp" type="datetime" :placeholder="t('cron.schedAtPlaceholder')" size="small" style="width:100%" format="yyyy-MM-dd HH:mm" />
            </n-config-provider>
          </div>

          <!-- 高级设置折叠区 -->
          <div class="advanced-section">
            <div class="advanced-toggle" @click="showAdvanced = !showAdvanced">
              <svg viewBox="0 0 24 24" width="12" height="12" fill="none" :style="{ transform: showAdvanced ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }">
                <polyline points="9,6 15,12 9,18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <span>{{ t('cron.advancedSettings') }}</span>
              <span class="advanced-summary" v-if="!showAdvanced">{{ advancedSummary }}</span>
            </div>
            <div v-show="showAdvanced" class="advanced-body">
              <!-- Session 会话模式 -->
              <div class="form-group">
                <label class="form-label">{{ t('cron.fieldSession') }}</label>
                <div class="radio-desc-group">
                  <label class="radio-desc-item" :class="{ active: form.session === 'isolated' }" @click="form.session = 'isolated'">
                    <span class="radio-dot" :class="{ checked: form.session === 'isolated' }"></span>
                    <div>
                       <span class="rdi-title">{{ t('cron.sessionIsolatedTitle') }} <span class="rdi-badge recommended">{{ t('cron.sessionIsolatedRecommended') }}</span></span>
                       <span class="rdi-desc">{{ t('cron.sessionIsolatedDesc') }}</span>
                    </div>
                  </label>
                  <label class="radio-desc-item" :class="{ active: form.session === 'main' }" @click="form.session = 'main'">
                    <span class="radio-dot" :class="{ checked: form.session === 'main' }"></span>
                    <div>
                       <span class="rdi-title">{{ t('cron.sessionMainTitle') }}</span>
                       <span class="rdi-desc">{{ t('cron.sessionMainDesc') }}</span>
                    </div>
                  </label>
                </div>
              </div>

              <!-- Wake Mode 唤醒模式 -->
              <div class="form-group">
                <label class="form-label">{{ t('cron.fieldWakeMode') }}</label>
                <div class="radio-desc-group">
                  <label class="radio-desc-item" :class="{ active: form.wakeMode === 'heartbeat' }" @click="form.wakeMode = 'heartbeat'">
                    <span class="radio-dot" :class="{ checked: form.wakeMode === 'heartbeat' }"></span>
                    <div>
                       <span class="rdi-title">{{ t('cron.wakeHeartbeatTitle') }} <span class="rdi-badge recommended">{{ t('cron.wakeHeartbeatRecommended') }}</span></span>
                       <span class="rdi-desc">{{ t('cron.wakeHeartbeatDesc') }}</span>
                    </div>
                  </label>
                  <label class="radio-desc-item" :class="{ active: form.wakeMode === 'now' }" @click="form.wakeMode = 'now'">
                    <span class="radio-dot" :class="{ checked: form.wakeMode === 'now' }"></span>
                    <div>
                       <span class="rdi-title">{{ t('cron.wakeNowTitle') }}</span>
                       <span class="rdi-desc">{{ t('cron.wakeNowDesc') }}</span>
                    </div>
                  </label>
                </div>
              </div>


              <div class="form-group">
                <label class="form-label">{{ t('cron.fieldNotify') }}</label>
                <n-radio-group v-model:value="form.deliveryMode" size="small" :theme-overrides="segmentTheme">
                  <n-radio-button value="none">{{ t('cron.notifyNone') }}</n-radio-button>
                  <n-radio-button value="webhook">Webhook</n-radio-button>
                </n-radio-group>
              </div>
              <div v-if="form.deliveryMode === 'webhook'" class="form-group">
                <n-input v-model:value="form.deliveryTo" placeholder="https://your-webhook.example.com/hook" size="small" />
              </div>
            </div>
          </div>
        </div>
        <template #footer>
          <div class="form-footer">
            <n-button @click="showForm = false" size="small" quaternary>{{ t('cron.cancel') }}</n-button>
            <n-button type="primary" @click="submitForm" :loading="submitting" size="small">{{ editingJob ? t('cron.saveEdit') : t('cron.createJob') }}</n-button>
          </div>
        </template>
      </n-modal>

      <!-- 运行历史弹窗 -->
      <n-modal v-model:show="showRuns" preset="card" :title="t('cron.historyTitle') + ' — ' + runsJobName" style="max-width: 580px;" :bordered="false">
        <div v-if="loadingRuns" class="loading-state"><div class="loading-spinner"></div></div>
        <div v-else-if="!runs.length" class="empty-hint">{{ t('cron.noHistory') }}</div>
        <div v-else class="runs-list">
          <div
            v-for="(run, i) in runs"
            :key="i"
            class="run-row"
            :class="[run.status, { expanded: expandedRuns.has(i) }]"
            @click="toggleRunExpand(i)"
          >
            <div class="run-status-bar"></div>
            <div class="run-content">
              <div class="run-top">
                <span class="run-status-badge" :class="run.status">{{ run.status || '—' }}</span>
                <span class="run-summary">{{ run.summary || run.jobId || '—' }}</span>
              </div>
              <div class="run-bottom">
                <span class="run-index">#{{ runs.length - i }}</span>
                <span class="run-time">{{ formatTime(run.runAtMs || run.ts) }}</span>
                <span v-if="run.durationMs" class="run-duration">⏱ {{ formatDuration(run.durationMs) }}</span>
              </div>
            </div>
          </div>
        </div>
      </n-modal>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NSwitch, NTooltip, NModal, NInput, NInputNumber, NButton, NRadioGroup, NRadioButton, NSelect, NDatePicker, NConfigProvider } from 'naive-ui'
import { zhCN, dateZhCN } from 'naive-ui'
import {
  listCronJobs, addCronJob, editCronJob, removeCronJob,
  enableCronJob, disableCronJob, runCronJob, getCronRuns
} from '@/api/cron'
import cache from '@/stores/cache'

const gm = window.$gm || {}
const { t } = useI18n()

// ===== Segmented Control 主题覆盖（去掉所有 Naive UI 边框） =====
const segmentTheme = {
  buttonBorderColor: 'transparent',
  buttonBorderColorActive: 'transparent',
  buttonBorderColorHover: 'transparent',
  buttonBoxShadow: 'none',
  buttonBoxShadowFocus: 'none',
  buttonBoxShadowHover: 'none',
}

// ========== 任务列表 ==========
const jobs = ref([])
const loading = ref(false)

async function fetchJobs() {
  loading.value = true
  try {
    const res = await listCronJobs()
    jobs.value = res?.jobs || []
    cache.cronJobs = [...jobs.value]
  } catch (e) {
    gm.message?.error?.(t('cron.fetchFailed') + ': ' + (e.message || ''))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (cache.cronJobs !== null) {
    jobs.value = [...cache.cronJobs]
    loading.value = false
    return
  }
  fetchJobs()
})

// ========== 启用/禁用 ==========
const togglingId = ref('')

async function toggleJob(job, enabled) {
  togglingId.value = job.id
  try {
    if (enabled) {
      await enableCronJob({ jobId: job.id })
    } else {
      await disableCronJob({ jobId: job.id })
    }
    await fetchJobs()
  } catch (e) {
    gm.message?.error?.(t('cron.toggleFailed') + ': ' + (e.message || ''))
  } finally {
    togglingId.value = ''
  }
}

// ========== 手动执行 ==========
const runningId = ref('')

async function doRunJob(jobId) {
  runningId.value = jobId
  try {
    await runCronJob({ jobId })
    gm.message?.success?.(t('cron.runSuccess'))
    await fetchJobs()
  } catch (e) {
    gm.message?.error?.(t('cron.runFailed') + ': ' + (e.message || ''))
  } finally {
    runningId.value = ''
  }
}

// ========== 删除 ==========
const deletingId = ref('')

async function doRemoveJob(job) {
  if (deletingId.value) return
  deletingId.value = job.id
  try {
    await removeCronJob({ jobId: job.id })
    gm.message?.success?.(`${job.name} ${t('cron.deleteSuccess')}`)
    await fetchJobs()
  } catch (e) {
    gm.message?.error?.(t('cron.deleteFailed') + ': ' + (e.message || ''))
  } finally {
    deletingId.value = ''
  }
}

// ========== 新增/编辑表单 ==========
const showForm = ref(false)
const editingJob = ref(null)
const submitting = ref(false)

const intervalUnits = computed(() => [
  { label: t('cron.unitMinute'), value: 'm' },
  { label: t('cron.unitHour'), value: 'h' },
  { label: t('cron.unitDay'), value: 'd' },
])

const defaultForm = () => ({
  name: '',
  desc: '',
  scheduleKind: 'cron',
  cron: '',
  everyValue: 10,
  everyUnit: 'm',
  atTimestamp: null,
  session: 'isolated',
  wakeMode: 'heartbeat',
  payloadKind: 'text',
  message: '',
  eventJson: '',
  deliveryMode: 'none',
  deliveryTo: '',
})

const form = ref(defaultForm())
const showAdvanced = ref(false)

const advancedSummary = computed(() => {
  const f = form.value
  const parts = []
  parts.push(f.session === 'main' ? 'Main' : 'Isolated')
  parts.push(f.wakeMode === 'now' ? 'Now' : 'Next heartbeat')
  if (f.deliveryMode === 'webhook') parts.push('Webhook')
  return parts.join(' · ')
})

const contentPlaceholder = computed(() =>
  form.value.payloadKind === 'event'
    ? '{"event": "check_news", "topic": "AI"}'
    : t('cron.contentPlaceholder')
)

function openAddForm() {
  editingJob.value = null
  form.value = defaultForm()
  showForm.value = true
}

function openEditForm(job) {
  editingJob.value = job
  // 解析 everyMs 为 value+unit
  let everyValue = 10, everyUnit = 'm'
  if (job.schedule?.kind === 'every' && job.schedule?.everyMs) {
    const ms = job.schedule.everyMs
    if (ms >= 86400000) { everyValue = ms / 86400000; everyUnit = 'd' }
    else if (ms >= 3600000) { everyValue = ms / 3600000; everyUnit = 'h' }
    else { everyValue = ms / 60000; everyUnit = 'm' }
  }

  // 判断 payload 类型
  const isEvent = job.payload?.event !== undefined
  form.value = {
    name: job.name || '',
    desc: job.description || '',
    scheduleKind: job.schedule?.kind || 'cron',
    cron: job.schedule?.kind === 'cron' ? (job.schedule?.expr || '') : '',
    everyValue,
    everyUnit,
    atTimestamp: job.schedule?.atMs || null,
    session: job.sessionTarget || 'isolated',
    wakeMode: job.wakeMode || 'heartbeat',
    payloadKind: isEvent ? 'event' : 'text',
    message: isEvent ? '' : (job.payload?.message || job.payload?.text || ''),
    eventJson: isEvent ? JSON.stringify(job.payload, null, 2) : '',
    deliveryMode: job.delivery?.mode || 'none',
    deliveryTo: job.delivery?.to || '',
  }
  showForm.value = true
}

async function submitForm() {
  if (!form.value.name) { gm.message?.warning?.(t('cron.validateName')); return }
  submitting.value = true
  try {
    const f = form.value
    const params = {
      name: f.name,
      scheduleKind: f.scheduleKind,
      session: f.session,
      wakeMode: f.wakeMode,
      deliveryMode: f.deliveryMode,
      deliveryTo: f.deliveryTo,
    }
    if (f.payloadKind === 'event') {
      try { params.payload = JSON.parse(f.eventJson) } catch { gm.message?.warning?.(t('cron.invalidJson')); submitting.value = false; return }
    } else {
      params.message = f.message
    }

    // 调度参数
    if (f.scheduleKind === 'cron') {
      params.cron = f.cron
    } else if (f.scheduleKind === 'every') {
      params.every = `${f.everyValue}${f.everyUnit}`
    } else if (f.scheduleKind === 'at') {
      if (f.atTimestamp) {
        params.at = new Date(f.atTimestamp).toISOString()
      }
    }

    if (editingJob.value) {
      params.jobId = editingJob.value.id
      await editCronJob(params)
      gm.message?.success?.(t('cron.updateSuccess'))
    } else {
      await addCronJob(params)
      gm.message?.success?.(t('cron.createSuccess'))
    }
    showForm.value = false
    await fetchJobs()
  } catch (e) {
    gm.message?.error?.((editingJob.value ? t('cron.editFailed') : t('cron.saveFailed')) + ': ' + (e.message || ''))
  } finally {
    submitting.value = false
  }
}

// ========== 运行历史 ==========
const showRuns = ref(false)
const runs = ref([])
const runsJobName = ref('')
const loadingRuns = ref(false)
const expandedRuns = ref(new Set())

function toggleRunExpand(index) {
  const newSet = new Set(expandedRuns.value)
  if (newSet.has(index)) {
    newSet.delete(index)
  } else {
    newSet.add(index)
  }
  expandedRuns.value = newSet
}

async function viewRuns(job) {
  runsJobName.value = job.name
  runs.value = []
  expandedRuns.value.clear()
  showRuns.value = true
  loadingRuns.value = true
  try {
    const res = await getCronRuns({ jobId: job.id })
    runs.value = res?.runs || []
  } catch (e) {
    runs.value = []
  } finally {
    loadingRuns.value = false
  }
}

// ========== 工具函数 ==========
function truncate(s, n) { return s && s.length > n ? s.slice(0, n) + '...' : s }

function formatDuration(ms) {
  if (!ms && ms !== 0) return ''
  if (ms < 1000) return ms + 'ms'
  if (ms < 60000) return (ms / 1000).toFixed(1) + 's'
  return Math.round(ms / 60000) + 'min'
}

function formatSchedule(schedule) {
  if (!schedule) return '—'
  switch (schedule.kind) {
    case 'cron': return `⏰ ${schedule.expr || ''}${schedule.tz ? ' (' + schedule.tz + ')' : ''}`
    case 'every': return `🔄 ${t('cron.everyPrefix')} ${formatEveryMs(schedule.everyMs)}`
    case 'at': return `📌 ${schedule.at ? formatTime(new Date(schedule.at).getTime()) : formatTime(schedule.atMs)}`
    default: return schedule.kind
  }
}

function formatEveryMs(ms) {
  if (!ms) return ''
  if (ms >= 86400000) return (ms / 86400000) + ' ' + t('cron.durationDay')
  if (ms >= 3600000) return (ms / 3600000) + ' ' + t('cron.durationHour')
  if (ms >= 60000) return (ms / 60000) + ' ' + t('cron.durationMinute')
  return (ms / 1000) + ' ' + t('cron.durationSecond')
}

function formatTime(ms) {
  if (!ms) return '—'
  const d = new Date(ms)
  const now = new Date()
  const diff = d.getTime() - now.getTime()
  if (diff > 0 && diff < 86400000) {
    const mins = Math.round(diff / 60000)
    if (mins < 60) return t('cron.timeMinutesLater').replace('{n}', mins)
    return t('cron.timeHoursLater').replace('{n}', Math.round(mins / 60))
  }
  const pad = n => String(n).padStart(2, '0')
  return `${d.getMonth()+1}/${d.getDate()} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
</script>

<style scoped>
.cron-page { width: 100%; height: 100%; padding: 24px 24px; box-sizing: border-box; overflow-y: auto; }
.cron-container { max-width: 100%; margin: 0 auto; }

/* Fade in */
.fade-in-up { animation: fadeInUp 0.4s cubic-bezier(0.22, 1, 0.36, 1) both; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: none; } }

/* Header */
.cron-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px; }
.header-left { display: flex; flex-direction: column; gap: 2px; }
.page-title { font-size: 16px; font-weight: 700; color: var(--jm-accent-7); display: flex; align-items: center; gap: 8px; margin: 0; letter-spacing: -0.01em; }
.header-hint { font-size: 11px; color: var(--jm-accent-4); margin-left: 28px; }

.add-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 7px 16px; font-size: 12px; font-weight: 600;
  background: rgba(var(--jm-primary-1-rgb), 0.12); color: var(--jm-primary-2);
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.2);
  border-radius: 10px; cursor: pointer;
  backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 0 12px rgba(var(--jm-primary-1-rgb), 0.06), var(--jm-glass-inner-glow);
}
.add-btn:hover { transform: translateY(-1px); background: rgba(var(--jm-primary-1-rgb), 0.18); box-shadow: 0 4px 16px rgba(var(--jm-primary-1-rgb), 0.12); }
.add-btn.small { margin-top: 12px; padding: 5px 12px; font-size: 11px; }

/* Loading / Empty */
.loading-state { display: flex; justify-content: center; padding: 60px 0; }
.loading-spinner { width: 22px; height: 22px; border: 2px solid var(--jm-glass-border); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.6s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.empty-state { display: flex; flex-direction: column; align-items: center; padding: 80px 0; color: var(--jm-accent-4); }
.empty-state p { margin: 12px 0 0; font-size: 13px; }
.empty-hint { padding: 24px; text-align: center; color: var(--jm-accent-4); font-size: 12px; }

/* ===== Job List — Acrylic Cards ===== */
.job-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px; }
.job-card {
  background: var(--jm-glass-bg);
  backdrop-filter: blur(var(--jm-glass-blur)); -webkit-backdrop-filter: blur(var(--jm-glass-blur));
  border: 1px solid var(--jm-glass-border);
  border-radius: 14px;
  padding: 14px 16px;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.06),
    0 6px 16px rgba(0, 0, 0, 0.04);
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
/* 激活态左边框 — 仅3px色条 */
.job-card::before {
  content: ''; position: absolute; left: 0; top: 12px; bottom: 12px; width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--jm-primary-1);
  opacity: 0; transition: opacity 0.3s;
}
.job-card:not(.disabled)::before { opacity: 1; }
.job-card.disabled::before { opacity: 0; }

.job-card:hover {
  border-color: var(--jm-glass-border-hover);
  transform: translateY(-2px);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.08),
    0 12px 28px rgba(0, 0, 0, 0.06);
}
.job-card.disabled { opacity: 0.45; }

/* Compact card layout */
.job-top-row {
  display: flex; align-items: center; justify-content: space-between; gap: 8px;
  padding-left: 8px;
}
.job-title-area { display: flex; align-items: center; gap: 6px; min-width: 0; flex: 1; }
.job-name { font-size: 14px; font-weight: 600; color: var(--jm-accent-7); letter-spacing: -0.01em; }

.schedule-badge {
  font-size: 10px; padding: 2px 8px; border-radius: 10px; white-space: nowrap;
  font-family: var(--jm-font-mono);
  background: rgba(var(--jm-accent-2-rgb), 0.2);
  border: 1px solid var(--jm-glass-border);
  color: var(--jm-accent-5);
}
.session-badge { font-size: 10px; padding: 2px 8px; border-radius: 10px; font-weight: 500; }
.session-badge.isolated { background: rgba(99,102,241,0.1); color: #818cf8; border: 1px solid rgba(99,102,241,0.15); }
.session-badge.main { background: rgba(251,191,36,0.1); color: #f59e0b; border: 1px solid rgba(251,191,36,0.15); }

/* 调度信息行 */
.job-schedule-line {
  display: flex; align-items: center; gap: 8px;
  padding-left: 8px;
  font-size: 10px; color: var(--jm-accent-4);
}
.job-time-row {
  display: flex; align-items: center; gap: 10px;
  padding-left: 8px;
  font-size: 10px; color: var(--jm-accent-4);
}
.next-run, .last-run { display: flex; align-items: center; gap: 3px; }

.job-meta { padding-left: 8px; }
.job-message { font-size: 11px; color: var(--jm-accent-5); display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.job-desc { font-size: 11px; color: var(--jm-accent-4); font-style: italic; }

/* 操作按钮胶囊 — 默认半透明，hover 显现 */
.job-actions-capsule {
  display: flex; gap: 2px;
  padding: 2px;
  padding-left: 6px;
  border-radius: 10px;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  opacity: 0;
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  transform: translateY(4px);
}
.job-card:hover .job-actions-capsule {
  opacity: 1;
  transform: translateY(0);
}
.icon-btn {
  width: 26px; height: 26px; display: flex; align-items: center; justify-content: center;
  background: transparent; border: none;
  border-radius: 7px; color: var(--jm-accent-5); cursor: pointer;
  transition: all 0.2s;
}
.icon-btn:hover { background: var(--jm-glass-bg-hover); color: var(--jm-accent-7); }
.icon-btn.danger:hover { color: #ef4444; background: rgba(239,68,68,0.06); }
.icon-btn:disabled { opacity: 0.3; cursor: not-allowed; }

.mini-spinner { width: 12px; height: 12px; border: 1.5px solid var(--jm-glass-border); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.6s linear infinite; }

/* ===== Form — Glassmorphic ===== */
.form-body { display: flex; flex-direction: column; gap: 20px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-label { font-size: 11px; font-weight: 600; color: var(--jm-accent-4); letter-spacing: 0.05em; text-transform: uppercase; }
.form-label .required { color: #ef4444; }

.schedule-detail {
  margin-top: 8px;
  padding: 14px;
  background: rgba(var(--jm-accent-1-rgb), 0.06);
  border: none;
  border-radius: 12px;
  box-shadow: inset 0 2px 6px rgba(0, 0, 0, 0.03);
  transition: all 0.3s ease;
}

/* Cron hint chips — Neumorphic */
.cron-hints { display: flex; gap: 6px; margin-top: 8px; flex-wrap: wrap; }
.hint-chip {
  font-size: 10px; padding: 4px 12px; border-radius: 8px; cursor: pointer;
  font-family: var(--jm-font-mono);
  background: var(--jm-glass-bg-hover) !important;
  border: none !important;
  color: var(--jm-accent-5);
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.08),
    inset 0 -1px 0 rgba(0, 0, 0, 0.05);
}
.hint-chip:hover {
  transform: translateY(-1px);
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  border-color: rgba(var(--jm-primary-1-rgb), 0.2);
  color: var(--jm-primary-2);
  box-shadow: 0 4px 12px rgba(var(--jm-primary-1-rgb), 0.1);
}
.hint-chip:active {
  transform: translateY(1px);
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.2);
}

.interval-row { display: flex; align-items: center; gap: 6px; }
.interval-label { font-size: 12px; color: var(--jm-accent-5); white-space: nowrap; }

.form-footer { display: flex; justify-content: flex-end; gap: 8px; padding-top: 12px; border-top: 1px solid var(--jm-glass-border); }

/* ===== Radio Cards — 3D with Spring ===== */
.radio-desc-group { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.radio-desc-item {
  display: flex; align-items: flex-start; gap: 10px; padding: 12px 14px;
  border: 1px solid transparent; border-radius: 12px; cursor: pointer;
  background: var(--jm-glass-bg-hover);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.04);
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}
.radio-desc-item:hover {
  border-color: var(--jm-glass-border-hover);
  transform: translateY(-1px);
}
.radio-desc-item.active {
  background: rgba(var(--jm-primary-1-rgb), 0.03);
  border-color: rgba(var(--jm-primary-1-rgb), 0.2);
  box-shadow:
    0 10px 20px rgba(var(--jm-primary-1-rgb), 0.08),
    inset 0 0 0 1px rgba(var(--jm-primary-1-rgb), 0.1);
  transform: scale(1.02);
}
.radio-desc-item > div { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.rdi-title { font-size: 12px; font-weight: 600; color: var(--jm-accent-7); display: flex; align-items: center; gap: 4px; }
.rdi-desc  { font-size: 10px; color: var(--jm-accent-4); line-height: 1.4; }
.rdi-badge { font-size: 9px; padding: 1px 6px; border-radius: 8px; font-weight: 600; }
.rdi-badge.recommended { background: rgba(var(--jm-primary-1-rgb), 0.12); color: var(--jm-primary-1); }

/* Radio dot — glow when checked */
.radio-dot {
  width: 14px; height: 14px; flex-shrink: 0; border-radius: 50%;
  border: 1.5px solid var(--jm-accent-3); background: transparent;
  margin-top: 2px; transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1); box-sizing: border-box;
}
.radio-dot.checked {
  border-color: var(--jm-primary-1); background: var(--jm-primary-1);
  box-shadow: inset 0 0 0 3px var(--jm-bg-color), 0 0 12px rgba(var(--jm-primary-1-rgb), 0.4);
}

/* Advanced section */
.advanced-section {
  border: 1px solid var(--jm-glass-border); border-radius: 12px; overflow: hidden;
  background: rgba(var(--jm-accent-1-rgb), 0.08);
}
.advanced-toggle {
  display: flex; align-items: center; gap: 6px; padding: 10px 14px;
  cursor: pointer; font-size: 11px; font-weight: 500; color: var(--jm-accent-4);
  background: transparent; user-select: none; transition: all 0.2s;
}
.advanced-toggle:hover { background: var(--jm-glass-bg); color: var(--jm-accent-6); }
.advanced-summary { margin-left: auto; font-size: 10px; color: var(--jm-accent-3); font-weight: 400; }
.advanced-body { display: flex; flex-direction: column; gap: 16px; padding: 16px 14px; border-top: 1px solid var(--jm-glass-border); }

.form-hint { font-size: 10px; color: var(--jm-accent-4); margin-top: 4px; }

/* ===== Run History — Frosted Cards ===== */
.runs-list { max-height: 440px; overflow-y: auto; display: flex; flex-direction: column; gap: 8px; padding: 2px 0; }
.run-row {
  display: flex; align-items: stretch; gap: 0;
  background: var(--jm-glass-bg);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--jm-glass-border);
  border-radius: 12px; overflow: hidden;
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  cursor: pointer;
}
.run-row:hover { border-color: var(--jm-glass-border-hover); transform: translateY(-1px); }

.run-status-bar { width: 3px; flex-shrink: 0; }
.run-row.ok    .run-status-bar { background: #3cff7a; box-shadow: 0 0 6px rgba(60,255,122,0.3); }
.run-row.failed .run-status-bar,
.run-row.error  .run-status-bar { background: #ef4444; box-shadow: 0 0 6px rgba(239,68,68,0.3); }
.run-row.running .run-status-bar { background: #818cf8; box-shadow: 0 0 6px rgba(129,140,248,0.3); }
.run-row:not(.ok):not(.failed):not(.error):not(.running) .run-status-bar { background: var(--jm-accent-3); }

.run-content { flex: 1; padding: 10px 14px; min-width: 0; display: flex; flex-direction: column; gap: 4px; }

.run-top { display: flex; align-items: center; gap: 8px; min-width: 0; }
.run-status-badge {
  flex-shrink: 0; font-size: 10px; font-weight: 600; padding: 2px 8px;
  border-radius: 10px; text-transform: lowercase;
}
.run-status-badge.ok      { background: rgba(60,255,122,0.1); color: #3cff7a; }
.run-status-badge.failed,
.run-status-badge.error   { background: rgba(239,68,68,0.1);  color: #ef4444; }
.run-status-badge.running { background: rgba(129,140,248,0.1); color: #818cf8; }
.run-status-badge:not(.ok):not(.failed):not(.error):not(.running) { background: rgba(var(--jm-accent-2-rgb), 0.3); color: var(--jm-accent-5); }

.run-summary {
  font-size: 12px; color: var(--jm-accent-6); overflow: hidden;
  text-overflow: ellipsis; white-space: nowrap; flex: 1;
  transition: all 0.2s; cursor: pointer;
}
.run-row.expanded { background: var(--jm-glass-bg-hover); }
.run-row.expanded .run-summary {
  white-space: pre-wrap; word-break: break-word; text-overflow: clip; line-height: 1.5;
  margin-top: 4px; padding-right: 4px; max-height: 240px; overflow-y: auto;
}
.run-row.expanded .run-summary::-webkit-scrollbar { width: 4px; }
.run-row.expanded .run-summary::-webkit-scrollbar-thumb { background: var(--jm-accent-3); border-radius: 4px; }

.run-bottom { display: flex; align-items: center; gap: 10px; }
.run-index   { font-size: 10px; color: var(--jm-accent-3); font-variant-numeric: tabular-nums; font-family: var(--jm-font-mono); }
.run-time    { font-size: 10px; color: var(--jm-accent-4); font-family: var(--jm-font-mono); }
.run-duration{ font-size: 10px; color: var(--jm-accent-4); font-family: var(--jm-font-mono); }

/* ===== Naive UI Overrides ===== */
:deep(.n-date-panel-date--selected .n-date-panel-date__trigger) {
  background: rgba(var(--jm-primary-1-rgb), 0.12) !important;
  border-radius: 8px !important;
  box-shadow: inset 0 0 0 2px var(--jm-primary-1) !important;
}
:deep(.n-date-panel-date--selected .n-date-panel-date__date),
:deep(.n-date-panel-date--selected .n-date-panel-date__trigger .n-date-panel-date__date) {
  color: var(--jm-primary-1) !important; font-weight: 600 !important;
}
:deep(.n-date-panel-date--current .n-date-panel-date__trigger) { border-color: var(--jm-primary-1) !important; }
:deep(.n-date-panel .n-button--primary-type) { background-color: var(--jm-primary-1) !important; border-color: var(--jm-primary-1) !important; border-radius: 8px !important; }
:deep(.n-date-panel .n-button--primary-type:hover) { background-color: var(--jm-primary-2) !important; border-color: var(--jm-primary-2) !important; }

:deep(.n-modal .n-button--primary-type) { background-color: var(--jm-primary-1) !important; border-color: var(--jm-primary-1) !important; border-radius: 10px !important; }
:deep(.n-modal .n-button--primary-type:hover) { background-color: var(--jm-primary-2) !important; border-color: var(--jm-primary-2) !important; }

/* Modal glass override */
:deep(.n-card) {
  background: var(--jm-glass-bg) !important;
  backdrop-filter: blur(25px) !important; -webkit-backdrop-filter: blur(25px) !important;
  border: 1px solid var(--jm-glass-border) !important;
  border-radius: 16px !important;
  box-shadow: var(--jm-shadow-elevation-3) !important;
}
:deep(.n-card-header) { border-bottom-color: var(--jm-glass-border) !important; }
:deep(.n-card__footer) { border-top-color: var(--jm-glass-border) !important; }

/* Input focus — subtle gray ring, no neon */
:deep(.n-input) { border-radius: 10px !important; transition: box-shadow 0.3s !important; }
:deep(.n-input--focus) { box-shadow: 0 0 0 2px rgba(var(--jm-accent-3-rgb, 150,150,150), 0.2) !important; }

:deep(.n-radio-group) {
  /* 容器：物理凹槽，无缝滑块 */
  background: rgba(var(--jm-accent-1-rgb), 0.12) !important;
  border-radius: 12px !important;
  padding: 3px !important;
  display: inline-flex !important;
  align-items: center !important;
  gap: 0 !important;
  border: 1px solid rgba(0, 0, 0, 0.03) !important;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.05) !important;
  width: fit-content !important;
}

:deep(.n-radio-button) {
  --n-button-border-color: transparent !important;
  --n-button-box-shadow: none !important;
  background: transparent !important;
  border: none !important;
  border-radius: 9px !important;
  margin: 0 !important;
  height: 32px !important;
  line-height: 32px !important;
  padding: 0 20px !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  color: var(--jm-accent-4) !important;
}

:deep(.n-radio-button:hover) {
  color: var(--jm-accent-7) !important;
  background: rgba(255, 255, 255, 0.08) !important;
}

:deep(.n-radio-button--checked) {
  /* 选中态：实色悬浮滑块，双层阴影 + 微弹 */
  background: var(--jm-glass-bg-hover) !important;
  color: var(--jm-primary-1) !important;
  font-weight: 600 !important;
  box-shadow:
    0 3px 8px rgba(0, 0, 0, 0.12),
    0 1px 2px rgba(0, 0, 0, 0.04) !important;
  transform: translateY(-0.5px);
}

/* 彻底杀掉 Naive UI 的所有分割线和边框伪元素 */
:deep(.n-radio-button .n-radio-button__state-border),
:deep(.n-radio-button::before),
:deep(.n-radio-button::after) {
  display: none !important;
  opacity: 0 !important;
  width: 0 !important;
}
:deep(.n-radio-group .n-radio-button__border) {
  box-shadow: none !important;
  border: none !important;
  opacity: 0 !important;
}

/* Advanced body — subtle nesting */
.advanced-body {
  background: rgba(var(--jm-accent-1-rgb), 0.05) !important;
  margin: 0 -1px -1px !important;
  border-radius: 0 0 11px 11px !important;
}

/* add-btn */
.add-btn { border-radius: 10px !important; }
.add-btn:hover { filter: none !important; }
</style>


