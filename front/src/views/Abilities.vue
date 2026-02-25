<template>
  <div class="skills-page">
    <div class="skills-container fade-in-up">
      <!-- 顶部 -->
      <div class="skills-header">
        <div class="header-left">
          <h2 class="page-title">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none">
              <polygon points="13,2 3,14 12,14 11,22 21,10 12,10" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/>
            </svg>
            能力管理
          </h2>
          <span class="header-hint">管理 OpenClaw 的内置技能和社区市场技能</span>
        </div>
        <button class="refresh-btn" @click="fetchBuiltin()" :disabled="loadingBuiltin" title="刷新">
          <svg :class="{ spinning: loadingBuiltin }" viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M1 4v6h6M23 20v-6h-6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
      </div>

      <!-- 一级 Tab: 内置 / 市场 -->
      <div class="main-tab-bar">
        <button class="main-tab" :class="{ active: mainTab === 'builtin' }" @click="mainTab = 'builtin'">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><rect x="3" y="3" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.5"/><rect x="14" y="3" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.5"/><rect x="3" y="14" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.5"/><rect x="14" y="14" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.5"/></svg>
          内置技能
          <span v-if="builtinSkills.length" class="tab-badge">{{ builtinReadyCount }}/{{ builtinSkills.length }}</span>
        </button>
        <button class="main-tab" :class="{ active: mainTab === 'market' }" @click="mainTab = 'market'">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/><polygon points="16.2,7.8 14.5,14.5 7.8,16.2 9.5,9.5" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>
          技能市场
        </button>
      </div>

      <!-- ========== 内置技能面板 ========== -->
      <div v-if="mainTab === 'builtin'">
        <div v-if="loadingBuiltin" class="loading-state"><div class="loading-spinner"></div></div>
        <div v-else-if="builtinSkills.length === 0" class="empty-hint">暂无内置技能</div>
        <div v-else class="card-grid">
          <div v-for="skill in builtinSkills" :key="skill.name" class="skill-card-v2">
            <div class="card-top">
              <span class="card-name">{{ skill.name }}</span>
              <n-switch size="small" :value="skill.enabled" :loading="builtinLoading === skill.name" @update:value="v => toggleBuiltin(skill.name, v)" />
            </div>
            <p class="card-desc">{{ truncate(skill.description, 80) }}</p>
            <div class="card-footer">
              <span class="card-badge" :class="{ ready: skill.enabled }">{{ skill.enabled ? '✓ 已安装' : '✗ 未安装' }}</span>
              <span class="card-source">openclaw-bundled</span>
            </div>
          </div>
        </div>
      </div>

      <!-- ========== 技能市场面板 ========== -->
      <div v-if="mainTab === 'market'">
        <!-- 搜索栏 -->
        <div class="search-bar">
          <n-input
            v-model:value="searchQuery"
            placeholder="搜索技能... 例如: github, video, notion"
            clearable
            @keyup.enter="doSearch"
            size="medium"
          >
            <template #prefix>
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" style="color: var(--jm-accent-4)">
                <circle cx="11" cy="11" r="8" stroke="currentColor" stroke-width="1.5"/>
                <line x1="21" y1="21" x2="16.65" y2="16.65" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </template>
          </n-input>
          <n-button type="primary" @click="doSearch" :loading="searching" :disabled="!searchQuery.trim()">搜索</n-button>
        </div>

        <!-- 推荐技能 -->
        <div class="recommend-section">
          <span class="recommend-title">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none"><polygon points="13,2 3,14 12,14 11,22 21,10 12,10" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
            推荐技能
          </span>
          <div class="recommend-grid">
            <button v-for="r in recommendedSkills" :key="r.slug" class="recommend-chip" @click="doInstall(r.slug)" :disabled="installingSlug === r.slug">
              <span class="chip-icon" v-html="r.icon"></span>
              <span class="chip-info">
                <span class="chip-name">{{ r.name }}</span>
                <span class="chip-slug">{{ r.slug }}</span>
              </span>
            </button>
          </div>
        </div>

        <!-- 二级 Tab 切换 -->
        <div class="tab-bar">
          <button class="tab-btn" :class="{ active: activeTab === 'search' }" @click="activeTab = 'search'">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><circle cx="11" cy="11" r="7" stroke="currentColor" stroke-width="1.5"/><line x1="16.5" y1="16.5" x2="21" y2="21" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            搜索结果
            <span v-if="searchResults.length" class="tab-count">{{ searchResults.length }}</span>
          </button>
          <button class="tab-btn" :class="{ active: activeTab === 'installed' }" @click="activeTab = 'installed'; fetchInstalled()">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z" stroke="currentColor" stroke-width="1.5"/><path d="M3.27 6.96L12 12.01l8.73-5.05M12 22.08V12" stroke="currentColor" stroke-width="1.5"/></svg>
            已安装
            <span v-if="installedSkills.length" class="tab-count">{{ installedSkills.length }}</span>
          </button>
          <button class="tab-btn" :class="{ active: activeTab === 'explore' }" @click="activeTab = 'explore'; fetchExplore()">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/><polygon points="16.2,7.8 14.5,14.5 7.8,16.2 9.5,9.5" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>
            发现
          </button>
        </div>

        <!-- 搜索结果 -->
        <div v-if="activeTab === 'search'">
          <div v-if="searching" class="loading-state"><div class="loading-spinner"></div></div>
          <div v-else-if="searchResults.length === 0 && hasSearched" class="empty-hint">没有找到相关技能</div>
          <div v-else-if="searchResults.length === 0 && !hasSearched" class="empty-hint">输入关键词搜索技能</div>
          <div v-else class="card-grid">
            <div v-for="skill in searchResults" :key="skill.slug" class="skill-card-v2">
              <div class="card-top">
                <span class="card-name">{{ skill.name || skill.slug }}</span>
                <div class="card-right">
                  <n-button size="tiny" quaternary @click="viewDetail(skill.slug)">详情</n-button>
                  <n-button size="tiny" type="primary" ghost @click="doInstall(skill.slug)" :loading="installingSlug === skill.slug">安装</n-button>
                </div>
              </div>
              <p class="card-desc">{{ skill.slug }}</p>
              <div class="card-footer">
                <span class="card-badge">v{{ skill.version }}</span>
                <span v-if="skill.score" class="card-score">⭐ {{ skill.score }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 已安装 -->
        <div v-if="activeTab === 'installed'">
          <div v-if="loadingInstalled" class="loading-state"><div class="loading-spinner"></div></div>
          <div v-else-if="installedSkills.length === 0" class="empty-hint">暂无已安装的技能</div>
          <div v-else class="card-grid">
            <div v-for="skill in installedSkills" :key="skill.slug" class="skill-card-v2">
              <div class="card-top">
                <span class="card-name">{{ skill.slug }}</span>
                <div class="card-right">
                  <n-button size="tiny" quaternary @click="viewDetail(skill.slug)">详情</n-button>
                  <n-button size="tiny" type="error" quaternary @click="doUninstall(skill.slug)" :loading="installingSlug === skill.slug">卸载</n-button>
                </div>
              </div>
              <div class="card-footer">
                <span v-if="skill.version" class="card-badge">v{{ skill.version }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 最新推荐 -->
        <div v-if="activeTab === 'explore'">
          <div v-if="loadingExplore" class="loading-state"><div class="loading-spinner"></div></div>
          <div v-else-if="exploreSkillsData.length === 0" class="empty-hint">暂无推荐</div>
          <div v-else class="card-grid">
            <div v-for="skill in exploreSkillsData" :key="skill.slug" class="skill-card-v2">
              <div class="card-top">
                <span class="card-name">{{ skill.slug }}</span>
                <div class="card-right">
                  <n-button size="tiny" quaternary @click="viewDetail(skill.slug)">详情</n-button>
                  <n-button size="tiny" type="primary" ghost @click="doInstall(skill.slug)" :loading="installingSlug === skill.slug">安装</n-button>
                </div>
              </div>
              <p class="card-desc">{{ truncate(skill.description, 80) }}</p>
              <div class="card-footer">
                <span class="card-badge">v{{ skill.version }}</span>
                <span v-if="skill.timeAgo" class="card-time">{{ skill.timeAgo }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 详情弹窗 -->
      <div v-if="detailData" class="detail-overlay" @click.self="detailData = null">
        <div class="detail-panel">
          <div class="detail-header">
            <h3>{{ detailData.name || detailData.slug }}</h3>
            <button class="close-btn" @click="detailData = null">&times;</button>
          </div>
          <div class="detail-body">
            <div v-if="detailLoading" class="loading-state"><div class="loading-spinner"></div></div>
            <template v-else>
              <div class="detail-row" v-if="detailData.slug"><span class="detail-label">Slug</span><span class="detail-val">{{ detailData.slug }}</span></div>
              <div class="detail-row" v-if="detailData.summary"><span class="detail-label">简介</span><span class="detail-val">{{ detailData.summary }}</span></div>
              <div class="detail-row" v-if="detailData.version"><span class="detail-label">版本</span><span class="detail-val">v{{ detailData.version }}</span></div>
              <div class="detail-row" v-if="detailData.owner"><span class="detail-label">作者</span><span class="detail-val">{{ detailData.owner }}</span></div>
              <div class="detail-row" v-if="detailData.updated"><span class="detail-label">更新</span><span class="detail-val">{{ detailData.updated }}</span></div>
              <div class="detail-row" v-if="detailData.tags"><span class="detail-label">标签</span><span class="detail-val">{{ detailData.tags }}</span></div>
              <div class="detail-actions">
                <n-button type="primary" @click="doInstall(detailData.slug); detailData = null" size="small">安装此技能</n-button>
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- 风险警告弹窗 -->
      <div v-if="suspiciousData" class="detail-overlay" @click.self="suspiciousData = null">
        <div class="risk-panel">
          <div class="risk-header">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" class="risk-icon-svg"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" stroke="#fc8181" stroke-width="1.5"/><line x1="12" y1="9" x2="12" y2="13" stroke="#fc8181" stroke-width="1.5" stroke-linecap="round"/><circle cx="12" cy="17" r="0.5" fill="#fc8181"/></svg>
            <h3>检测到该技能存在风险</h3>
          </div>
          <div class="risk-body">
            <p class="risk-slug">{{ suspiciousData.slug }}</p>
            <div class="risk-detail">
              <p v-for="(line, i) in suspiciousData.warning.split('\n')" :key="i" class="risk-line">• {{ line }}</p>
            </div>
            <p class="risk-hint">该技能被 VirusTotal 标记为可疑，可能包含风险代码。请审查后再决定是否安装。</p>
          </div>
          <div class="risk-footer">
            <n-button quaternary size="small" @click="suspiciousData = null">取消</n-button>
            <n-button type="error" size="small" @click="doForceInstall" :loading="installingSlug === suspiciousData.slug">我已了解风险，强行安装</n-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { NInput, NButton, NSwitch } from 'naive-ui'
import {
  searchSkills, inspectSkill, installSkill, uninstallSkill,
  listInstalledSkills, exploreSkills,
  listBuiltinSkills, installBuiltinSkill, uninstallBuiltinSkill
} from '@/api/skill'
import gm from '@/utils/gmssh'
import cache from '@/stores/cache'

const mainTab = ref('builtin')

function truncate(str, len) {
  if (!str) return ''
  return str.length > len ? str.slice(0, len) + '...' : str
}

// 内置技能国际化
const builtinI18n = {
  '1password': '配置并使用 1Password 命令行工具。用于管理密钥、集成桌面应用以及处理账号登录。',
  'apple-notes': '在 macOS 上管理苹果备忘录。支持创建、查看、编辑、删除、搜索及导出笔记。',
  'apple-reminders': '在 macOS 上管理苹果提醒事项。支持列表显示、添加、编辑、标记完成及删除。',
  'bear-notes': '通过命令行创建、搜索并管理 Bear (熊掌记) 笔记。',
  'bird': 'X (原 Twitter) 助手。支持阅读、搜索、发推以及基于 Cookie 的互动。',
  'blogwatcher': '博客监控助手。自动监测并提醒 RSS/Atom 订阅源的更新。',
  'blucli': 'BluOS 音响控制工具。支持设备发现、播放控制、音箱分组及音量调节。',
  'bluebubbles': 'BlueBubbles 外部频道插件。用于在 OpenClaw 中集成 iMessage 消息收发。',
  'camsnap': '摄像头监控抓拍。从 RTSP 或 ONVIF 协议的摄像头中提取画面帧或录制片段。',
  'clawhub': 'ClawHub 技能商店。一键搜索、安装、更新或发布 OpenClaw 的智能体技能。',
  'coding-agent': '代码助手。集成 Codex、Claude Code 等引擎，实现自动化编程和代码控制。',
  'docker-essentials': 'Docker 核心工具。用于容器管理、镜像操作及系统排错的必备命令。',
  'eightctl': 'Eight Sleep 智能床垫控制。监控状态、调节温度、设置闹钟和计划任务。',
  'gemini': 'Gemini 助手。支持快速问答、文本总结及内容生成。',
  'gifgrep': 'GIF 动图搜索工具。搜索、下载动图并支持提取静态帧。',
  'github': 'GitHub 深度集成。通过官方命令行处理 Issue、PR (拉取请求) 及查看流水线。',
  'gog': 'Google 全家桶助手。一键管理邮件、日历、云端硬盘及各类办公文档。',
  'goplaces': 'Google 地点搜索。查询餐厅、地标等详细信息及用户评价。',
  'healthcheck': '系统健康检查。提供安全审计、防火墙加固、风险评估及版本自动巡检。',
  'himalaya': 'Himalaya 邮件管理器。在终端通过 IMAP/SMTP 协议收发和管理多账号邮件。',
  'imsg': 'iMessage/短信助手。支持查看聊天记录、监控新消息及直接发送短讯。',
  'local-places': '本地周边搜索。通过本地代理快速查找附近的餐厅、咖啡馆等生活服务。',
  'mcporter': 'MCP 协议搬运工。直接配置和调用各类 MCP 协议服务，支持命令行和类型生成。',
  'model-usage': '模型用量统计。实时汇总各个模型（如 Claude/Codex）的消耗金额和频次。',
  'nano-banana-pro': 'Nano Banana 画师。基于 Gemini 3 Pro 技术生成或编辑图像内容。',
  'nano-pdf': 'PDF 自然语言编辑器。直接用大白话下指令来修改或编辑 PDF 文件。',
  'notion': 'Notion 笔记管理。通过 API 自动化创建和维护页面、数据库及内容块。',
  'obsidian': 'Obsidian 助手。直接操作本地 Markdown 库，实现笔记自动化整理。',
  'openai-image-gen': 'OpenAI 批量绘图。批量生成图片并自动创建一个 HTML 网页画廊。',
  'openai-whisper': 'Whisper 本地语音转文字。离线运行，无需 API 密钥即可识别语音。',
  'openai-whisper-api': 'Whisper 云端转录。使用 OpenAI 官方接口进行高精度的语音转文字。',
  'openhue': '飞利浦 Hue 灯光控制。调节智能灯泡的状态、亮度和场景模式。',
  'oracle': 'Oracle 指导专家。提供提示词封装、文件绑定及附件关联的最佳实践建议。',
  'ordercli': '外卖订单查询。支持查询 Foodora 的历史订单及当前外卖配送状态。',
  'peekaboo': 'macOS UI 自动化。捕获系统界面元素并实现自动点击、录制等操作。',
  'sag': "ElevenLabs 语音合成。提供类似 Mac 系统 'say' 命令的高品质人声播报。",
  'session-logs': '会话日志分析。使用强大的 jq 语法搜索和审计历史聊天记录。',
  'sherpa-onnx-tts': '完全离线语音合成。基于 ONNX 技术的本地 TTS，不依赖云端，保护隐私。',
  'skill-creator': '技能创作工具。帮助你设计、封装并打包新的智能体技能及相关素材。',
  'slack': 'Slack 助手。在 OpenClaw 中直接操控 Slack，包括消息回复和频道置顶。',
  'songsee': '音频可视化工具。自动将音频文件转化为频谱图和动态分析面板。',
  'sonoscli': 'Sonos 音响管理器。控制音响的播放、音量、多房间同步等功能。',
  'spotify-player': 'Spotify 终端播放器。在命令行搜索并播放你喜欢的音乐。',
  'summarize': '万能总结助手。自动从网页链接、播客或视频中提取文字摘要。',
  'things-mac': 'Things 3 任务管理。在 macOS 上快速添加待办、查看收件箱及今日计划。',
  'tmux': 'Tmux 远程控制器。通过模拟按键和捕获面板输出来自动化操作终端。',
  'trello': 'Trello 看板助手。自动化管理任务卡片、看板列表及项目状态。',
  'video-frames': '视频抽帧助手。利用 ffmpeg 从视频中快速提取截图或剪辑短片。',
  'voice-call': '语音通话插件。通过 OpenClaw 插件直接发起网络语音呼叫。',
  'wacli': 'WhatsApp 助手。支持发送消息给联系人、搜索和同步聊天历史。',
  'weather': '天气预报助手。获取当前天气和未来预测，无需 API 密钥。',
}

function skillDesc(skill) {
  return builtinI18n[skill.name] || skill.description || ''
}

// ========== 内置技能 ==========
// 通过外部 cache 缓存
const builtinSkills = computed(() => cache.builtinSkills || [])
const loadingBuiltin = ref(false)
const builtinLoading = ref('')

const builtinReadyCount = computed(() => builtinSkills.value.filter(s => s.enabled).length)

async function fetchBuiltin() {
  loadingBuiltin.value = true
  try {
    const res = await listBuiltinSkills()
    const allSkills = res?.skills || []

    // 用 i18n 词条 key 筛选内置技能，同时保留 openclaw skills list 返回的状态
    const skillMap = {}
    for (const s of allSkills) {
      skillMap[s.name] = s
    }

    cache.builtinSkills = Object.keys(builtinI18n).map(name => {
      const remote = skillMap[name]
      return {
        name,
        icon: remote?.icon || '',
        description: builtinI18n[name],
        enabled: remote?.enabled || false,
        source: remote?.source || 'openclaw-bundled',
      }
    })
  } catch (e) {
    cache.builtinSkills = Object.keys(builtinI18n).map(name => ({
      name,
      description: builtinI18n[name],
      enabled: false,
    }))
  } finally {
    loadingBuiltin.value = false
  }
}

async function toggleBuiltin(name, enabled) {
  if (enabled) {
    await doInstallBuiltin(name)
  } else {
    await doUninstallBuiltin(name)
  }
}

async function doInstallBuiltin(name) {
  builtinLoading.value = name
  try {
    const res = await installBuiltinSkill({ name })
    if (res?.success) gm.success(res.message || `${name} 安装成功`)
    await fetchBuiltin()
  } catch (e) {
    gm.error('安装失败: ' + (e.message || ''))
  } finally {
    builtinLoading.value = ''
  }
}

async function doUninstallBuiltin(name) {
  builtinLoading.value = name
  try {
    await uninstallBuiltinSkill({ name })
    gm.success(`${name} 已卸载`)
    await fetchBuiltin()
  } catch (e) {
    gm.error('卸载失败: ' + (e.message || ''))
  } finally {
    builtinLoading.value = ''
  }
}

// ========== 市场技能 ==========
const activeTab = ref('search')
const searchQuery = ref('')
const searching = ref(false)
const hasSearched = ref(false)
const searchResults = ref([])
const installedSkills = ref([])
const loadingInstalled = ref(false)
const loadingExplore = ref(false)
const exploreSkillsData = ref([])
const installingSlug = ref('')
const detailData = ref(null)
const detailLoading = ref(false)
const suspiciousData = ref(null)

const recommendedSkills = [
  { slug: 'claw-shell', name: '执行命令', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><rect x="2" y="3" width="20" height="18" rx="2" stroke="currentColor" stroke-width="1.3"/><path d="M6 9l4 3-4 3M13 15h5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>' },
  { slug: 'sql-toolkit', name: '数据库', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><ellipse cx="12" cy="5" rx="9" ry="3" stroke="currentColor" stroke-width="1.3"/><path d="M21 12c0 1.66-4.03 3-9 3s-9-1.34-9-3M21 5v14c0 1.66-4.03 3-9 3s-9-1.34-9-3V5" stroke="currentColor" stroke-width="1.3"/></svg>' },
  { slug: 'docker-essentials', name: 'Docker', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><rect x="1" y="8" width="22" height="12" rx="2" stroke="currentColor" stroke-width="1.3"/><rect x="5" y="11" width="3" height="3" stroke="currentColor" stroke-width="1"/><rect x="10" y="11" width="3" height="3" stroke="currentColor" stroke-width="1"/><rect x="15" y="11" width="3" height="3" stroke="currentColor" stroke-width="1"/><path d="M10 4h4v4h-4z" stroke="currentColor" stroke-width="1.3"/></svg>' },
  { slug: 'playwright-scraper-skill', name: '网页爬虫', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.3"/><path d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10A15.3 15.3 0 0112 2" stroke="currentColor" stroke-width="1.3"/></svg>' },
  { slug: 'tavily-search', name: '智能搜索', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><circle cx="11" cy="11" r="7" stroke="currentColor" stroke-width="1.3"/><line x1="16.5" y1="16.5" x2="21" y2="21" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/><path d="M11 8v6M8 11h6" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>' },
  { slug: 'clawdbot-filesystem', name: '文件系统', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" stroke="currentColor" stroke-width="1.3"/></svg>' },
  { slug: 'agentmail', name: 'Email', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><rect x="2" y="4" width="20" height="16" rx="2" stroke="currentColor" stroke-width="1.3"/><path d="M22 6l-10 7L2 6" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round"/></svg>' },
  { slug: 'openclaw-memorybox', name: '终身记忆', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><path d="M12 2a7 7 0 017 7c0 2.38-1.19 4.47-3 5.74V17a2 2 0 01-2 2h-4a2 2 0 01-2-2v-2.26C6.19 13.47 5 11.38 5 9a7 7 0 017-7z" stroke="currentColor" stroke-width="1.3"/><line x1="9" y1="21" x2="15" y2="21" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>' },
  { slug: 'find-skills', name: '找技能', icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none"><circle cx="11" cy="11" r="7" stroke="currentColor" stroke-width="1.3"/><line x1="16.5" y1="16.5" x2="21" y2="21" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/><polygon points="11,7 12.5,10 15,10.5 13,12.5 13.5,15 11,13.5 8.5,15 9,12.5 7,10.5 9.5,10" stroke="currentColor" stroke-width="1" fill="none"/></svg>' },
]

async function doSearch() {
  if (!searchQuery.value.trim()) return
  searching.value = true
  hasSearched.value = true
  try {
    const res = await searchSkills({ query: searchQuery.value.trim() })
    searchResults.value = res?.skills || []
  } catch (e) { gm.error('搜索失败: ' + (e.message || '')) }
  finally { searching.value = false }
}

async function fetchInstalled() {
  loadingInstalled.value = true
  try {
    const res = await listInstalledSkills()
    installedSkills.value = res?.skills || []
  } catch (e) { gm.error('获取已安装技能失败: ' + (e.message || '')) }
  finally { loadingInstalled.value = false }
}

async function fetchExplore() {
  loadingExplore.value = true
  try {
    const res = await exploreSkills()
    exploreSkillsData.value = res?.skills || []
  } catch (e) { gm.error('获取推荐失败: ' + (e.message || '')) }
  finally { loadingExplore.value = false }
}

async function viewDetail(slug) {
  detailData.value = { slug }
  detailLoading.value = true
  try {
    const res = await inspectSkill({ slug })
    detailData.value = { ...detailData.value, ...res }
  } catch (e) { gm.error('获取详情失败: ' + (e.message || '')); detailData.value = null }
  finally { detailLoading.value = false }
}

async function doInstall(slug) {
  installingSlug.value = slug
  try {
    const res = await installSkill({ slug })
    if (res?.suspicious) { suspiciousData.value = { slug: res.slug || slug, warning: res.warning || '该技能被标记为可疑' }; return }
    if (res?.success) { gm.success(res.message || `${slug} 安装成功`); await fetchInstalled() }
  } catch (e) { gm.error('安装失败: ' + (e.message || '')) }
  finally { installingSlug.value = '' }
}

async function doForceInstall() {
  const slug = suspiciousData.value?.slug
  if (!slug) return
  suspiciousData.value = null
  installingSlug.value = slug
  try {
    const res = await installSkill({ slug, force: true })
    if (res?.success) { gm.success(res.message || `${slug} 强制安装成功`); await fetchInstalled() }
    else gm.error(res?.message || '安装失败')
  } catch (e) { gm.error('强制安装失败: ' + (e.message || '')) }
  finally { installingSlug.value = '' }
}

async function doUninstall(slug) {
  const gmApi = gm.getGmApi()
  const doIt = async () => {
    installingSlug.value = slug
    try { await uninstallSkill({ slug }); gm.success(`${slug} 已卸载`); await fetchInstalled() }
    catch (e) { gm.error('卸载失败: ' + (e.message || '')) }
    finally { installingSlug.value = '' }
  }
  if (gmApi?.dialog) {
    gmApi.dialog.warning({ title: '卸载技能', content: `确定卸载「${slug}」吗？`, positiveText: '确定', negativeText: '取消', onPositiveClick: doIt })
  } else { if (confirm(`确定卸载「${slug}」？`)) doIt() }
}

onMounted(() => {
  if (cache.builtinSkills !== null) return
  fetchBuiltin()
})
</script>

<style scoped>
.skills-page { width: 100%; height: 100%; overflow-y: auto; padding: 20px; }
.skills-container { max-width: 960px; margin: 0 auto; display: flex; flex-direction: column; gap: 16px; }

.skills-header { display: flex; align-items: flex-start; justify-content: space-between; }
.header-left { display: flex; flex-direction: column; gap: 4px; }
.page-title { display: flex; align-items: center; gap: 8px; font-size: 18px; font-weight: 600; color: var(--jm-accent-7); margin: 0; }
.header-hint { font-size: 12px; color: var(--jm-accent-4); padding-left: 28px; }

/* 一级 Tab */
.main-tab-bar { display: flex; gap: 6px; background: rgba(var(--jm-accent-1-rgb), 0.4); border-radius: 10px; padding: 4px; }
.main-tab {
  flex: 1; padding: 10px 16px; border: none; border-radius: 8px;
  background: transparent; color: var(--jm-accent-5); font-size: 13px; font-weight: 500;
  cursor: pointer; transition: all 0.2s;
  display: flex; align-items: center; justify-content: center; gap: 6px;
}
.main-tab.active { background: rgba(var(--jm-primary-1-rgb), 0.12); color: var(--jm-primary-2); box-shadow: 0 1px 4px rgba(0,0,0,0.08); }
.main-tab:hover:not(.active) { color: var(--jm-accent-7); background: rgba(var(--jm-accent-1-rgb), 0.6); }
.tab-badge {
  font-size: 11px; padding: 1px 8px; border-radius: 10px;
  background: rgba(var(--jm-primary-1-rgb), 0.15); color: var(--jm-primary-2); font-weight: 600;
}

/* ========== 卡片网格 ========== */
.card-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.skill-card-v2 {
  display: flex; flex-direction: column; gap: 8px;
  padding: 14px 16px; border-radius: 10px;
  border: 1px solid rgba(var(--jm-accent-2-rgb), 0.6);
  background: rgba(var(--jm-accent-1-rgb), 0.35);
  transition: border-color 0.2s, box-shadow 0.2s;
}
.skill-card-v2:hover { border-color: var(--jm-primary-2); box-shadow: 0 2px 12px rgba(var(--jm-primary-1-rgb), 0.08); }
.card-top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.card-name { font-size: 13px; font-weight: 700; color: var(--jm-accent-7); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.card-right { display: flex; gap: 4px; flex-shrink: 0; }
.card-desc {
  margin: 0; font-size: 12px; color: var(--jm-accent-5); line-height: 1.5;
  display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.card-icon { margin-right: 6px; }
.card-footer { display: flex; align-items: center; gap: 8px; }
.card-badge {
  font-size: 10px; padding: 2px 8px; border-radius: 4px; font-weight: 600;
  background: rgba(var(--jm-accent-2-rgb), 0.4); color: var(--jm-accent-4);
}
.card-badge.ready { background: rgba(72,199,142,0.12); color: #48c78e; }
.card-source { font-size: 10px; color: var(--jm-accent-4); }
.card-score { font-size: 10px; color: var(--jm-warning-color, #ff9800); }
.card-time { font-size: 10px; color: var(--jm-accent-4); margin-left: auto; }

/* 搜索栏 */
.search-bar { display: flex; gap: 8px; margin-bottom: 12px; }
.search-bar .n-input { flex: 1; }

/* 推荐技能 */
.recommend-section { display: flex; flex-direction: column; gap: 8px; margin-bottom: 12px; }
.recommend-title { font-size: 12px; color: var(--jm-accent-4); display: flex; align-items: center; gap: 4px; }
.recommend-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.recommend-chip {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 12px; border: 1px solid var(--jm-accent-2); border-radius: 8px;
  background: rgba(var(--jm-accent-1-rgb), 0.4); cursor: pointer;
  transition: all 0.15s; text-align: left; color: var(--jm-accent-6);
}
.recommend-chip:hover:not(:disabled) { border-color: var(--jm-primary-2); background: rgba(var(--jm-primary-1-rgb), 0.06); }
.recommend-chip:disabled { opacity: 0.5; cursor: wait; }
.chip-icon { flex-shrink: 0; display: flex; align-items: center; color: var(--jm-primary-2); }
.chip-info { display: flex; flex-direction: column; min-width: 0; }
.chip-name { font-size: 12px; font-weight: 600; color: var(--jm-accent-7); }
.chip-slug { font-size: 10px; color: var(--jm-accent-4); font-family: monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* 二级 Tab */
.tab-bar { display: flex; gap: 4px; border-bottom: 1px solid var(--jm-accent-2); padding-bottom: 0; margin-bottom: 12px; }
.tab-btn {
  padding: 8px 14px; border: none; background: transparent;
  color: var(--jm-accent-4); font-size: 13px; cursor: pointer;
  border-bottom: 2px solid transparent; transition: all 0.2s;
  display: flex; align-items: center; gap: 6px;
}
.tab-btn.active { color: var(--jm-primary-1); border-bottom-color: var(--jm-primary-1); }
.tab-btn:hover:not(.active) { color: var(--jm-accent-6); }
.tab-count { background: rgba(var(--jm-primary-1-rgb), 0.1); color: var(--jm-primary-2); padding: 0 6px; border-radius: 8px; font-size: 11px; }

/* 空状态 & 加载 */
.empty-hint { text-align: center; color: var(--jm-accent-4); font-size: 13px; padding: 40px 0; }
.loading-state { display: flex; justify-content: center; padding: 40px; }
.loading-spinner { width: 24px; height: 24px; border: 2px solid var(--jm-accent-2); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* 详情弹窗 */
.detail-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.detail-panel { background: var(--jm-bg-1, #1e1e2e); border: 1px solid var(--jm-accent-2); border-radius: 12px; width: 420px; max-height: 80vh; overflow-y: auto; box-shadow: 0 16px 48px rgba(0,0,0,0.4); }
.detail-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; border-bottom: 1px solid var(--jm-accent-2); }
.detail-header h3 { margin: 0; font-size: 15px; font-weight: 600; color: var(--jm-accent-7); }
.close-btn { width: 28px; height: 28px; border-radius: 6px; border: none; background: transparent; color: var(--jm-accent-4); font-size: 18px; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.close-btn:hover { background: rgba(var(--jm-accent-1-rgb), 0.6); color: var(--jm-accent-6); }
.detail-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 10px; }
.detail-row { display: flex; gap: 12px; }
.detail-label { font-size: 11px; color: var(--jm-accent-4); min-width: 48px; flex-shrink: 0; }
.detail-val { font-size: 13px; color: var(--jm-accent-6); word-break: break-word; }
.detail-actions { padding-top: 8px; display: flex; justify-content: flex-end; }

/* 风险面板 */
.risk-panel { background: var(--jm-bg-1, #1e1e2e); border: 1px solid #e53e3e; border-radius: 12px; width: 440px; box-shadow: 0 16px 48px rgba(229,62,62,0.2); }
.risk-header { display: flex; align-items: center; gap: 10px; padding: 16px 20px; border-bottom: 1px solid rgba(229,62,62,0.3); background: rgba(229,62,62,0.08); border-radius: 12px 12px 0 0; }
.risk-header h3 { margin: 0; font-size: 15px; font-weight: 600; color: #fc8181; }
.risk-icon-svg { flex-shrink: 0; }
.risk-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 10px; }
.risk-slug { margin: 0; font-family: monospace; font-size: 13px; color: var(--jm-accent-6); }
.risk-detail { background: rgba(229,62,62,0.06); border: 1px solid rgba(229,62,62,0.15); border-radius: 8px; padding: 10px 14px; }
.risk-line { margin: 4px 0; font-size: 12px; color: #fc8181; line-height: 1.5; }
.risk-hint { margin: 0; font-size: 11px; color: var(--jm-accent-4); line-height: 1.5; }
.risk-footer { padding: 12px 20px; display: flex; justify-content: flex-end; gap: 8px; border-top: 1px solid var(--jm-accent-2); }
</style>
