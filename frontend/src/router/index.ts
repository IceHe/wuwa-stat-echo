import {createRouter, createWebHistory} from 'vue-router'
import { getStoredAuthToken } from '@/auth'
import HomeView from '@/views/HomeView.vue'
import AnalysisView from '@/views/AnalysisView.vue'
import EchoView from "@/views/EchoView.vue";
import EchoViewerView from "@/views/EchoViewerView.vue";
import SubstatView from "@/views/SubstatView.vue";
import EchoBoardView from "@/views/EchoBoardView.vue";
import EchoDcritCountView from "@/views/EchoDcritCountView.vue";
import LoginView from '@/views/LoginView.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            redirect: '/home',
        },
        {
            path: '/login',
            name: 'login',
            component: LoginView,
            meta: { public: true },
        },
        {
            path: '/home',
            name: 'home',
            component: HomeView,
        },
        {
            path: '/substat',
            name: 'substat',
            component: SubstatView,
        },
        {
            path: '/echo',
            name: 'echo',
            component: EchoView,
        },
        {
            path: '/analysis',
            name: 'analysis',
            component: AnalysisView,
        },
        {
            path: '/echo_board',
            name: 'echo_board',
            component: EchoBoardView,
        },
        {
            path: '/echo_dcrit_count',
            name: 'echo_dcrit_count',
            component: EchoDcritCountView,
        },
        {
            path: '/echo-viewer',
            name: 'echo-viewer',
            component: EchoViewerView,
            meta: { public: true },
        },
    ],
})

router.beforeEach((to) => {
    if (to.meta.public) {
        return true
    }

    if (getStoredAuthToken()) {
        return true
    }

    return {
        name: 'login',
        query: {
            redirect: to.fullPath,
        },
    }
})

export default router
