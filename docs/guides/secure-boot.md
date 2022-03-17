### Secure Boot

A ["secure boot"](https://www.keyfactor.com/blog/what-is-secure-boot-its-where-iot-security-starts/) capability may be
offered by Arduino boards platforms.

The compiled sketch is signed and encrypted by a [tool](../platform-specification.md#tools) before being flashed to the
target board. The bootloader of the board is then responsible for starting the compiled sketch only if the matching keys
are used.

To be able to correctly carry out all the operations at the end of the build we can leverage the
[post build hooks](../platform-specification.md#pre-and-post-build-hooks-since-arduino-ide-165) to sign and encrypt a
binary by using `recipe.hooks.objcopy.postobjcopy.NUMBER.pattern` key in
[`platform.txt`](../platform-specification.md#platformtxt). The security keys used are defined in the
[`boards.txt`](../platform-specification.md#boardstxt) file, this way there could be different keys for different
boards.

```
[...]
## Create secure image (bin file)
recipe.hooks.objcopy.postobjcopy.1.pattern={build.postbuild.cmd}

#
# IMGTOOL
#
tools.imgtool.cmd=imgtool
tools.imgtool.flags=sign --key "{build.keys.keychain}/{build.keys.sign_key}" --encrypt "{build.keys.keychain}/{build.keys.encrypt_key}" "{build.path}/{build.project_name}.bin" "{build.path}/{build.project_name}.bin" --align {build.alignment} --max-align {build.alignment} --version {build.version} --header-size {build.header_size} --pad-header --slot-size {build.slot_size}
[...]

```

By having only `tools.TOOL_NAME.cmd` and `tools.TOOL_NAME.falgs`, we can customize the behavior with a
[custom board option](../platform-specification.md#custom-board-options). Then in the
[`boards.txt`](../platform-specification.md#boardstxt) we can define the new option to use a different `postbuild.cmd`:

```
[...]
menu.security=Security setting

envie_m7.menu.security.none=None
envie_m7.menu.security.sien=Signature + Encryption

envie_m7.menu.security.sien.build.postbuild.cmd="{tools.imgtool.cmd}" {tools.imgtool.flags}
envie_m7.menu.security.none.build.postbuild.cmd="{tools.imgtool.cmd}" exit

envie_m7.menu.security.sien.build.keys.keychain={runtime.hardware.path}/Default_Keys
envie_m7.menu.security.sien.build.keys.sign_key=default-signing-key.pem
envie_m7.menu.security.sien.build.keys.encrypt_key=default-encrypt-key.pem
[...]
```

The security keys can be added with:

- `build.keys.keychain` indicates the path of the dir where to search for the custom keys to sign and encrypt a binary.
- `build.keys.sign_key` indicates the name of the custom signing key to use to sign a binary during the compile process.
- `build.keys.encrypt_key` indicates the name of the custom encryption key to use to encrypt a binary during the compile
  process.

It's suggested to use the property names mentioned before, because they can be overridden respectively with
`--keys-keychain`, `--sign-key` and ``--encrypt-key` Arduino CLI [compile flags](../commands/arduino-cli_compile.md).

For example, by using the following command, the sketch is compiled and the resulting binary is signed and encrypted
with the specified keys located in `/home/user/Arduino/keys` directory:

```
arduino-cli compile -b arduino:mbed_portenta:envie_m7:security=sien --keys-keychain /home/user/Arduino/keys --sign-key ecsdsa-p256-signing-key.pem --encrypt-key ecsdsa-p256-encrypt-key.pem /home/user/Arduino/MySketch
```
