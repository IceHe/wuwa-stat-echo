import { ref } from 'vue'

const storageKey = 'echo-substat-color-mode'

const readStoredColorMode = () => {
  if (typeof window === 'undefined') {
    return false
  }
  return window.localStorage.getItem(storageKey) === 'colorful'
}

export const isEchoSubstatColorful = ref(readStoredColorMode())

export const toggleEchoSubstatColorMode = () => {
  isEchoSubstatColorful.value = !isEchoSubstatColorful.value
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(
      storageKey,
      isEchoSubstatColorful.value ? 'colorful' : 'default',
    )
  }
}
