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
            定时任务
          </h2>
          <span class="header-hint">管理 OpenClaw 的 Cron 定时任务调度</span>
        </div>
        <div class="header-actions">
          <button class="refresh-btn" @click="fetchJobs()" :disabled="loading" title="刷新">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </button>
          <button class="add-btn" @click="openAddForm">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none">
            <line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
          新增任务
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
        <p>暂无定时任务</p>
        <button class="add-btn small" @click="openAddForm">创建第一个任务</button>
      </div>
      <div v-else class="job-list">
        <div v-for="job in jobs" :key="job.id" class="job-card" :class="{ disabled: !job.enabled }">
          <div class="job-main">
            <div class="job-left">
              <div class="job-title-row">
                <span class="job-name">{{ job.name }}</span>
                <span class="schedule-badge">{{ formatSchedule(job.schedule) }}</span>
                <span v-if="job.sessionTarget" class="session-badge" :class="job.sessionTarget">{{ job.sessionTarget === 'isolated' ? '独立会话' : '主会话' }}</span>
              </div>
              <div class="job-meta">
                <span v-if="job.payload?.message" class="job-message">{{ truncate(job.payload.message, 60) }}</span>
                <span v-if="job.payload?.text" class="job-message">{{ truncate(job.payload.text, 60) }}</span>
                <span v-if="job.description" class="job-desc">{{ job.description }}</span>
              </div>
              <div class="job-times">
                <span v-if="job.state?.nextRunAtMs" class="time-info">
                  <svg viewBox="0 0 24 24" width="11" height="11" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><polyline points="12,6 12,12 16,14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
                  下次: {{ formatTime(job.state.nextRunAtMs) }}
                </span>
                <span v-if="job.state?.lastRunAtMs" class="time-info">
                  最近: {{ formatTime(job.state.lastRunAtMs) }}
                </span>
              </div>
            </div>
            <div class="job-right">
              <n-switch size="small" :value="job.enabled" :loading="togglingId === job.id" @update:value="v => toggleJob(job, v)" />
              <div class="job-actions">
                <n-tooltip trigger="hover"><template #trigger><button class="icon-btn" @click="doRunJob(job.id)" :disabled="runningId === job.id">
                  <svg v-if="runningId !== job.id" viewBox="0 0 24 24" width="14" height="14" fill="none"><polygon points="5,3 19,12 5,21" fill="currentColor"/></svg>
                  <div v-else class="mini-spinner"></div>
                </button></template>手动执行</n-tooltip>
                <n-tooltip trigger="hover"><template #trigger><button class="icon-btn" @click="openEditForm(job)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                </button></template>编辑</n-tooltip>
                <n-tooltip trigger="hover"><template #trigger><button class="icon-btn" @click="viewRuns(job)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><path d="M12 8v4l3 3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/></svg>
                </button></template>运行历史</n-tooltip>
                <n-tooltip trigger="hover"><template #trigger><button class="icon-btn danger" @click="doRemoveJob(job)" :disabled="deletingId === job.id">
                  <div v-if="deletingId === job.id" class="mini-spinner"></div>
                  <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none"><polyline points="3 6 5 6 21 6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                </button></template>删除</n-tooltip>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 新增/编辑弹窗 -->
      <n-modal v-model:show="showForm" preset="card" :title="editingJob ? '编辑任务' : '新增定时任务'" style="max-width: 480px;" :bordered="false" :mask-closable="false" :segmented="{ footer: 'soft' }">
        <div class="form-body">
          <!-- 任务名称 -->
          <div class="form-group">
            <label class="form-label">任务名称 <span class="required">*</span></label>
            <n-input v-model:value="form.name" placeholder="如：早间摘要、定时清理" size="small" />
          </div>

          <!-- 描述 -->
          <div class="form-group">
            <label class="form-label">描述</label>
            <n-input v-model:value="form.desc" placeholder="任务用途说明（可选）" size="small" />
          </div>

          <!-- 请求方式 -->
          <div class="form-group">
            <label class="form-label">请求方式</label>
            <n-radio-group v-model:value="form.payloadKind" size="small" class="schedule-tabs">
              <n-radio-button value="text">文本</n-radio-button>
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
            <label class="form-label">调度方式</label>
            <n-radio-group v-model:value="form.scheduleKind" size="small" class="schedule-tabs">
              <n-radio-button value="cron">Cron 表达式</n-radio-button>
              <n-radio-button value="every">固定间隔</n-radio-button>
              <n-radio-button value="at">一次性</n-radio-button>
            </n-radio-group>
          </div>

          <!-- Cron 表达式 -->
          <div v-if="form.scheduleKind === 'cron'" class="form-group schedule-detail">
            <n-input v-model:value="form.cron" placeholder="0 7 * * *  (分 时 日 月 周)" size="small" />
            <div class="cron-hints">
              <span class="hint-chip" @click="form.cron = '0 * * * *'">每小时</span>
              <span class="hint-chip" @click="form.cron = '0 7 * * *'">每天7点</span>
              <span class="hint-chip" @click="form.cron = '0 9 * * 1'">每周一9点</span>
              <span class="hint-chip" @click="form.cron = '0 0 1 * *'">每月1号</span>
            </div>
          </div>

          <!-- 固定间隔 -->
          <div v-if="form.scheduleKind === 'every'" class="form-group schedule-detail">
            <div class="interval-row">
              <span class="interval-label">每</span>
              <n-input-number v-model:value="form.everyValue" :min="1" placeholder="10" size="small" style="width: 100px" />
              <n-select v-model:value="form.everyUnit" :options="intervalUnits" size="small" style="width: 90px" />
              <span class="interval-label">执行一次</span>
            </div>
          </div>

          <!-- 一次性 -->
          <div v-if="form.scheduleKind === 'at'" class="form-group schedule-detail">
            <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
              <n-date-picker v-model:value="form.atTimestamp" type="datetime" placeholder="选择执行时间" size="small" style="width:100%" format="yyyy-MM-dd HH:mm" />
            </n-config-provider>
          </div>

          <!-- 高级设置折叠区 -->
          <div class="advanced-section">
            <div class="advanced-toggle" @click="showAdvanced = !showAdvanced">
              <svg viewBox="0 0 24 24" width="12" height="12" fill="none" :style="{ transform: showAdvanced ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }">
                <polyline points="9,6 15,12 9,18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <span>高级设置</span>
              <span class="advanced-summary" v-if="!showAdvanced">{{ advancedSummary }}</span>
            </div>
            <div v-show="showAdvanced" class="advanced-body">
              <!-- Session 会话模式 -->
              <div class="form-group">
                <label class="form-label">会话模式</label>
                <div class="radio-desc-group">
                  <label class="radio-desc-item" :class="{ active: form.session === 'isolated' }" @click="form.session = 'isolated'">
                    <span class="radio-dot" :class="{ checked: form.session === 'isolated' }"></span>
                    <div>
                      <span class="rdi-title">Isolated <span class="rdi-badge recommended">推荐</span></span>
                      <span class="rdi-desc">独立沙箱，后台静默执行，不影响聊天记录</span>
                    </div>
                  </label>
                  <label class="radio-desc-item" :class="{ active: form.session === 'main' }" @click="form.session = 'main'">
                    <span class="radio-dot" :class="{ checked: form.session === 'main' }"></span>
                    <div>
                      <span class="rdi-title">Main</span>
                      <span class="rdi-desc">在主会话中执行，适合定时提醒/剧情推演</span>
                    </div>
                  </label>
                </div>
              </div>

              <!-- Wake Mode 唤醒模式 -->
              <div class="form-group">
                <label class="form-label">唤醒模式</label>
                <div class="radio-desc-group">
                  <label class="radio-desc-item" :class="{ active: form.wakeMode === 'heartbeat' }" @click="form.wakeMode = 'heartbeat'">
                    <span class="radio-dot" :class="{ checked: form.wakeMode === 'heartbeat' }"></span>
                    <div>
                      <span class="rdi-title">Next heartbeat <span class="rdi-badge recommended">推荐</span></span>
                      <span class="rdi-desc">等待下个调度周期执行，稳定不突然触发</span>
                    </div>
                  </label>
                  <label class="radio-desc-item" :class="{ active: form.wakeMode === 'now' }" @click="form.wakeMode = 'now'">
                    <span class="radio-dot" :class="{ checked: form.wakeMode === 'now' }"></span>
                    <div>
                      <span class="rdi-title">Now</span>
                      <span class="rdi-desc">保存后立即执行一次，适合即时测试</span>
                    </div>
                  </label>
                </div>
              </div>


              <div class="form-group">
                <label class="form-label">通知</label>
                <n-radio-group v-model:value="form.deliveryMode" size="small">
                  <n-radio-button value="none">不通知</n-radio-button>
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
            <n-button @click="showForm = false" size="small" quaternary>取消</n-button>
            <n-button type="primary" @click="submitForm" :loading="submitting" size="small">{{ editingJob ? '保存修改' : '创建任务' }}</n-button>
          </div>
        </template>
      </n-modal>

      <!-- 运行历史弹窗 -->
      <n-modal v-model:show="showRuns" preset="card" :title="`运行历史 — ${runsJobName}`" style="max-width: 580px;" :bordered="false">
        <div v-if="loadingRuns" class="loading-state"><div class="loading-spinner"></div></div>
        <div v-else-if="!runs.length" class="empty-hint">暂无运行记录</div>
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
import { NSwitch, NTooltip, NModal, NInput, NInputNumber, NButton, NRadioGroup, NRadioButton, NSelect, NDatePicker, NConfigProvider } from 'naive-ui'
import { zhCN, dateZhCN } from 'naive-ui'
import {
  listCronJobs, addCronJob, editCronJob, removeCronJob,
  enableCronJob, disableCronJob, runCronJob, getCronRuns
} from '@/api/cron'
import cache from '@/stores/cache'

const gm = window.$gm || {}

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
    gm.message?.error?.('获取任务列表失败: ' + (e.message || ''))
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
    gm.message?.error?.('操作失败: ' + (e.message || ''))
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
    gm.message?.success?.('任务已触发执行')
    await fetchJobs()
  } catch (e) {
    gm.message?.error?.('执行失败: ' + (e.message || ''))
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
    gm.message?.success?.(`${job.name} 已删除`)
    await fetchJobs()
  } catch (e) {
    gm.message?.error?.('删除失败: ' + (e.message || ''))
  } finally {
    deletingId.value = ''
  }
}

// ========== 新增/编辑表单 ==========
const showForm = ref(false)
const editingJob = ref(null)
const submitting = ref(false)

const intervalUnits = [
  { label: '分钟', value: 'm' },
  { label: '小时', value: 'h' },
  { label: '天', value: 'd' },
]

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
    : '和 AI 的对话内容，例如：搜索今日新闻并生成摘要'
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
  if (!form.value.name) { gm.message?.warning?.('请输入任务名称'); return }
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
      try { params.payload = JSON.parse(f.eventJson) } catch { gm.message?.warning?.('系统事件 JSON 格式不正确'); submitting.value = false; return }
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
      gm.message?.success?.('任务已更新')
    } else {
      await addCronJob(params)
      gm.message?.success?.('任务创建成功')
    }
    showForm.value = false
    await fetchJobs()
  } catch (e) {
    gm.message?.error?.((editingJob.value ? '编辑' : '创建') + '失败: ' + (e.message || ''))
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
    case 'every': return `🔄 每 ${formatEveryMs(schedule.everyMs)}`
    case 'at': return `📌 ${schedule.at ? formatTime(new Date(schedule.at).getTime()) : formatTime(schedule.atMs)}`
    default: return schedule.kind
  }
}

function formatEveryMs(ms) {
  if (!ms) return ''
  if (ms >= 86400000) return (ms / 86400000) + ' 天'
  if (ms >= 3600000) return (ms / 3600000) + ' 小时'
  if (ms >= 60000) return (ms / 60000) + ' 分钟'
  return (ms / 1000) + ' 秒'
}

function formatTime(ms) {
  if (!ms) return '—'
  const d = new Date(ms)
  const now = new Date()
  const diff = d.getTime() - now.getTime()
  if (diff > 0 && diff < 86400000) {
    const mins = Math.round(diff / 60000)
    if (mins < 60) return `${mins} 分钟后`
    return `${Math.round(mins / 60)} 小时后`
  }
  const pad = n => String(n).padStart(2, '0')
  return `${d.getMonth()+1}/${d.getDate()} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
</script>

<style scoped>
.cron-page { width: 100%; height: 100%; padding: 20px; box-sizing: border-box; overflow-y: auto; }
.cron-container { max-width: 880px; margin: 0 auto; }

/* Fade in animation */
.fade-in-up { animation: fadeInUp 0.3s ease; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: none; } }

/* Header */
.cron-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 20px; }
.header-left { display: flex; flex-direction: column; gap: 2px; }
.page-title { font-size: 16px; font-weight: 600; color: var(--jm-accent-8); display: flex; align-items: center; gap: 6px; margin: 0; }
.header-hint { font-size: 11px; color: var(--jm-accent-4); margin-left: 26px; }

.add-btn {
  display: flex; align-items: center; gap: 5px;
  padding: 6px 14px; font-size: 12px; font-weight: 500;
  background: var(--jm-primary-1); color: #fff; border: none; border-radius: 6px; cursor: pointer;
  transition: all 0.15s;
}
.add-btn:hover { filter: brightness(1.15); }
.add-btn.small { margin-top: 12px; padding: 5px 12px; font-size: 11px; }

/* Loading / Empty */
.loading-state { display: flex; justify-content: center; padding: 60px 0; }
.loading-spinner { width: 24px; height: 24px; border: 2px solid var(--jm-accent-2); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.6s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.empty-state { display: flex; flex-direction: column; align-items: center; padding: 60px 0; color: var(--jm-accent-4); }
.empty-state p { margin: 12px 0 0; font-size: 13px; }
.empty-hint { padding: 24px; text-align: center; color: var(--jm-accent-4); font-size: 12px; }

/* Job list */
.job-list { display: flex; flex-direction: column; gap: 8px; }
.job-card {
  background: var(--jm-bg-2); border: 1px solid var(--jm-accent-1); border-radius: 8px;
  padding: 14px 16px; transition: all 0.15s;
}
.job-card:hover { border-color: var(--jm-primary-1); box-shadow: 0 0 0 1px rgba(var(--jm-primary-rgb, 99,102,241), 0.1); }
.job-card.disabled { opacity: 0.55; }

.job-main { display: flex; justify-content: space-between; gap: 16px; }
.job-left { flex: 1; min-width: 0; }
.job-right { display: flex; flex-direction: column; align-items: flex-end; gap: 8px; flex-shrink: 0; }

.job-title-row { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.job-name { font-size: 13px; font-weight: 600; color: var(--jm-accent-8); }
.schedule-badge { font-size: 10px; padding: 1px 6px; background: var(--jm-accent-1); color: var(--jm-accent-6); border-radius: 4px; white-space: nowrap; }
.session-badge { font-size: 10px; padding: 1px 6px; border-radius: 4px; }
.session-badge.isolated { background: rgba(99,102,241,0.12); color: #818cf8; }
.session-badge.main { background: rgba(251,191,36,0.12); color: #f59e0b; }

.job-meta { margin-top: 4px; }
.job-message { font-size: 11px; color: var(--jm-accent-5); display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.job-desc { font-size: 11px; color: var(--jm-accent-4); font-style: italic; }

.job-times { display: flex; gap: 12px; margin-top: 6px; }
.time-info { font-size: 10px; color: var(--jm-accent-4); display: flex; align-items: center; gap: 3px; }

.job-actions { display: flex; gap: 2px; }
.icon-btn {
  width: 26px; height: 26px; display: flex; align-items: center; justify-content: center;
  background: none; border: 1px solid transparent; border-radius: 4px; color: var(--jm-accent-5);
  cursor: pointer; transition: all 0.15s;
}
.icon-btn:hover { background: var(--jm-accent-1); color: var(--jm-accent-7); border-color: var(--jm-accent-2); }
.icon-btn.danger:hover { color: #ef4444; background: rgba(239,68,68,0.08); }
.icon-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.mini-spinner { width: 12px; height: 12px; border: 1.5px solid var(--jm-accent-3); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.6s linear infinite; }

/* ===== Form ===== */
.form-body { display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 5px; }
.form-label { font-size: 12px; font-weight: 500; color: var(--jm-accent-6); }
.form-label .required { color: #ef4444; }
.schedule-detail { padding: 10px 12px; background: var(--jm-accent-0, rgba(255,255,255,0.02)); border: 1px solid var(--jm-accent-1); border-radius: 6px; }

.cron-hints { display: flex; gap: 6px; margin-top: 6px; flex-wrap: wrap; }
.hint-chip {
  font-size: 10px; padding: 2px 8px; border-radius: 10px; cursor: pointer;
  background: var(--jm-accent-1); color: var(--jm-accent-5); transition: all 0.15s;
}
.hint-chip:hover { background: var(--jm-primary-1); color: #fff; }

.interval-row { display: flex; align-items: center; gap: 6px; }
.interval-label { font-size: 12px; color: var(--jm-accent-5); white-space: nowrap; }

.form-footer { display: flex; justify-content: flex-end; gap: 8px; }

/* Radio desc card group */
.radio-desc-group { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; }
.radio-desc-item {
  display: flex; align-items: flex-start; gap: 8px; padding: 9px 11px;
  border: 1px solid var(--jm-accent-2); border-radius: 7px; cursor: pointer;
  transition: all 0.15s; background: var(--jm-bg-2);
}
.radio-desc-item:hover { border-color: var(--jm-accent-4); }
.radio-desc-item.active { border-color: #5361F4; background: rgba(83,97,244,0.06); }
.radio-desc-item > div { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.rdi-title { font-size: 11px; font-weight: 600; color: var(--jm-accent-8); display: flex; align-items: center; gap: 4px; }
.rdi-desc  { font-size: 10px; color: var(--jm-accent-4); line-height: 1.4; }
.rdi-badge { font-size: 9px; padding: 0 5px; border-radius: 6px; font-weight: 500; }
.rdi-badge.recommended { background: rgba(83,97,244,0.15); color: #5361F4; }

/* Custom radio dot */
.radio-dot {
  width: 13px; height: 13px; flex-shrink: 0; border-radius: 50%;
  border: 1.5px solid var(--jm-accent-3); background: transparent;
  margin-top: 1px; transition: all 0.15s; box-sizing: border-box;
}
.radio-dot.checked { border-color: #5361F4; background: #5361F4; box-shadow: inset 0 0 0 3px var(--jm-bg-2); }

/* Advanced section */
.advanced-section { border: 1px solid var(--jm-accent-1); border-radius: 7px; overflow: hidden; }
.advanced-toggle {
  display: flex; align-items: center; gap: 6px; padding: 8px 12px;
  cursor: pointer; font-size: 11px; font-weight: 500; color: var(--jm-accent-5);
  background: var(--jm-accent-0, rgba(255,255,255,0.02)); user-select: none;
  transition: background 0.15s;
}
.advanced-toggle:hover { background: var(--jm-accent-1); color: var(--jm-accent-7); }
.advanced-summary { margin-left: auto; font-size: 10px; color: var(--jm-accent-3); font-weight: 400; }
.advanced-body { display: flex; flex-direction: column; gap: 14px; padding: 14px 12px; border-top: 1px solid var(--jm-accent-1); }

.form-hint { font-size: 10px; color: var(--jm-accent-4); margin-top: 4px; }

/* Runs */
.runs-list { max-height: 440px; overflow-y: auto; display: flex; flex-direction: column; gap: 6px; padding: 2px 0; }
.run-row {
  display: flex; align-items: stretch; gap: 0;
  background: var(--jm-bg-2); border: 1px solid var(--jm-accent-1);
  border-radius: 7px; overflow: hidden; transition: all 0.15s;
  cursor: pointer;
}
.run-row:hover { border-color: var(--jm-accent-3); }

/* 左侧状态色条 */
.run-status-bar { width: 3px; flex-shrink: 0; }
.run-row.ok    .run-status-bar { background: #22c55e; }
.run-row.failed .run-status-bar,
.run-row.error  .run-status-bar { background: #ef4444; }
.run-row.running .run-status-bar { background: #818cf8; }
.run-row:not(.ok):not(.failed):not(.error):not(.running) .run-status-bar { background: var(--jm-accent-2); }

.run-content { flex: 1; padding: 9px 12px; min-width: 0; display: flex; flex-direction: column; gap: 4px; }

.run-top { display: flex; align-items: center; gap: 8px; min-width: 0; }
.run-status-badge {
  flex-shrink: 0; font-size: 10px; font-weight: 600; padding: 1px 7px;
  border-radius: 10px; text-transform: lowercase;
}
.run-status-badge.ok      { background: rgba(34,197,94,0.12); color: #22c55e; }
.run-status-badge.failed,
.run-status-badge.error   { background: rgba(239,68,68,0.12);  color: #ef4444; }
.run-status-badge.running { background: rgba(99,102,241,0.12); color: #818cf8; }
.run-status-badge:not(.ok):not(.failed):not(.error):not(.running) { background: var(--jm-accent-1); color: var(--jm-accent-5); }

.run-summary {
  font-size: 12px; color: var(--jm-accent-7); overflow: hidden;
  text-overflow: ellipsis; white-space: nowrap; flex: 1;
  transition: all 0.2s; cursor: pointer;
}
.run-row.expanded {
  background: var(--jm-bg-3, rgba(255, 255, 255, 0.03));
}
.run-row.expanded .run-summary {
  white-space: pre-wrap; word-break: break-word; text-overflow: clip; line-height: 1.5;
  margin-top: 4px; padding-right: 4px;
  max-height: 240px; overflow-y: auto;
}
.run-row.expanded .run-summary::-webkit-scrollbar {
  width: 4px; height: 4px;
}
.run-row.expanded .run-summary::-webkit-scrollbar-thumb {
  background: var(--jm-accent-3); border-radius: 4px;
}

.run-bottom { display: flex; align-items: center; gap: 10px; }
.run-index   { font-size: 10px; color: var(--jm-accent-3); font-variant-numeric: tabular-nums; }
.run-time    { font-size: 10px; color: var(--jm-accent-4); }
.run-duration{ font-size: 10px; color: var(--jm-accent-4); }

/* ===== Naive UI 主题色覆盖：#5361F4 ===== */

/* 日期选择器 — 选中日期（边框样式） */
:deep(.n-date-panel-date--selected .n-date-panel-date__trigger) {
  background: rgba(83, 97, 244, 0.12) !important;
  border-radius: 8px !important;
  box-shadow: inset 0 0 0 2px #5361F4 !important;
}
:deep(.n-date-panel-date--selected .n-date-panel-date__date),
:deep(.n-date-panel-date--selected .n-date-panel-date__trigger .n-date-panel-date__date) {
  color: #5361F4 !important;
  font-weight: 600 !important;
}
/* 日期选择器 — 今日边框 */
:deep(.n-date-panel-date--current .n-date-panel-date__trigger) {
  border-color: #5361F4 !important;
}
/* 日期选择器 — 「此刻」「确认」按钮 */
:deep(.n-date-panel .n-button--primary-type) {
  background-color: #5361F4 !important;
  border-color: #5361F4 !important;
  border-radius: 8px !important;
}
:deep(.n-date-panel .n-button--primary-type:hover) {
  background-color: #6a78f5 !important;
  border-color: #6a78f5 !important;
}

/* 弹窗 footer — 「创建任务」「保存修改」主色按钮 */
:deep(.n-modal .n-button--primary-type) {
  background-color: #5361F4 !important;
  border-color: #5361F4 !important;
  border-radius: 8px !important;
}
:deep(.n-modal .n-button--primary-type:hover) {
  background-color: #6a78f5 !important;
  border-color: #6a78f5 !important;
}

/* add-btn 统一圆角 */
.add-btn { border-radius: 8px !important; background: #5361F4 !important; }
.add-btn:hover { background: #6a78f5 !important; filter: none !important; }
</style>


