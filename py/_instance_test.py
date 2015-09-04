#!/usr/bin/env python


class PythonTest():

    a = ''

    def __init__(self):
        self.a = 'initialized'

    def logger(self, s):
        self.a += '_' + s
        return self.a
