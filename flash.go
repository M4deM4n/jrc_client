package main

import (
	"syscall"
	"unsafe"
)

var (
	hwnd         uintptr
	conWinProc   *syscall.LazyProc
	flashWinProc *syscall.LazyProc
)

func initFlash() {
	k32 := syscall.NewLazyDLL("kernel32.dll")
	u32 := syscall.NewLazyDLL("user32.dll")

	flashWinProc = u32.NewProc("FlashWindow")
	conWinProc = k32.NewProc("GetConsoleWindow")

}

func flashWindow(flash bool) {
	hwnd, _, _ = conWinProc.Call()
	_, _, _ = flashWinProc.Call(hwnd, uintptr(unsafe.Pointer(&flash)))
}
