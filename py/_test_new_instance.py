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
        return PythonTest3()

    @classmethod
    def get_class_value(cls):
        return cls.v

    def get_instance_str(self):
        return "instance method"


class ChildClass(PythonTest3):
    v = "instance_value"
    # ChildClass.get_class_value() will return "instance_value"
