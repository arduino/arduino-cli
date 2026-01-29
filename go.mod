module github.com/arduino/arduino-cli

go 1.24.4

// We must use this fork until https://github.com/mailru/easyjson/pull/372 is merged
replace github.com/mailru/easyjson => github.com/cmaglie/easyjson v0.8.1

require (
	github.com/arduino/go-paths-helper v1.12.1
	github.com/arduino/go-properties-orderedmap v1.7.1
	github.com/arduino/go-timeutils v0.0.0-20171220113728-d1dd9e313b1b
	github.com/arduino/go-win32-utils v0.0.0-20180330194947-ed041402e83b
	github.com/cmaglie/pb v1.0.27
	github.com/djherbis/buffer v1.1.0
	github.com/djherbis/nio/v3 v3.0.1
	github.com/fatih/color v1.7.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/leonelquinteros/gotext v1.4.0
	github.com/mailru/easyjson v0.7.7
	github.com/marcinbor85/gohex v0.0.0-20210308104911-55fb1c624d84
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.14
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmylund/sortutil v0.0.0-20120526081524-abeda66eb583
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/schollz/closestmatch v2.1.0+incompatible
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.2.1
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.10.0
	go.bug.st/cleanup v1.0.0
	go.bug.st/downloader/v2 v2.1.1
	go.bug.st/relaxed-semver v0.9.0
	go.bug.st/serial v1.3.2
	golang.org/x/crypto v0.37.0
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/text v0.24.0
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/codeclysm/extract/v4 v4.0.0
	github.com/go-git/go-git/v5 v5.16.4
	github.com/rogpeppe/go-internal v1.14.1
	go.bug.st/testifyjson v1.1.1
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/itchyny/gojq v0.12.8 // indirect
	github.com/itchyny/timefmt-go v0.1.3 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/juju/errors v0.0.0-20181118221551-089d3ea4e4d5 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.15.13 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
