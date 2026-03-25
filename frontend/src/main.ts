import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import { installAxiosAuth } from './auth'
import router from './router'

const app = createApp(App)

installAxiosAuth()

app.use(createPinia())
app.use(router)

app.mount('#app')
