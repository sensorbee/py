class TestClass(object):
    def __init__(self):
        pass

    def write(self, value):
        return 'called! arg is "{}"'.format(str(value))


class TestClass2(object):
    def __init__(self, **params):
        self.v1 = params['v1']
        self.v2 = params['v2']

    def confirm(self):
        return 'constructor init arg is v1={}, v2={}'.format(
            str(self.v1), str(self.v2))


class TestClass3(object):
    def __init__(self, a, b='b', **c):
        self.a = a
        self.b = b
        self.c = c

    def confirm(self):
        return 'constructor init arg is a={}, b={}, c={}'.format(
            self.a, self.b, self.c)
