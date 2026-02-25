// 前端数据缓存 — 模块单例，组件卸载后数据仍然保留
import { reactive } from 'vue'

const store = reactive({
    // 能力管理
    builtinSkills: null,    // null = 未加载, [] = 已加载
    // Agent 人格
    agentFiles: null,       // { IDENTITY, USER, SOUL }
    agentOriginals: null,
    agentTemplates: null,
    // 三方平台
    channels: null,
    // 定时任务
    cronJobs: null,
})

export default store
