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

    @staticmethod
    def get_str():
        return "staticmethod"

    @staticmethod
    def get_instance():
        return PythonTest3()

    def get_instance_str(self):
        return "instance method"
