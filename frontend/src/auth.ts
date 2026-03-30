import axios from 'axios'
import { reactive } from 'vue'

import { API_BASE_URL } from '@/stores/constants'


const AUTH_TOKEN_KEY = 'wuwa_auth_token'

let interceptorInstalled = false
let currentUserPromise: Promise<AuthUser> | null = null

export type Permission = 'manage' | 'view' | 'edit'

export interface AuthUser {
  id: number
  name: string
  expires_at: string
  remark?: string
  permissions: Permission[]
  created_at: string
  updated_at: string
}

export const authState = reactive({
  token: typeof window === 'undefined' ? '' : localStorage.getItem(AUTH_TOKEN_KEY) || '',
  user: null as AuthUser | null,
})

export const getStoredAuthToken = (): string => {
  if (typeof window === 'undefined') {
    return ''
  }
  return localStorage.getItem(AUTH_TOKEN_KEY) || ''
}

export const setStoredAuthToken = (token: string) => {
  const trimmed = token.trim()
  localStorage.setItem(AUTH_TOKEN_KEY, trimmed)
  authState.token = trimmed
}

export const clearStoredAuthToken = () => {
  localStorage.removeItem(AUTH_TOKEN_KEY)
  authState.token = ''
  authState.user = null
  currentUserPromise = null
}

const getRedirectTarget = () => {
  const fullPath = `${window.location.pathname}${window.location.search}${window.location.hash}`
  if (!fullPath || fullPath.startsWith('/login')) {
    return ''
  }
  return fullPath
}

export const redirectToLogin = () => {
  const redirect = getRedirectTarget()
  const loginUrl = redirect ? `/login?redirect=${encodeURIComponent(redirect)}` : '/login'
  if (window.location.pathname !== '/login') {
    window.location.assign(loginUrl)
  }
}

export const installAxiosAuth = () => {
  if (interceptorInstalled) {
    return
  }

  interceptorInstalled = true

  axios.interceptors.request.use((config) => {
    const token = getStoredAuthToken()
    if (token) {
      config.headers = config.headers || {}
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })

  axios.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error?.response?.status === 401) {
        clearStoredAuthToken()
        redirectToLogin()
      }
      return Promise.reject(error)
    },
  )
}

export const loginWithToken = async (token: string): Promise<AuthUser> => {
  const trimmed = token.trim()
  const response = await axios.post<AuthUser>(`${API_BASE_URL}/auth/login`, { token: trimmed })
  setStoredAuthToken(trimmed)
  authState.user = response.data
  return response.data
}

export const fetchCurrentUser = async (forceRefresh = false): Promise<AuthUser> => {
  const token = getStoredAuthToken()
  if (!token) {
    throw new Error('missing token')
  }

  if (!forceRefresh && authState.user && authState.token === token) {
    return authState.user
  }

  if (!forceRefresh && currentUserPromise) {
    return currentUserPromise
  }

  currentUserPromise = axios
    .get<AuthUser>(`${API_BASE_URL}/auth/me`)
    .then((response) => {
      authState.user = response.data
      authState.token = token
      return response.data
    })
    .finally(() => {
      currentUserPromise = null
    })

  return currentUserPromise
}

export const restoreAuthSession = async (): Promise<AuthUser | null> => {
  const token = getStoredAuthToken()
  if (!token) {
    return null
  }

  authState.token = token
  try {
    return await fetchCurrentUser()
  } catch {
    clearStoredAuthToken()
    return null
  }
}
