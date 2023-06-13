// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

//go:build darwin && cgo

package i18n

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#import <Foundation/Foundation.h>

const char* getLocaleIdentifier() {
    NSString *cs = [[NSLocale currentLocale] localeIdentifier];
    const char *cstr = [cs UTF8String];
    return cstr;
}

*/
import "C"

func getLocaleIdentifier() string {
	if envLocale := getLocaleIdentifierFromEnv(); envLocale != "" {
		return envLocale
	}
	return C.GoString(C.getLocaleIdentifier())
}
