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

def return_bytearray():
    return bytearray('abcdefg')

def return_array():
    return [1, 2, {"key": 3}]

def return_map():
    return {"key1": 123, "key2": "str"}

def return_nested_map():
    return {"key1": {"key2": 123}}

def return_none():
    return None

def return_timestamp():
    return datetime.datetime(2015, 4, 1, 14, 27, 0, 500*1000, None)
