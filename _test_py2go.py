#!/usr/bin/env python
import datetime


def return_true():
    return True


def return_false():
    return False


def return_int():
    return 123


def return_float():
    return 1.0


def return_string():
    return "ABC"


def return_unicode():
    return u"hello"


def return_bytearray():
    return bytearray('abcdefg')


def return_array():
    return [1, 2, {"key": 3}]


def return_map():
    return {"key1": 123, u"key2": "str"}


def return_nested_map():
    return {"key1": {"key2": 123}}


def return_none():
    return None


def return_timestamp():
    return datetime.datetime(2015, 4, 1, 14, 27, 0, 500*1000, None)


class TEST_TZ(datetime.tzinfo):
    def utcoffset(self, dt):
        return datetime.timedelta(hours=9, minutes=3)

    def dst(self, dt):
        return datetime.timedelta(0)

    def tzname(self, dt):
        return 'TEST'


def return_timestamp_with_tz():
    return datetime.datetime(2015, 4, 1, 14, 27, 0, 500*1000, TEST_TZ())


def return_onetuple():
    return ('a', {'key1': 1}, [1, 2])


def return_astuple():
    return 'a', {'key1': 1}, [1, 2]


def return_object():
    class FailureTest(object):
        def __init__(self):
            pass

    return FailureTest()
