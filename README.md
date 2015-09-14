pystate supports Python module and instance using `Py_Object`.

## Set up to link python

Go codes in `py` package use cgo and call `Py_Object`, cgo code is here:

```go
/*
#cgo pkg-config: python-2.7
#include "Python.h"
*/
```

Currently pystate use pkg-config to link "Python.h". User needs to set up pkg-config and "python-2.7.pc".

* [TODO] currently only support darwin and linux, need to support windows
* [TODO] support python3

### Example

If user uses pyenv, "python-2.7.pc" would be installed at `~/.pyenv/versions/<version>/lib/pkgconfig`, and user can use this file.

Default pkg-config path (`PKG_CONFIG_PATH`) would be `/usr/local/lib/pkgconfig/`, then

```sh
ln -s ~/.pyenv/versions/<version>/lib/pkgconfig/python-2.7.pc /usr/local/lib/pkgconfig/
```
