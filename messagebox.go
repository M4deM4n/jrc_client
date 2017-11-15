package main

import (
	"syscall"
	"unsafe"
)

const MB_ABORTRETRYIGNORE = 0x00000002
const MB_CANCELTRYCONTINUE = 0x00000006
const MB_HELP = 0x00004000
const MB_OK = 0x00000000
const MB_CANCEL = 0x00000001
const MB_RETRYCANCEL = 0x00000005
const MB_YESNO = 0x00000004
const MB_YESNOCANCEL = 0x00000003

var (
	msgBoxProc *syscall.LazyProc
)

func msgBoxInit() {
	mod := syscall.NewLazyDLL("user32.dll")
	msgBoxProc = mod.NewProc("MessageBoxW")
}

func MessageBox(title string, message string, buttons int) int {
	ret, _, _ := msgBoxProc.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(buttons))

	return int(ret)
}
