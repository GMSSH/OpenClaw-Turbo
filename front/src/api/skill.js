import gm from '@/utils/gmssh'

export function searchSkills(params) {
    return gm.request('searchSkills', params)
}

export function inspectSkill(params) {
    return gm.request('inspectSkill', params)
}

export function installSkill(params) {
    return gm.request('installSkill', params)
}

export function uninstallSkill(params) {
    return gm.request('uninstallSkill', params)
}

export function listInstalledSkills() {
    return gm.request('listInstalledSkills')
}

export function exploreSkills() {
    return gm.request('exploreSkills')
}

export function listBuiltinSkills() {
    return gm.request('listBuiltinSkills')
}

export function installBuiltinSkill(params) {
    return gm.request('installBuiltinSkill', params)
}

export function uninstallBuiltinSkill(params) {
    return gm.request('uninstallBuiltinSkill', params)
}

export function getActiveSkillCount(params) {
    return gm.request('getActiveSkillCount', params)
}
