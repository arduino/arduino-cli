#include <testlib1.h>
#include "subfolder/other.h"
#include "src/subfolder/other.h"

MyClass myClass;

void setup() {
    myClass.init ( &Serial );
}

void loop() {
}
