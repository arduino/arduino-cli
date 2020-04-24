# I18N

## Usage

In the source code, use the function `i18n.Tr("message", ...args)` to get a localized string. This tool parses the source using the `go/ast` package to generate the `en` locale using these messages.

## Updating messages to reflect code changes

The following command updates the locales present in the source code to reflect addition/removal of messages.

```sh
task i18n:update
```