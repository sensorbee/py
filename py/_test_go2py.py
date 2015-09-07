#!/usr/bin/env python


def go2py(arg):
    ret = arg['string']
    ret += '_' + str(arg['int'])
    ret += '_' + str(arg['float'])
    ret += '_' + str(arg['byte'])
    ret += '_' + str(arg['bool'])
    ret += '_' + str(arg['null'])
    for v in arg['array']:
        ret += '_' + str(v)
    return ret
