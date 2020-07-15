# Integration tests

This dir contains integration tests, aimed to test the Command Line Interface
and its output from a pure user point of view.

## Installation

See also [Contributing][0].

To run the integration tests you must install [Poetry][poetry-website].

```shell
pip3 install --user poetry
```

For more installation options read the [official documentation][poetry-docs].

## Running tests

To run all the tests from the project's root folder:

```shell
task test-integration
```

This will create and install all necessary dependencies if not already existing and then run integrations tests.

To run specific modules you must run `pytest` from the virtual environment created by Poetry.
If dependencies have not already been installed first run `poetry install`.

```shell
poetry run pytest test/test_lib.py
```

To run very specific test functions:

```shell
poetry run pytest test/test_lib.py::test_list
```

You can avoid writing the `poetry run` prefix each time by creating a new shell inside the virtual environment:

```shell
poetry shell
pytest test_lib.py
pytest test_lib.py::test_list
```

[0]: ../docs/CONTRIBUTING.md
[poetry-website]: https://python-poetry.org/
[poetry-docs]: https://python-poetry.org/docs/
