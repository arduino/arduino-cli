This specification is a 3rd party Hardware format to be used in Arduino IDE starting from 1.5.x series. \
This specification allows a 3rd party vendor/maintainer to add support for new boards inside the Arduino IDE by providing a file to unzip into the *hardware* folder of Arduino's sketchbook folder. \
It is also possible to add new 3rd party boards by providing just one configuration file.

# Hardware Folders structure

The new hardware folders have a hierarchical structure organized in two levels:
  - the first level is the vendor/maintainer
  - the second level is the supported architecture

A vendor/maintainer can have multiple supported architectures. For example, below we have three hardware vendors called "arduino", "yyyyy" and "xxxxx":

    hardware/arduino/avr/...     - Arduino - AVR Boards
    hardware/arduino/sam/...     - Arduino - SAM (32bit ARM) Boards
    hardware/yyyyy/avr/...       - Yyy - AVR
    hardware/xxxxx/avr/...       - Xxx - AVR

The vendor "arduino" has two supported architectures (AVR and SAM), while "xxxxx" and "yyyyy" have only AVR.

If possible, follow existing architecture name conventions when creating hardware packages. The architecture folder name is used to determine library compatibility and also to permit referencing resources from another core of the same architecture so using a non-standard architecture name can only be harmful to your users. Architecture values are case sensitive (e.g. `AVR` != `avr`). Use the vendor folder name to differentiate your package, **NOT** the architecture name.

# Architecture configurations

Each architecture must be configured through a set of configuration files:

* **platform.txt** contains definitions for the CPU architecture used (compiler, build process parameters, tools used for upload, etc.)
* **boards.txt** contains definitions for the boards (board name, parameters for building and uploading sketches, etc.)
* **programmers.txt** contains definitions for external programmers (typically used to burn bootloaders or sketches on a blank CPU/board)

## Configuration files format

A configuration file is a list of "key=value" properties. The **value** of a property can be expressed using the value of another property by putting its name inside brackets "{" "}". For example:

    compiler.path=/tools/g++_arm_none_eabi/bin/
    compiler.c.cmd=arm-none-eabi-gcc
    [....]
    recipe.c.o.pattern={compiler.path}{compiler.c.cmd}

In this example the property **recipe.c.o.pattern** will be set to **/tools/g++_arm_none_eabi/bin/arm-none-eabi-gcc** that is the composition of the two properties **compiler.path** and **compiler.c.cmd**.

### Comments

Lines starting with **#** are treated as comments and will be ignored.

    # Like in this example
    # --------------------
    # I'm a comment!

### Automatic property override for specific OS

We can specify an OS-specific value for a property. For example the following file:

    tools.bossac.cmd=bossac
    tools.bossac.cmd.windows=bossac.exe

will set the property **tools.bossac.cmd** to the value **bossac** on Linux and Mac OS and **bossac.exe** on Windows. Suffixes [supported](https://github.com/arduino/Arduino/blob/1.8.10/arduino-core/src/processing/app/helpers/PreferencesMap.java#L110-L112) are `.linux`, `.windows` and `.macosx`.

### Global Predefined properties

The Arduino IDE sets the following properties that can be used globally in all configurations files:

* `{runtime.platform.path}`: the absolute path of the [board platform](#platform-terminology) folder (i.e. the folder containing boards.txt)
* `{runtime.hardware.path}`: the absolute path of the hardware folder (i.e. the folder containing the [board platform](#platform-terminology) folder)
* `{runtime.ide.path}`: the absolute path of the Arduino IDE folder
* `{runtime.ide.version}`: the version number of the Arduino IDE as a number (this uses two digits per version number component, and removes the points and leading zeroes, so Arduino IDE 1.8.3 becomes `01.08.03` which becomes `runtime.ide.version=10803`).
* `{ide_version}`: Compatibility alias for `{runtime.ide.version}`
* `{runtime.os}`: the running OS ("linux", "windows", "macosx")

Compatibility note: Versions before 1.6.0 only used one digit per version number component in `{runtime.ide.version}` (so 1.5.9 was `159`, not `10509`).

# platform.txt

The platform.txt file contains information about a platform's specific aspects (compilers
command line flags, paths, system libraries, etc.).

The following meta-data must be defined:

    name=Arduino AVR Boards
    version=1.5.3

The **name** will be shown in the Boards menu of the Arduino IDE. \
The **version** is currently unused, it is reserved for future use (probably together with the libraries manager to handle dependencies on cores).

## Build process

The platform.txt file is used to configure the build process performed by the Arduino IDE. This is done through a list of **recipes**. Each recipe is a command line expression that explains how to call the compiler (or other tools) for every build step and which parameter should be passed.

The Arduino IDE, before starting the build, determines the list of files to compile. The list is composed by:
- the user's Sketch
- source code in the selected board's Core
- source code in the Libraries used in the sketch

The IDE creates a temporary folder to store the build artifacts whose path is available through the global property **{build.path}**. A property **{build.project_name}** with the name of the project and a property **{build.arch}** with the name of the architecture is set as well.

* `{build.path}`: The path to the temporary folder to store build artifacts
* `{build.project_name}`: The project name
* `{build.arch}`: The MCU architecture (avr, sam, etc...)

There are some other **{build.xxx}** properties available, that are explained in the boards.txt section of this guide.

### Recipes to compile source code

We said that the Arduino IDE determines a list of files to compile. Each file can be source code written in C (.c files), C++ (.cpp files) or Assembly (.S files). Every language is compiled using its respective **recipe**:

* `recipe.c.o.pattern`: for C files
* `recipe.cpp.o.pattern`: for CPP files
* `recipe.S.o.pattern`: for Assembly files

The recipes can be built concatenating other properties set by the IDE (for each file compiled):

* `{includes}`: the list of include paths in the format "-I/include/path -I/another/path...."
* `{source_file}`: the path to the source file
* `{object_file}`: the path to the output file

For example the following is used for AVR:

    ## Compiler global definitions
    compiler.path={runtime.ide.path}/tools/avr/bin/
    compiler.c.cmd=avr-gcc
    compiler.c.flags=-c -g -Os -w -ffunction-sections -fdata-sections -MMD

    [......]

    ## Compile c files
    recipe.c.o.pattern="{compiler.path}{compiler.c.cmd}" {compiler.c.flags} -mmcu={build.mcu} -DF_CPU={build.f_cpu} -DARDUINO={runtime.ide.version} -DARDUINO_{build.board} -DARDUINO_ARCH_{build.arch} {build.extra_flags} {includes} "{source_file}" -o "{object_file}"

Note that some properties, like **{build.mcu}** for example, are taken from the **boards.txt** file which is documented later in this specification.

### Recipes to build the core.a archive file

The core of the selected board is compiled as described in the previous paragraph, but the object files obtained from the compile are also archived into a static library named *core.a* using the **recipe.ar.pattern**.

The recipe can be built concatenating the following properties set by the IDE:

* `{object_file}`: the object file to include in the archive
* `{archive_file_path}`: fully qualified archive file (ex. "/path/to/core.a"). This property was added in Arduino IDE 1.6.6/arduino builder 1.0.0-beta12 as a replacement for `{build.path}/{archive_file}`.
* `{archive_file}`: the name of the resulting archive (ex. "core.a")

For example, Arduino provides the following for AVR:

    compiler.ar.cmd=avr-ar
    compiler.ar.flags=rcs

    [......]

    ## Create archives
    recipe.ar.pattern="{compiler.path}{compiler.ar.cmd}" {compiler.ar.flags} "{archive_file_path}" "{object_file}"

### Recipes for linking

All the artifacts produced by the previous steps (sketch object files, libraries object files and core.a archive) are linked together using the **recipe.c.combine.pattern**.

The recipe can be built concatenating the following properties set by the IDE:

* `{object_files}`: the list of object files to include in the archive ("file1.o file2.o ....")
* `{archive_file_path}`: fully qualified archive file (ex. "/path/to/core.a"). This property was added in Arduino IDE 1.6.6/arduino builder 1.0.0-beta12 as a replacement for `{build.path}/{archive_file}`.
* `{archive_file}`: the name of the core archive file (ex. "core.a")

For example the following is used for AVR:

    compiler.c.elf.flags=-Os -Wl,--gc-sections
    compiler.c.elf.cmd=avr-gcc

    [......]

    ## Combine gc-sections, archives, and objects
    recipe.c.combine.pattern="{compiler.path}{compiler.c.elf.cmd}" {compiler.c.elf.flags} -mmcu={build.mcu} -o "{build.path}/{build.project_name}.elf" {object_files} "{archive_file_path}" "-L{build.path}" -lm

### Recipes for extraction of executable files and other binary data

An arbitrary number of extra steps can be performed by the IDE at the end of objects linking.
These steps can be used to extract binary data used for upload and they are defined by a set of recipes with the following format:

    recipe.objcopy.FILE_EXTENSION_1.pattern=[.....]
    recipe.objcopy.FILE_EXTENSION_2.pattern=[.....]
    [.....]

`FILE_EXTENSION_x` must be replaced with the extension of the extracted file, for example the AVR platform needs two files a `.hex` and a `.eep`, so we made two recipes like:

    recipe.objcopy.eep.pattern=[.....]
    recipe.objcopy.hex.pattern=[.....]

There are no specific properties set by the IDE here.
A full example for the AVR platform can be:

    ## Create eeprom
    recipe.objcopy.eep.pattern="{compiler.path}{compiler.objcopy.cmd}" {compiler.objcopy.eep.flags} "{build.path}/{build.project_name}.elf" "{build.path}/{build.project_name}.eep"

    ## Create hex
    recipe.objcopy.hex.pattern="{compiler.path}{compiler.elf2hex.cmd}" {compiler.elf2hex.flags} "{build.path}/{build.project_name}.elf" "{build.path}/{build.project_name}.hex"

### Recipes to compute binary sketch size

At the end of the build the Arduino IDE shows the final binary sketch size to the user. The size is calculated using the recipe **recipe.size.pattern**. The output of the command executed using the recipe is parsed through the regular expression set in the property **recipe.size.regex**. The regular expression must match the sketch size.

For AVR we have:

    compiler.size.cmd=avr-size
    [....]
    ## Compute size
    recipe.size.pattern="{compiler.path}{compiler.size.cmd}" -A "{build.path}/{build.project_name}.hex"
    recipe.size.regex=Total\s+([0-9]+).*

### Recipes to export compiled binary

When you do a **Sketch > Export compiled Binary**, the compiled binary is copied from the build folder to the sketch folder. Two binaries are copied; the standard binary, and a binary that has been merged with the bootloader file (identified by the `.with_bootloader` in the filename).

Two recipes affect how **Export compiled Binary** works:
* **recipe.output.tmp_file**: Defines the binary's filename in the build folder.
* **recipe.output.save_file**: Defines the filename to use when copying the binary file to the sketch folder.

As with other processes, there are pre and post build hooks for **Export compiled Binary**.

The **recipe.hooks.savehex.presavehex.NUMBER.pattern** and **recipe.hooks.savehex.postsavehex.NUMBER.pattern** hooks (but not **recipe.output.tmp_file** and **recipe.output.save_file**) can be built concatenating the following properties set by the IDE:

    {sketch_path}              - the absolute path of the sketch folder

### Recipe to run the preprocessor
For detecting what libraries to include in the build, and for generating function prototypes, the Arduino IDE must be able to run (just) the preprocessor. For this, the **recipe.preproc.macros** recipe exists. This recipe must run the preprocessor on a given source file, writing the preprocessed output to a given output file, and generate (only) preprocessor errors on standard output. This preprocessor run should happen with the same defines and other preprocessor-influencing-options as for normally compiling the source files.

The recipes can be built concatenating other properties set by the IDE (for each file compiled):

* `{includes}`: the list of include paths in the format "-I/include/path -I/another/path...."
* `{source_file}`: the path to the source file
* `{preprocessed_file_path}`: the path to the output file

For example the following is used for AVR:

    preproc.macros.flags=-w -x c++ -E -CC
    recipe.preproc.macros="{compiler.path}{compiler.cpp.cmd}" {compiler.cpp.flags} {preproc.macros.flags} -mmcu={build.mcu} -DF_CPU={build.f_cpu} -DARDUINO={runtime.ide.version} -DARDUINO_{build.board} -DARDUINO_ARCH_{build.arch} {compiler.cpp.extra_flags} {build.extra_flags} {includes} "{source_file}" -o "{preprocessed_file_path}"

Note that the `{preprocessed_file_path}` might point to (your operating system's equivalent) of `/dev/null`. In this case, also passing `-MMD` to gcc is problematic, as it will try to generate a dependency file called `/dev/null.d`, which will usually result in a permission error. Since platforms typically include `{compiler.cpp.flags}` here, which includes `-MMD`, the Arduino IDE automatically filters out the `-MMD` option from the `recipe.preproc.macros` recipe to prevent this error.

Note that older IDE versions used the **recipe.preproc.includes** recipe to determine includes, which is undocumented here. Since Arduino IDE 1.6.7 (arduino-builder 1.2.0) this was changed and **recipe.preproc.includes** is no longer used.

### Pre and post build hooks (since IDE 1.6.5)

You can specify pre and post actions around each recipe. These are called "hooks". Here is the complete list of available hooks:
* `recipe.hooks.sketch.prebuild.NUMBER.pattern` (called before sketch compilation)
* `recipe.hooks.sketch.postbuild.NUMBER.pattern` (called after sketch compilation)
* `recipe.hooks.libraries.prebuild.NUMBER.pattern` (called before libraries compilation)
* `recipe.hooks.libraries.postbuild.NUMBER.pattern` (called after libraries compilation)
* `recipe.hooks.core.prebuild.NUMBER.pattern` (called before core compilation)
* `recipe.hooks.core.postbuild.NUMBER.pattern` (called after core compilation)
* `recipe.hooks.linking.prelink.NUMBER.pattern` (called before linking)
* `recipe.hooks.linking.postlink.NUMBER.pattern` (called after linking)
* `recipe.hooks.objcopy.preobjcopy.NUMBER.pattern` (called before objcopy recipes execution)
* `recipe.hooks.objcopy.postobjcopy.NUMBER.pattern` (called after objcopy recipes execution)
* `recipe.hooks.savehex.presavehex.NUMBER.pattern` (called before savehex recipe execution)
* `recipe.hooks.savehex.postsavehex.NUMBER.pattern` (called after savehex recipe execution)

Example: you want to execute 2 commands before sketch compilation and 1 after linking. You'll add to your platform.txt
```
recipe.hooks.sketch.prebuild.1.pattern=echo sketch compilation started at
recipe.hooks.sketch.prebuild.2.pattern=date

recipe.hooks.linking.postlink.1.pattern=echo linking is complete
```

Warning: hooks recipes are sorted before execution. If you need to write more than 10 recipes for a single hook, pad the number with a zero, for example:
```
recipe.hooks.sketch.prebuild.01.pattern=echo 1
recipe.hooks.sketch.prebuild.02.pattern=echo 2
...
recipe.hooks.sketch.prebuild.11.pattern=echo 11
```

# Global platform.txt

Properties defined in a platform.txt created in the **hardware** subfolder of the IDE installation folder will be used for all platforms and will override local properties.

# platform.local.txt

Introduced in Arduino IDE 1.5.7. This file can be used to override properties defined in platform.txt or define new properties without modifying platform.txt (e.g. when platform.txt is tracked by a version control system). It should be placed in the architecture folder.


# boards.txt

This file contains definitions and meta-data for the boards supported. Every board must be referred through its short name, the board ID. The settings for a board are defined through a set of properties
with keys having the board ID as prefix.

For example the board ID chosen for the Arduino Uno board is "uno". An extract of the Uno board configuration in boards.txt looks like:

    [......]
    uno.name=Arduino Uno
    uno.build.mcu=atmega328p
    uno.build.f_cpu=16000000L
    uno.build.board=AVR_UNO
    uno.build.core=arduino
    uno.build.variant=standard
    [......]

Note that all the relevant keys start with the board ID **uno.xxxxx**.

The **uno.name** property contains the name of the board shown in the Board menu of the Arduino IDE.

The **uno.build.board** property is used to set a compile-time variable **ARDUINO_{build.board}** to allow use of conditional code between `#ifdef`s. The Arduino IDE automatically generates a **build.board** value if not defined. In this case the variable defined at compile time will be `ARDUINO_AVR_UNO`.

The other properties will override the corresponding global properties of the IDE when the user selects the board. These properties will be globally available, in other configuration files too, without the board ID prefix:

    uno.build.mcu           =>   build.mcu
    uno.build.f_cpu         =>   build.f_cpu
    uno.build.board         =>   build.board
    uno.build.core          =>   build.core
    uno.build.variant       =>   build.variant

This explains the presence of **{build.mcu}** or **{build.board}** in the platform.txt recipes: their value is overwritten respectively by **{uno.build.mcu}** and **{uno.build.board}** when the Uno board is selected!
Moreover the IDE automatically provides the following properties:

* `{build.core.path}`: The path to the selected board's core folder (inside the [core platform](#platform-terminology), for example hardware/arduino/avr/core/arduino)
* `{build.system.path}`: The path to the [core platform](#platform-terminology)'s system folder if available (for example hardware/arduino/sam/system)
* `{build.variant.path}`: The path to the selected board variant folder (inside the [variant platform](#platform-terminology), for example hardware/arduino/avr/variants/micro)

## Cores

Cores are placed inside the **cores** subfolder. Many different cores can be provided within a single platform. For example the following could be a valid platform layout:

* `hardware/arduino/avr/cores/`: Cores folder for "avr" architecture, package "arduino"
* `hardware/arduino/avr/cores/arduino`: the Arduino Core
* `hardware/arduino/avr/cores/rtos`: an hypothetical RTOS Core

The board's property **build.core** is used by the Arduino IDE to find the core that must be compiled and linked when the board is selected. For example if a board needs the Arduino core the **build.core** variable should be set to:

    uno.build.core=arduino

or if the RTOS core is needed, to:

    uno.build.core=rtos

In any case the contents of the selected core folder are compiled and the core folder path is added to the include files search path.

## Core Variants

Sometimes a board needs some tweaking on default core configuration (different pin mapping is a typical example). A core variant folder is an additional folder that is compiled together with the core and allows to easily add specific configurations.

Variants must be placed inside the **variants** folder in the current architecture.
For example, Arduino AVR Boards uses:

* `hardware/arduino/avr/cores`: Core folder for "avr" architecture, "arduino" package
* `hardware/arduino/avr/cores/arduino`: The Arduino core
* `hardware/arduino/avr/variants/`: Variant folder for "avr" architecture, "arduino" package
* `hardware/arduino/avr/variants/standard`: ATmega328 based variants
* `hardware/arduino/avr/variants/leonardo`: ATmega32U4 based variants

In this example, the Arduino Uno board needs the *standard* variant so the **build.variant** property is set to *standard*:

    [.....]
    uno.build.core=arduino
    uno.build.variant=standard
    [.....]

instead, the Arduino Leonardo board needs the *leonardo* variant:

    [.....]
    leonardo.build.core=arduino
    leonardo.build.variant=leonardo
    [.....]

In the example above, both Uno and Leonardo share the same core but use different variants. \
In any case, the contents of the selected variant folder path is added to the include search path and its contents are compiled and linked with the sketch.

The parameter **build.variant.path** is automatically found by the IDE.

# Tools

The Arduino IDE uses external command line tools to upload the compiled sketch to the board or to burn bootloaders using external programmers. Currently *avrdude* is used for AVR based boards and *bossac* for SAM based boards, but there is no limit, any command line executable can be used. The command line parameters are specified using **recipes** in the same way used for platform build process.

Tools are configured inside the platform.txt file. Every Tool is identified by a short name, the Tool ID.
A tool can be used for different purposes:

- **upload** a sketch to the target board (using a bootloader preinstalled on the board)
- **program** a sketch to the target board using an external programmer
- **erase** the target board's flash memory using an external programmer
- burn a **bootloader** into the target board using an external programmer

Each action has its own recipe and its configuration is done through a set of properties having key starting with **tools** prefix followed by the tool ID and the action:

    [....]
    tools.avrdude.upload.pattern=[......]
    [....]
    tools.avrdude.program.pattern=[......]
    [....]
    tools.avrdude.erase.pattern=[......]
    [....]
    tools.avrdude.bootloader.pattern=[......]
    [.....]

A tool may have some actions not defined (it's not mandatory to define all four actions). \
Let's look at how the **upload** action is defined for avrdude:

    tools.avrdude.path={runtime.tools.avrdude.path}
    tools.avrdude.cmd.path={path}/bin/avrdude
    tools.avrdude.config.path={path}/etc/avrdude.conf

    tools.avrdude.upload.pattern="{cmd.path}" "-C{config.path}" -p{build.mcu} -c{upload.protocol} -P{serial.port} -b{upload.speed} -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"

A **{runtime.tools.TOOL_NAME.path}** and **{runtime.tools.TOOL_NAME-TOOL_VERSION.path}** property is generated for the tools of Arduino AVR Boards and any other platform installed via Boards Manager. **{runtime.tools.TOOL_NAME.path}** points to the latest version of the tool available.

The Arduino IDE makes the tool configuration properties available globally without the prefix. For example, the **tools.avrdude.cmd.path** property can be used as **{cmd.path}** inside the recipe, and the same happens for all the other avrdude configuration variables.

### Verbose parameter

It is possible for the user to enable verbosity from the Arduino IDE's Preferences panel. This preference is transferred to the command line by the IDE using the **ACTION.verbose** property (where ACTION is the action we are considering). \
When the verbose mode is enabled the **tools.TOOL_ID.ACTION.params.verbose** property is copied into **ACTION.verbose**. When the verbose mode is disabled, the **tools.TOOL_ID.ACTION.params.quiet** property is copied into **ACTION.verbose**. Confused? Maybe an example will clear things:

    tools.avrdude.upload.params.verbose=-v -v -v -v
    tools.avrdude.upload.params.quiet=-q -q
    tools.avrdude.upload.pattern="{cmd.path}" "-C{config.path}" {upload.verbose} -p{build.mcu} -c{upload.protocol} -P{serial.port} -b{upload.speed} -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"

In this example if the user enables verbose mode, then **{upload.params.verbose}** is used in **{upload.verbose}**:

    tools.avrdude.upload.params.verbose    =>    upload.verbose

If the user didn't enable verbose mode, the **{upload.params.quiet}** is used in **{upload.verbose}**:

    tools.avrdude.upload.params.quiet      =>    upload.verbose

## Sketch upload configuration

The Upload action is triggered when the user clicks on the "Upload" button on the IDE toolbar.
The Arduino IDE selects the tool to be used for upload by looking at the **upload.tool** property.
A specific **upload.tool** property should be defined for every board in boards.txt:

    [......]
    uno.upload.tool=avrdude
    [......]
    leonardo.upload.tool=avrdude
    [......]

Also other upload parameters can be defined together, for example in the Arduino boards.txt we have:

    [.....]
    uno.name=Arduino Uno
    uno.upload.tool=avrdude
    uno.upload.protocol=arduino
    uno.upload.maximum_size=32256
    uno.upload.speed=115200
    [.....]
    leonardo.name=Arduino Leonardo
    leonardo.upload.tool=avrdude
    leonardo.upload.protocol=avr109
    leonardo.upload.maximum_size=28672
    leonardo.upload.speed=57600
    leonardo.upload.use_1200bps_touch=true
    leonardo.upload.wait_for_upload_port=true
    [.....]

Most **{upload.XXXX}** variables are used later in the avrdude upload recipe in platform.txt:

    [.....]
    tools.avrdude.upload.pattern="{cmd.path}" "-C{config.path}" {upload.verbose} -p{build.mcu} -c{upload.protocol} -P{serial.port} -b{upload.speed} -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"
    [.....]

### 1200bps bootloader reset
Most Arduino boards use a dedicated USB-to-serial chip, that takes care of restarting the main MCU (starting the bootloader) when the serial port is opened. However, boards that have a native USB connection (such as the Leonardo or Zero) will have to disconnect from USB when rebooting into the bootloader (after which the bootloader reconnects to USB and offers a new serial port for uploading). After the upload is complete, the bootloader disconnects from USB again, starts the sketch, which then reconnects to USB. Because of these reconnections, the standard restart-on-serial open will not work, since that would cause the serial port to disappear and be closed again. Instead, the sketch running on these boards interpret a bitrate of 1200bps as a signal the bootloader should be started.

To let the IDE perform these steps, two board parameters can be set:

* `use_1200bps_touch` causes the IDE to briefly open the selected serial port at 1200bps (8N1) before starting the upload.
* `wait_for_upload_port` causes the IDE to wait for the serial port to (re)appear before and after the upload. This is only used when `use_1200bps_touch` is also set.  When set, after doing the 1200bps touch, the IDE will wait for a new serial port to appear and use that as the port for uploads. Alternatively, if the original port does not disappear within a few seconds, the upload continues with the original port (which can be the case if the board was already put into bootloader manually, or the IDE missed the disconnect and reconnect). Additionally, after the upload is complete, the IDE again waits for a new port to appear (or the originally selected port to be present).

Note that the IDE implementation of this 1200bps touch has some peculiarities, and the newer `arduino-cli` implementation also seems different (does not wait for the port after the reset, which is probably only needed in the IDE to prevent opening the wrong port on the serial monitor, and does not have a shorter timeout when the port never disappears).

## Serial port

The Arduino IDE auto-detects all available serial ports on the running system and lets the user choose one from the GUI. The selected port is available as a configuration property **{serial.port}**.

## Upload using an external programmer

**TODO...**
The platform.txt associated with the selected programmer will be used.

## Burn Bootloader

**TODO...**
The platform.txt associated with the selected board will be used.

# Custom board menus

The Arduino IDE allows adding extra menu items under the Tools menu. With these sub-menus the user can select different configurations for a specific board (for example a board could be provided in two or more variants with different CPUs, or may have different crystal speed based on the board model, and so on...).

Let's see an example of how a custom menu is implemented.
The board used in the example is the Arduino Duemilanove. This board was produced in two models, one with an ATmega168 CPU and another with an ATmega328P. \
We are going then to define a custom menu "Processor" that allows the user to choose between the two
different microcontrollers.

We must first define a set of **menu.MENU_ID=Text** properties. Text is what is displayed on the GUI for every custom menu we are going to create and must be declared at the beginning of the boards.txt file:

    menu.cpu=Processor
    [.....]

in this case we declare only one custom menu "Processor" which we refer using the "cpu" MENU_ID. \
Now let's add, always in the boards.txt file, the default configuration (common to all processors) for the duemilanove board:

    menu.cpu=Processor
    [.....]
    duemilanove.name=Arduino Duemilanove
    duemilanove.upload.tool=avrdude
    duemilanove.upload.protocol=arduino
    duemilanove.build.f_cpu=16000000L
    duemilanove.build.board=AVR_DUEMILANOVE
    duemilanove.build.core=arduino
    duemilanove.build.variant=standard
    [.....]

Now let's define the options to show in the "Processor" menu:

    [.....]
    duemilanove.menu.cpu.atmega328=ATmega328P
    [.....]
    duemilanove.menu.cpu.atmega168=ATmega168
    [.....]

We have defined two options: "ATmega328P" and "ATmega168". \
Note that the property keys must follow the format **BOARD_ID.menu.MENU_ID.OPTION_ID=Text**. \
Finally, the specific configuration for every option:

    [.....]
    ## Arduino Duemilanove w/ ATmega328P
    duemilanove.menu.cpu.atmega328=ATmega328P
    duemilanove.menu.cpu.atmega328.upload.maximum_size=30720
    duemilanove.menu.cpu.atmega328.upload.speed=57600
    duemilanove.menu.cpu.atmega328.build.mcu=atmega328p

    ## Arduino Duemilanove w/ ATmega168
    duemilanove.menu.cpu.atmega168=ATmega168
    duemilanove.menu.cpu.atmega168.upload.maximum_size=14336
    duemilanove.menu.cpu.atmega168.upload.speed=19200
    duemilanove.menu.cpu.atmega168.build.mcu=atmega168
    [.....]

Note that when the user selects an option, all the "sub properties" of that option are copied in the global configuration. For example when the user selects "ATmega168" from the "Processor" menu the Arduino IDE makes the configuration under atmega168 available globally:

    duemilanove.menu.cpu.atmega168.upload.maximum_size     =>   upload.maximum_size
    duemilanove.menu.cpu.atmega168.upload.speed            =>   upload.speed
    duemilanove.menu.cpu.atmega168.build.mcu               =>   build.mcu

There is no limit to the number of custom menus that can be defined.

**TODO: add an example with more than one submenu**

# Referencing another core, variant or tool

Inside the boards.txt we can define a board that uses a core provided by another vendor/mantainer using the syntax **VENDOR_ID:CORE_ID**. For example, if we want to define a board that uses the "arduino" core from the "arduino" vendor we should write:

    [....]
    myboard.name=My Wonderful Arduino Compatible board
    myboard.build.core=arduino:arduino
    [....]

Note that we don't need to specify any architecture since the same architecture of "myboard" is used, so we just say "arduino:arduino" instead of "arduino:avr:arduino".

The platform.txt settings are inherited from the referenced core platform, thus there is no need to provide a platform.txt unless there are some specific properties that need to be overridden.

The libraries from the referenced platform are used, thus there is no need to provide a libraries. If libraries are provided the list of available libraries are the sum of the 2 libraries where the referencing platform has priority over the referenced platform.

In the same way we can use variants and tools defined on another platform:

    [....]
    myboard.build.variant=arduino:standard
    myboard.upload.tool=arduino:avrdude
    myboard.bootloader.tool=arduino:avrdude
    [....]

Using this syntax allows us to reduce the minimum set of files needed to define a new "hardware" to just the boards.txt file.

Note that referencing a variant in another platform does *not* inherit any properties from that platform's platform.txt (like referencing a core does).

## Platform Terminology
Because boards can reference cores, variants and tools in different platforms, this means that a single build or upload can use data from up to four different platforms. To keep this clear, the following terminology is used:
* The "board platform" is the platform that defines the currently selected board (e.g. the platform that contains the board.txt the board is defined in.
* The "core platform" is the the platform that contains the core to be used.
* The "variant platform" is the platform that contains the variant to be used.
* The "tool platform" is the platform that contains the tool used for the current operation.

In the most common case, without any references, all of these will refer to the same platform.

Note that the above terminology is not in widespread use, but was invented for clarity within this document. In the actual arduino-cli code, the "board platform" is called `targetPlatform`, the "core platform" is called `actualPlatform`, the others are pretty much nameless.

# boards.local.txt

Introduced in Arduino IDE 1.6.6. This file can be used to override properties defined in boards.txt or define new properties without modifying boards.txt.


# keywords.txt

As of Arduino IDE 1.6.6, per-platform keywords can be defined by adding a keywords.txt file to the platform's architecture folder. These keywords are only highlighted when one of the boards of that platform are selected. This file follows the [same format](https://github.com/arduino/Arduino/wiki/Arduino-IDE-1.5:-Library-specification#keywords) as the keywords.txt used in libraries. Each keyword must be separated from the keyword identifier by a tab.
