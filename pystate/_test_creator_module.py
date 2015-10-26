class TestClass(object):
    def __init__(self):
        pass

    def write(self, value):
        return 'called! arg is "{}"'.format(str(value))


class TestClass2(object):
    def __init__(self, params):
        self.v1 = params['v1']

    def confirm(self):
        return 'constructor init arg is "{}"'.format(str(self.v1))
