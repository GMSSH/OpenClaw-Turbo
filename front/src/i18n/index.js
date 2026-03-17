import { createI18n } from 'vue-i18n'
import zh from './zh.js'
import en from './en.js'

// Create i18n with 'zh' as safe default.
// The actual locale is applied in main.js after $gm is confirmed ready.
const i18n = createI18n({
    legacy: false,        // Composition API mode
    locale: 'zh',
    fallbackLocale: 'zh',
    messages: { zh, en },
})

/**
 * Call this after $gm is confirmed available.
 * Reads $gm.lang and updates the active locale.
 * Normalizes e.g. 'en-US' → 'en', 'zh-CN' → 'zh'.
 */
export function applyGmLang() {
    const rawLang = window.$gm?.lang || ''
    console.log('[i18n] $gm.lang raw value:', JSON.stringify(rawLang))

    let lang = 'zh' // default
    if (rawLang.toLowerCase().startsWith('en')) lang = 'en'
    else if (rawLang.toLowerCase().startsWith('zh')) lang = 'zh'

    console.log('[i18n] applying locale:', lang)
    i18n.global.locale.value = lang
}

export default i18n
