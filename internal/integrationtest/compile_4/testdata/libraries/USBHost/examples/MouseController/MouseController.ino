/*
 Mouse Controller Example

 Shows the output of a USB Mouse connected to
 the Native USB port on an Arduino Due Board.

 created 8 Oct 2012
 by Cristian Maglie

 https://www.arduino.cc/en/Tutorial/MouseController

 This sample code is part of the public domain.
 */

// Require mouse control library
#include <MouseController.h>

// Initialize USB Controller
USBHost usb;

// Attach mouse controller to USB
MouseController mouse(usb);

// variables for mouse button states
bool leftButton = false;
bool middleButton = false;
bool rightButton = false;

// This function intercepts mouse movements
void mouseMoved() {
  SERIAL_PORT_MONITOR.print("Move: ");
  SERIAL_PORT_MONITOR.print(mouse.getXChange());
  SERIAL_PORT_MONITOR.print(", ");
  SERIAL_PORT_MONITOR.println(mouse.getYChange());
}

// This function intercepts mouse movements while a button is pressed
void mouseDragged() {
  SERIAL_PORT_MONITOR.print("DRAG: ");
  SERIAL_PORT_MONITOR.print(mouse.getXChange());
  SERIAL_PORT_MONITOR.print(", ");
  SERIAL_PORT_MONITOR.println(mouse.getYChange());
}

// This function intercepts mouse button press
void mousePressed() {
  SERIAL_PORT_MONITOR.print("Pressed: ");
  if (mouse.getButton(LEFT_BUTTON)) {
    SERIAL_PORT_MONITOR.print("L");
    leftButton = true;
  }
  if (mouse.getButton(MIDDLE_BUTTON)) {
    SERIAL_PORT_MONITOR.print("M");
    middleButton = true;
  }
  if (mouse.getButton(RIGHT_BUTTON)) {
    SERIAL_PORT_MONITOR.print("R");
    rightButton = true;
  }
  SERIAL_PORT_MONITOR.println();
}

// This function intercepts mouse button release
void mouseReleased() {
  SERIAL_PORT_MONITOR.print("Released: ");
  if (!mouse.getButton(LEFT_BUTTON) && leftButton == true) {
    SERIAL_PORT_MONITOR.print("L");
    leftButton = false;
  }
  if (!mouse.getButton(MIDDLE_BUTTON) && middleButton == true) {
    SERIAL_PORT_MONITOR.print("M");
    middleButton = false;
  }
  if (!mouse.getButton(RIGHT_BUTTON) && rightButton == true) {
    SERIAL_PORT_MONITOR.print("R");
    rightButton = false;
  }
  SERIAL_PORT_MONITOR.println();
}

void setup()
{
  SERIAL_PORT_MONITOR.begin( 115200 );
  while (!SERIAL_PORT_MONITOR); // Wait for serial port to connect - used on Leonardo, Teensy and other boards with built-in USB CDC serial connection
  SERIAL_PORT_MONITOR.println("Mouse Controller Program started");

  if (usb.Init() == -1)
      SERIAL_PORT_MONITOR.println("OSC did not start.");

  delay( 20 );
}

void loop()
{
  // Process USB tasks
  usb.Task();
}
