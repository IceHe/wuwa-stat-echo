<template>
  <div class="login-page">
    <form class="login-card" @submit.prevent="handleLogin">
      <div class="eyebrow">WUWA ECHO</div>
      <h1>Token 登录</h1>
      <p>使用 `~/wuwa/auth` 管理的 token 进入工具。</p>

      <label class="field">
        <span>Token</span>
        <input
          v-model="tokenInput"
          type="password"
          autocomplete="current-password"
          placeholder="请输入 token"
          :disabled="loading"
        />
      </label>

      <button type="submit" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </form>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { loginWithToken, restoreAuthSession } from '@/auth'


const route = useRoute()
const router = useRouter()

const tokenInput = ref('')
const loading = ref(false)
const errorMessage = ref('')

const getRedirectPath = () => {
  const redirect = route.query.redirect
  return typeof redirect === 'string' && redirect.startsWith('/') ? redirect : '/home'
}

const handleLogin = async () => {
  const token = tokenInput.value.trim()
  if (!token) {
    errorMessage.value = '请输入 token'
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    await loginWithToken(token)
    await router.replace(getRedirectPath())
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const status = error.response?.status
      if (status === 401) {
        errorMessage.value = 'token 无效或已过期'
      } else if (status === 503) {
        errorMessage.value = '鉴权服务不可用'
      } else {
        errorMessage.value = '登录失败'
      }
    } else {
      errorMessage.value = '登录失败'
    }
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const currentUser = await restoreAuthSession()
  if (currentUser) {
    await router.replace(getRedirectPath())
  }
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    radial-gradient(circle at top, rgba(255, 210, 155, 0.55), transparent 32%),
    linear-gradient(160deg, #f4efe6 0%, #dce8e3 48%, #edf4f8 100%);
}

.login-card {
  width: min(440px, 100%);
  padding: 32px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(20, 20, 20, 0.08);
  box-shadow: 0 24px 60px rgba(44, 52, 64, 0.16);
}

.eyebrow {
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.18em;
  color: #87623d;
}

h1 {
  margin: 10px 0 8px;
  font-size: 32px;
}

p {
  margin: 0;
  color: #5b6170;
}

.field {
  display: grid;
  gap: 8px;
  margin-top: 24px;
}

.field span {
  font-size: 14px;
  font-weight: 700;
}

input {
  width: 100%;
  border: 1px solid rgba(18, 24, 32, 0.12);
  border-radius: 14px;
  padding: 14px 16px;
  font-size: 16px;
  background: rgba(250, 250, 250, 0.95);
}

button {
  width: 100%;
  margin-top: 18px;
  border: 0;
  border-radius: 14px;
  padding: 14px 16px;
  font-size: 16px;
  font-weight: 800;
  color: #fff;
  background: linear-gradient(135deg, #1e5060 0%, #2f7662 100%);
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.7;
}

.error {
  margin-top: 14px;
  color: #b42318;
}
</style>
