export type ScoreTemplateChangeEvent = {
  field: 'resonator' | 'cost'
  value: string
  resonator: string
  cost: string
}

const channelName = 'wuwa-echo-score-template-sync'
const storageKey = 'wuwa-echo-score-template-sync'

let channel: BroadcastChannel | null = null

const getChannel = (): BroadcastChannel | null => {
  if (typeof window === 'undefined' || typeof BroadcastChannel === 'undefined') {
    return null
  }
  if (!channel) {
    channel = new BroadcastChannel(channelName)
  }
  return channel
}

export const publishScoreTemplateChange = (payload: ScoreTemplateChangeEvent) => {
  const currentChannel = getChannel()
  if (currentChannel) {
    currentChannel.postMessage(payload)
    return
  }
  if (typeof window === 'undefined') {
    return
  }
  localStorage.setItem(storageKey, JSON.stringify({
    ...payload,
    timestamp: Date.now(),
  }))
}

export const subscribeScoreTemplateChange = (
  handler: (payload: ScoreTemplateChangeEvent) => void,
) => {
  const currentChannel = getChannel()
  if (currentChannel) {
    const listener = (event: MessageEvent<ScoreTemplateChangeEvent>) => {
      if (event.data) {
        handler(event.data)
      }
    }
    currentChannel.addEventListener('message', listener)
    return () => currentChannel.removeEventListener('message', listener)
  }

  if (typeof window === 'undefined') {
    return () => {}
  }

  const listener = (event: StorageEvent) => {
    if (event.key !== storageKey || !event.newValue) {
      return
    }
    try {
      const payload = JSON.parse(event.newValue)
      handler(payload)
    } catch {
      // Ignore malformed cross-tab sync payloads.
    }
  }
  window.addEventListener('storage', listener)
  return () => window.removeEventListener('storage', listener)
}
