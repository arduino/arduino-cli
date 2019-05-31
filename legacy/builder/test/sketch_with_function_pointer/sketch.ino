#include "CallbackBug.h"
Task t1(&t1Callback);
void t1Callback() {}
void setup() {}
void loop() {}
