import six


class TestClass(object):

    @staticmethod
    def create():
        self = TestClass()
        return self

    def write(self, value):
        return 'called! arg is "{}"'.format(str(value))


class TestClass2(object):

    @staticmethod
    def create(**params):
        self = TestClass2()
        self.v1 = params['v1']
        self.v2 = params['v2']
        return self

    def confirm(self):
        return 'constructor init arg is v1={}, v2={}'.format(
            str(self.v1), str(self.v2))


class TestClass3(object):

    @staticmethod
    def create(a, b='b', **c):
        self = TestClass3()
        self.a = a
        self.b = b
        self.c = c
        return self

    def confirm(self):
        return 'constructor init arg is a={}, b={}, c={}'.format(
            self.a, self.b, self.c)


class TestClass4(object):

    @staticmethod
    def create(*args, **params):
        self = TestClass4()
        self.params = params
        return self

    @staticmethod
    def load(filepath, *args, **kwargs):
        with open(filepath, 'r') as f:
            return six.moves.cPickle.load(f)

    def modify_params(self):
        self.params["a"] = 2
        self.params["b"] = "moge"

    def confirm(self):
        return self.params

    def save(self, filepath, *args, **kwargs):
        with open(filepath, 'w') as f:
            six.moves.cPickle.dump(self, f)


class TestClassTerminateError(object):

    @staticmethod
    def create():
        self = TestClassTerminateError()
        return self

    def terminate(self):
        return 1 / 0  # cause ZeroDivisionError on purpose
