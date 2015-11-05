/*
Package py provides tools to interact with Python. This package is mainly
for SensorBee but can be used for other purposes.

py package depends on py/mainthread package. mainthread requires that it
initializes the Python interpreter so that it can keep the main thread under
its control. As a result, py package may not work with other packages which
initializes Python.
*/
package py

/*
Internal: GIL(mainthread.Exec) rules

1. All private functions or methods MUST NOT acquire the GIL.
2. All public functions or methods MUST acquire the GIL.
3. However, public functions or methods having a "NoGIL" suffix MUST NOT
   acquire the GIL.

Therefore, public functions or methods must not call other public ones which
don't have "NoGIL" suffix.
*/
