# The three pillars of the Arduino CLI

The Arduino CLI is an open source Command Line Application written in [Golang] that can be used from a terminal to
compile, verify and upload sketches to Arduino boards and that’s capable of managing all the software and tools needed
in the process. But don’t get fooled by its name: Arduino CLI can do much more than the average console application, as
shown by [Arduino IDE 2.x][arduino ide 2.x] and [Arduino Cloud], which rely on it for similar purposes but each one in a
completely different way from the other. In this article we introduce the three pillars of the Arduino CLI, explaining
how we designed the software so that it can be effectively leveraged under different scenarios.

## The first pillar: command line interface

### Console applications for humans

As you might expect, the first way to use the Arduino CLI is from a terminal and by a human, and user experience plays a
key role here. The UX is under a continuous improvement process as we want the tool to be powerful without being too
complicated. We heavily rely on sub-commands to provide a rich set of different operations logically grouped together,
so that users can easily explore the interface while getting very specific contextual help (even in Chinese!).

```
$ LANG=zh arduino-cli
Arduino 命令行界面 (arduino-cli)

用法：
  arduino-cli [command]

示例：
  arduino-cli <命令> [参数...]

可用命令：
  board           Arduino 开发板命令
  burn-bootloader 上传引导加载程序。
  cache           Arduino 缓存命令。
  compile         编译 Arduino 项目
  completion      已生成脚本
  config          Arduino 配置命令。
  core            Arduino 内核操作。
  daemon          在端口上作为守护进程运行：50051
  debug           调试 Arduino 项目
  help            Help about any command
  lib             Arduino 关于库的命令。
  monitor         开启开发板的通信端口。
  outdated        列出可以升级的内核和库
  sketch          Arduino CLI 项目命令
  update          更新内核和库的索引
  upgrade         升级已安装的内核和库。
  upload          上传 Arduino 项目。
  version         显示 Arduino CLI 的版本号。

参数：
      --additional-urls strings   以逗号分隔的开发板管理器附加地址列表。
      --config-file string        自定义配置文件（如果未指定，将使用默认值）。
      --format string             日志的输出格​​式，可以是：text, json, jsonmini, yaml (default "text")
  -h, --help                      help for arduino-cli
      --log                       在标准输出上打印日志。
      --log-file string           写入日志的文件的路径。
      --log-format string         日志的输出格​​式，可以是：text, json
      --log-level string          记录此级别及以上的消息。有效级别为 trace, debug, info, warn, error, fatal, panic
      --no-color                  Disable colored output.

使用 "arduino-cli [command] --help" 获取有关命令的更多信息。
```

### Console applications for robots

Humans are not the only type of customers we want to support and the Arduino CLI was also designed to be used
programmatically - think about automation pipelines or a [CI][continuous integration]/[CD][continuous deployment]
system. There are some niceties to observe when you write software that’s supposed to be easy to run when unattended and
one in particular is the ability to run without a configuration file. This is possible because every configuration
option you find in the arduino-cli.yaml configuration file can be provided either through a command line flag or by
setting an environment variable. To give an example, the following commands are all equivalent and will fetch the
external package index for ESP32 platforms:

```
$ cat ~/.arduino15/arduino-cli.yaml
board_manager:
  additional_urls:
    - https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json

$ arduino-cli core update-index
Downloading index: package_index.tar.bz2 downloaded
Downloading index: package_esp32_index.json downloaded
```

or:

```
$ export ARDUINO_BOARD_MANAGER_ADDITIONAL_URLS="https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json"
$ arduino-cli core update-index
Downloading index: package_index.tar.bz2 downloaded
Downloading index: package_esp32_index.json downloaded
```

or:

```
$ arduino-cli core update-index --additional-urls="https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json"
Downloading index: package_index.tar.bz2 downloaded
Downloading index: package_esp32_index.json downloaded
```

See the [configuration documentation] for details about Arduino CLI's configuration system.

Consistent with the previous paragraph, when it comes to providing output the Arduino CLI aims to be user friendly but
also slightly verbose, something that doesn’t play well with robots. This is why we added an option to provide output
that’s easy to parse. For example, the following figure shows what getting the result in [JSON] format and filtering
using the `jq` tools may look like:

```
$ arduino-cli lib search FlashStorage --format json | jq .libraries[0].latest
{
  "author": "Various",
  "version": "1.0.0",
  "maintainer": "Arduino <info@arduino.cc>",
  "sentence": "The FlashStorage library aims to provide a convenient way to store and retrieve user's data using the non-volatile flash memory of microcontrollers.",
  "paragraph": "Useful if the EEPROM is not available or too small. Currently, ATSAMD21 and ATSAMD51 cpu are supported (and consequently every board based on this cpu like the Arduino Zero or Aduino MKR1000).",
  "website": "https://github.com/cmaglie/FlashStorage",
  "category": "Data Storage",
  "architectures": [
    "samd"
  ],
  "types": [
    "Contributed"
  ],
  "resources": {
    "url": "https://downloads.arduino.cc/libraries/github.com/cmaglie/FlashStorage-1.0.0.zip",
    "archive_filename": "FlashStorage-1.0.0.zip",
    "checksum": "SHA-256:2f5a349e1c5dc4ec7f7e22268c0f998af3f471b98ed873abd6e671ac67940e39",
    "size": 12265,
    "cache_path": "libraries"
  }
}
```

Even if not related to software design, one last feature that’s worth mentioning is the availability of a one-line
[installation script] that can be used to make the latest version of the Arduino CLI available on most systems with an
HTTP client like curl or wget and a shell like bash.

For more information on Arduino CLI's command line interface, see the [command reference].

## The second pillar: gRPC interface

[gRPC] is a high performance [RPC] framework that can efficiently connect client and server applications. The Arduino
CLI can act as a gRPC server (we call it [daemon mode]), exposing a set of procedures that implement the very same set
of features of the command line interface and waiting for clients to connect and use them. To give an idea, the
following is some [Golang] code capable of retrieving the version number of a remote running Arduino CLI server
instance:

```go
// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package main

import (
	"context"
	"log"
	"time"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Establish a connection with the gRPC server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	conn, err := grpc.DialContext(ctx, "localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	cancel()
	if err != nil {
		log.Println(err)
		log.Fatal("error connecting to arduino-cli rpc server, you can start it by running `arduino-cli daemon`")
	}
	defer conn.Close()

	// Create an instance of the gRPC clients.
	cli := rpc.NewArduinoCoreServiceClient(conn)

	// Now we can call various methods of the API...
	versionResp, err := cli.Version(context.Background(), &rpc.VersionRequest{})
	if err != nil {
		log.Fatalf("Error getting version: %s", err)
	}
	log.Printf("arduino-cli version: %v", versionResp.GetVersion())
}
```

gRPC is language agnostic: even if the example is written in Golang, the programming language used for the client can be
Python, JavaScript or any of the many [supported ones][grpc supported languages], leading to a variety of possible
scenarios. [Arduino IDE 2.x][arduino ide 2.x] is a good example of how to leverage the daemon mode of the Arduino CLI
with a clean separation of concerns: the IDE knows nothing about how to download a core, compile a sketch or talk to an
Arduino board and it demands all these features of an Arduino CLI instance. Conversely, the Arduino CLI doesn’t even
know that the client that’s connected is the Arduino IDE, and neither does it care.

For more information on Arduino CLI's gRPC interface, see the [gRPC interface reference].

## The third pillar: embedding

Arduino CLI is written in [Golang] and the code is organized in a way that makes it easy to use it as a library by
including the modules you need in another Golang application at compile time. Both the first and second pillars rely on
a common Golang API, based on the gRPC protobuf definitions: a set of functions that abstract all the functionalities
offered by the Arduino CLI, so that when we provide a fix or a new feature, they are automatically available to both the
command line and gRPC interfaces. The source modules implementing this API are implemented through the `commands`
package, and it can be imported in other Golang programs to embed a full-fledged Arduino CLI. For example, this is how
some backend services powering [Arduino Cloud] can compile sketches and manage libraries. Just to give you a taste of
what it means to embed the Arduino CLI, here is how to search for a core using the API:

```go
// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a new ArduinoCoreServer
	srv := commands.NewArduinoCoreServer()

	// Disable logging
	logrus.SetOutput(io.Discard)

	// Create a new instance in the server
	ctx := context.Background()
	resp, err := srv.Create(ctx, &rpc.CreateRequest{})
	if err != nil {
		log.Fatal("Error creating instance:", err)
	}
	instance := resp.GetInstance()

	// Defer the destruction of the instance
	defer func() {
		if _, err := srv.Destroy(ctx, &rpc.DestroyRequest{Instance: instance}); err != nil {
			log.Fatal("Error destroying instance:", err)
		}
		fmt.Println("Instance successfully destroyed")
	}()

	// Initialize the instance
	initStream := commands.InitStreamResponseToCallbackFunction(ctx, func(r *rpc.InitResponse) error {
		fmt.Println("INIT> ", r)
		return nil
	})
	if err := srv.Init(&rpc.InitRequest{Instance: instance}, initStream); err != nil {
		log.Fatal("Error during initialization:", err)
	}

	// Search for platforms and output the result
	searchResp, err := srv.PlatformSearch(ctx, &rpc.PlatformSearchRequest{Instance: instance})
	if err != nil {
		log.Fatal("Error searching for platforms:", err)
	}
	for _, platformSummary := range searchResp.GetSearchOutput() {
		installed := platformSummary.GetInstalledRelease()
		meta := platformSummary.GetMetadata()
		fmt.Printf("%30s %8s %s\n", meta.GetId(), installed.GetVersion(), installed.GetName())
	}
}
```

Embedding the Arduino CLI is limited to Golang applications and requires a deep knowledge of its internals. For the
average use case, the gRPC interface might be a better alternative. Nevertheless, this remains a valid option that we
use and provide support for.

## Conclusions

You can start playing with the Arduino CLI right away. The code is open source and [the repo][arduino cli repository]
contains [example code showing how to implement a gRPC client][grpc client example]. If you’re curious about how we
designed the low level API, have a look at the [commands package] and don’t hesitate to leave feedback on the [issue
tracker] if you’ve got a use case that doesn’t fit one of the three pillars.

[golang]: https://go.dev/
[arduino ide 2.x]: https://github.com/arduino/arduino-ide
[arduino cloud]: https://cloud.arduino.cc/home
[continuous integration]: https://en.wikipedia.org/wiki/Continuous_integration
[continuous deployment]: https://en.wikipedia.org/wiki/Continuous_deployment
[configuration documentation]: configuration.md
[json]: https://www.json.org
[installation script]: installation.md#use-the-install-script
[command reference]: commands/arduino-cli.md
[grpc]: https://grpc.io/
[rpc]: https://en.wikipedia.org/wiki/Remote_procedure_call
[daemon mode]: commands/arduino-cli_daemon.md
[grpc interface reference]: rpc/commands.md
[grpc supported languages]: https://grpc.io/docs/languages/
[arduino cli repository]: https://github.com/arduino/arduino-cli
[grpc client example]: https://github.com/arduino/arduino-cli/blob/master/rpc/internal/client_example
[commands package]: https://github.com/arduino/arduino-cli/tree/master/commands
[issue tracker]: https://github.com/arduino/arduino-cli/issues
[contextual help screenshot]: img/CLI_contextual_help_screenshot.png
