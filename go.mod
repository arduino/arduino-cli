module github.com/arduino/arduino-cli

go 1.21

// We must use this fork until https://github.com/mailru/easyjson/pull/372 is merged
replace github.com/mailru/easyjson => github.com/cmaglie/easyjson v0.8.1

require (
	github.com/ProtonMail/go-crypto v0.0.0-20230828082145-3c4c8a2d2371
	github.com/arduino/go-paths-helper v1.9.0
	github.com/arduino/go-properties-orderedmap v1.8.0
	github.com/arduino/go-timeutils v0.0.0-20171220113728-d1dd9e313b1b
	github.com/arduino/go-win32-utils v1.0.0
	github.com/cmaglie/pb v1.0.27
	github.com/codeclysm/extract/v3 v3.1.1
	github.com/djherbis/buffer v1.1.0
	github.com/djherbis/nio/v3 v3.0.1
	github.com/fatih/color v1.7.0
	github.com/gofrs/uuid/v5 v5.0.0
	github.com/leonelquinteros/gotext v1.4.0
	github.com/mailru/easyjson v0.7.7
	github.com/marcinbor85/gohex v0.0.0-20210308104911-55fb1c624d84
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.14
	github.com/pkg/errors v0.9.1
	github.com/pmylund/sortutil v0.0.0-20120526081524-abeda66eb583
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/rogpeppe/go-internal v1.3.0
	github.com/schollz/closestmatch v2.1.0+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.2.1
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.8.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.bug.st/cleanup v1.0.0
	go.bug.st/downloader/v2 v2.1.1
	go.bug.st/relaxed-semver v0.11.0
	go.bug.st/serial v1.6.1
	go.bug.st/testifyjson v1.1.1
	golang.org/x/term v0.6.0
	golang.org/x/text v0.8.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230526203410-71b5a4ffd15e
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/itchyny/gojq v0.12.8 // indirect
	github.com/itchyny/timefmt-go v0.1.3 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/juju/errors v0.0.0-20181118221551-089d3ea4e4d5 // indirect
	github.com/juju/loggo v1.0.0 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/klauspost/compress v1.15.13 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/ulikunitz/xz v0.5.11 // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
