def bit_count(bits: int) -> int:
    cnt = 0
    while bits:
        cnt += 1
        bits &= bits - 1
    return cnt


def bit_pos(bits: int) -> int:
    pos = 0
    while bits:
        if bits & 1:
            return pos
        bits >>= 1
        pos += 1
    return -1
