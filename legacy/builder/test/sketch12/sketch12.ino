// See: https://github.com/arduino/arduino-builder/issues/68

// The following avoid duplicate definitions of min and max
#undef min
#undef max

#include <memory>

void setup() {
  test();
}

void loop() {}

void test() {}

