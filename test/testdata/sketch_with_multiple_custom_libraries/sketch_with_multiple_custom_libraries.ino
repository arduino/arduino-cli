#include "lib1.h"
#include "lib2.h"

#if LIB2_SOME_SIZE != 42
#error should be 42 per global configuration
#endif

void setup() {
}

void loop() {
}
