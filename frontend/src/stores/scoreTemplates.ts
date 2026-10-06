import axios from 'axios'
import {reactive} from 'vue'

import {API_BASE_URL} from '@/stores/constants.ts'
import type {ResonatorTemplate} from '@/utils/echoScore'

const storageKey = 'wuwa-echo-score-template-config'
const contextKey = 'wuwa-echo-score-template-context'
const builtinVersion = 'builtin-xwuid-2026-09-14-compat-2026-10-06-v2'

const normalizeResonatorName = (resonator: string) => resonator === '清霄' ? '清宵' : resonator

// Keep the UI-side calculator aligned with the legacy weight retained by the
// backend when an older cached or fallback template is loaded.
const substatWeightOverrides: Record<string, Record<string, number>> = {
    丽贝卡: {攻击固定值: 0.11},
}

const applySubstatWeightOverrides = (templates: Record<string, ResonatorTemplate>) =>
    Object.fromEntries(Object.entries(templates).map(([key, template]) => {
        const name = normalizeResonatorName(template.name || key)
        const overrides = substatWeightOverrides[name]
        if (!overrides) {
            return [key, template]
        }
        return [key, {
            ...template,
            substat_weight: {
                ...(template.substat_weight || {}),
                ...overrides,
            },
        }]
    })) as Record<string, ResonatorTemplate>

const buildTemplate = (
    name: string,
    echoMax: Record<string, number>,
    mainstat: Record<string, number>,
    weights: Record<string, number>,
): ResonatorTemplate => {
    const baseWeights: Record<string, number> = {
        暴击: 2.0,
        暴击伤害: 1.0,
        攻击: 1.1,
        攻击固定值: 0.1,
    }
    Object.entries(weights).forEach(([key, value]) => {
        baseWeights[key] = value
    })
    return {
        name,
        echo_max_score: echoMax,
        mainstat_max_score: mainstat,
        substat_weight: baseWeights,
    }
}

const builtinTemplates: Record<string, ResonatorTemplate> = {
    '': buildTemplate('通用', {4: 80, 3: 80, 1: 80}, {
        '4C': 8.86,
        '3C属伤': 6.78,
        '3C攻击': 6.73,
        '3C其它': 1.57,
        '1C': 4.76
    }, {共鸣效率: 0.3, 普攻: 0.05, 重击: 0.05, 共鸣技能: 0.05, 共鸣解放: 0.05}),
    通用: buildTemplate('通用', {4: 80, 3: 80, 1: 80}, {
        '4C': 8.86,
        '3C属伤': 6.78,
        '3C攻击': 6.73,
        '3C其它': 1.57,
        '1C': 4.76
    }, {共鸣效率: 0.3, 普攻: 0.05, 重击: 0.05, 共鸣技能: 0.05, 共鸣解放: 0.05}),
    暗主: buildTemplate('暗主', {4: 82.527, 3: 78.527, 1: 74.977}, {
        '4C': 8.93,
        '3C属伤': 6.84,
        '3C攻击': 6.84,
        '3C其它': 1.59,
        '1C': 4.8
    }, {共鸣效率: 0.5, 普攻: 0.275, 共鸣技能: 0.22, 共鸣解放: 0.605}),
    椿: buildTemplate('椿', {4: 83.8, 3: 79.8, 1: 76.25}, {
        '4C': 8.79,
        '3C属伤': 6.8,
        '3C攻击': 6.8,
        '3C其它': 1.6,
        '1C': 4.72
    }, {共鸣效率: 0.15, 普攻: 0.715, 共鸣解放: 0.275}),
    珂莱塔: buildTemplate('珂莱塔', {4: 86.066, 3: 82.066, 1: 78.516}, {
        '4C': 8.56,
        '3C属伤': 6.54,
        '3C攻击': 6.54,
        '3C其它': 1.52,
        '1C': 4.58
    }, {共鸣效率: 0.2, 共鸣技能: 0.91}),
    今汐: buildTemplate('今汐', {4: 83.8, 3: 79.8, 1: 76.25}, {
        '4C': 8.79,
        '3C属伤': 6.72,
        '3C攻击': 6.72,
        '3C其它': 1.56,
        '1C': 4.72
    }, {共鸣效率: 0.25, 共鸣技能: 0.715, 共鸣解放: 0.33}),
    长离: buildTemplate('长离', {4: 83.17, 3: 79.17, 1: 75.62}, {
        '4C': 8.86,
        '3C属伤': 6.78,
        '3C攻击': 6.78,
        '3C其它': 1.57,
        '1C': 4.76
    }, {共鸣效率: 0.3, 共鸣技能: 0.66, 共鸣解放: 0.44}),
    坎特蕾拉: buildTemplate('坎特蕾拉', {4: 83.17, 3: 79.17, 1: 75.62}, {
        '4C': 8.86,
        '3C属伤': 6.78,
        '3C攻击': 6.78,
        '3C其它': 1.57,
        '1C': 4.76
    }, {共鸣效率: 0.5, 普攻: 0.66}),
    折枝: buildTemplate('折枝', {4: 81.89, 3: 77.89, 1: 74.34}, {
        '4C': 8.99,
        '3C属伤': 6.89,
        '3C攻击': 6.89,
        '3C其它': 1.6,
        '1C': 4.84
    }, {共鸣效率: 0.2, 普攻: 0.55, 重击: 0.22, 共鸣技能: 0.22}),
    忌炎: buildTemplate('忌炎', {4: 83.8, 3: 79.8, 1: 76.25}, {
        '4C': 8.79,
        '3C属伤': 6.72,
        '3C攻击': 6.72,
        '3C其它': 1.56,
        '1C': 4.72
    }, {共鸣效率: 0.3, 普攻: 0.165, 重击: 0.715, 共鸣技能: 0.33}),
    相里要: buildTemplate('相里要', {4: 83.8, 3: 79.8, 1: 76.25}, {
        '4C': 8.79,
        '3C属伤': 6.72,
        '3C攻击': 6.72,
        '3C其它': 1.56,
        '1C': 4.72
    }, {共鸣效率: 0.3, 普攻: 0.165, 共鸣技能: 0.22, 共鸣解放: 0.715}),
    洛可可: buildTemplate('洛可可', {4: 85.25, 3: 81.25, 1: 77.7}, {
        '4C': 8.64,
        '3C属伤': 6.6,
        '3C攻击': 6.6,
        '3C其它': 1.53,
        '1C': 4.63
    }, {共鸣效率: 0.3, 重击: 0.84}),
    布兰特: buildTemplate('布兰特', {4: 77.33, 3: 74.03, 1: 71.88}, {
        '4C': 8.17,
        '3C属伤': 6.31,
        '3C攻击': 6.31,
        '3C其它': 6.31,
        '1C': 5
    }, {攻击: 0.44, 攻击固定值: 0.044, 共鸣效率: 0.8, 普攻: 0.66, 共鸣解放: 0.165}),
    菲比: buildTemplate('菲比', {4: 78.76, 3: 74.76, 1: 71.21}, {
        '4C': 9.36,
        '3C属伤': 7.18,
        '3C攻击': 6.78,
        '3C其它': 1.57,
        '1C': 5.05
    }, {暴击: 1.58, 共鸣效率: 0.1, 普攻: 0.088, 重击: 0.66, 共鸣技能: 0.055, 共鸣解放: 0.187}),
    赞妮: buildTemplate('赞妮', {4: 83.8, 3: 79.8, 1: 76.25}, {
        '4C': 8.79,
        '3C属伤': 6.72,
        '3C攻击': 6.72,
        '3C其它': 1.56,
        '1C': 4.72
    }, {共鸣效率: 0.3, 重击: 0.715, 共鸣解放: 0.154}),
    夏空: buildTemplate('夏空', {4: 82.78, 3: 78.78, 1: 75.23}, {
        '4C': 8.9,
        '3C属伤': 6.81,
        '3C攻击': 6.81,
        '3C其它': 1.58,
        '1C': 4.78
    }, {共鸣效率: 0.3, 普攻: 0.506, 重击: 0.363, 共鸣解放: 0.627}),
    卡提希娅: buildTemplate('卡提希娅', {4: 79.726, 3: 76.871, 1: 78.986}, {
        '4C': 6.89,
        '3C属伤': 5.46,
        '3C攻击': 5.46,
        '3C其它': 0,
        '1C': 6.48
    }, {攻击: 0, 攻击固定值: 0, 生命: 1.1, 生命固定值: 0.01, 共鸣效率: 0.1, 普攻: 0.704, 共鸣解放: 0.308}),
    露帕: buildTemplate('露帕', {4: 84.059, 3: 80.059, 1: 76.509}, {
        '4C': 8.77,
        '3C属伤': 6.71,
        '3C攻击': 6.71,
        '3C其它': 1.56,
        '1C': 4.7
    }, {共鸣效率: 0.2, 普攻: 0.077, 重击: 0.055, 共鸣技能: 0.231, 共鸣解放: 0.737}),
    弗洛洛: buildTemplate('弗洛洛', {4: 84.059, 3: 80.059, 1: 76.509}, {
        '4C': 8.77,
        '3C属伤': 6.71,
        '3C攻击': 6.71,
        '3C其它': 1.56,
        '1C': 4.7
    }, {共鸣技能: 0.737}),
    奥古斯塔: buildTemplate('奥古斯塔', {4: 85.161, 3: 81.161, 1: 77.611}, {
        '4C': 8.65,
        '3C属伤': 6.62,
        '3C攻击': 6.62,
        '3C其它': 1.54,
        '1C': 4.63
    }, {重击: 0.832, 共鸣效率: 0.2}),
    尤诺: buildTemplate('尤诺', {4: 83.804, 3: 79.804, 1: 76.254}, {
        '4C': 8.79,
        '3C属伤': 6.72,
        '3C攻击': 6.72,
        '3C其它': 1.56,
        '1C': 4.72
    }, {共鸣效率: 0.2, 共鸣解放: 0.715}),
    嘉贝莉娜: buildTemplate('嘉贝莉娜', {4: 80.358, 3: 76.358, 1: 72.808}, {
        '4C': 8.79,
        '3C属伤': 7.03,
        '3C攻击': 7.03,
        '3C其它': 1.63,
        '1C': 4.94
    }, {共鸣效率: 0.2, 重击: 0.418}),
    仇远: buildTemplate('仇远', {4: 82.017, 3: 78.017, 1: 74.467}, {
        '4C': 8.98,
        '3C属伤': 6.88,
        '3C攻击': 6.88,
        '3C其它': 1.6,
        '1C': 4.83
    }, {重击: 0.561, 共鸣效率: 0.2}),
    琳奈: buildTemplate('琳奈', {4: 84.117, 3: 80.117, 1: 76.567}, {
        '4C': 8.75,
        '3C属伤': 6.7,
        '3C攻击': 6.7,
        '3C其它': 1.56,
        '1C': 4.7
    }, {攻击: 1.05, 普攻: 0.792, 共鸣解放: 0.253, 共鸣效率: 0.2}),
    莫宁: buildTemplate('莫宁', {4: 64.515, 3: 64.099, 1: 63.698}, {
        '4C': 9.71,
        '3C属伤': 5.85,
        '3C攻击': 0,
        '3C其它': 8.73,
        '1C': 8.47
    }, {防御: 1.25, 防御固定值: 0.1, 共鸣效率: 1.3, 暴击: 0.1, 暴击伤害: 0.3, 共鸣解放: 0.44}),
    陆赫斯: buildTemplate('陆赫斯', {4: 85.915, 3: 81.915, 1: 78.365}, {
        '4C': 8.58,
        '3C属伤': 6.55,
        '3C攻击': 6.55,
        '3C其它': 1.52,
        '1C': 4.59
    }, {攻击: 1.15, 普攻: 0.847, 共鸣效率: 0.15}),
    爱弥斯: buildTemplate('爱弥斯', {4: 85.642, 3: 81.642, 1: 78.092}, {
        '4C': 8.6,
        '3C属伤': 6.58,
        '3C攻击': 6.58,
        '3C其它': 1.53,
        '1C': 4.6
    }, {攻击固定值: 0.12, 共鸣解放: 0.77, 共鸣效率: 0.2}),
    西格莉卡: buildTemplate('西格莉卡', {4: 83.23, 3: 81.43, 1: 79.68}, {
        '4C': 5.28 + 2.25,
        '3C属伤': 5.06 + 1.53,
        '3C攻击': 5.06 + 1.53,
        '3C其它': 4.42 + 1.53,
        '1C': 5.64
    }, {共鸣效率: 0.8}),
    绯雪: buildTemplate('绯雪', {4: 83.589, 3: 81.53, 1: 82.63}, {
        '4C': 6.31 + 1.79,
        '3C属伤': 5.51 + 1.53,
        '3C攻击': 5.51 + 1.53,
        '3C其它': 1.53,
        '1C': 7.62
    }, {共鸣效率: 0.25, 共鸣解放: 0.8, 暴击: 1.5, 攻击: 1.5, 攻击固定值: 0.11}),
    达妮娅: buildTemplate('达妮娅', {4: 85.939, 3: 83.88, 1: 84.979}, {
        '4C': 6.14 + 1.74,
        '3C属伤': 5.36 + 1.49,
        '3C攻击': 5.36 + 1.49,
        '3C其它': 1.49,
        '1C': 7.41,
    }, {攻击: 1.2, 攻击固定值: 0.11, 共鸣效率: 0.2, 共鸣解放: 0.85}),
    丽贝卡: buildTemplate('丽贝卡', {4: 85.475, 3: 83.416, 1: 82.715}, {
        '4C': 5.18 + 1.47,
        '3C属伤': 5.39 + 1.49,
        '3C攻击': 5.39 + 1.49,
        '3C其它': 1.49,
        '1C': 6.52,
    }, {攻击: 1.2, 攻击固定值: 0.11, 共鸣效率: 0.25, 共鸣解放: 0.09, 普攻: 0.81}),
    露西: buildTemplate('露西', {4: 86.518, 3: 84.46, 1: 83.758}, {
        '4C': 6.1 + 1.73,
        '3C属伤': 5.32 + 1.47,
        '3C攻击': 5.32 + 1.47,
        '3C其它': 1.47,
        '1C': 6.44,
    }, {攻击: 1.2, 攻击固定值: 0.11, 共鸣效率: 0.2, 重击: 0.9}),
    校长霜渐: buildTemplate('校长霜渐', {4: 87.71, 3: 84.46, 1: 83.758}, {
        '4C': 6.27 + 2.13,
        '3C属伤': 5.32 + 1.47,
        '3C攻击': 5.32 + 1.47,
        '3C其它': 1.47,
        '1C': 6.44,
    }, {攻击: 1.2, 攻击固定值: 0.11, 普攻: 0.9}),
    校长声骸: buildTemplate('校长声骸', {4: 74.609, 3: 71.8, 1: 71.098}, {
        '4C': 7.07 + 2.51,
        '3C属伤': 6.26 + 1.74,
        '3C攻击': 6.26 + 1.74,
        '3C其它': 1.74,
        '1C': 7.59,
    }, {暴击: 1.4, 暴击伤害: 1, 攻击: 1.3, 攻击固定值: 0.12, 普攻: 0.2, 共鸣技能: 0.12}),
    千咲: buildTemplate('千咲', {4: 82.527, 3: 78.527, 1: 79.977}, {
        '4C': 5.49 + 2.27,
        '3C属伤': 5.25 + 1.59,
        '3C攻击': 5.25 + 1.59,
        '3C其它': 1.59,
        '1C': 4.8,
    }, {共鸣效率: 0.25, 共鸣解放: 0.605}),
    秧秧玄翎: buildTemplate('秧秧玄翎', {4: 86.473, 3: 82.473, 1: 78.923}, {
        '4C': 6.36 + 2.16,
        '3C属伤': 5 + 1.51,
        '3C攻击': 5 + 1.51,
        '3C其它': 1.51,
        '1C': 4.56,
    }, {暴击: 1.75, 攻击: 1.2, 攻击固定值: 0.12, 共鸣效率: 0.15, 普攻: 0.11, 重击: 0.968}),
    穗穗: buildTemplate('穗穗', {4: 54.427, 3: 52.478, 1: 58.154}, {
        '4C': 10.61,
        '3C属伤': 7.84 + 5.29,
        '3C攻击': 7.84 + 5.29,
        '3C其它': 7.84 + 5.29,
        '1C': 7.84 + 5.29,
    }, {暴击: 0.1, 暴击伤害: 0.33, 攻击: 0, 攻击固定值: 0, 共鸣效率: 1, 共鸣技能: 0.33, 生命: 1.2, 生命固定值: 0.01}),
    清宵: buildTemplate('清宵', {4: 83.051, 3: 79.801, 1: 79.1}, {
        '4C': 6.62 + 2.25,
        '3C属伤': 5.63 + 1.56,
        '3C攻击': 5.63 + 1.56,
        '3C其它': 1.56,
        '1C': 6.82,
    }, {暴击: 1.7, 暴击伤害: 1, 攻击: 1.2, 攻击固定值: 0.11, 共鸣效率: 0.2, 重击: 0.77, 普攻: 0.44, 共鸣解放: 0.605}),
}

type RemoteTemplatePayload = {
    version?: string
    updated_at?: string
    templates?: ResonatorTemplate[]
    resonator_templates?: Record<string, ResonatorTemplate>
}

const readStorage = () => {
    if (typeof window === 'undefined') {
        return null
    }
    const raw = window.localStorage.getItem(storageKey)
    if (!raw) {
        return null
    }
    try {
        return JSON.parse(raw) as RemoteTemplatePayload
    } catch {
        return null
    }
}

const writeStorage = (payload: RemoteTemplatePayload) => {
    if (typeof window === 'undefined') {
        return
    }
    window.localStorage.setItem(storageKey, JSON.stringify(payload))
}

const readContext = () => {
    if (typeof window === 'undefined') {
        return null
    }
    const raw = window.localStorage.getItem(contextKey)
    if (!raw) {
        return null
    }
    try {
        return JSON.parse(raw) as { resonator?: string; cost?: string; mainstat?: string }
    } catch {
        return null
    }
}

const writeContext = (payload: { resonator: string; cost: string; mainstat: string }) => {
    if (typeof window === 'undefined') {
        return
    }
    window.localStorage.setItem(contextKey, JSON.stringify(payload))
}

export const scoreTemplateState = reactive({
    templates: builtinTemplates as Record<string, ResonatorTemplate>,
    version: builtinVersion,
    updatedAt: '',
    source: 'builtin',
    loading: false,
    initialized: false,
    error: '',
})

export const scoreTemplateContext = reactive({
    resonator: '',
    cost: '',
    mainstat: '',
})

const applyTemplatePayload = (payload: RemoteTemplatePayload | null, source: string) => {
    const templateRows = payload?.templates
    const templates = templateRows?.length
      ? Object.fromEntries(templateRows.map((template) => [template.name || '', template]))
      : payload?.resonator_templates
    if (!templates || Object.keys(templates).filter(Boolean).length === 0) {
        return false
    }
    if (templates['清霄'] && !templates['清宵']) {
        templates['清宵'] = {...templates['清霄'], name: '清宵'}
        delete templates['清霄']
    }
    scoreTemplateState.templates = applySubstatWeightOverrides(templates)
    scoreTemplateState.version = payload?.version || builtinVersion
    scoreTemplateState.updatedAt = payload?.updated_at || ''
    scoreTemplateState.source = source
    scoreTemplateState.error = ''
    scoreTemplateState.initialized = true
    return true
}

const hydrate = () => {
    const stored = readStorage()
    if (!applyTemplatePayload(stored, 'cache')) {
        scoreTemplateState.initialized = true
    }
    const storedContext = readContext()
    if (storedContext) {
        scoreTemplateContext.resonator = normalizeResonatorName(storedContext.resonator || '')
        scoreTemplateContext.cost = storedContext.cost || ''
        scoreTemplateContext.mainstat = storedContext.mainstat || ''
    }
}

hydrate()

export const setScoreTemplateContext = (payload: { resonator?: string; cost?: string; mainstat?: string }) => {
    const next = {
        resonator: normalizeResonatorName(typeof payload.resonator === 'string' ? payload.resonator : scoreTemplateContext.resonator),
        cost: typeof payload.cost === 'string' ? payload.cost : scoreTemplateContext.cost,
        mainstat: typeof payload.mainstat === 'string' ? payload.mainstat : scoreTemplateContext.mainstat,
    }
    scoreTemplateContext.resonator = next.resonator
    scoreTemplateContext.cost = next.cost
    scoreTemplateContext.mainstat = next.mainstat
    writeContext(next)
}

export const getResonatorTemplate = (resonator: string) =>
    scoreTemplateState.templates[normalizeResonatorName(resonator)] ||
    scoreTemplateState.templates['通用'] ||
    scoreTemplateState.templates[''] ||
    null

export const getScoreTemplateOptions = () =>
    Object.entries(scoreTemplateState.templates)
        .filter(([name, template]) => Boolean(name) && Boolean(template?.name))
        .map(([name, template]) => ({ name, template }))

export const sortScoreTemplateOptions = (
    options: Array<{ name: string; template: ResonatorTemplate }>,
) => {
    // Newest limited characters appear first; permanent and launch characters
    // are assigned to the separate trailing group below.
    const releaseOrder = [
        '心', '锁暝', '景燃', '清宵', '秧秧玄翎', '穗穗', '洛瑟菈', '绯雪', '珂莱塔',
        '折枝', '弗洛洛', '暗主', '坎特蕾拉', '洛可可', '椿', '长离', '夏空', '西格莉卡',
        '仇远', '尤诺', '卡提希娅', '风主', '今汐', '相里要', '奥古斯塔', '雷主',
        '露西', '丽贝卡', '达妮娅', '爱弥斯', '莫宁', '嘉贝莉娜', '露帕', '布兰特',
        '陆·赫斯', '琳奈', '千咲', '赞妮', '菲比', '守岸人', '灯灯', '光主',
        '秧秧', '秋水', '忌炎', '鉴心', '吟霖', '卡卡罗', '维里奈', '凌阳', '安可',
        '桃祈', '丹瑾', '渊武', '莫特斐', '炽霞', '白芷', '散华', '釉瑚',
    ]
    const permanentFiveStars = new Set(['维里奈', '卡卡罗', '凌阳', '安可', '鉴心', '吟霖'])
    const initialFourStars = new Set(['秧秧', '白芷', '炽霞', '散华', '莫特斐', '渊武', '丹瑾', '桃祈'])
    const releaseIndex = new Map(releaseOrder.map((name, index) => [name, index]))
    const baseName = (name: string) => {
        const normalized = name.replace(/（\d+）$/, '')
        if (normalized.startsWith('洛瑟菈-')) return '洛瑟菈'
        if (normalized === '秧秧·玄翎') return '秧秧玄翎'
        if (normalized === '角色') return '角色'
        return normalized
    }
    const rank = (name: string) => {
        const base = baseName(name)
        if (permanentFiveStars.has(base) || initialFourStars.has(base)) {
            return 10000 + (permanentFiveStars.has(base) ? 0 : 1000) + (releaseIndex.get(base) ?? 0)
        }
        if (base === '角色') return 20000
        return releaseIndex.get(base) ?? 5000
    }
    return [...options].sort((a, b) => {
        const rankDiff = rank(a.name) - rank(b.name)
        return rankDiff || a.name.localeCompare(b.name, 'zh-Hans-CN')
    })
}

export const refreshScoreTemplates = async (force = false) => {
    if (scoreTemplateState.loading) {
        return
    }
    scoreTemplateState.loading = true
    scoreTemplateState.error = ''
    try {
        const response = await axios.get(`${API_BASE_URL}/score_templates${force ? '?force=1' : ''}`)
        const payload = response.data?.data as RemoteTemplatePayload | undefined
        if (!applyTemplatePayload(payload || null, 'remote')) {
            throw new Error('invalid score template payload')
        }
        writeStorage(payload || {})
    } catch (error) {
        scoreTemplateState.error = error instanceof Error ? error.message : 'failed to load score templates'
    } finally {
        scoreTemplateState.loading = false
        scoreTemplateState.initialized = true
    }
}

export const ensureScoreTemplatesLoaded = async () => {
    if (scoreTemplateState.loading) {
        return
    }
    await refreshScoreTemplates(false)
}
