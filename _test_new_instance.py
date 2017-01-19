#!/usr/bin/env python


class PythonTest():

    a = ''

    def __init__(self):
        self.a = 'initialized'

    def logger(self, s):
        self.a += '_' + s
        return self.a


class PythonTest2(object):

    def __init__(self, s):
        self.a = s

    def get_a(self):
        return self.a


class PythonTest3():
    v = "class_value"

    @staticmethod
    def get_static_value():
        return PythonTest3.v

    @staticmethod
    def get_instance():
        ins = PythonTest3()
        ins.val1 = "test1"
        return ins

    @classmethod
    def get_class_value(cls):
        return cls.v

    def get_instance_str(self):
        return "instance method " + self.val1

    @staticmethod
    def get_instance2(a, b=5, **c):
        ins = PythonTest3()
        ins.a = a
        ins.b = b
        ins.c1 = c['v1']
        return ins

    def confirm(self):
        return str(self.a) + '_' + str(self.b) + '_' + str(self.c1)


class ChildClass(PythonTest3):
    v = "instance_value"
    # ChildClass.get_class_value() will return "instance_value"


class PythonTestForKwd(object):

    def __init__(self, a, b=5, **c):
        self.a = a
        self.b = b
        self.c = c['c'] if 'c' in c else ''
        self.d = c['d'] if 'd' in c else ''
        self.e = c['e'] if 'e' in c else ''

    def confirm_init(self):
        return str(self.a) + '_' + str(self.b) + '_' + str(self.c) + '_' + \
            str(self.d) + '_' + str(self.e)
