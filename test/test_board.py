# This file is part of arduino-cli.
#
# Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
#
# This software is released under the GNU General Public License version 3,
# which covers the main part of arduino-cli.
# The terms of this license can be found at:
# https://www.gnu.org/licenses/gpl-3.0.en.html
#
# You can be released from the requirements of the above licenses by purchasing
# a commercial license. Buying such a license is mandatory if you want to modify or
# otherwise use the software for commercial activities involving the Arduino
# software without disclosing the source code of your own applications. To purchase
# a commercial license, send an email to license@arduino.cc.
from pathlib import Path
from git import Repo
import os
import glob
import simplejson as json
import semver
import pytest

from .common import running_on_ci


gold_board = """
    {
  "fqbn": "arduino:samd:nano_33_iot",
  "name": "Arduino NANO 33 IoT",
  "version": "1.8.6",
  "properties_id": "nano_33_iot",
  "official": true,
  "package": {
    "maintainer": "Arduino",
    "url": "https://downloads.arduino.cc/packages/package_index.tar.bz2",
    "website_url": "http://www.arduino.cc/",
    "email": "packages@arduino.cc",
    "name": "arduino",
    "help": {
      "online": "http://www.arduino.cc/en/Reference/HomePage"
    }
  },
  "platform": {
    "architecture": "samd",
    "category": "Arduino",
    "url": "http://downloads.arduino.cc/cores/samd-1.8.6.tar.bz2",
    "archive_filename": "samd-1.8.6.tar.bz2",
    "checksum": "SHA-256:68a4fffa6fe6aa7886aab2e69dff7d3f94c02935bbbeb42de37f692d7daf823b",
    "size": 2980953,
    "name": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)"
  },
  "toolsDependencies": [
    {
      "packager": "arduino",
      "name": "arm-none-eabi-gcc",
      "version": "7-2017q4",
      "systems": [
        {
          "checksum": "SHA-256:34180943d95f759c66444a40b032f7dd9159a562670fc334f049567de140c51b",
          "host": "arm-linux-gnueabihf",
          "archive_filename": "gcc-arm-none-eabi-7-2019-q4-major-linuxarm.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-7-2019-q4-major-linuxarm.tar.bz2",
          "size": 96613739
        },
        {
          "checksum": "SHA-256:6fb5752fb4d11012bd0a1ceb93a19d0641ff7cf29d289b3e6b86b99768e66f76",
          "host": "aarch64-linux-gnu",
          "archive_filename": "gcc-arm-none-eabi-7-2018-q2-update-linuxarm64.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-7-2018-q2-update-linuxarm64.tar.bz2",
          "size": 99558726
        },
        {
          "checksum": "SHA-256:96dd0091856f4d2eb21046eba571321feecf7d50b9c156f708b2a8b683903382",
          "host": "i686-mingw32",
          "archive_filename": "gcc-arm-none-eabi-7-2017-q4-major-win32-arduino1.zip",
          "url": "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-7-2017-q4-major-win32-arduino1.zip",
          "size": 131761924
        },
        {
          "checksum": "SHA-256:89b776c7cf0591c810b5b60067e4dc113b5b71bc50084a536e71b894a97fdccb",
          "host": "x86_64-apple-darwin",
          "archive_filename": "gcc-arm-none-eabi-7-2017-q4-major-mac.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-7-2017-q4-major-mac.tar.bz2",
          "size": 104550003
        },
        {
          "checksum": "SHA-256:96a029e2ae130a1210eaa69e309ea40463028eab18ba19c1086e4c2dafe69a6a",
          "host": "x86_64-pc-linux-gnu",
          "archive_filename": "gcc-arm-none-eabi-7-2017-q4-major-linux64.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-7-2017-q4-major-linux64.tar.bz2",
          "size": 99857645
        },
        {
          "checksum": "SHA-256:090a0bc2b1956bc49392dff924a6c30fa57c88130097b1972204d67a45ce3cf3",
          "host": "i686-pc-linux-gnu",
          "archive_filename": "gcc-arm-none-eabi-7-2018-q2-update-linux32.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/gcc-arm-none-eabi-7-2018-q2-update-linux32.tar.bz2",
          "size": 97427309
        }
      ]
    },
    {
      "packager": "arduino",
      "name": "bossac",
      "version": "1.7.0-arduino3",
      "systems": [
        {
          "checksum": "SHA-256:62745cc5a98c26949ec9041ef20420643c561ec43e99dae659debf44e6836526",
          "host": "i686-mingw32",
          "archive_filename": "bossac-1.7.0-arduino3-windows.tar.gz",
          "url": "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-windows.tar.gz",
          "size": 3607421
        },
        {
          "checksum": "SHA-256:adb3c14debd397d8135e9e970215c6972f0e592c7af7532fa15f9ce5e64b991f",
          "host": "x86_64-apple-darwin",
          "archive_filename": "bossac-1.7.0-arduino3-osx.tar.gz",
          "url": "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-osx.tar.gz",
          "size": 75510
        },
        {
          "checksum": "SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100",
          "host": "x86_64-pc-linux-gnu",
          "archive_filename": "bossac-1.7.0-arduino3-linux64.tar.gz",
          "url": "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz",
          "size": 207271
        },
        {
          "checksum": "SHA-256:4ac4354746d1a09258f49a43ef4d1baf030d81c022f8434774268b00f55d3ec3",
          "host": "i686-pc-linux-gnu",
          "archive_filename": "bossac-1.7.0-arduino3-linux32.tar.gz",
          "url": "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux32.tar.gz",
          "size": 193577
        },
        {
          "checksum": "SHA-256:626c6cc548046901143037b782bf019af1663bae0d78cf19181a876fb9abbb90",
          "host": "arm-linux-gnueabihf",
          "archive_filename": "bossac-1.7.0-arduino3-linuxarm.tar.gz",
          "url": "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linuxarm.tar.gz",
          "size": 193941
        },
        {
          "checksum": "SHA-256:a098b2cc23e29f0dc468416210d097c4a808752cd5da1a7b9b8b7b931a04180b",
          "host": "aarch64-linux-gnu",
          "archive_filename": "bossac-1.7.0-arduino3-linuxaarch64.tar.gz",
          "url": "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linuxaarch64.tar.gz",
          "size": 268365
        }
      ]
    },
    {
      "packager": "arduino",
      "name": "openocd",
      "version": "0.10.0-arduino7",
      "systems": [
        {
          "checksum": "SHA-256:f8e0d783e80a3d5f75ee82e9542315871d46e1e283a97447735f1cbcd8986b06",
          "host": "arm-linux-gnueabihf",
          "archive_filename": "openocd-0.10.0-arduino7-static-arm-linux-gnueabihf.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/openocd-0.10.0-arduino7-static-arm-linux-gnueabihf.tar.bz2",
          "size": 1638575
        },
        {
          "checksum": "SHA-256:d47d728a9a8d98f28dc22e31d7127ced9de0d5e268292bf935e050ef1d2bdfd0",
          "host": "aarch64-linux-gnu",
          "archive_filename": "openocd-0.10.0-arduino7-static-aarch64-linux-gnu.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/openocd-0.10.0-arduino7-static-aarch64-linux-gnu.tar.bz2",
          "size": 1580739
        },
        {
          "checksum": "SHA-256:1e539a587a0c54a551ce0dc542af10a2520b1c93bbfe2ca4ebaef4c83411df1a",
          "host": "i386-apple-darwin11",
          "archive_filename": "openocd-0.10.0-arduino7-static-x86_64-apple-darwin13.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/openocd-0.10.0-arduino7-static-x86_64-apple-darwin13.tar.bz2",
          "size": 1498970
        },
        {
          "checksum": "SHA-256:91d418bd309ec1e98795c622cd25c936aa537c0b3828fa5bcb191389378a1b27",
          "host": "x86_64-linux-gnu",
          "archive_filename": "openocd-0.10.0-arduino7-static-x86_64-ubuntu12.04-linux-gnu.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/openocd-0.10.0-arduino7-static-x86_64-ubuntu12.04-linux-gnu.tar.bz2",
          "size": 1701581
        },
        {
          "checksum": "SHA-256:08a18f39d72a5626383503053a30a5da89eed7fdccb6f514b20b77403eb1b2b4",
          "host": "i686-linux-gnu",
          "archive_filename": "openocd-0.10.0-arduino7-static-i686-ubuntu12.04-linux-gnu.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/openocd-0.10.0-arduino7-static-i686-ubuntu12.04-linux-gnu.tar.bz2",
          "size": 1626347
        },
        {
          "checksum": "SHA-256:f251aec5471296e18aa540c3078d66475357a76a77c16c06a2d9345f4e12b3d5",
          "host": "i686-mingw32",
          "archive_filename": "openocd-0.10.0-arduino7-static-i686-w64-mingw32.zip",
          "url": "http://downloads.arduino.cc/tools/openocd-0.10.0-arduino7-static-i686-w64-mingw32.zip",
          "size": 2016965
        }
      ]
    },
    {
      "packager": "arduino",
      "name": "CMSIS",
      "version": "4.5.0",
      "systems": [
        {
          "checksum": "SHA-256:cd8f7eae9fc7c8b4a1b5e40b89b9666d33953b47d3d2eb81844f5af729fa224d",
          "host": "i686-mingw32",
          "archive_filename": "CMSIS-4.5.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-4.5.0.tar.bz2",
          "size": 31525196
        },
        {
          "checksum": "SHA-256:cd8f7eae9fc7c8b4a1b5e40b89b9666d33953b47d3d2eb81844f5af729fa224d",
          "host": "x86_64-apple-darwin",
          "archive_filename": "CMSIS-4.5.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-4.5.0.tar.bz2",
          "size": 31525196
        },
        {
          "checksum": "SHA-256:cd8f7eae9fc7c8b4a1b5e40b89b9666d33953b47d3d2eb81844f5af729fa224d",
          "host": "x86_64-pc-linux-gnu",
          "archive_filename": "CMSIS-4.5.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-4.5.0.tar.bz2",
          "size": 31525196
        },
        {
          "checksum": "SHA-256:cd8f7eae9fc7c8b4a1b5e40b89b9666d33953b47d3d2eb81844f5af729fa224d",
          "host": "i686-pc-linux-gnu",
          "archive_filename": "CMSIS-4.5.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-4.5.0.tar.bz2",
          "size": 31525196
        },
        {
          "checksum": "SHA-256:cd8f7eae9fc7c8b4a1b5e40b89b9666d33953b47d3d2eb81844f5af729fa224d",
          "host": "arm-linux-gnueabihf",
          "archive_filename": "CMSIS-4.5.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-4.5.0.tar.bz2",
          "size": 31525196
        },
        {
          "checksum": "SHA-256:cd8f7eae9fc7c8b4a1b5e40b89b9666d33953b47d3d2eb81844f5af729fa224d",
          "host": "aarch64-linux-gnu",
          "archive_filename": "CMSIS-4.5.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-4.5.0.tar.bz2",
          "size": 31525196
        },
        {
          "checksum": "SHA-256:cd8f7eae9fc7c8b4a1b5e40b89b9666d33953b47d3d2eb81844f5af729fa224d",
          "host": "all",
          "archive_filename": "CMSIS-4.5.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-4.5.0.tar.bz2",
          "size": 31525196
        }
      ]
    },
    {
      "packager": "arduino",
      "name": "CMSIS-Atmel",
      "version": "1.2.0",
      "systems": [
        {
          "checksum": "SHA-256:5e02670be7e36be9691d059bee0b04ee8b249404687531f33893922d116b19a5",
          "host": "i686-mingw32",
          "archive_filename": "CMSIS-Atmel-1.2.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-Atmel-1.2.0.tar.bz2",
          "size": 2221805
        },
        {
          "checksum": "SHA-256:5e02670be7e36be9691d059bee0b04ee8b249404687531f33893922d116b19a5",
          "host": "x86_64-apple-darwin",
          "archive_filename": "CMSIS-Atmel-1.2.0.tar.bz2",
          "url": "http://downloads.arduino.cc/CMSIS-Atmel-1.2.0.tar.bz2",
          "size": 2221805
        },
        {
          "checksum": "SHA-256:5e02670be7e36be9691d059bee0b04ee8b249404687531f33893922d116b19a5",
          "host": "x86_64-pc-linux-gnu",
          "archive_filename": "CMSIS-Atmel-1.2.0.tar.bz2",
          "url": "https://downloads.arduino.cc/CMSIS-Atmel-1.2.0.tar.bz2",
          "size": 2221805
        },
        {
          "checksum": "SHA-256:5e02670be7e36be9691d059bee0b04ee8b249404687531f33893922d116b19a5",
          "host": "i686-pc-linux-gnu",
          "archive_filename": "CMSIS-Atmel-1.2.0.tar.bz2",
          "url": "https://downloads.arduino.cc/CMSIS-Atmel-1.2.0.tar.bz2",
          "size": 2221805
        },
        {
          "checksum": "SHA-256:5e02670be7e36be9691d059bee0b04ee8b249404687531f33893922d116b19a5",
          "host": "arm-linux-gnueabihf",
          "archive_filename": "CMSIS-Atmel-1.2.0.tar.bz2",
          "url": "https://downloads.arduino.cc/CMSIS-Atmel-1.2.0.tar.bz2",
          "size": 2221805
        },
        {
          "checksum": "SHA-256:5e02670be7e36be9691d059bee0b04ee8b249404687531f33893922d116b19a5",
          "host": "aarch64-linux-gnu",
          "archive_filename": "CMSIS-Atmel-1.2.0.tar.bz2",
          "url": "https://downloads.arduino.cc/CMSIS-Atmel-1.2.0.tar.bz2",
          "size": 2221805
        },
        {
          "checksum": "SHA-256:5e02670be7e36be9691d059bee0b04ee8b249404687531f33893922d116b19a5",
          "host": "all",
          "archive_filename": "CMSIS-Atmel-1.2.0.tar.bz2",
          "url": "https://downloads.arduino.cc/CMSIS-Atmel-1.2.0.tar.bz2",
          "size": 2221805
        }
      ]
    },
    {
      "packager": "arduino",
      "name": "arduinoOTA",
      "version": "1.2.1",
      "systems": [
        {
          "checksum": "SHA-256:2ffdf64b78486c1d0bf28dc23d0ca36ab75ca92e84b9487246da01888abea6d4",
          "host": "i686-linux-gnu",
          "archive_filename": "arduinoOTA-1.2.1-linux_386.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/arduinoOTA-1.2.1-linux_386.tar.bz2",
          "size": 2133779
        },
        {
          "checksum": "SHA-256:5b82310d53688480f34a916aac31cd8f2dd2be65dd8fa6c2445262262e1948f9",
          "host": "x86_64-linux-gnu",
          "archive_filename": "arduinoOTA-1.2.1-linux_amd64.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/arduinoOTA-1.2.1-linux_amd64.tar.bz2",
          "size": 2257689
        },
        {
          "checksum": "SHA-256:ad54b3dcd586212941fd992bab573b53d13207a419a3f2981c970a085ae0e9e0",
          "host": "arm-linux-gnueabihf",
          "archive_filename": "arduinoOTA-1.2.1-linux_arm.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/arduinoOTA-1.2.1-linux_arm.tar.bz2",
          "size": 2093132
        },
        {
          "checksum": "SHA-256:ad54b3dcd586212941fd992bab573b53d13207a419a3f2981c970a085ae0e9e0",
          "host": "aarch64-linux-gnu",
          "archive_filename": "arduinoOTA-1.2.1-linux_arm.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/arduinoOTA-1.2.1-linux_arm.tar.bz2",
          "size": 2093132
        },
        {
          "checksum": "SHA-256:93a6d9f9c0c765d237be1665bf7a0a8e2b0b6d2a8531eae92db807f5515088a7",
          "host": "i386-apple-darwin11",
          "archive_filename": "arduinoOTA-1.2.1-darwin_amd64.tar.bz2",
          "url": "http://downloads.arduino.cc/tools/arduinoOTA-1.2.1-darwin_amd64.tar.bz2",
          "size": 2244088
        },
        {
          "checksum": "SHA-256:e1ebf21f2c073fce25c09548c656da90d4ef6c078401ec6f323e0c58335115e5",
          "host": "i686-mingw32",
          "archive_filename": "arduinoOTA-1.2.1-windows_386.zip",
          "url": "http://downloads.arduino.cc/tools/arduinoOTA-1.2.1-windows_386.zip",
          "size": 2237511
        }
      ]
    }
  ],
  "identification_properties": [
    {
      "properties": {
        "vid": "0x2341",
        "pid": "0x8057"
      }
    },
    {
      "properties": {
        "vid": "0x2341",
        "pid": "0x0057"
      }
    }
  ],
  "programmers": [
    {
      "platform": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)",
      "id": "edbg",
      "name": "Atmel EDBG"
    },
    {
      "platform": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)",
      "id": "atmel_ice",
      "name": "Atmel-ICE"
    },
    {
      "platform": "Arduino SAMD Boards (32-bits ARM Cortex-M0+)",
      "id": "sam_ice",
      "name": "Atmel SAM-ICE"
    }
  ]
}
"""  # noqa: E501


def test_board_details(run_command):
    run_command(["core", "update-index"])
    # Download samd core pinned to 1.8.6
    run_command(["core", "install", "arduino:samd@1.8.6"])

    # Test board listall with and without showing hidden elements
    result = run_command(["board", "listall", "MIPS", "--format", "json"])
    assert result.ok
    assert result.stdout == "{}\n"

    result = run_command(["board", "listall", "MIPS", "-a", "--format", "json"])
    assert result.ok
    result = json.loads(result.stdout)
    assert result["boards"][0]["name"] == "Arduino Tian (MIPS Console port)"

    result = run_command(["board", "details", "-b", "arduino:samd:nano_33_iot", "--format", "json"])
    assert result.ok
    # Sort everything before compare
    result = json.loads(result.stdout)
    gold_board_details = json.loads(gold_board)

    assert result["fqbn"] == gold_board_details["fqbn"]
    assert result["name"] == gold_board_details["name"]
    assert result["version"] == gold_board_details["version"]
    assert result["properties_id"] == gold_board_details["properties_id"]
    assert result["official"] == gold_board_details["official"]
    assert result["package"] == gold_board_details["package"]
    assert result["platform"] == gold_board_details["platform"]
    for usb_id in gold_board_details["identification_properties"]:
        assert usb_id in result["identification_properties"]
    for programmer in gold_board_details["programmers"]:
        assert programmer in result["programmers"]

    # Download samd core pinned to 1.8.8
    run_command(["core", "install", "arduino:samd@1.8.8"])

    result = run_command(["board", "details", "-b", "arduino:samd:nano_33_iot", "--format", "json"])
    assert result.ok
    result = json.loads(result.stdout)
    assert result["debugging_supported"] is True


def test_board_details_no_flags(run_command):
    run_command(["core", "update-index"])
    # Download samd core pinned to 1.8.6
    run_command(["core", "install", "arduino:samd@1.8.6"])
    result = run_command(["board", "details"])
    assert not result.ok
    assert 'Error: required flag(s) "fqbn" not set' in result.stderr
    assert result.stdout == ""


def test_board_details_list_programmers_without_flag(run_command):
    run_command(["core", "update-index"])
    # Download samd core pinned to 1.8.6
    run_command(["core", "install", "arduino:samd@1.8.6"])
    result = run_command(["board", "details", "-b", "arduino:samd:nano_33_iot"], hide=True)
    assert result.ok
    lines = [l.strip().split() for l in result.stdout.splitlines()]
    assert ["Programmers:", "Id", "Name"] in lines
    assert ["edbg", "Atmel", "EDBG"] in lines
    assert ["atmel_ice", "Atmel-ICE"] in lines
    assert ["sam_ice", "Atmel", "SAM-ICE"] in lines


def test_board_details_list_programmers_flag(run_command):
    run_command(["core", "update-index"])
    # Download samd core pinned to 1.8.6
    run_command(["core", "install", "arduino:samd@1.8.6"])
    result = run_command(["board", "details", "-b", "arduino:samd:nano_33_iot", "--list-programmers"], hide=True)
    assert result.ok

    lines = [l.strip() for l in result.stdout.splitlines()]
    assert "Id        Programmer name" in lines
    assert "edbg      Atmel EDBG" in lines
    assert "atmel_ice Atmel-ICE" in lines
    assert "sam_ice   Atmel SAM-ICE" in lines


def test_board_search(run_command, data_dir):
    assert run_command(["update"])

    res = run_command(["board", "search", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    # Verifies boards are returned
    assert len(data) > 0
    # Verifies no board has FQBN set since no platform is installed
    assert len([board["fqbn"] for board in data if "fqbn" in board]) == 0
    names = [board["name"] for board in data if "name" in board]
    assert "Arduino Uno" in names
    assert "Arduino Yún" in names
    assert "Arduino Zero" in names
    assert "Arduino Nano 33 BLE" in names
    assert "Arduino Portenta H7" in names

    # Search in non installed boards
    res = run_command(["board", "search", "--format", "json", "nano", "33"])
    assert res.ok
    data = json.loads(res.stdout)
    # Verifies boards are returned
    assert len(data) > 0
    # Verifies no board has FQBN set since no platform is installed
    assert len([board["fqbn"] for board in data if "fqbn" in board]) == 0
    names = [board["name"] for board in data if "name" in board]
    assert "Arduino Nano 33 BLE" in names
    assert "Arduino Nano 33 IoT" in names

    # Install a platform from index
    assert run_command(["core", "install", "arduino:avr@1.8.3"])

    res = run_command(["board", "search", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) > 0
    # Verifies some FQBNs are now returned after installing a platform
    assert len([board["fqbn"] for board in data if "fqbn" in board]) == 26
    installed_boards = {board["fqbn"]: board for board in data if "fqbn" in board}
    assert "arduino:avr:uno" in installed_boards
    assert "Arduino Uno" == installed_boards["arduino:avr:uno"]["name"]
    assert "arduino:avr:yun" in installed_boards
    assert "Arduino Yún" == installed_boards["arduino:avr:yun"]["name"]

    res = run_command(["board", "search", "--format", "json", "arduino", "yun"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) > 0
    installed_boards = {board["fqbn"]: board for board in data if "fqbn" in board}
    assert "arduino:avr:yun" in installed_boards
    assert "Arduino Yún" == installed_boards["arduino:avr:yun"]["name"]

    # Manually installs a core in sketchbooks hardware folder
    git_url = "https://github.com/arduino/ArduinoCore-samd.git"
    repo_dir = Path(data_dir, "hardware", "arduino-beta-development", "samd")
    assert Repo.clone_from(git_url, repo_dir, multi_options=["-b 1.8.11"])

    res = run_command(["board", "search", "--format", "json"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) > 0
    # Verifies some FQBNs are now returned after installing a platform
    assert len([board["fqbn"] for board in data if "fqbn" in board]) == 43
    installed_boards = {board["fqbn"]: board for board in data if "fqbn" in board}
    assert "arduino:avr:uno" in installed_boards
    assert "Arduino Uno" == installed_boards["arduino:avr:uno"]["name"]
    assert "arduino:avr:yun" in installed_boards
    assert "Arduino Yún" == installed_boards["arduino:avr:yun"]["name"]
    assert "arduino-beta-development:samd:mkrwifi1010" in installed_boards
    assert "Arduino MKR WiFi 1010" == installed_boards["arduino-beta-development:samd:mkrwifi1010"]["name"]
    assert "arduino-beta-development:samd:mkr1000" in installed_boards
    assert "Arduino MKR1000" == installed_boards["arduino-beta-development:samd:mkr1000"]["name"]
    assert "arduino-beta-development:samd:mkrzero" in installed_boards
    assert "Arduino MKRZERO" == installed_boards["arduino-beta-development:samd:mkrzero"]["name"]
    assert "arduino-beta-development:samd:nano_33_iot" in installed_boards
    assert "Arduino NANO 33 IoT" == installed_boards["arduino-beta-development:samd:nano_33_iot"]["name"]
    assert "arduino-beta-development:samd:arduino_zero_native" in installed_boards

    res = run_command(["board", "search", "--format", "json", "mkr1000"])
    assert res.ok
    data = json.loads(res.stdout)
    assert len(data) > 0
    # Verifies some FQBNs are now returned after installing a platform
    installed_boards = {board["fqbn"]: board for board in data if "fqbn" in board}
    assert "arduino-beta-development:samd:mkr1000" in installed_boards
    assert "Arduino MKR1000" == installed_boards["arduino-beta-development:samd:mkr1000"]["name"]


def test_board_attach_without_sketch_json(run_command, data_dir):
    run_command(["update"])

    sketch_name = "BoardAttachWithoutSketchJson"
    sketch_path = Path(data_dir, sketch_name)
    fqbn = "arduino:avr:uno"

    # Create a test sketch
    assert run_command(["sketch", "new", sketch_path])

    assert run_command(["board", "attach", "-b", fqbn, sketch_path])


def test_board_search_with_outdated_core(run_command):
    assert run_command(["update"])

    # Install an old core version
    assert run_command(["core", "install", "arduino:samd@1.8.6"])

    res = run_command(["board", "search", "arduino:samd:mkrwifi1010", "--format", "json"])

    data = json.loads(res.stdout)
    assert len(data) == 1
    board = data[0]
    assert board["name"] == "Arduino MKR WiFi 1010"
    assert board["fqbn"] == "arduino:samd:mkrwifi1010"
    samd_core = board["platform"]
    assert samd_core["id"] == "arduino:samd"
    installed_version = semver.parse_version_info(samd_core["installed"])
    latest_version = semver.parse_version_info(samd_core["latest"])
    # Installed version must be older than latest
    assert installed_version.compare(latest_version) == -1
