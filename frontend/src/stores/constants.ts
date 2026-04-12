const API_PORT = '8888'
const IS_PUBLIC_ECHO_HOST = typeof window !== 'undefined' && window.location.hostname === 'echo.icehe.life'
const API_HOST = typeof window === 'undefined'
    ? '127.0.0.1'
    : window.location.hostname

export const API_BASE_URL = IS_PUBLIC_ECHO_HOST
    ? `${window.location.origin}/api`
    : `${window.location.protocol}//${API_HOST}:${API_PORT}`

export const getSubstatColor = (bits: number) => {
    for (let i = 0; i < 13; i++) {
        if ((bits >> i) & 1) {
            return SUBSTAT[i].font_color
        }
    }
    return ''
}

export const CLASS_COLORS: Record<string, string> = {
    听唤语义之愿: '#6c63a8',
    流金溯真之式: '#e9b439',
    长路启航之星: 'red',
    星构寻辉之环: 'green',
    逆光跃彩之约: '#d96ea8',
    命理崩毁之弦: 'black',
    焚羽猎魔之影: 'red',
    息界同调之律: '#6fdfd9',
    荣斗铸锋之冠: '#e9b439',
    失序彼岸之梦: '#4e3d60',
    奔狼燎原: 'red',
    荣光之旅: '#6fdfd9',
    凝夜白霜: 'blue',
    熔山裂谷: 'red',
    彻空冥雷: 'purple',
    啸谷长风: '#6fdfd9',
    浮星祛暗: '#a17819',
    沉日劫明: '#4e3d60',
    隐世回光: 'green',
    轻云出月: 'gray',
    不绝余音: '#9a1b4a',
    凌冽决断之心: '#2960ce',
    此间永驻之光: '#e9b439',
    幽夜隐匿之帷: '#704688',
    高天共奏之曲: '#253d82',
    无惧浪涛之勇: '#898785',
    流云逝尽之空: '#196965',
}


export const CLASSES: string[] = [
    '听唤语义之愿',
    '流金溯真之式',
    '长路启航之星',
    '星构寻辉之环',
    '逆光跃彩之约',
    '命理崩毁之弦',
    '焚羽猎魔之影',
    '息界同调之律',
    '荣斗铸锋之冠',
    '失序彼岸之梦',
    '奔狼燎原',
    '荣光之旅',
    '凝夜白霜',
    '熔山裂谷',
    '彻空冥雷',
    '啸谷长风',
    '浮星祛暗',
    '沉日劫明',
    '隐世回光',
    '轻云出月',
    '不绝余音',
    '凌冽决断之心',
    '此间永驻之光',
    '幽夜隐匿之帷',
    '高天共奏之曲',
    '无惧浪涛之勇',
    '流云逝尽之空',
]

export const RESONATORS: string[] = [
    '西格莉卡',
    '爱弥斯',
    '陆赫斯',
    '莫宁',
    '琳奈',
    '仇远',
    '嘉贝莉娜',
    '尤诺',
    '奥古斯塔',
    '弗洛洛',
    '露帕',
    '卡提希娅',
    '椿',
    '珂莱塔',
    '今汐',
    '长离',
    '坎特蕾拉',
    '折枝',
    '忌炎',
    '相里要',
    '洛可可',
    '布兰特',
    '菲比',
    '赞妮',
    '夏空',
    '暗主',
]

class Substat {
    public bitmap: number

    constructor(
        public num: number,
        public name: string,
        public font_color: string,
    ) {
        this.bitmap = 1 << num
    }
}

// export const SUBSTAT: Substat[] = [
//   new Substat(0, '暴击', '#CC3333'),
//   new Substat(1, '暴击伤害', 'orange'),
//   new Substat(2, '攻击', '#dc266a'),
//   new Substat(3, '防御', '#00CC00'),
//   new Substat(4, '生命', '#0099CC'),
//   new Substat(5, '攻击固定值', 'purple'),
//   new Substat(6, '防御固定值', '#dc266a'),
//   new Substat(7, '生命固定值', '#6666CC'),
//   new Substat(8, '共鸣效率', '#CC3333'),
//   new Substat(9, '普攻', 'orange'),
//   new Substat(10, '重击', '#dc266a'),
//   new Substat(11, '共鸣技能', '#00CC00'),
//   new Substat(12, '共鸣解放', '#0099CC'),
// ]

export const SUBSTAT: Substat[] = [
    new Substat(0, '暴击', '#CC3333'),
    new Substat(1, '暴伤', 'orange'),
    new Substat(2, '攻击', '#dc266a'),
    new Substat(3, '防御', '#00CC00'),
    new Substat(4, '生命', '#0099CC'),
    new Substat(5, '小攻', 'purple'),
    new Substat(6, '小防', '#dc266a'),
    new Substat(7, '小生', '#6666CC'),
    new Substat(8, '共效', '#CC3333'),
    new Substat(9, '普攻', 'orange'),
    new Substat(10, '重击', '#dc266a'),
    new Substat(11, '共技', '#00CC00'),
    new Substat(12, '共解', '#0099CC'),
]

class SubstatValue {
    constructor(
        public substat_number: number,
        public value_number: number,
        public desc: string,
        public desc_full: string,
    ) {}
}

export const SUBSTAT_VALUE_MAP: Record<number, SubstatValue[]> = {
    0: [
        new SubstatValue(0, 0, '6.3%', '暴击 6.3%'),
        new SubstatValue(0, 1, '6.9%', '暴击 6.9%'),
        new SubstatValue(0, 2, '7.5%', '暴击 7.5%'),
        new SubstatValue(0, 3, '8.1%', '暴击 8.1%'),
        new SubstatValue(0, 4, '8.7%', '暴击 8.7%'),
        new SubstatValue(0, 5, '9.3%', '暴击 9.3%'),
        new SubstatValue(0, 6, '9.9%', '暴击 9.9%'),
        new SubstatValue(0, 7, '10.5%', '暴击 10.5%'),
    ],
    1: [
        new SubstatValue(1, 0, '12.6%', '暴击伤害 12.6%'),
        new SubstatValue(1, 1, '13.8%', '暴击伤害 13.8%'),
        new SubstatValue(1, 2, '15.0%', '暴击伤害 15.0%'),
        new SubstatValue(1, 3, '16.2%', '暴击伤害 16.2%'),
        new SubstatValue(1, 4, '17.4%', '暴击伤害 17.4%'),
        new SubstatValue(1, 5, '18.6%', '暴击伤害 18.6%'),
        new SubstatValue(1, 6, '19.8%', '暴击伤害 19.8%'),
        new SubstatValue(1, 7, '21.0%', '暴击伤害 21.0%'),
    ],
    2: [
        new SubstatValue(2, 0, '6.4%', '攻击 6.4%'),
        new SubstatValue(2, 1, '7.1%', '攻击 7.1%'),
        new SubstatValue(2, 2, '7.9%', '攻击 7.9%'),
        new SubstatValue(2, 3, '8.6%', '攻击 8.6%'),
        new SubstatValue(2, 4, '9.4%', '攻击 9.4%'),
        new SubstatValue(2, 5, '10.1%', '攻击 10.1%'),
        new SubstatValue(2, 6, '10.9%', '攻击 10.9%'),
        new SubstatValue(2, 7, '11.6%', '攻击 11.6%'),
    ],
    3: [
        new SubstatValue(3, 0, '8.1%', '防御 8.1%'),
        new SubstatValue(3, 1, '9.0%', '防御 9.0%'),
        new SubstatValue(3, 2, '10.0%', '防御 10.0%'),
        new SubstatValue(3, 3, '10.9%', '防御 10.9%'),
        new SubstatValue(3, 4, '11.8%', '防御 11.8%'),
        new SubstatValue(3, 5, '12.8%', '防御 12.8%'),
        new SubstatValue(3, 6, '13.8%', '防御 13.8%'),
        new SubstatValue(3, 7, '14.7%', '防御 14.7%'),
    ],
    4: [
        new SubstatValue(4, 0, '6.4%', '生命 6.4%'),
        new SubstatValue(4, 1, '7.1%', '生命 7.1%'),
        new SubstatValue(4, 2, '7.9%', '生命 7.9%'),
        new SubstatValue(4, 3, '8.6%', '生命 8.6%'),
        new SubstatValue(4, 4, '9.4%', '生命 9.4%'),
        new SubstatValue(4, 5, '10.1%', '生命 10.1%'),
        new SubstatValue(4, 6, '10.9%', '生命 10.9%'),
        new SubstatValue(4, 7, '11.6%', '生命 11.6%'),
    ],
    5: [
        new SubstatValue(5, 0, '30', '攻击固定值 30'),
        new SubstatValue(5, 1, '40', '攻击固定值 40'),
        new SubstatValue(5, 2, '50', '攻击固定值 50'),
        new SubstatValue(5, 3, '60', '攻击固定值 60'),
    ],
    6: [
        new SubstatValue(6, 0, '40', '防御固定值 40'),
        new SubstatValue(6, 1, '50', '防御固定值 50'),
        new SubstatValue(6, 2, '60', '防御固定值 60'),
        new SubstatValue(6, 3, '70', '防御固定值 70'),
    ],
    7: [
        new SubstatValue(7, 0, '320', '生命固定值 320'),
        new SubstatValue(7, 1, '360', '生命固定值 360'),
        new SubstatValue(7, 2, '390', '生命固定值 390'),
        new SubstatValue(7, 3, '430', '生命固定值 430'),
        new SubstatValue(7, 4, '470', '生命固定值 470'),
        new SubstatValue(7, 5, '510', '生命固定值 510'),
        new SubstatValue(7, 6, '540', '生命固定值 540'),
        new SubstatValue(7, 7, '580', '生命固定值 580'),
    ],
    8: [
        new SubstatValue(8, 0, '6.8%', '共鸣效率 6.8%'),
        new SubstatValue(8, 1, '7.6%', '共鸣效率 7.6%'),
        new SubstatValue(8, 2, '8.4%', '共鸣效率 8.4%'),
        new SubstatValue(8, 3, '9.2%', '共鸣效率 9.2%'),
        new SubstatValue(8, 4, '10.0%', '共鸣效率 10.0%'),
        new SubstatValue(8, 5, '10.8%', '共鸣效率 10.8%'),
        new SubstatValue(8, 6, '11.6%', '共鸣效率 11.6%'),
        new SubstatValue(8, 7, '12.4%', '共鸣效率 12.4%'),
    ],
    9: [
        new SubstatValue(9, 0, '6.4%', '普攻 6.4%'),
        new SubstatValue(9, 1, '7.1%', '普攻 7.1%'),
        new SubstatValue(9, 2, '7.9%', '普攻 7.9%'),
        new SubstatValue(9, 3, '8.6%', '普攻 8.6%'),
        new SubstatValue(9, 4, '9.4%', '普攻 9.4%'),
        new SubstatValue(9, 5, '10.1%', '普攻 10.1%'),
        new SubstatValue(9, 6, '10.9%', '普攻 10.9%'),
        new SubstatValue(9, 7, '11.6%', '普攻 11.6%'),
    ],
    10: [
        new SubstatValue(10, 0, '6.4%', '重击 6.4%'),
        new SubstatValue(10, 1, '7.1%', '重击 7.1%'),
        new SubstatValue(10, 2, '7.9%', '重击 7.9%'),
        new SubstatValue(10, 3, '8.6%', '重击 8.6%'),
        new SubstatValue(10, 4, '9.4%', '重击 9.4%'),
        new SubstatValue(10, 5, '10.1%', '重击 10.1%'),
        new SubstatValue(10, 6, '10.9%', '重击 10.9%'),
        new SubstatValue(10, 7, '11.6%', '重击 11.6%'),
    ],
    11: [
        new SubstatValue(11, 0, '6.4%', '共鸣技能 6.4%'),
        new SubstatValue(11, 1, '7.1%', '共鸣技能 7.1%'),
        new SubstatValue(11, 2, '7.9%', '共鸣技能 7.9%'),
        new SubstatValue(11, 3, '8.6%', '共鸣技能 8.6%'),
        new SubstatValue(11, 4, '9.4%', '共鸣技能 9.4%'),
        new SubstatValue(11, 5, '10.1%', '共鸣技能 10.1%'),
        new SubstatValue(11, 6, '10.9%', '共鸣技能 10.9%'),
        new SubstatValue(11, 7, '11.6%', '共鸣技能 11.6%'),
    ],
    12: [
        new SubstatValue(12, 0, '6.4%', '共鸣解放 6.4%'),
        new SubstatValue(12, 1, '7.1%', '共鸣解放 7.1%'),
        new SubstatValue(12, 2, '7.9%', '共鸣解放 7.9%'),
        new SubstatValue(12, 3, '8.6%', '共鸣解放 8.6%'),
        new SubstatValue(12, 4, '9.4%', '共鸣解放 9.4%'),
        new SubstatValue(12, 5, '10.1%', '共鸣解放 10.1%'),
        new SubstatValue(12, 6, '10.9%', '共鸣解放 10.9%'),
        new SubstatValue(12, 7, '11.6%', '共鸣解放 11.6%'),
    ],
}

export const ECHO_COST: string[] = [
    '4C',
    '3C属伤',
    '3C攻击',
    '3C其它',
    '1C',
]
