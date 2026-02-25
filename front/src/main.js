import { createApp } from 'vue'
import { createDiscreteApi } from 'naive-ui'
import App from './App.vue'
import router from './router'
import './theme/theme.css'
import './style.css'

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

        const app = createApp(App)
        app.use(router)
        app.mount('#app')

        // 生命周期：应用被宿主关闭时卸载
        if (window.$gm && window.$gm.childDestroyedListener) {
            window.$gm.childDestroyedListener(() => {
                app.unmount()
            })
        }
    })()
