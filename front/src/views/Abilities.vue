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
            {{ t('skills.title') }}
          </h2>
          <span class="header-hint">{{ t('skills.subtitle') }}</span>
        </div>
        <button class="refresh-btn" @click="refreshAll()" :disabled="loadingBuiltin" title="刷新">
          <svg :class="{ spinning: loadingBuiltin }" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/></svg>
        </button>
      </div>

      <!-- 已安装 / 市场 Tab -->
      <div class="main-tab-bar">
        <button class="main-tab" :class="{ active: mainTab === 'builtin' }" @click="mainTab = 'builtin'">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><rect x="3" y="3" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.5"/><rect x="14" y="3" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.5"/><rect x="3" y="14" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.5"/><rect x="14" y="14" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.5"/></svg>
          {{ t('skills.builtin') }}
        </button>
        <button class="main-tab" :class="{ active: mainTab === 'recommend' }" @click="mainTab = 'recommend'">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>
          推荐
        </button>
        <button class="main-tab" :class="{ active: mainTab === 'market' }" @click="mainTab = 'market'; checkClawHub()">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/><polygon points="16.2,7.8 14.5,14.5 7.8,16.2 9.5,9.5" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>
          {{ t('skills.market') }}
        </button>
        <button class="main-tab" :class="{ active: mainTab === 'installed' }" @click="mainTab = 'installed'; fetchInstalled()">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z" stroke="currentColor" stroke-width="1.5"/><path d="M3.27 6.96L12 12.01l8.73-5.05M12 22.08V12" stroke="currentColor" stroke-width="1.5"/></svg>
          {{ t('skills.installed') }}
          <span v-if="installedSkills.length" class="tab-badge">{{ installedSkills.length }}</span>
        </button>
      </div>

      <!-- 搜索和筛选 -->
      <div v-if="mainTab === 'builtin'" class="filter-bar">
        <n-input v-model:value="filterQuery" :placeholder="t('skills.searchPlaceholder')" clearable size="medium" class="filter-input">
          <template #prefix>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" style="color:var(--jm-accent-4)"><circle cx="11" cy="11" r="8" stroke="currentColor" stroke-width="1.5"/><line x1="21" y1="21" x2="16.65" y2="16.65" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          </template>
        </n-input>
        <div v-if="mainTab === 'builtin'" class="filter-tabs">
          <button class="filter-tab" :class="{ active: builtinFilter === 'all' }" @click="builtinFilter = 'all'">{{ t('skills.allFilter', { count: builtinSkills.length }) }}</button>
          <button class="filter-tab" :class="{ active: builtinFilter === 'enabled' }" @click="builtinFilter = 'enabled'">{{ t('skills.enabledFilter', { count: builtinReadyCount }) }}</button>
          <button class="filter-tab" :class="{ active: builtinFilter === 'disabled' }" @click="builtinFilter = 'disabled'">{{ t('skills.disabledFilter', { count: builtinSkills.length - builtinReadyCount }) }}</button>
        </div>
      </div>

      <!-- ========== 内置技能面板 ========== -->
      <div v-if="mainTab === 'builtin'">
        <div v-if="loadingBuiltin" class="loading-state"><div class="loading-spinner"></div></div>
        <div v-else-if="filteredBuiltinSkills.length === 0" class="empty-hint">{{ t('skills.noMatch') }}</div>
        <div v-else class="card-grid">
          <div v-for="skill in filteredBuiltinSkills" :key="skill.name" class="skill-card-v2" @click="openSkillDetail(skill)">
            <div class="card-top">
              <div class="card-name-row">
                <span class="card-emoji">{{ skill.icon || '🔧' }}</span>
                <span class="card-name">{{ skill.name }}</span>
              </div>
              <n-switch size="small" :value="skill.enabled" :loading="builtinLoading === skill.name" @update:value="v => toggleBuiltin(skill.name, v)" @click.stop />
            </div>
            <p class="card-desc">{{ truncate(skillDesc(skill), 80) }}</p>
            <div class="card-footer">
              <span class="card-badge">v1.0.0</span>
            </div>
          </div>
        </div>
      </div>

      <!-- ========== 推荐技能面板 ========== -->
      <div v-if="mainTab === 'recommend'">
        <div v-if="loadingRecommend" class="loading-state"><div class="loading-spinner"></div></div>
        <div v-else-if="recommendedSkillsData.length === 0" class="empty-hint">{{ t('skills.noRecommend') }}</div>
        <div v-else class="card-grid">
          <div
            v-for="r in recommendedSkillsData"
            :key="r.slug"
            class="skill-card-v2"
            @click="openRecommendDetail(r)"
          >
            <div class="card-top">
              <div class="card-name-row">
                <span class="card-emoji">📦</span>
                <span class="card-name">{{ r.slug }}</span>
              </div>
              <n-switch
                size="small"
                :value="r.installed"
                :loading="installingRecommend === r.slug"
                :disabled="!r.zipExists"
                @update:value="v => toggleRecommended(r.slug, v)"
                @click.stop
              />
            </div>
            <p class="card-desc">{{ truncate(locale === 'zh' ? r.descCn : r.descEn, 80) }}</p>
            <div class="card-footer">
              <span class="card-badge">{{ r.zipExists ? 'local' : 'unavailable' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- ========== 技能市场面板 ========== -->
      <div v-if="mainTab === 'market'">
        <!-- clawhub 未安装提示 -->
        <div v-if="clawHubChecked && !clawHubReady" class="clawhub-banner">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none">
              <polygon points="13,2 3,14 12,14 11,22 21,10 12,10" stroke="var(--jm-primary-1)" stroke-width="1.5" stroke-linejoin="round"/>
            </svg>
          </div>
          <div class="banner-text">
            <strong>{{ t('skills.clawhubRequired') }}</strong>
            <span>{{ t('skills.clawhubDesc') }}</span>
          </div>
          <n-button type="primary" size="small" @click="doInstallClawHub" :loading="installingHub">
            {{ installingHub ? t('skills.clawhubInstalling') : t('skills.clawhubInstall') }}
          </n-button>
        </div>

        <!-- 搜索栏 -->
        <div class="search-bar">
          <n-input
            v-model:value="searchQuery"
            :placeholder="t('skills.searchMarketPlaceholder')"
            clearable
            @keyup.enter="doSearch"
            size="medium"
            :disabled="!clawHubReady"
          >
            <template #prefix>
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" style="color: var(--jm-accent-4)">
                <circle cx="11" cy="11" r="8" stroke="currentColor" stroke-width="1.5"/>
                <line x1="21" y1="21" x2="16.65" y2="16.65" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </template>
          </n-input>
          <n-button type="primary" @click="doSearch" :loading="searching" :disabled="!searchQuery.trim() || !clawHubReady">{{ t('skills.searchBtn') }}</n-button>
        </div>

        <!-- 限速提示 -->
        <div class="ratelimit-banner">
          <div class="ratelimit-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
              <circle cx="12" cy="12" r="10" stroke="var(--jm-warning-color, #f59e0b)" stroke-width="1.5"/>
              <line x1="12" y1="8" x2="12" y2="12" stroke="var(--jm-warning-color, #f59e0b)" stroke-width="1.5" stroke-linecap="round"/>
              <circle cx="12" cy="16" r="0.8" fill="var(--jm-warning-color, #f59e0b)"/>
            </svg>
          </div>
          <div class="ratelimit-text">
            <span class="ratelimit-title">{{ t('skills.ratelimitTitle') }}</span>
            <span class="ratelimit-desc">
              {{ t('skills.ratelimitDesc') }}
              <a href="https://clawhub.ai/" target="_blank" class="ratelimit-link">clawhub.ai</a>
              {{ t('skills.ratelimitDesc2') }}
              <span v-if="skillsDir" class="ratelimit-path" @click="openSkillsDir" title="点击打开目录">{{ skillsDir }}
                <svg viewBox="0 0 24 24" width="11" height="11" fill="none" style="margin-left:2px;flex-shrink:0">
                  <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" stroke="currentColor" stroke-width="1.5"/>
                </svg>
              </span>
              <span v-else class="ratelimit-path-placeholder">{{ t('skills.ratelimitPathPlaceholder') }}</span>
              {{ t('skills.ratelimitSuffix') }}
            </span>
          </div>
        </div>


        <!-- 二级 Tab 切换 -->
        <div class="tab-bar">
          <button class="tab-btn" :class="{ active: activeTab === 'search' }" @click="activeTab = 'search'">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><circle cx="11" cy="11" r="7" stroke="currentColor" stroke-width="1.5"/><line x1="16.5" y1="16.5" x2="21" y2="21" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            {{ t('skills.searchResults') }}
            <span v-if="searchResults.length" class="tab-count">{{ searchResults.length }}</span>
          </button>
          <button class="tab-btn" :class="{ active: activeTab === 'explore' }" @click="activeTab = 'explore'; fetchExplore()">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/><polygon points="16.2,7.8 14.5,14.5 7.8,16.2 9.5,9.5" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>
            {{ t('skills.discover') }}
          </button>
        </div>

        <!-- 搜索结果 -->
        <div v-if="activeTab === 'search'">
          <div v-if="searching" class="loading-state"><div class="loading-spinner"></div></div>
          <div v-else-if="searchResults.length === 0 && hasSearched" class="empty-hint">{{ t('skills.noSearchResults') }}</div>
          <div v-else-if="searchResults.length === 0 && !hasSearched" class="empty-hint">{{ t('skills.searchHint') }}</div>
          <div v-else class="card-grid">
            <div v-for="skill in searchResults" :key="skill.slug" class="skill-card-v2">
              <div class="card-top">
                <span class="card-name">{{ skill.name || skill.slug }}</span>
                <div class="card-right">
                  <n-button size="tiny" quaternary @click="viewDetail(skill.slug)">{{ t('common.detail') }}</n-button>
                  <n-button size="tiny" type="primary" ghost @click="doInstall(skill.slug)" :loading="installingSlug === skill.slug">{{ t('common.install') }}</n-button>
                </div>
              </div>
              <p class="card-desc">{{ skill.slug }}</p>
              <div class="card-footer">
                <span v-if="skill.score" class="card-score">⭐ {{ skill.score }}</span>
                <span class="card-badge">v{{ skill.version }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 最新推荐 -->
        <div v-if="activeTab === 'explore'">
          <div v-if="loadingExplore" class="loading-state"><div class="loading-spinner"></div></div>
          <div v-else-if="exploreSkillsData.length === 0" class="empty-hint">{{ t('skills.noRecommend') }}</div>
          <div v-else class="card-grid">
            <div v-for="skill in exploreSkillsData" :key="skill.slug" class="skill-card-v2">
              <div class="card-top">
                <span class="card-name">{{ skill.slug }}</span>
                <div class="card-right">
                  <n-button size="tiny" quaternary @click="viewDetail(skill.slug)">{{ t('common.detail') }}</n-button>
                  <n-button size="tiny" type="primary" ghost @click="doInstall(skill.slug)" :loading="installingSlug === skill.slug">{{ t('common.install') }}</n-button>
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

      <!-- ========== 已安装面板 ========== -->
      <div v-if="mainTab === 'installed'">
        <div v-if="loadingInstalled" class="loading-state"><div class="loading-spinner"></div></div>
        <div v-else-if="installedSkills.length === 0" class="empty-hint">{{ t('skills.noInstalled') }}</div>
        <div v-else class="card-grid">
          <div v-for="skill in installedSkills" :key="skill.slug" class="skill-card-v2" @click="openInstalledDetail(skill)">
            <div class="card-top">
              <div class="card-name-row">
                <span class="card-emoji">📦</span>
                <span class="card-name">{{ skill.name || skill.slug }}</span>
              </div>
              <n-button size="tiny" type="error" quaternary @click.stop="doUninstall(skill.slug)" :loading="installingSlug === skill.slug">{{ t('skills.uninstallBtn') }}</n-button>
            </div>
            <p v-if="skill.description" class="card-desc">{{ truncate(skill.description, 80) }}</p>
            <div class="card-footer">
              <span class="card-badge" style="opacity:1">{{ skill.version === 'unknown' ? t('skills.versionUnknown') : 'v' + skill.version }}</span>
              <span v-if="skill.author" class="card-source">— {{ skill.author }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 技能详情弹框 -->
      <n-modal v-model:show="showSkillDetail" preset="card" :bordered="false" size="small" style="max-width: 480px;" :title="selectedSkill?.name || selectedSkill?.slug || ''" :mask-closable="true">
        <!-- Tab 切换 -->
        <div class="detail-tabs">
          <button class="dtab" :class="{ active: detailTab === 'info' }" @click="detailTab = 'info'">{{ t('skills.infoTab') }}</button>
          <button v-if="!selectedSkill?.slug || isSelectedInstalled" class="dtab" :class="{ active: detailTab === 'config' }" @click="detailTab = 'config'; loadEnvVars()">{{ t('skills.configTab') }}</button>
        </div>

        <!-- 信息 Tab -->
        <div v-if="detailTab === 'info'" class="detail-content">
          <!-- 描述 -->
          <div class="info-section">
            <div v-if="detailLoading" class="loading-inline"><div class="loading-spinner-sm"></div> {{ t('common.loading') }}</div>
            <p v-else class="info-desc">{{ skillDesc(selectedSkill) || selectedSkill?.summary || selectedSkill?.description || t('common.noData') }}</p>
          </div>
          <!-- 元信息 -->
          <div class="info-meta">
            <div class="meta-row">
              <span class="meta-key">{{ t('skills.versionLabel') }}</span>
              <span class="meta-val badge">{{ selectedSkill?.version || '1.0.0' }}</span>
            </div>
            <div class="meta-row">
              <span class="meta-key">{{ t('skills.sourceLabel') }}</span>
              <span class="meta-val badge" :class="selectedSkill?.slug ? 'badge-market' : 'badge-builtin'">{{ selectedSkill?.slug ? t('skills.sourceMarket') : t('skills.sourceBuiltin') }}</span>
            </div>
            <div v-if="selectedSkill?.name" class="meta-row">
              <span class="meta-key">{{ t('skills.identLabel') }}</span>
              <span class="meta-val mono">{{ selectedSkill.name }}</span>
            </div>
          </div>
        </div>

        <!-- 配置 Tab -->
        <div v-if="detailTab === 'config'" class="detail-content">
          <p class="config-hint-top">{{ t('skills.configHint') }}</p>

          <div class="env-list">
            <div v-for="(item, i) in envVars" :key="i" class="env-row">
              <n-input v-model:value="item.key" placeholder="KEY" size="small" class="env-key" />
              <span class="env-eq">=</span>
              <n-input v-model:value="item.value" placeholder="value" size="small" class="env-value" />
              <button class="env-del" @click="envVars.splice(i, 1)" title="删除">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
              </button>
            </div>
            <button class="env-add" @click="envVars.push({ key: '', value: '' })">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none"><line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
              {{ t('skills.addEnvVar') }}
            </button>
          </div>

          <div style="display: flex; justify-content: flex-end; margin-top: 8px;">
            <n-button type="primary" size="small" :loading="savingEnv" @click="doSaveEnvVars">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" style="margin-right: 4px;"><path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z" stroke="currentColor" stroke-width="1.5"/><path d="M17 21v-8H7v8M7 3v5h8" stroke="currentColor" stroke-width="1.5"/></svg>
              {{ t('common.saveConfig') }}
            </n-button>
          </div>
        </div>

        <!-- 底部状态 -->
        <template #footer>
          <div class="detail-footer" v-if="!selectedSkill?.slug">
            <div class="status-row" :class="{ on: selectedSkill?.enabled }">
              <svg v-if="selectedSkill?.enabled" viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M22 11.08V12a10 10 0 11-5.93-9.14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><polyline points="22 4 12 14.01 9 11.01" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07" stroke="currentColor" stroke-width="2"/></svg>
              {{ selectedSkill?.enabled ? t('skills.enabled') : t('skills.disabled') }}
            </div>
            <n-switch :value="selectedSkill?.enabled" :loading="builtinLoading === selectedSkill?.name" @update:value="v => { toggleBuiltin(selectedSkill.name, v) }" />
          </div>
          <div class="detail-footer" v-else-if="isSelectedInstalled">
            <span class="status-row on">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none"><path d="M22 11.08V12a10 10 0 11-5.93-9.14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><polyline points="22 4 12 14.01 9 11.01" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              {{ t('skills.alreadyInstalled') }}
            </span>
            <n-button size="small" type="error" quaternary @click="doUninstall(selectedSkill.slug); showSkillDetail = false" :loading="installingSlug === selectedSkill?.slug">{{ t('skills.uninstallBtn') }}</n-button>
          </div>
          <div class="detail-footer" v-else>
            <span></span>
            <n-button type="primary" size="small" @click="doInstall(selectedSkill.slug); showSkillDetail = false" :loading="installingSlug === selectedSkill?.slug">{{ t('skills.installBtn') }}</n-button>
          </div>
        </template>
      </n-modal>


    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { NInput, NButton, NSwitch, NModal } from 'naive-ui'
import {
  searchSkills, inspectSkill, installSkill, uninstallSkill,
  listInstalledSkills, exploreSkills,
  listBuiltinSkills, installBuiltinSkill, uninstallBuiltinSkill, toggleBuiltinSkill,
  listRecommendedSkills, installRecommendedSkill, uninstallRecommendedSkill,
  listEnvVars, saveEnvVars,
  isClawHubInstalled, installClawHub, getSkillsDir
} from '@/api/skill'
import gm from '@/utils/gmssh'
import cache from '@/stores/cache'

const { t } = useI18n()

const mainTab = ref('builtin')
const filterQuery = ref('')
const builtinFilter = ref('all')

// clawhub 全局安装检测
const clawHubChecked = ref(false)
const clawHubReady = ref(true) // 默认不拦截，等检测完再决定
const installingHub = ref(false)

async function checkClawHub() {
  try {
    const res = await isClawHubInstalled()
    clawHubReady.value = !!res?.installed
  } catch { clawHubReady.value = false }
  clawHubChecked.value = true
}

async function doInstallClawHub() {
  installingHub.value = true
  try {
    await installClawHub()
    clawHubReady.value = true
    gm.success('clawhub ' + t('skills.clawhubInstallSuccess'))
  } catch (e) {
    gm.error(t('skills.installFailed') + ': ' + (e.message || ''))
  } finally {
    installingHub.value = false
  }
}

// 技能详情弹框
const showSkillDetail = ref(false)
const selectedSkill = ref(null)
const detailTab = ref('info')

// 环境变量配置
const envVars = ref([])
const savingEnv = ref(false)

async function loadEnvVars() {
  try {
    const res = await listEnvVars()
    envVars.value = (res?.vars || []).map(v => ({ key: v.key, value: v.value }))
  } catch {}
}

async function doSaveEnvVars() {
  // 过滤掉空 key 的行
  const vars = envVars.value.filter(v => v.key.trim())
  savingEnv.value = true
  try {
    await saveEnvVars({ vars })
    gm.success(t('common.saveConfig'))
  } catch (e) {
    gm.error(t('common.saveFailed') + ': ' + (e.message || ''))
  } finally {
    savingEnv.value = false
  }
}

function openSkillDetail(skill) {
  selectedSkill.value = skill
  detailTab.value = 'info'
  showSkillDetail.value = true
}

const isSelectedInstalled = computed(() => {
  if (!selectedSkill.value?.slug) return false
  return installedSkills.value.some(s => s.slug === selectedSkill.value.slug)
})

function refreshAll() {
  fetchBuiltin()
}

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

const filteredBuiltinSkills = computed(() => {
  let list = builtinSkills.value
  if (builtinFilter.value === 'enabled') list = list.filter(s => s.enabled)
  else if (builtinFilter.value === 'disabled') list = list.filter(s => !s.enabled)
  if (filterQuery.value.trim()) {
    const q = filterQuery.value.trim().toLowerCase()
    list = list.filter(s => s.name.toLowerCase().includes(q) || (s.description || '').toLowerCase().includes(q))
  }
  return list
})

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
  await doToggleBuiltin(name, enabled)
}

async function doToggleBuiltin(skillKey, enabled) {
  builtinLoading.value = skillKey
  try {
    const res = await toggleBuiltinSkill({ skillKey, enabled })
    if (res?.success) gm.success(res.message || `${skillKey} ${enabled ? t('skills.installSuccess') : t('skills.uninstallSuccess')}`)
    await fetchBuiltin()
    // 同步更新弹框中选中的技能状态
    if (selectedSkill.value?.name === skillKey) {
      selectedSkill.value = { ...selectedSkill.value, enabled }
    }
  } catch (e) {
    gm.error((enabled ? t('skills.installFailed') : t('skills.uninstallFailed')) + ': ' + (e.message || ''))
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


// ========== 推荐技能（从后端 scripts/ 加载） ==========
const recommendedSkillsData = ref([])
const loadingRecommend = ref(false)
const installingRecommend = ref('')
const locale = computed(() => t('_locale') === 'en' ? 'en' : 'zh')

async function fetchRecommended() {
  loadingRecommend.value = true
  try {
    const res = await listRecommendedSkills()
    recommendedSkillsData.value = res?.skills || []
  } catch (e) {
    gm.error(t('common.loadFailed') + ': ' + (e.message || ''))
  } finally {
    loadingRecommend.value = false
  }
}

async function toggleRecommended(slug, install) {
  installingRecommend.value = slug
  try {
    const res = install
      ? await installRecommendedSkill({ slug })
      : await uninstallRecommendedSkill({ slug })
    if (res?.success) {
      gm.success(res.message || `${slug} ${install ? t('skills.installSuccess') : t('skills.uninstallSuccess')}`)
      await fetchRecommended()
      await fetchInstalled()
      // 同步弹框中的状态
      if (selectedSkill.value?.name === slug) {
        selectedSkill.value = { ...selectedSkill.value, enabled: install }
      }
    }
  } catch (e) {
    gm.error((install ? t('skills.installFailed') : t('skills.uninstallFailed')) + ': ' + (e.message || ''))
  } finally {
    installingRecommend.value = ''
  }
}

function openRecommendDetail(r) {
  selectedSkill.value = {
    name: r.slug,
    icon: '📦',
    description: r.descCn || r.descEn,
    enabled: r.installed,
    source: 'recommend',
  }
  detailTab.value = 'info'
  showSkillDetail.value = true
}

async function doSearch() {
  if (!searchQuery.value.trim()) return
  searching.value = true
  hasSearched.value = true
  try {
    const res = await searchSkills({ query: searchQuery.value.trim() })
    searchResults.value = res?.skills || []
  } catch (e) { gm.error(t('skills.installFailed') + ': ' + (e.message || '')) }
  finally { searching.value = false }
}

async function fetchInstalled() {
  loadingInstalled.value = true
  try {
    const res = await listInstalledSkills()
    installedSkills.value = res?.skills || []
  } catch (e) { gm.error(t('skills.fetchInstalledFailed') + ': ' + (e.message || '')) }
  finally { loadingInstalled.value = false }
}

async function fetchExplore() {
  loadingExplore.value = true
  try {
    const res = await exploreSkills()
    exploreSkillsData.value = res?.skills || []
  } catch (e) { gm.error(t('common.loadFailed') + ': ' + (e.message || '')) }
  finally { loadingExplore.value = false }
}

// 点击已安装技能：直接展示配置 Tab，不请求 clawhub
function openInstalledDetail(skill) {
  selectedSkill.value = {
    slug:        skill.slug,
    name:        skill.name || skill.slug,
    description: skill.description || '',
    version:     skill.version,
    author:      skill.author,
    source:      skill.source,
  }
  detailTab.value = 'config'
  showSkillDetail.value = true
  loadEnvVars()
}

async function viewDetail(slug) {
  selectedSkill.value = { name: slug, slug }
  detailTab.value = 'info'
  showSkillDetail.value = true
  detailLoading.value = true
  try {
    const res = await inspectSkill({ slug })
    selectedSkill.value = { ...selectedSkill.value, ...res, name: res.name || slug, description: res.summary || res.description || '' }
  } catch (e) { gm.error(t('skills.fetchDetailFailed') + ': ' + (e.message || '')); showSkillDetail.value = false }
  finally { detailLoading.value = false }
}

async function doInstall(slug) {
  installingSlug.value = slug
  try {
    const res = await installSkill({ slug })
    if (res?.success) { gm.success(res.message || `${slug} ${t('skills.installSuccess')}`); await fetchInstalled() }
  } catch (e) { gm.error(t('skills.installFailed') + ': ' + (e.message || '')) }
  finally { installingSlug.value = '' }
}

async function doUninstall(slug) {
  const gmApi = gm.getGmApi()
  const doIt = async () => {
    installingSlug.value = slug
    try { await uninstallSkill({ slug }); gm.success(`${slug} ${t('skills.uninstallSuccess')}`); await fetchInstalled() }
    catch (e) { gm.error(t('skills.uninstallFailed') + ': ' + (e.message || '')) }
    finally { installingSlug.value = '' }
  }
  if (gmApi?.dialog) {
    gmApi.dialog.warning({ title: t('skills.uninstallTitle'), content: t('skills.uninstallConfirm', { slug }), positiveText: t('common.confirm'), negativeText: t('common.cancel'), onPositiveClick: doIt })
  } else { if (confirm(t('skills.uninstallConfirm', { slug }))) doIt() }
}

// ========== 技能目录 ==========
const skillsDir = ref('')
const skillsDirMode = ref('local')

async function loadSkillsDir() {
  try {
    const res = await getSkillsDir()
    skillsDir.value = res?.path || ''
    skillsDirMode.value = res?.mode || 'local'
  } catch {}
}

function openSkillsDir() {
  if (!skillsDir.value) return
  const gmApi = gm.getGmApi()
  if (gmApi?.openFolder) {
    gmApi.openFolder(skillsDir.value)
  } else {
    // fallback: 复制路径到剪贴板
    navigator.clipboard?.writeText(skillsDir.value).then(() => {
      gm.success(t('skills.skillsDir'))
    }).catch(() => {
        gm.info(t('skills.skillsDir') + ': ' + skillsDir.value)
    })
  }
}

onMounted(() => {
  checkClawHub()
  loadSkillsDir()
  // 预取已安装数量，用于角标显示
  fetchInstalled()
  // 载入推荐技能列表
  fetchRecommended()
  if (cache.builtinSkills !== null) return
  fetchBuiltin()
})
</script>

<style scoped>
.skills-page { position: relative; width: 100%; height: 100%; overflow-y: auto; padding: 20px 24px; }
.skills-container { max-width: 100%; margin: 0 auto; display: flex; flex-direction: column; gap: 16px; }

/* clawhub 安装提示横幅 */
.clawhub-banner {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 16px; border-radius: 10px;
  background: linear-gradient(135deg, rgba(var(--jm-primary-1-rgb), 0.06), rgba(var(--jm-primary-1-rgb), 0.02));
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.15);
  margin-bottom: 12px;
}
.banner-icon {
  width: 36px; height: 36px; border-radius: 8px;
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.banner-text {
  flex: 1; display: flex; flex-direction: column; gap: 2px;
}
.banner-text strong { font-size: 13px; color: var(--jm-accent-7); }
.banner-text span { font-size: 11px; color: var(--jm-accent-4); }

.skills-header { display: flex; align-items: flex-start; justify-content: space-between; }
.header-left { display: flex; flex-direction: column; gap: 4px; }
.page-title { display: flex; align-items: center; gap: 8px; font-size: 18px; font-weight: 600; color: var(--jm-accent-7); margin: 0; }
.header-hint { font-size: 12px; color: var(--jm-accent-4); padding-left: 28px; }
.refresh-btn {
  display: flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: 8px; border: 1px solid var(--jm-glass-border);
  background: transparent; color: var(--jm-accent-5); cursor: pointer; transition: all 0.2s;
}
.refresh-btn:hover { border-color: var(--jm-accent-3); color: var(--jm-accent-7); }
.refresh-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.spinning { animation: spin 0.8s linear infinite; }

/* 一级 Tab */
.main-tab-bar {
  display: inline-flex; gap: 0; border-radius: 12px; padding: 3px; width: fit-content;
  background: rgba(var(--jm-accent-1-rgb), 0.12);
  border: 1px solid rgba(0, 0, 0, 0.03);
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.05);
}
.main-tab {
  padding: 8px 24px; border: none; border-radius: 9px;
  background: transparent; color: var(--jm-accent-4); font-size: 13px; font-weight: 500;
  cursor: pointer; transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex; align-items: center; justify-content: center; gap: 6px;
}
.main-tab.active {
  background: var(--jm-glass-bg-hover); color: var(--jm-primary-1); font-weight: 600;
  box-shadow: 0 3px 8px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.04);
  transform: translateY(-0.5px);
}
.main-tab:hover:not(.active) { color: var(--jm-accent-7); background: rgba(255, 255, 255, 0.08); }
.tab-badge {
  font-size: 11px; padding: 1px 8px; border-radius: 10px;
  background: rgba(var(--jm-primary-1-rgb), 0.15); color: var(--jm-primary-2); font-weight: 600;
}

/* ========== 卡片网格 ========== */
.card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 14px; }
.skill-card-v2 {
  display: flex; flex-direction: column; gap: 8px;
  padding: 16px; border-radius: 14px;
  border: 1px solid var(--jm-glass-border);
  background: rgba(var(--jm-accent-1-rgb), 0.3);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow:
    
    var(--jm-glass-inner-glow),
    0 1px 3px rgba(0, 0, 0, 0.04),
    0 4px 12px rgba(0, 0, 0, 0.03);
  position: relative;
  overflow: hidden;
}
.skill-card-v2:hover {
  border-color: var(--jm-glass-border-hover);
  transform: translateY(-3px);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.06),
    0 12px 28px rgba(0, 0, 0, 0.06),
    0 0 20px rgba(var(--jm-primary-1-rgb), 0.04);
}
.card-top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.card-name-row { display: flex; align-items: center; gap: 8px; min-width: 0; }
.card-emoji { font-size: 22px; flex-shrink: 0; }
.card-name { font-size: 13px; font-weight: 600; color: var(--jm-accent-7); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; letter-spacing: -0.01em; }
.card-right { display: flex; gap: 4px; flex-shrink: 0; }
.card-desc {
  margin: 0; font-size: 12px; color: var(--jm-accent-5); line-height: 1.5;
  display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.card-icon { margin-right: 6px; }
.card-footer { display: flex; align-items: center; gap: 8px; margin-top: auto; }
.card-badge {
  font-size: 10px; padding: 2px 8px; border-radius: 4px; font-weight: 600;
  background: rgba(var(--jm-accent-2-rgb), 0.4); color: var(--jm-accent-4);
  opacity: 0; transition: opacity 0.2s;
}
.skill-card-v2:hover .card-badge { opacity: 1; }
.card-badge.ready { background: rgba(72,199,142,0.12); color: #48c78e; }
.card-source { font-size: 10px; color: var(--jm-accent-4); }
.card-score { font-size: 10px; color: var(--jm-warning-color, #ff9800); }
.card-time { font-size: 10px; color: var(--jm-accent-4); margin-left: auto; }

/* 推荐面板安装按钮 */
.install-btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 5px 14px; border-radius: 8px; border: none;
  font-size: 12px; font-weight: 500; cursor: pointer;
  background: rgba(var(--jm-primary-1-rgb), 0.1);
  color: var(--jm-primary-1);
  transition: all 0.2s;
}
.install-btn:hover:not(:disabled) {
  background: var(--jm-primary-1); color: #fff;
  box-shadow: 0 3px 10px rgba(var(--jm-primary-1-rgb), 0.3);
}
.install-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.uninstall-btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 5px 14px; border-radius: 8px; border: none;
  font-size: 12px; font-weight: 500; cursor: pointer;
  background: rgba(var(--jm-error-color-rgb, 239,68,68), 0.08);
  color: var(--jm-error-color, #ef4444);
  transition: all 0.2s;
}
.uninstall-btn:hover:not(:disabled) {
  background: rgba(var(--jm-error-color-rgb, 239,68,68), 0.18);
  box-shadow: 0 3px 8px rgba(239,68,68, 0.2);
}
.uninstall-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.badge-installed {
  display: inline-flex; align-items: center;
  font-size: 10px; padding: 2px 8px; border-radius: 10px; font-weight: 600;
  background: rgba(72,199,142,0.12); color: #48c78e; flex-shrink: 0;
}
.spin { animation: spin 0.7s linear infinite; }

/* 搜索筛选 */
.filter-bar { display: flex; align-items: center; gap: 12px; }
.filter-input { flex: 1; }
:deep(.filter-input .n-input) { border-radius: 10px !important; transition: box-shadow 0.3s !important; }
:deep(.filter-input .n-input--focus) { box-shadow: 0 0 0 2px rgba(var(--jm-primary-1-rgb), 0.12), 0 0 12px rgba(var(--jm-primary-1-rgb), 0.06) !important; }
.filter-tabs { display: flex; gap: 4px; flex-shrink: 0; }
.filter-tab {
  padding: 6px 14px; border-radius: 20px; border: 1px solid var(--jm-glass-border);
  background: transparent; color: var(--jm-accent-5); font-size: 12px;
  cursor: pointer; transition: all 0.2s; white-space: nowrap;
}
.filter-tab.active { background: var(--jm-primary-1); color: #fff; border-color: var(--jm-primary-1); }
.filter-tab:hover:not(.active) { border-color: var(--jm-accent-3); color: var(--jm-accent-6); }

/* 详情弹框 */
.detail-tabs { display: flex; border: 1px solid var(--jm-glass-border); border-radius: 8px; overflow: hidden; margin-bottom: 16px; }
.dtab { flex: 1; padding: 10px; border: none; background: transparent; color: var(--jm-accent-5); font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.2s; }
.dtab.active { background: rgba(var(--jm-accent-1-rgb), 0.6); color: var(--jm-accent-7); font-weight: 600; }
.dtab:hover:not(.active) { background: rgba(var(--jm-accent-1-rgb), 0.3); }
.detail-content { display: flex; flex-direction: column; gap: 16px; }
.info-section { }
.info-desc {
  margin: 0; font-size: 13px; color: var(--jm-accent-6);
  line-height: 1.7; word-break: break-word;
}
.info-meta {
  display: flex; flex-direction: column; gap: 0;
  border: 1px solid var(--jm-glass-border); border-radius: 8px; overflow: hidden;
}
.meta-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--jm-accent-2);
}
.meta-row:last-child { border-bottom: none; }
.meta-key { font-size: 12px; color: var(--jm-accent-4); font-weight: 500; }
.meta-val { font-size: 12px; color: var(--jm-accent-6); }
.meta-val.badge {
  padding: 2px 10px; border-radius: 4px;
  background: rgba(var(--jm-accent-1-rgb), 0.5); font-weight: 500;
}
.meta-val.badge-builtin { background: rgba(85,105,250,0.1); color: #5569FA; }
.meta-val.badge-market { background: rgba(72,199,142,0.1); color: #48c78e; }
.meta-val.mono { font-family: 'SF Mono','Fira Code',monospace; font-size: 11px; color: var(--jm-accent-5); }
.config-hint-top { margin: 0; font-size: 12px; color: var(--jm-accent-4); line-height: 1.6; }
.env-list { display: flex; flex-direction: column; gap: 8px; }
.env-row { display: flex; align-items: center; gap: 6px; }
.env-key { flex: 2; }
.env-eq { color: var(--jm-accent-4); font-family: monospace; font-size: 14px; flex-shrink: 0; }
.env-value { flex: 3; }
.env-del {
  width: 28px; height: 28px; border-radius: 6px; border: none;
  background: transparent; color: var(--jm-accent-4); cursor: pointer;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  transition: all 0.15s;
}
.env-del:hover { background: rgba(229,62,62,0.1); color: #fc8181; }
.env-add {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 12px; border: 1px dashed var(--jm-accent-2); border-radius: 8px;
  background: transparent; color: var(--jm-accent-4); font-size: 12px;
  cursor: pointer; transition: all 0.15s;
}
.env-add:hover { border-color: var(--jm-primary-2); color: var(--jm-primary-2); }
.detail-footer { display: flex; align-items: center; justify-content: space-between; }
.status-row { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--jm-accent-4); }
.status-row.on { color: #22c55e; }

/* 可点击卡片 */
.skill-card-v2 { cursor: pointer; }

/* 内联加载 */
.loading-inline { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--jm-accent-4); }
.loading-spinner-sm { width: 14px; height: 14px; border: 2px solid var(--jm-accent-2); border-top-color: var(--jm-primary-1); border-radius: 50%; animation: spin 0.8s linear infinite; }

/* 搜索栏 */
.search-bar { display: flex; gap: 8px; margin-bottom: 12px; }
.search-bar .n-input { flex: 1; }

/* 限速提示 */
.ratelimit-banner {
  display: flex; align-items: flex-start; gap: 10px;
  padding: 10px 14px; border-radius: 10px; margin-bottom: 12px;
  background: rgba(245, 158, 11, 0.06);
  border: 1px solid rgba(245, 158, 11, 0.18);
}
.ratelimit-icon {
  flex-shrink: 0; margin-top: 1px;
  display: flex; align-items: center;
}
.ratelimit-text {
  flex: 1; display: flex; flex-wrap: wrap; align-items: baseline; gap: 4px;
  font-size: 12px; color: var(--jm-accent-5); line-height: 1.6;
}
.ratelimit-title {
  font-weight: 600; color: var(--jm-warning-color, #f59e0b); white-space: nowrap;
}
.ratelimit-desc { flex: 1; min-width: 0; }
.ratelimit-link {
  color: var(--jm-primary-2); text-decoration: none; font-weight: 500;
}
.ratelimit-link:hover { text-decoration: underline; }
.ratelimit-path {
  display: inline-flex; align-items: center; gap: 2px;
  font-family: 'SF Mono', 'Fira Code', monospace; font-size: 11px;
  color: var(--jm-primary-2); background: rgba(var(--jm-primary-1-rgb), 0.08);
  padding: 1px 6px; border-radius: 4px; cursor: pointer;
  border: 1px solid rgba(var(--jm-primary-1-rgb), 0.2);
  transition: all 0.15s;
}
.ratelimit-path:hover { background: rgba(var(--jm-primary-1-rgb), 0.14); }
.ratelimit-path-placeholder {
  font-family: 'SF Mono', 'Fira Code', monospace; font-size: 11px;
  color: var(--jm-accent-4);
}

/* 推荐技能 */
.recommend-section { display: flex; flex-direction: column; gap: 8px; margin-bottom: 12px; }
.recommend-title { font-size: 12px; color: var(--jm-accent-4); display: flex; align-items: center; gap: 4px; }
.recommend-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 8px; }
.recommend-chip {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 12px; border: 1px solid var(--jm-glass-border); border-radius: 10px;
  background: var(--jm-glass-bg); cursor: pointer;
  backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1); text-align: left; color: var(--jm-accent-6);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
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
.detail-overlay { position: fixed; inset: 0; background: var(--jm-overlay-bg); z-index: 100; display: flex; align-items: center; justify-content: center; }
.detail-panel { background: var(--jm-bg-color); border: 1px solid var(--jm-glass-border); border-radius: 12px; width: 420px; max-height: 80vh; overflow-y: auto; box-shadow: 0 16px 48px var(--jm-shadow-color); }
.detail-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; border-bottom: 1px solid var(--jm-accent-2); }
.detail-header h3 { margin: 0; font-size: 15px; font-weight: 600; color: var(--jm-accent-7); }
.close-btn { width: 28px; height: 28px; border-radius: 6px; border: none; background: transparent; color: var(--jm-accent-4); font-size: 18px; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.close-btn:hover { background: rgba(var(--jm-accent-1-rgb), 0.6); color: var(--jm-accent-6); }
.detail-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 10px; }
.detail-row { display: flex; gap: 12px; }
.detail-label { font-size: 11px; color: var(--jm-accent-4); min-width: 48px; flex-shrink: 0; }
.detail-val { font-size: 13px; color: var(--jm-accent-6); word-break: break-word; }
.detail-actions { padding-top: 8px; display: flex; justify-content: flex-end; }

</style>
