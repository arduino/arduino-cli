void setup() {}

void loop() {
  // Visualize leds via Adalight
  int8_t newData = adalight();

}


//#define ADALIGHT_USE_TEMPLATE

#ifdef ADALIGHT_USE_TEMPLATE
int16_t adalight()
#else
int8_t adalight()
#endif
{
  // Flag if the leds got a new frame that should be updated
  return 0;
}
