#!/usr/bin/env python
import msgpack
import pickle


def tenTimes(x):
    return x * 10


def logger():
    return "called"


def twoLogger():
    return "called1", "called2"


def plusSuffix(s):
    return s + "_through_python"


def loadMsgPack(row):
    maps = msgpack.unpackb(row)

    model = ""
    byte = maps['model']
    if byte == '':
        model = "TEST"
    else:
        model = pickle.loads(byte)
        model += "_re"

    pic = pickle.dumps(model, -1)

    retmap = {'model': pic, 'log': 'done'}

    packed = msgpack.packb(retmap)
    return packed


def dict(arg):
    return arg['string'] + str(arg['int']) + str(arg['byte'])
