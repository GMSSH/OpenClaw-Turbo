import { createApp } from 'vue'
import { createDiscreteApi } from 'naive-ui'
import App from './App.vue'
import router from './router'
import i18n, { applyGmLang } from './i18n/index.js'
import './theme/theme.css'
import './style.css'
import '@fontsource/jetbrains-mono/400.css'
import '@fontsource/jetbrains-mono/500.css'
import '@fontsource/jetbrains-mono/600.css'

// 创建离散化 API，用于非组件场景
const { message, dialog } = createDiscreteApi(['message', 'dialog'])
window.$message = message
window.$dialog = dialog

    // 应用初始化：先调用 SDK init，再挂载 Vue
    ; (async () => {
        // 注册并调用 $gm.init
        if (window.$gm) {
            window.$gm.init = function () {
                return window.$gm.request('/api/center/check_status', {
                    method: 'post',
                    data: {
                        app_name: window?.$gm?.name,
                        version: window?.$gm?.version,
                        communication_type: window?.$gm?.communicationType,
                    },
                })
            }
            try {
                await window.$gm.init()
            } catch (e) {
                console.warn('[GMClaw] SDK init failed:', e)
            }
        }

        // Apply host language setting before mounting
        applyGmLang()

        const app = createApp(App)

        // Vue 全局错误处理：仅记录日志，不打断用户界面
        app.config.errorHandler = (err, instance, info) => {
            console.error('[GMClaw] Vue Error:', info, err)
        }
        window.addEventListener('unhandledrejection', (event) => {
            // 忽略 vue-i18n 的内部警告
            if (event.reason && String(event.reason).includes('[intlify]')) return
            console.error('[GMClaw] Unhandled Rejection:', event.reason)
        })

        app.use(router)
        app.use(i18n)
        app.mount('#app')

        // 生命周期：应用被宿主关闭时卸载
        if (window.$gm && window.$gm.childDestroyedListener) {
            window.$gm.childDestroyedListener(() => {
                app.unmount()
            })
        }
    })()
