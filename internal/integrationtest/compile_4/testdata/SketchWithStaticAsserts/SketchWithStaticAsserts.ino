// https://github.com/arduino/arduino-builder/issues/68

const int a = 10;
const int b = 20;

static_assert(a < b, "bar");

void setup() {
  test();
}

void loop() {
}

void test() {
}
