module github.com/arduino/arduino-cli

go 1.22

toolchain go1.22.3

// We must use this fork until https://github.com/mailru/easyjson/pull/372 is merged
replace github.com/mailru/easyjson => github.com/cmaglie/easyjson v0.8.1

require (
	github.com/arduino/board-discovery v0.0.0-20180823133458-1ba29327fb0c
	github.com/arduino/go-paths-helper v1.12.1
	github.com/arduino/go-properties-orderedmap v1.7.1
	github.com/arduino/go-timeutils v0.0.0-20171220113728-d1dd9e313b1b
	github.com/arduino/go-win32-utils v0.0.0-20180330194947-ed041402e83b
	github.com/cmaglie/pb v1.0.27
	github.com/codeclysm/cc v1.2.2 // indirect
	github.com/djherbis/buffer v1.1.0
	github.com/djherbis/nio/v3 v3.0.1
	github.com/fatih/color v1.7.0
	github.com/fluxio/iohelpers v0.0.0-20160419043813-3a4dd67a94d2 // indirect
	github.com/fluxio/multierror v0.0.0-20160419044231-9c68d39025e5 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leonelquinteros/gotext v1.4.0
	github.com/mailru/easyjson v0.7.7
	github.com/marcinbor85/gohex v0.0.0-20210308104911-55fb1c624d84
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.14
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/miekg/dns v1.1.43 // indirect
	github.com/oleksandr/bonjour v0.0.0-20160508152359-5dcf00d8b228 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmylund/sortutil v0.0.0-20120526081524-abeda66eb583
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/schollz/closestmatch v2.1.0+incompatible
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.2.1
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.9.0
	go.bug.st/cleanup v1.0.0
	go.bug.st/downloader/v2 v2.1.1
	go.bug.st/relaxed-semver v0.9.0
	go.bug.st/serial v1.3.2
	go.bug.st/serial.v1 v0.0.0-20180827123349-5f7892a7bb45 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20210505024714-0287a6fb4125 // indirect
	golang.org/x/text v0.3.6
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/codeclysm/extract/v4 v4.0.0
	github.com/rogpeppe/go-internal v1.3.0
	go.bug.st/testifyjson v1.1.1
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/itchyny/gojq v0.12.8 // indirect
	github.com/itchyny/timefmt-go v0.1.3 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/juju/errors v0.0.0-20181118221551-089d3ea4e4d5 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/klauspost/compress v1.15.13 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	golang.org/x/sys v0.16.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
