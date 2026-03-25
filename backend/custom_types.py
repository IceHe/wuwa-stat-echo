from collections import defaultdict


class SubstatItem:
    def __init__(
            self,
            number: int,
            name: str,
            name_cn: str,
    ):
        self.number = number
        self.name = name
        self.name_cn = name_cn
        self.total = 0
        self.percent = 0.0
        self.value_dict = defaultdict(None)
        self.value_dict['all'] = SubstatValueStat(0, 'all', '所有档位')
        self.cur_pos_percent = ''


class SubstatValueStat:
    def __init__(
            self,
            value_number: int,
            value_desc: str,
            value_desc_full: str,
    ):
        self.value_number = value_number
        self.value_desc = value_desc
        self.value_desc_full = value_desc_full
        self.total = 0
        self.percent = 0.0
        self.position_dict: dict[str, SubstatValuePositionStat] = defaultdict(None)
        self.position_dict['0'] = SubstatValuePositionStat(0)
        self.position_dict['1'] = SubstatValuePositionStat(1)
        self.position_dict['2'] = SubstatValuePositionStat(2)
        self.position_dict['3'] = SubstatValuePositionStat(3)
        self.position_dict['4'] = SubstatValuePositionStat(4)


class SubstatValuePositionStat:
    def __init__(
            self,
            position: int,
    ):
        self.position = position
        self.total = 0
        self.percent = 0.0
