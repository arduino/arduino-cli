# Arduino Sketch Build Process

## What is a sketch.

A sketch is a program made in C/C++ using the Arduino SDK (libraries and platform). A sketch is made by a main file with
the extension `.ino` and, possibly, other source files with the `.ino` extension, or the classic `.cpp` and `.h`. The
sketch must be in a directory named exactly as the main `.ino` file, for example a sketch named `Blink.ino` must be
placed in a directory named `Blink`.

### .ino file composition.

If a sketch contains multiple `.ino` files, these are concatenated in a single file before the actual build. The order
of concatenation is as follow:

1. The first line will always be `#include <Arduino.h>`.
1. The main sketch file.
1. The remaining files in alphabetic order.

For example a `Test` sketch composed by these two files:

- Test/Test.ino:
  ```
  void setup() {
    Serial.begin(9600);
  }
  ```
- Test/OtherFile.ino
  ```
  void loop() {
    Serial.println("hello!");
    delay(1000);
  }
  ```

will be concatenated as the following "merged" `Test.ino.cpp` file:

```
#include <Arduino.h>
#line 1 "/home/Arduino/Test/Test.ino"
void setup() {
  Serial.begin(9600);
}

#line 1 "/home/Arduino/Test/OtherFile.ino"
void loop() {
  Serial.println("hello!");
  delay(1000);
}
```

### The .ino "language" and the Arduino Preprocessing.

The `.ino` file has the same syntax of a C/C++ program with the main difference that it should not require forward
declaration of the functions. A program like this:

```c++
// Uncomment the following line to make it valid C/C++.
//void other_func();

void func() {
    other_func();
}

void other_func() {
    // ....
}
```

is not valid C/C++, but it's a valid sketch `.ino`. The Arduino CLI build system **should** add the missing forward
declaration for you in the so called "Arduino Preprocessing" phase. It "should" because adding the forward declarations
it's not always possible and it may fail, in particular if the sketch uses complex C++ syntax and constructs. In such
cases, the user must manually add the missing forward declarations or use directly `.cpp` and `.h` files.

## The sketch build process

The Arduino CLI build is composed by multiple phases:

1. [Preparation of the build path, with sketch merge and composition](#preparation-of-the-build-path-with-sketch-merge-and-composition).
1. [Auto discovery of used libraries](#auto-discovery-of-used-libraries).
1. Arduino Preprocessing of the sketch.
1. Compilation of the sketch.
1. Compilation of the used libraries.
1. Compilation of the core platform.
1. Linking of all the compiled files (object files) into the final executable.

### Preparation of the build path, with sketch merge and composition.

A sub-directory called `sketch` is created in the build folder, with a copy of the sketch combined .ino files, and all
the .cpp and .h files.

```
build/sketch/Test.ino.cpp
             other.cpp
             other.h
```

The `Test.ino.cpp` is composed following the rules described in [.ino file composition](#ino-file-composition) section.

### Auto discovery of used libraries.

In this phase the build system tests if the sketch uses external libraries that are not part of the build. The algorithm
in principle is very simple: it just tries to C/C++ Preprocess each compilation unit, and check if this operation
results in a `Missing include <xxx.h>` error. In such cases the builder adds a library that provides the missing `xxx.h`
include and retries, until the C/C++ Preprocess is succesful or there are errors different from `Missing include ...`.

The C/C++ Preprocessing consists in the macro expansion of the C language, i.e. the substitution of the `#include` and
`#define` and `#ifdef/endif` macros. We are doing only this operation during the auto discovery of used libraries,
becuase it's much faster than fully compiling a compilation unit, and is sufficient to detect missing includes, which is
what we are interested in.

The command line flags to tell `gcc` to make a Preprocess-only instead of a compile are
`gcc -E ... file.cpp -o /dev/null`:

- `-E` will tell `gcc` to run only the macro expansion
- `-o /dev/null` will not write any output. To add a library path the flag `-I/path/to/the/library` is used, the `-I`
  flag may be used multiple times.

The full algorithm for auto discovery of libraries used in the builder is the following:

1. Initialize a **queue** of file to check for missing libraries, and put all the sketch sources in the queue.
1. Add the platform core paths in the **include paths list**.
1. Pick the first source file in the queue.
1. Repeat the following:
   1. Do macro expansion on the source file (C/C++ Preprocessing)
   1. If the operation is successful:
      - Remove the file from the list;
      - If there are no file left in the queue we are done;
      - Otherwise pick the next file in the queue and repeat.
   1. If the operation fails with an error different from `Missing Include...`:
      - Print the error and exit with a failed compile.
   1. If the operation fails with a `Missing include <lib.h>` error:
      - Search for a library that provides `lib.h`
      - If such library do not exists, print the error and exit with a failed compile.
      - Add the new library path to include path list.
      - Add the new library source files to the queue of files to check for missing libraries.
      - Repeat the macro expansion on the current source file.

## TODO: Add missing sections
