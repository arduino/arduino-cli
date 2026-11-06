// LibA includes <B.h> only in A.cpp: the library discovery finds LibB after the sketch.
#include <A.h>
#if __has_include(<B.h>)
#include <C.h>
#endif

void setup() {
  a();
  c();
}

void loop() {}
