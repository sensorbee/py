#!/usr/bin/env python


def go2py_tostr(arg):
    return str(arg)


def go2py_toutf8(arg):
    return arg.decode('utf-8')


def go2py_mapinmap(arg):
    ret = arg['string']
    ret += '_' + arg['map']['instr']
    for v in arg['array']:
        ret += '_' + str(v)
    return ret


def go2py_arrayinmap(arg):
    ret = arg[0][0]
    ret += '_' + str(arg[0][1])
    ret += '_' + arg[1]['map']
    return ret
