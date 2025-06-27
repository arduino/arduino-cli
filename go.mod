module github.com/arduino/arduino-cli

go 1.24.1

// We must use this fork until https://github.com/mailru/easyjson/pull/372 is merged
replace github.com/mailru/easyjson => github.com/cmaglie/easyjson v0.8.1

require (
	fortio.org/safecast v1.0.0
	github.com/ProtonMail/go-crypto v1.3.0
	github.com/arduino/go-paths-helper v1.13.1
	github.com/arduino/go-properties-orderedmap v1.8.1
	github.com/arduino/go-serial-utils v0.1.2
	github.com/arduino/go-timeutils v0.0.0-20171220113728-d1dd9e313b1b
	github.com/arduino/go-win32-utils v1.0.0
	github.com/arduino/pluggable-discovery-protocol-handler/v2 v2.2.1
	github.com/arduino/pluggable-monitor-protocol-handler v0.9.2
	github.com/cmaglie/pb v1.0.27
	github.com/codeclysm/extract/v4 v4.0.0
	github.com/djherbis/buffer v1.2.0
	github.com/djherbis/nio/v3 v3.0.1
	github.com/fatih/color v1.18.0
	github.com/go-git/go-git/v5 v5.16.0
	github.com/gofrs/uuid/v5 v5.3.2
	github.com/leonelquinteros/gotext v1.7.1
	github.com/mailru/easyjson v0.7.7
	github.com/marcinbor85/gohex v0.0.0-20210308104911-55fb1c624d84
	github.com/mattn/go-colorable v0.1.14
	github.com/mattn/go-isatty v0.0.20
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/rogpeppe/go-internal v1.14.1
	github.com/schollz/closestmatch v2.1.0+incompatible
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.bug.st/cleanup v1.0.0
	go.bug.st/downloader/v2 v2.2.0
	go.bug.st/f v0.4.0
	go.bug.st/relaxed-semver v0.15.0
	go.bug.st/testifyjson v1.3.0
	golang.org/x/sys v0.33.0
	golang.org/x/term v0.32.0
	golang.org/x/text v0.26.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v3 v3.0.1
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/gojq v0.12.8 // indirect
	github.com/itchyny/timefmt-go v0.1.3 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/juju/errors v1.0.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	go.bug.st/serial v1.6.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/exp v0.0.0-20241217172543-b2144cdd0a67 // indirect
	golang.org/x/net v0.39.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)
