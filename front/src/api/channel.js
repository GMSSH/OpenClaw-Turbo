import gm from '@/utils/gmssh'

export function getChannels() {
    return gm.request('getChannels')
}

export function saveChannel(params) {
    return gm.request('saveChannel', params)
}

export function deleteChannel(params) {
    return gm.request('deleteChannel', params)
}

export function toggleChannel(params) {
    return gm.request('toggleChannel', params)
}
