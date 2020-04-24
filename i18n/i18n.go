package i18n

// Tr returns msg translated to the selected locale
// the msg argument must be a literal string
func Tr(msg string, args ...interface{}) string {
	return po.Get(msg, args...)
}
