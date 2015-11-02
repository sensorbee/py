#!/usr/bin/env python


class negator(object):
    def __call__(self, x):
        return -x


def divideByZero(n):
    return n / 0


class alwaysFail(object):
    def __call__(self):
        return divideByZero(42)

not_func_attr = 'test'
