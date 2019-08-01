# Integration tests

This dir contains integration tests, the aim is to test the Command Line Interface and its output
from a pure user point of view.

## Installation

See also [Contributing][0].

```shell
cd test
virtualenv --python=python3 venv
source venv/bin/activate
pip install -r requirements.txt
```

## Running tests

To run all the tests:

```shell
pytest
```

To run specific modules:

```shell
pytest test_lib.py
```

To run very specific test functions:

```shell
pytest test_lib.py::test_list
```

[0]: ../CONTRIBUTING.md
