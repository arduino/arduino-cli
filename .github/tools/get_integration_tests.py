from pathlib import Path
import json


if __name__ == "__main__":
    import sys

    tests_path = sys.argv[1]

    test_files = [str(f) for f in Path(tests_path).glob("test_*.py")]
    print(json.dumps(test_files))
