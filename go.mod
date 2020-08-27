module github.com/arduino/arduino-cli

go 1.14

// This one must be kept until https://github.com/GeertJohan/go.rice/pull/159 is merged
replace github.com/GeertJohan/go.rice => github.com/cmaglie/go.rice v1.0.1

require (
	bou.ke/monkey v1.0.1
	github.com/GeertJohan/go.rice v1.0.0
	github.com/arduino/board-discovery v0.0.0-20180823133458-1ba29327fb0c
	github.com/arduino/go-paths-helper v1.2.0
	github.com/arduino/go-properties-orderedmap v1.3.0
	github.com/arduino/go-timeutils v0.0.0-20171220113728-d1dd9e313b1b
	github.com/arduino/go-win32-utils v0.0.0-20180330194947-ed041402e83b
	github.com/cmaglie/pb v1.0.27
	github.com/codeclysm/cc v1.2.2 // indirect
	github.com/codeclysm/extract/v3 v3.0.1
	github.com/fatih/color v1.7.0
	github.com/fluxio/iohelpers v0.0.0-20160419043813-3a4dd67a94d2 // indirect
	github.com/fluxio/multierror v0.0.0-20160419044231-9c68d39025e5 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/protobuf v1.4.1
	github.com/h2non/filetype v1.0.8 // indirect
	github.com/imjasonmiller/godice v0.1.2
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/leonelquinteros/gotext v1.4.0
	github.com/marcinbor85/gohex v0.0.0-20200531163658-baab2527a9a2
	github.com/mattn/go-colorable v0.1.2
	github.com/mattn/go-isatty v0.0.8
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/miekg/dns v1.0.5 // indirect
	github.com/oleksandr/bonjour v0.0.0-20160508152359-5dcf00d8b228 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmylund/sortutil v0.0.0-20120526081524-abeda66eb583
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/schollz/closestmatch v2.1.0+incompatible
	github.com/segmentio/stats/v4 v4.5.3
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.0.1-0.20200710201246-675ae5f5a98c
	github.com/spf13/jwalterweatherman v1.0.0
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.6.1
	go.bug.st/cleanup v1.0.0
	go.bug.st/downloader/v2 v2.0.1
	go.bug.st/relaxed-semver v0.0.0-20190922224835-391e10178d18
	go.bug.st/serial v1.0.0
	go.bug.st/serial.v1 v0.0.0-20180827123349-5f7892a7bb45 // indirect
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	golang.org/x/text v0.3.2
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce // indirect
	gopkg.in/yaml.v2 v2.3.0
)
