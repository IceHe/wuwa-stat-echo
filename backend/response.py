class Success:
    def __init__(self, data: dict = None, message: str = ""):
        self.code = 200
        self.message = message
        self.data = data


class Error:
    def __init__(self, message: str = "", code: int = 500):
        self.code = code
        self.message = message


class Page:
    def __init__(
            self,
            message: str = "",
            data: list = None,
            data_total: int = 0,
            page: int = 0,
            page_size: int = 0,
    ):
        if data is None:
            data = []
        self.code = 200
        self.message = message
        self.data = data
        self.data_total = data_total
        self.page = page
        self.page_size = page_size
        self.page_total = data_total // page_size + 1
