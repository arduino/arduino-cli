#include "FastLED.h"

#define DATA_PIN    7
#define CLK_PIN   6
#define LED_TYPE    APA102
#define COLOR_ORDER GRB
#define NUM_LEDS    79
CRGB leds[NUM_LEDS];

#define BRIGHTNESS          96
#define FRAMES_PER_SECOND  120
void setup() {

  FastLED.addLeds<LED_TYPE, DATA_PIN, CLK_PIN, COLOR_ORDER>(leds, NUM_LEDS).setCorrection(TypicalLEDStrip);
}

void loop() {

}

typedef void (*SimplePatternList[])();
//SimplePatternList gPatterns = { rainbow, rainbowWithGlitter, confetti, sinelon, juggle, bpm };
SimplePatternList gPatterns = {sinelon};

void sinelon()
{
}
