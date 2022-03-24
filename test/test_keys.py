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

from ecdsa import VerifyingKey, SigningKey
from pathlib import Path


def test_keys_generate(run_command, working_dir):
    # Create security keys without specifying the keychain dir (by default in the working directory)
    sign_key_name = "ecdsa-p256-signing-key.pem"
    sign_header_name = "ecdsa-p256-signing-key.h"
    result = run_command(["keys", "generate", "--key-name", sign_key_name])
    assert result.ok
    assert f"Keys created in: {working_dir}" in result.stdout
    assert Path(working_dir, f"pub_{sign_key_name}").is_file()
    assert Path(working_dir, f"priv_{sign_key_name}").is_file()
    assert Path(working_dir, f"pub_{sign_header_name}").is_file()
    assert Path(working_dir, f"priv_{sign_header_name}").is_file()

    # Overwrite security keys
    result = run_command(["keys", "generate", "--key-name", sign_key_name])
    assert result.failed
    assert f"Error during generate: Cannot create file: File already exists: {working_dir}" in result.stderr

    # Create security keys in specified directory
    keychain_name = "keychain"
    keychain_path = Path(working_dir, keychain_name)
    result = run_command(["keys", "generate", "--key-name", sign_key_name, "--keys-keychain", keychain_path])
    assert result.ok
    assert f"Keys created in: {keychain_path}" in result.stdout
    assert Path(keychain_path, f"pub_{sign_key_name}").is_file()
    assert Path(keychain_path, f"priv_{sign_key_name}").is_file()
    assert Path(keychain_path, f"pub_{sign_header_name}").is_file()
    assert Path(keychain_path, f"priv_{sign_header_name}").is_file()

    # verify that keypar is valid by signing a message and then verify it
    with open(f"{keychain_path}/pub_{sign_key_name}") as f:
        vk = VerifyingKey.from_pem(f.read())
    with open(f"{keychain_path}/priv_{sign_key_name}") as f1:
        sk = SigningKey.from_pem(f1.read())

    signature = sk.sign(b"message")
    assert vk.verify(signature, b"message")

    # Create security keys without specifying --key-name
    result = run_command(["keys", "generate", "--keys-keychain", keychain_path])
    assert result.failed
    assert 'Error: required flag(s) "key-name" not set' in result.stderr

    # Create security keys with unsupported algorithm
    result = run_command(
        ["keys", "generate", "--key-name", sign_key_name, "--keys-keychain", keychain_path, "-t", "rsa"]
    )
    assert result.failed
    assert "Error during generate: Cannot create file: Unsupported algorithm: rsa" in result.stderr
