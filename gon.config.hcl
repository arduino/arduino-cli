source = ["dist/arduino_cli_osx_darwin_amd64/arduino-cli"]
bundle_id = "cc.arduino.arduino-cli"

sign {
  application_identity = "Developer ID Application: ARDUINO SA (7KT7ZWMCJT)"
}

zip {
  output_path = "arduino-cli.zip"
}
