package matchers

import "bytes"

var (
	TypeDoc  = newType("doc", "application/msword")
	TypeDocx = newType("docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	TypeXls  = newType("xls", "application/vnd.ms-excel")
	TypeXlsx = newType("xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	TypePpt  = newType("ppt", "application/vnd.ms-powerpoint")
	TypePptx = newType("pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
)

var Document = Map{
	TypeDoc:  Doc,
	TypeDocx: Docx,
	TypeXls:  Xls,
	TypeXlsx: Xlsx,
	TypePpt:  Ppt,
	TypePptx: Pptx,
}

func Doc(buf []byte) bool {
	return len(buf) > 7 &&
		buf[0] == 0xD0 && buf[1] == 0xCF &&
		buf[2] == 0x11 && buf[3] == 0xE0 &&
		buf[4] == 0xA1 && buf[5] == 0xB1 &&
		buf[6] == 0x1A && buf[7] == 0xE1
}

func Docx(buf []byte) bool {
	return len(buf) > 3 &&
		buf[0] == 0x50 && buf[1] == 0x4B &&
		buf[2] == 0x03 && buf[3] == 0x04 &&
		bytes.Contains(buf[:256], []byte(TypeDocx.MIME.Value))
}

func Xls(buf []byte) bool {
	return len(buf) > 7 &&
		buf[0] == 0xD0 && buf[1] == 0xCF &&
		buf[2] == 0x11 && buf[3] == 0xE0 &&
		buf[4] == 0xA1 && buf[5] == 0xB1 &&
		buf[6] == 0x1A && buf[7] == 0xE1
}

func Xlsx(buf []byte) bool {
	return len(buf) > 3 &&
		buf[0] == 0x50 && buf[1] == 0x4B &&
		buf[2] == 0x03 && buf[3] == 0x04 &&
		bytes.Contains(buf[:256], []byte(TypeXlsx.MIME.Value))
}

func Ppt(buf []byte) bool {
	return len(buf) > 7 &&
		buf[0] == 0xD0 && buf[1] == 0xCF &&
		buf[2] == 0x11 && buf[3] == 0xE0 &&
		buf[4] == 0xA1 && buf[5] == 0xB1 &&
		buf[6] == 0x1A && buf[7] == 0xE1
}

func Pptx(buf []byte) bool {
	return len(buf) > 3 &&
		buf[0] == 0x50 && buf[1] == 0x4B &&
		buf[2] == 0x07 && buf[3] == 0x08 &&
		bytes.Contains(buf[:256], []byte(TypePptx.MIME.Value))
}
