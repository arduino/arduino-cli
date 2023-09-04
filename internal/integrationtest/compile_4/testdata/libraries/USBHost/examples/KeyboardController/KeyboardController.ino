/*
 Keyboard Controller Example

 Shows the output of a USB Keyboard connected to
 the Native USB port on an Arduino Due Board.

 created 8 Oct 2012
 by Cristian Maglie

 https://www.arduino.cc/en/Tutorial/KeyboardController

 This sample code is part of the public domain.
 */

// Require keyboard control library
#include <KeyboardController.h>

// Initialize USB Controller
USBHost usb;

// Attach keyboard controller to USB
KeyboardController keyboard(usb);

void printKey();

// This function intercepts key press
void keyPressed() {
  SERIAL_PORT_MONITOR.print("Pressed:  ");
  printKey();
}

// This function intercepts key release
void keyReleased() {
  SERIAL_PORT_MONITOR.print("Released: ");
  printKey();
}

void printKey() {
  // getOemKey() returns the OEM-code associated with the key
  SERIAL_PORT_MONITOR.print(" key:");
  SERIAL_PORT_MONITOR.print(keyboard.getOemKey());

  // getModifiers() returns a bits field with the modifiers-keys
  int mod = keyboard.getModifiers();
  SERIAL_PORT_MONITOR.print(" mod:");
  SERIAL_PORT_MONITOR.print(mod);

  SERIAL_PORT_MONITOR.print(" => ");

  if (mod & LeftCtrl)
    SERIAL_PORT_MONITOR.print("L-Ctrl ");
  if (mod & LeftShift)
    SERIAL_PORT_MONITOR.print("L-Shift ");
  if (mod & Alt)
    SERIAL_PORT_MONITOR.print("Alt ");
  if (mod & LeftCmd)
    SERIAL_PORT_MONITOR.print("L-Cmd ");
  if (mod & RightCtrl)
    SERIAL_PORT_MONITOR.print("R-Ctrl ");
  if (mod & RightShift)
    SERIAL_PORT_MONITOR.print("R-Shift ");
  if (mod & AltGr)
    SERIAL_PORT_MONITOR.print("AltGr ");
  if (mod & RightCmd)
    SERIAL_PORT_MONITOR.print("R-Cmd ");

  // getKey() returns the ASCII translation of OEM key
  // combined with modifiers.
  SERIAL_PORT_MONITOR.write(keyboard.getKey());
  SERIAL_PORT_MONITOR.println();
}

void setup()
{
  SERIAL_PORT_MONITOR.begin( 115200 );
  while (!SERIAL_PORT_MONITOR); // Wait for serial port to connect - used on Leonardo, Teensy and other boards with built-in USB CDC serial connection
  SERIAL_PORT_MONITOR.println("Keyboard Controller Program started");

  if (usb.Init() == -1)
	  SERIAL_PORT_MONITOR.println("OSC did not start.");
  
  delay( 20 );
}

void loop()
{
  // Process USB tasks
  usb.Task();
}
