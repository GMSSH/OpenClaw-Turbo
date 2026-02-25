<template>
  <div class="env-check">
    <div class="check-container fade-in-up">
      <div class="check-header">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
            <rect x="2" y="4" width="20" height="16" rx="2.5" stroke="var(--jm-primary-1)" stroke-width="1.5" fill="rgba(var(--jm-primary-1-rgb), 0.08)"/>
            <rect x="5" y="13" width="3" height="4" rx="0.5" fill="var(--jm-primary-2)" opacity="0.7"/>
            <rect x="9" y="11" width="3" height="6" rx="0.5" fill="var(--jm-primary-1)"/>
            <rect x="13" y="9" width="3" height="8" rx="0.5" fill="var(--jm-primary-2)" opacity="0.7"/>
            <rect x="17" y="7" width="3" height="10" rx="0.5" fill="var(--jm-primary-1)"/>
          </svg>
        </div>
        <div>
          <h2>环境检测</h2>
          <p class="header-desc">{{ mode === 'local' ? '检测本地编译所需的运行环境' : '检测 OpenClaw 所需的运行环境' }}</p>
        </div>
      </div>

      <div class="check-list">
        <div
          v-for="(item, i) in checkItems"
          :key="item.key"
          class="check-item"
          :style="{ animationDelay: `${i * 0.1}s` }"
        >
          <div class="item-icon">
            <n-spin v-if="item.status === 'checking'" :size="16" />
            <svg v-else-if="item.status === 'success'" viewBox="0 0 24 24" width="18" height="18">
              <circle cx="12" cy="12" r="11" fill="var(--jm-success-color)" opacity="0.12"/>
              <path d="M8 12.5l2.5 2.5 5.5-5.5" stroke="var(--jm-success-color)" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <svg v-else-if="item.status === 'failed'" viewBox="0 0 24 24" width="18" height="18">
              <circle cx="12" cy="12" r="11" fill="var(--jm-error-color)" opacity="0.12"/>
              <path d="M9 9l6 6M15 9l-6 6" stroke="var(--jm-error-color)" stroke-width="2" fill="none" stroke-linecap="round"/>
            </svg>
            <div v-else class="dot"></div>
          </div>
          <div class="item-info">
            <span class="item-label">{{ item.label }}</span>
            <span class="item-desc">{{ item.desc }}</span>
          </div>
          <span class="item-status" :class="item.status">
            {{ { checking: '检测中', success: '就绪', failed: '未就绪', pending: '等待' }[item.status] }}
          </span>
        </div>
      </div>

      <div v-if="checkComplete" class="result fade-in-up">
        <template v-if="allReady">
          <div class="result-msg ok">
            <svg viewBox="0 0 24 24" width="16" height="16">
              <path d="M8 12.5l2.5 2.5 5.5-5.5" stroke="var(--jm-success-color)" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span>环境已就绪</span>
          </div>
          <n-button type="primary" @click="$emit('passed')" class="go-btn" size="large">
            <template #icon>
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
                <path d="M13 5l7 7-7 7M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </template>
            开始配置部署
          </n-button>
        </template>
        <template v-else>
          <!-- Local 模式：提供一键安装 -->
          <template v-if="mode === 'local'">
            <div class="result-msg fail">
              <svg viewBox="0 0 24 24" width="16" height="16">
                <path d="M12 9v4M12 16v.5" stroke="var(--jm-warning-color)" stroke-width="2" stroke-linecap="round"/>
              </svg>
              <span>Node.js 环境未安装</span>
            </div>
            <n-button type="primary" @click="installNode" :loading="installing" class="go-btn" size="large">
              🔧 一键安装 Node.js 环境
            </n-button>
            <n-button v-if="!installing" quaternary type="primary" @click="retryCheck" size="small">重新检测</n-button>
          </template>
          <!-- Docker 模式 -->
          <template v-else>
            <div class="result-msg fail">
              <svg viewBox="0 0 24 24" width="16" height="16">
                <path d="M12 9v4M12 16v.5" stroke="var(--jm-warning-color)" stroke-width="2" stroke-linecap="round"/>
              </svg>
              <span>请先通过应用中心安装 Docker 管理器</span>
            </div>
            <n-button quaternary type="primary" @click="retryCheck" size="small">重新检测</n-button>
          </template>
        </template>
      </div>

      <!-- 安装日志 -->
      <div v-if="installing || installLogs.length" class="install-log-box">
        <div class="install-log-title">安装日志</div>
        <div class="install-log-content" ref="logBox">
          <div v-for="(log, i) in installLogs" :key="i" class="log-line">{{ log }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, defineEmits, defineProps, nextTick, watch } from 'vue'
import { NButton, NSpin } from 'naive-ui'
import { checkEnvironment, installNodeEnv, getDeployLogs } from '@/api/deploy'
import gm from '@/utils/gmssh'

const props = defineProps({
  mode: { type: String, default: 'docker' }  // 'docker' | 'local'
})
defineEmits(['passed'])

const checkComplete = ref(false)
const allReady = ref(false)
const installing = ref(false)
const installLogs = ref([])
const logBox = ref(null)

// 根据模式决定检测项
const dockerItems = [
  { key: 'pluginPath', label: 'Docker 插件', desc: '检测 Docker 管理器是否已安装', status: 'pending' },
  { key: 'docker', label: 'Docker 引擎', desc: '检测 docker 命令是否可用', status: 'pending' },
  { key: 'dockerCompose', label: 'Docker Compose', desc: '检测 docker compose 是否可用', status: 'pending' },
]
const localItems = [
  { key: 'node', label: 'Node.js', desc: '检测 node 命令是否可用', status: 'pending' },
  { key: 'pnpm', label: 'pnpm', desc: '检测 pnpm 包管理器是否可用', status: 'pending' },
]

const checkItems = reactive(props.mode === 'local' ? localItems : dockerItems)

async function runCheck() {
  checkComplete.value = false
  for (const item of checkItems) { item.status = 'checking' }

  try {
    await sleep(500)
    const r = await checkEnvironment()

    if (props.mode === 'local') {
      checkItems[0].status = r.nodeReady ? 'success' : 'failed'
      if (r.nodeReady && r.nodeVersion) {
        checkItems[0].desc = `已安装 ${r.nodeVersion}`
      }
      await sleep(250)
      checkItems[1].status = r.pnpmReady ? 'success' : 'failed'
      allReady.value = r.nodeReady && r.pnpmReady
    } else {
      checkItems[0].status = r.pluginPathExists ? 'success' : 'failed'
      await sleep(250)
      checkItems[1].status = r.dockerReady ? 'success' : 'failed'
      await sleep(250)
      checkItems[2].status = r.dockerComposeReady ? 'success' : 'failed'
      allReady.value = r.allReady
    }
  } catch {
    checkItems.forEach(i => { i.status = 'failed' })
    allReady.value = false
  }
  await sleep(200)
  checkComplete.value = true
}

const sleep = ms => new Promise(r => setTimeout(r, ms))

function retryCheck() {
  checkItems.forEach(i => { i.status = 'pending' })
  checkComplete.value = false
  installLogs.value = []
  runCheck()
}

async function installNode() {
  installing.value = true
  installLogs.value = []

  try {
    await installNodeEnv()

    // 轮询安装日志
    const poll = setInterval(async () => {
      try {
        const logs = await getDeployLogs()
        installLogs.value = logs.logs || []
        await nextTick()
        if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight

        if (logs.finished) {
          clearInterval(poll)
          installing.value = false
          if (logs.success) {
            gm.success('Node.js 环境安装完成！')
            retryCheck()
          } else {
            gm.error('安装失败，请查看日志')
          }
        }
      } catch {
        // ignore polling errors
      }
    }, 1500)
  } catch (e) {
    installing.value = false
    gm.error('启动安装失败: ' + e.message)
  }
}

onMounted(runCheck)
</script>

<style scoped>
.env-check {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 24px;
}

.check-container { width: 100%; max-width: 440px; }

.check-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}

.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: 10px;
  background: rgba(var(--jm-primary-1-rgb), 0.06);
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.1);
  flex-shrink: 0;
}

.check-header h2 { font-size: 16px; font-weight: 600; color: var(--jm-accent-7); margin: 0 0 2px; }
.header-desc { font-size: 12px; color: var(--jm-accent-4); margin: 0; }

.check-list { display: flex; flex-direction: column; gap: 4px; margin-bottom: 24px; }

.check-item {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 14px; border-radius: 10px;
  background: rgba(var(--jm-accent-1-rgb), 0.4);
  border: 1px solid rgba(var(--jm-accent-2-rgb, 255,255,255), 0.06);
  animation: fadeInUp 0.3s ease-out both;
}

.item-icon { width: 20px; height: 20px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.dot { width: 7px; height: 7px; border-radius: 50%; background: var(--jm-accent-3); }
.item-info { flex: 1; display: flex; flex-direction: column; gap: 1px; }
.item-label { font-size: 13px; font-weight: 500; color: var(--jm-accent-7); }
.item-desc { font-size: 11px; color: var(--jm-accent-4); }
.item-status { font-size: 11px; font-weight: 500; flex-shrink: 0; }
.item-status.checking { color: var(--jm-primary-2); }
.item-status.success { color: var(--jm-success-color); }
.item-status.failed { color: var(--jm-error-color); }
.item-status.pending { color: var(--jm-accent-4); }

.result { display: flex; flex-direction: column; align-items: center; gap: 14px; }
.result-msg { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--jm-accent-6); }

.go-btn {
  width: 100%; height: 40px; font-size: 14px; font-weight: 500;
  border-radius: 10px; transition: transform 0.15s, box-shadow 0.15s;
}
.go-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 16px rgba(var(--jm-primary-1-rgb), 0.25); }

/* 安装日志 */
.install-log-box {
  margin-top: 16px;
  border: 1px solid var(--jm-accent-2);
  border-radius: 8px;
  overflow: hidden;
}
.install-log-title {
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 500;
  color: var(--jm-accent-5);
  background: rgba(var(--jm-accent-1-rgb), 0.6);
  border-bottom: 1px solid var(--jm-accent-2);
}
.install-log-content {
  max-height: 240px;
  overflow-y: auto;
  padding: 8px 12px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 11px;
  line-height: 1.6;
  color: var(--jm-accent-6);
  background: rgba(0,0,0,0.02);
}
.log-line { white-space: pre-wrap; word-break: break-all; }

@keyframes fadeInUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
</style>
