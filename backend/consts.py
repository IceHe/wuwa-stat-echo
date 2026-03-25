from collections import defaultdict

from model import EchoLog
from util import bit_pos

SUBSTAT_BIT_WIDTH = 13


class SubstatValueE:
    def __init__(
            self,
            substat_number: int,
            value_number: int,
            value_desc: str,
            value_desc_full: str,
            value: float,
    ):
        self.substat_number = substat_number
        self.value_number = value_number
        self.value_bitmap = 1 << substat_number | 1 << (SUBSTAT_BIT_WIDTH + value_number)
        self.value_desc = value_desc
        self.value_desc_full = value_desc_full
        self.value = value


class SubstatE:
    def __init__(
            self,
            number: int,
            name: str,
            name_cn: str,
            value_dict: dict[str, SubstatValueE],
    ):
        self.number = number
        self.bitmap = 1 << number
        self.name_cn = name_cn
        self.name = name
        self.value_dict = value_dict


def init_substat_dict():
    substat_dict = defaultdict(None)
    all_substat = [
        SubstatE(number=0, name_cn="暴击", name="Crit.Rate", value_dict={
            "0": SubstatValueE(0, 0, "6.3%", "暴击 6.3%", 6.3),
            "1": SubstatValueE(0, 1, "6.9%", "暴击 6.9%", 6.9),
            "2": SubstatValueE(0, 2, "7.5%", "暴击 7.5%", 7.5),
            "3": SubstatValueE(0, 3, "8.1%", "暴击 8.1%", 8.1),
            "4": SubstatValueE(0, 4, "8.7%", "暴击 8.7%", 8.7),
            "5": SubstatValueE(0, 5, "9.3%", "暴击 9.3%", 9.3),
            "6": SubstatValueE(0, 6, "9.9%", "暴击 9.9%", 9.9),
            "7": SubstatValueE(0, 7, "10.5%", "暴击 10.5%", 10.5),
        }),
        SubstatE(number=1, name_cn="暴击伤害", name="Crit.DMG", value_dict={
            "0": SubstatValueE(1, 0, "12.6%", "暴击伤害 12.6%", 12.6),
            "1": SubstatValueE(1, 1, "13.8%", "暴击伤害 13.8%", 13.8),
            "2": SubstatValueE(1, 2, "15.0%", "暴击伤害 15.0%", 15.0),
            "3": SubstatValueE(1, 3, "16.2%", "暴击伤害 16.2%", 16.2),
            "4": SubstatValueE(1, 4, "17.4%", "暴击伤害 17.4%", 17.4),
            "5": SubstatValueE(1, 5, "18.6%", "暴击伤害 18.6%", 18.6),
            "6": SubstatValueE(1, 6, "19.8%", "暴击伤害 19.8%", 19.8),
            "7": SubstatValueE(1, 7, "21.0%", "暴击伤害 21.0%", 21.0),
        }),
        SubstatE(number=2, name_cn="攻击", name="ATK.Rate", value_dict={
            "0": SubstatValueE(2, 0, "6.4%", "攻击 6.4%", 6.4),
            "1": SubstatValueE(2, 1, "7.1%", "攻击 7.1%", 7.1),
            "2": SubstatValueE(2, 2, "7.9%", "攻击 7.9%", 7.9),
            "3": SubstatValueE(2, 3, "8.6%", "攻击 8.6%", 8.6),
            "4": SubstatValueE(2, 4, "9.4%", "攻击 9.4%", 9.4),
            "5": SubstatValueE(2, 5, "10.1%", "攻击 10.1%", 10.1),
            "6": SubstatValueE(2, 6, "10.9%", "攻击 10.9%", 10.9),
            "7": SubstatValueE(2, 7, "11.6%", "攻击 11.6%", 11.6),
        }),
        SubstatE(number=3, name_cn="防御", name="DEF.Rate", value_dict={
            "0": SubstatValueE(3, 0, "8.1%", "防御 8.1%", 8.1),
            "1": SubstatValueE(3, 1, "9.0%", "防御 9.0%", 9.0),
            "2": SubstatValueE(3, 2, "10.0%", "防御 10.0%", 10.0),
            "3": SubstatValueE(3, 3, "10.9%", "防御 10.9%", 10.9),
            "4": SubstatValueE(3, 4, "11.8%", "防御 11.8%", 11.8),
            "5": SubstatValueE(3, 5, "12.8%", "防御 12.8%", 12.8),
            "6": SubstatValueE(3, 6, "13.8%", "防御 13.8%", 13.8),
            "7": SubstatValueE(3, 7, "14.7%", "防御 14.7%", 14.7),
        }),
        SubstatE(number=4, name_cn="生命", name="HP.Rate", value_dict={
            "0": SubstatValueE(4, 0, "6.4%", "生命 6.4%", 6.4),
            "1": SubstatValueE(4, 1, "7.1%", "生命 7.1%", 7.1),
            "2": SubstatValueE(4, 2, "7.9%", "生命 7.9%", 7.9),
            "3": SubstatValueE(4, 3, "8.6%", "生命 8.6%", 8.6),
            "4": SubstatValueE(4, 4, "9.4%", "生命 9.4%", 9.4),
            "5": SubstatValueE(4, 5, "10.1%", "生命 10.1%", 10.1),
            "6": SubstatValueE(4, 6, "10.9%", "生命 10.9%", 10.9),
            "7": SubstatValueE(4, 7, "11.6%", "生命 11.6%", 11.6),
        }),
        SubstatE(number=5, name_cn="攻击固定值", name="ATK.Fixed", value_dict={
            "0": SubstatValueE(5, 0, "30", "攻击固定值 30", 30),
            "1": SubstatValueE(5, 1, "40", "攻击固定值 40", 40),
            "2": SubstatValueE(5, 2, "50", "攻击固定值 50", 50),
            "3": SubstatValueE(5, 3, "60", "攻击固定值 60", 60),
        }),
        SubstatE(number=6, name_cn="防御固定值", name="DEF.Fixed", value_dict={
            "0": SubstatValueE(6, 0, "40", "防御固定值 40", 40),
            "1": SubstatValueE(6, 1, "50", "防御固定值 50", 50),
            "2": SubstatValueE(6, 2, "60", "防御固定值 60", 60),
            "3": SubstatValueE(6, 3, "70", "防御固定值 70", 70),
        }),
        SubstatE(number=7, name_cn="生命固定值", name="HP.Fixed", value_dict={
            "0": SubstatValueE(7, 0, "320", "生命固定值 320", 320),
            "1": SubstatValueE(7, 1, "360", "生命固定值 360", 360),
            "2": SubstatValueE(7, 2, "390", "生命固定值 390", 390),
            "3": SubstatValueE(7, 3, "430", "生命固定值 430", 430),
            "4": SubstatValueE(7, 4, "470", "生命固定值 470", 470),
            "5": SubstatValueE(7, 5, "510", "生命固定值 510", 510),
            "6": SubstatValueE(7, 6, "540", "生命固定值 540", 540),
            "7": SubstatValueE(7, 7, "580", "生命固定值 580", 580),
        }),
        SubstatE(number=8, name_cn="共鸣效率", name="Energy Regen", value_dict={
            "0": SubstatValueE(8, 0, "6.8%", "共鸣效率 6.8%", 6.8),
            "1": SubstatValueE(8, 1, "7.6%", "共鸣效率 7.6%", 7.6),
            "2": SubstatValueE(8, 2, "8.4%", "共鸣效率 8.4%", 8.4),
            "3": SubstatValueE(8, 3, "9.2%", "共鸣效率 9.2%", 9.2),
            "4": SubstatValueE(8, 4, "10.0%", "共鸣效率 10.0%", 10.0),
            "5": SubstatValueE(8, 5, "10.8%", "共鸣效率 10.8%", 10.8),
            "6": SubstatValueE(8, 6, "11.6%", "共鸣效率 11.6%", 11.6),
            "7": SubstatValueE(8, 7, "12.4%", "共鸣效率 12.4%", 12.4),
        }),
        SubstatE(number=9, name_cn="普攻", name="Basic ATK", value_dict={
            "0": SubstatValueE(9, 0, "6.4%", "普攻 6.4%", 6.4),
            "1": SubstatValueE(9, 1, "7.1%", "普攻 7.1%", 7.1),
            "2": SubstatValueE(9, 2, "7.9%", "普攻 7.9%", 7.9),
            "3": SubstatValueE(9, 3, "8.6%", "普攻 8.6%", 8.6),
            "4": SubstatValueE(9, 4, "9.4%", "普攻 9.4%", 9.4),
            "5": SubstatValueE(9, 5, "10.1%", "普攻 10.1%", 10.1),
            "6": SubstatValueE(9, 6, "10.9%", "普攻 10.9%", 10.9),
            "7": SubstatValueE(9, 7, "11.6%", "普攻 11.6%", 11.6),
        }),
        SubstatE(number=10, name_cn="重击", name="Heavy ATK", value_dict={
            "0": SubstatValueE(10, 0, "6.4%", "重击 6.4%", 6.4),
            "1": SubstatValueE(10, 1, "7.1%", "重击 7.1%", 7.1),
            "2": SubstatValueE(10, 2, "7.9%", "重击 7.9%", 7.9),
            "3": SubstatValueE(10, 3, "8.6%", "重击 8.6%", 8.6),
            "4": SubstatValueE(10, 4, "9.4%", "重击 9.4%", 9.4),
            "5": SubstatValueE(10, 5, "10.1%", "重击 10.1%", 10.1),
            "6": SubstatValueE(10, 6, "10.9%", "重击 10.9%", 10.9),
            "7": SubstatValueE(10, 7, "11.6%", "重击 11.6%", 11.6),
        }),
        SubstatE(number=11, name_cn="共鸣技能", name="Skill", value_dict={
            "0": SubstatValueE(11, 0, "6.4%", "共鸣技能 6.4%", 6.4),
            "1": SubstatValueE(11, 1, "7.1%", "共鸣技能 7.1%", 7.1),
            "2": SubstatValueE(11, 2, "7.9%", "共鸣技能 7.9%", 7.9),
            "3": SubstatValueE(11, 3, "8.6%", "共鸣技能 8.6%", 8.6),
            "4": SubstatValueE(11, 4, "9.4%", "共鸣技能 9.4%", 9.4),
            "5": SubstatValueE(11, 5, "10.1%", "共鸣技能 10.1%", 10.1),
            "6": SubstatValueE(11, 6, "10.9%", "共鸣技能 10.9%", 10.9),
            "7": SubstatValueE(11, 7, "11.6%", "共鸣技能 11.6%", 11.6),
        }),
        SubstatE(number=12, name_cn="共鸣解放", name="Liberation", value_dict={
            "0": SubstatValueE(12, 0, "6.4%", "共鸣解放 6.4%", 6.4),
            "1": SubstatValueE(12, 1, "7.1%", "共鸣解放 7.1%", 7.1),
            "2": SubstatValueE(12, 2, "7.9%", "共鸣解放 7.9%", 7.9),
            "3": SubstatValueE(12, 3, "8.6%", "共鸣解放 8.6%", 8.6),
            "4": SubstatValueE(12, 4, "9.4%", "共鸣解放 9.4%", 9.4),
            "5": SubstatValueE(12, 5, "10.1%", "共鸣解放 10.1%", 10.1),
            "6": SubstatValueE(12, 6, "10.9%", "共鸣解放 10.9%", 10.9),
            "7": SubstatValueE(12, 7, "11.6%", "共鸣解放 11.6%", 11.6),
        }),
    ]
    for substat in all_substat:
        substat_dict[str(substat.number)] = substat
    return substat_dict


SUBSTAT_DICT = init_substat_dict()

# 声骸从 x 级升到 y 级需要多少声骸经验
EXP = defaultdict(lambda: defaultdict(int))
EXP[0][1] = 4500
EXP[0][2] = 16500
EXP[0][3] = 40000
EXP[0][4] = 79500
EXP[0][5] = 143000
# EXP[0][1] = 4500
EXP[1][2] = 12000
EXP[2][3] = 23500
EXP[3][4] = 39500
EXP[4][5] = 63500
EXP[2][4] = 63000

# 调谐器回收率
TUNER_RECYCLING_RATE = 0.3
TUNER_RECYCLED_PER_SUBSTAT = 3

# 密音筒声骸经验
EXP_GOLD = 5000
EXP_PURPLE = 2000
EXP_BLUE = 1000
EXP_GREEN = 500

# 声骸经验回收率
EXP_RETURN = defaultdict(int)
EXP_RETURN[1] = EXP_PURPLE * 1 + EXP_BLUE * 1  # = 3,000
EXP_RETURN[2] = EXP_GOLD * 2 + EXP_PURPLE * 1  # = 12,000
EXP_RETURN[3] = EXP_GOLD * 6  # = 30,000
EXP_RETURN[4] = EXP_GOLD * 11 + EXP_PURPLE * 2 + EXP_GREEN * 1  # = 5,9500
EXP_RETURN[5] = EXP_GOLD * 21 + EXP_BLUE * 1 + EXP_GREEN * 1  # = 10,6500


# 5级回收率  = 3000 / 4500 = 0.667
# 10级回收率 = 12000 / 16500 = 0.727
# 15级回收率 = 30000 / 40000 = 0.750
# 20级回收率 = 59500 / 79500 = 0.748
# 25级回收率 = 106500 / 143000 = 0.745


class EchoScore:
    def __init__(self):
        self.name = ''
        self.substat1 = 0.0
        self.substat2 = 0.0
        self.substat3 = 0.0
        self.substat4 = 0.0
        self.substat5 = 0.0
        self.substat_all = 0.0


class ResonatorTemplate:
    def __init__(
            self,
            name: str,
            echo_max_score: dict[str, float],
            mainstat_max_score: dict[str, float],
            substat_weight: dict[str, float],
    ):
        self.name: str = name

        self.echo_max_score: dict[str, float] = defaultdict(float)
        for x, y in echo_max_score.items():
            self.echo_max_score[str(x)] = y

        self.mainstat_max_score: dict[str, float] = defaultdict(float)
        for x, y in mainstat_max_score.items():
            self.mainstat_max_score[str(x)] = y

        self.substat_weight: dict[str, float] = defaultdict(float)
        self.substat_weight['暴击'] = 2.0
        self.substat_weight['暴击伤害'] = 1.0
        self.substat_weight['攻击'] = 1.1
        self.substat_weight['攻击固定值'] = 0.1
        for x, y in substat_weight.items():
            self.substat_weight[str(x)] = y

    def substat_score(self, substat: int) -> float:
        substat_num = bit_pos(substat)
        substat_element = SUBSTAT_DICT[str(substat_num)]
        value_num = bit_pos(substat >> SUBSTAT_BIT_WIDTH)
        value = substat_element.value_dict[str(value_num)].value
        return self.substat_weight[substat_element.name_cn] * value

    def echo_score(self, echo_log: EchoLog, cost: str) -> EchoScore:
        score = EchoScore()
        max_score = self.echo_max_score[cost[:1]]
        if max_score > 0:
            if echo_log.substat1 > 0:
                score.substat1 = round(self.substat_score(echo_log.substat1) / max_score * 50, 2)
            if echo_log.substat2 > 0:
                score.substat2 = round(self.substat_score(echo_log.substat2) / max_score * 50, 2)
            if echo_log.substat3 > 0:
                score.substat3 = round(self.substat_score(echo_log.substat3) / max_score * 50, 2)
            if echo_log.substat4 > 0:
                score.substat4 = round(self.substat_score(echo_log.substat4) / max_score * 50, 2)
            if echo_log.substat5 > 0:
                score.substat5 = round(self.substat_score(echo_log.substat5) / max_score * 50, 2)
            score_total = score.substat1 + score.substat2 + score.substat3 + score.substat4 + score.substat5
            score.substat_all = round(self.mainstat_max_score[cost] + score_total, 2)
        return score


def init_resonator_templates():
    resonator_templates = defaultdict(lambda: ResonatorTemplate(
        name='通用',
        echo_max_score={'4': 80, '3': 80, '1': 80},
        mainstat_max_score={
            '4C': 6.61 + 2.25,
            '3C属伤': 5.21 + 1.57,
            '3C攻击': 5.16 + 1.57,  # FIXME
            '3C其它': 1.57,
            '1C': 4.76,
        },
        substat_weight={
            '共鸣效率': 0.3,
            '普攻': 0.05,
            '重击': 0.05,
            '共鸣技能': 0.05,
            '共鸣解放': 0.05,
        }
    ))
    resonator_templates['暗主'] = ResonatorTemplate(
        name='暗主',
        echo_max_score={'4': 82.527, '3': 78.527, '1': 74.977},
        mainstat_max_score={
            '4C': 6.66 + 2.27,
            '3C属伤': 5.25 + 1.59,
            '3C攻击': 5.25 + 1.59,  # FIXME
            '3C其它': 1.59,
            '1C': 4.8,
        },
        substat_weight={
            '共鸣效率': 0.5,
            '普攻': 0.275,
            '共鸣技能': 0.22,
            '共鸣解放': 0.605,
        }
    )
    resonator_templates['椿'] = ResonatorTemplate(
        name='椿',
        echo_max_score={'4': 83.8, '3': 79.8, '1': 76.25},
        mainstat_max_score={
            '4C': 6.56 + 2.23,
            '3C属伤': 5.2 + 1.6,
            '3C攻击': 5.2 + 1.6,  # FIXME
            '3C其它': 1.6,
            '1C': 4.72,
        },
        substat_weight={
            '共鸣效率': 0.15,
            '普攻': 0.715,
            '共鸣解放': 0.275,
        }
    )
    resonator_templates['珂莱塔'] = ResonatorTemplate(
        name='珂莱塔',
        echo_max_score={'4': 86.066, '3': 82.066, '1': 78.516},
        mainstat_max_score={
            '4C': 6.39 + 2.17,
            '3C属伤': 5.02 + 1.52,
            '3C攻击': 5.02 + 1.52,  # FIXME
            '3C其它': 1.52,
            '1C': 4.58,
        },
        substat_weight={
            '共鸣效率': 0.2,
            '共鸣技能': 0.91,
        }
    )
    resonator_templates['今汐'] = ResonatorTemplate(
        name='今汐',
        echo_max_score={'4': 83.8, '3': 79.8, '1': 76.25},
        mainstat_max_score={
            '4C': 6.56 + 2.23,
            '3C属伤': 5.16 + 1.56,
            '3C攻击': 5.16 + 1.56,  # FIXME
            '3C其它': 1.56,
            '1C': 4.72,
        },
        substat_weight={
            '共鸣效率': 0.25,
            '共鸣技能': 0.715,
            '共鸣解放': 0.33,
        }
    )
    resonator_templates['长离'] = ResonatorTemplate(
        name='长离',
        echo_max_score={'4': 83.17, '3': 79.17, '1': 75.62},
        mainstat_max_score={
            '4C': 6.61 + 2.25,
            '3C属伤': 5.21 + 1.57,
            '3C攻击': 5.21 + 1.57,  # FIXME
            '3C其它': 1.57,
            '1C': 4.76,
        },
        substat_weight={
            '共鸣效率': 0.3,
            '共鸣技能': 0.66,
            '共鸣解放': 0.44,
        }
    )
    resonator_templates['坎特蕾拉'] = ResonatorTemplate(
        name='坎特蕾拉',
        echo_max_score={'4': 83.17, '3': 79.17, '1': 75.62},
        mainstat_max_score={
            '4C': 6.61 + 2.25,
            '3C属伤': 5.21 + 1.57,
            '3C攻击': 5.21 + 1.57,  # FIXME
            '3C其它': 1.57,
            '1C': 4.76,
        },
        substat_weight={
            '共鸣效率': 0.5,
            '普攻': 0.66,
        }
    )
    resonator_templates['折枝'] = ResonatorTemplate(
        name='折枝',
        echo_max_score={'4': 81.89, '3': 77.89, '1': 74.34},
        mainstat_max_score={
            '4C': 6.71 + 2.28,
            '3C属伤': 5.29 + 1.6,
            '3C攻击': 5.29 + 1.6,  # FIXME
            '3C其它': 1.6,
            '1C': 4.84,
        },
        substat_weight={
            '共鸣效率': 0.2,
            '普攻': 0.55,
            '重击': 0.22,
            '共鸣技能': 0.22,
        }
    )
    resonator_templates['忌炎'] = ResonatorTemplate(
        name='忌炎',
        echo_max_score={'4': 83.8, '3': 79.8, '1': 76.25},
        mainstat_max_score={
            '4C': 6.56 + 2.23,
            '3C属伤': 5.16 + 1.56,
            '3C攻击': 5.16 + 1.56,  # FIXME
            '3C其它': 1.56,
            '1C': 4.72,
        },
        substat_weight={
            '共鸣效率': 0.3,
            '普攻': 0.165,
            '重击': 0.715,
            '共鸣技能': 0.33,
        }
    )
    resonator_templates['相里要'] = ResonatorTemplate(
        name='相里要',
        echo_max_score={'4': 83.8, '3': 79.8, '1': 76.25},
        mainstat_max_score={
            '4C': 6.56 + 2.23,
            '3C属伤': 5.16 + 1.56,
            '3C攻击': 5.16 + 1.56,  # FIXME
            '3C其它': 1.56,
            '1C': 4.72,
        },
        substat_weight={
            '共鸣效率': 0.3,
            '普攻': 0.165,
            '共鸣技能': 0.22,
            '共鸣解放': 0.715,
        }
    )
    resonator_templates['洛可可'] = ResonatorTemplate(
        name='洛可可',
        echo_max_score={'4': 85.25, '3': 81.25, '1': 77.7},
        mainstat_max_score={
            '4C': 6.45 + 2.19,
            '3C属伤': 5.07 + 1.53,
            '3C攻击': 5.07 + 1.53,
            '3C其它': 1.53,
            '1C': 4.63,
        },
        substat_weight={
            '共鸣效率': 0.3,
            '重击': 0.84,
        }
    )
    resonator_templates['布兰特'] = ResonatorTemplate(
        name='布兰特',
        echo_max_score={'4': 77.33, '3': 74.03, '1': 71.88},
        mainstat_max_score={
            '4C': 7.11 + 1.06,
            '3C属伤': 5.57 + 0.74,
            '3C攻击': 5.57 + 0.74,  # FIXME
            '3C其它': 5.57 + 0.74,
            '1C': 5,
        },
        substat_weight={
            '攻击': 0.44,
            '攻击固定值': 0.044,
            '共鸣效率': 0.8,
            '普攻': 0.66,
            '共鸣解放': 0.165,
        }
    )
    resonator_templates['菲比'] = ResonatorTemplate(
        name='菲比',
        echo_max_score={'4': 78.76, '3': 74.76, '1': 71.21},
        mainstat_max_score={
            '4C': 6.98 + 2.38,
            '3C属伤': 5.51 + 1.67,
            '3C攻击': 5.21 + 1.57,
            '3C其它': 1.57,
            '1C': 5.05,
        },
        substat_weight={
            '暴击': 1.58,
            '共鸣效率': 0.1,
            '普攻': 0.088,
            '重击': 0.66,
            '共鸣技能': 0.055,
            '共鸣解放': 0.187,
        }
    )
    resonator_templates['赞妮'] = ResonatorTemplate(
        name='赞妮',
        echo_max_score={'4': 83.8, '3': 79.8, '1': 76.25},
        mainstat_max_score={
            '4C': 6.56 + 2.23,
            '3C属伤': 5.16 + 1.56,
            '3C攻击': 5.16 + 1.56,  # FIXME
            '3C其它': 1.56,
            '1C': 4.72,
        },
        substat_weight={
            '共鸣效率': 0.3,
            '重击': 0.715,
            '共鸣解放': 0.154,
        }
    )
    resonator_templates['夏空'] = ResonatorTemplate(
        name='夏空',
        echo_max_score={'4': 82.78, '3': 78.78, '1': 75.23},
        mainstat_max_score={
            '4C': 6.64 + 2.26,
            '3C属伤': 5.23 + 1.58,
            '3C攻击': 5.23 + 1.58,  # FIXME
            '3C其它': 1.58,
            '1C': 4.78,
        },
        substat_weight={
            '共鸣效率': 0.3,
            '普攻': 0.506,
            '重击': 0.363,
            '共鸣解放': 0.627,
        }
    )
    resonator_templates['卡提希娅'] = ResonatorTemplate(
        name='卡提希娅',
        echo_max_score={'4': 79.726, '3': 76.871, '1': 78.986},
        mainstat_max_score={
            '4C': 6.89,
            '3C属伤': 5.46,
            '3C攻击': 5.46,  # FIXME
            '3C其它': 0,
            '1C': 4.32 + 2.16,
        },
        substat_weight={
            '攻击': 0,
            '攻击固定值': 0,
            '生命': 1.1,
            '生命固定值': 0.01,
            '共鸣效率': 0.1,
            '普攻': 0.704,
            '共鸣解放': 0.308,
        }
    )
    resonator_templates['露帕'] = ResonatorTemplate(
        name='露帕',
        echo_max_score={'4': 84.059, '3': 80.059, '1': 76.509},
        mainstat_max_score={
            '4C': 6.54 + 2.23,
            '3C属伤': 5.15 + 1.56,
            '3C攻击': 5.15 + 1.56,
            '3C其它': 1.56,
            '1C': 4.7,
        },
        substat_weight={
            '共鸣效率': 0.2,
            '普攻': 0.077,
            '重击': 0.055,
            '共鸣技能': 0.231,
            '共鸣解放': 0.737,
        }
    )
    resonator_templates['弗洛洛'] = ResonatorTemplate(
        name='弗洛洛',
        echo_max_score={'4': 84.059, '3': 80.059, '1': 76.509},
        mainstat_max_score={
            '4C': 6.54 + 2.23,
            '3C属伤': 5.15 + 1.56,
            '3C攻击': 5.15 + 1.56,  # FIXME
            '3C其它': 1.56,  # FIXME
            '1C': 4.7,
        },
        substat_weight={
            '共鸣技能': 0.737,
        }
    )
    resonator_templates['奥古斯塔'] = ResonatorTemplate(
        name='奥古斯塔',
        echo_max_score={'4': 85.161, '3': 81.161, '1': 77.611},
        mainstat_max_score={
            '4C': 6.45 + 2.2,
            '3C属伤': 5.08 + 1.54,
            '3C攻击': 5.08 + 1.54,  # FIXME
            '3C其它': 1.54,  # FIXME
            '1C': 4.63,
        },
        substat_weight={
            '重击': 0.832,
            '共鸣效率': 0.2,
        }
    )
    resonator_templates['尤诺'] = ResonatorTemplate(
        name='尤诺',
        echo_max_score={'4': 83.804, '3': 79.804, '1': 76.254},
        mainstat_max_score={
            '4C': 6.56 + 2.23,
            '3C属伤': 5.16 + 1.56,
            '3C攻击': 5.16 + 1.56,  # FIXME
            '3C其它': 1.56,
            '1C': 4.72,
        },
        substat_weight={
            '共鸣效率': 0.2,
            '共鸣解放': 0.715,
        }
    )
    resonator_templates['嘉贝莉娜'] = ResonatorTemplate(
        name='嘉贝莉娜',
        echo_max_score={'4': 80.358, '3': 76.358, '1': 72.808},
        mainstat_max_score={
            '4C': 6.56 + 2.23,
            '3C属伤': 5.4+1.63,
            '3C攻击': 5.4+1.63,  # FIXME
            '3C其它': 1.63,
            '1C': 4.94,
        },
        substat_weight={
            '共鸣效率': 0.2,
            '重击': 0.418,
        }
    )
    resonator_templates['陆赫斯'] = ResonatorTemplate(
        name='陆赫斯',
        echo_max_score={'4': 85.915, '3': 81.915, '1': 78.365},
        mainstat_max_score={
            '4C': 6.4 + 2.18,
            '3C属伤': 5.03 + 1.52,
            '3C攻击': 5.03 + 1.52,  # FIXME
            '3C其它': 1.52,
            '1C': 4.59,
        },
        substat_weight={
            '攻击': 1.15,
            '普攻': 0.847,
            '共鸣效率': 0.15,
        }
    )
    resonator_templates['爱弥斯'] = ResonatorTemplate(
        name='爱弥斯',
        echo_max_score={'4': 85.642, '3': 81.642, '1': 78.092},
        mainstat_max_score={
            '4C': 6.42 + 2.18,
            '3C属伤': 5.05 + 1.53,
            '3C攻击': 5.05 + 1.53,  # FIXME
            '3C其它': 1.53,
            '1C': 4.6,
        },
        substat_weight={
            '攻击固定值': 0.12,
            '共鸣解放': 0.77,
            '共鸣效率': 0.2,
        }
    )
    return resonator_templates


RESONATOR_TEMPLATES = init_resonator_templates()
