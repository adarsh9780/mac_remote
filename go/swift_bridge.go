package main

/*
#cgo LDFLAGS: -L. system_helper.o -L/usr/lib/swift
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

float swift_getBrightness();
void swift_setBrightness(float val);
void swift_initMediaRemote();
char* swift_getMediaInfoJSON();
bool swift_sendMediaCommand(int32_t cmd);
void swift_moveMouse(double dx, double dy);
void swift_scrollMouse(int32_t dy, int32_t dx);
void swift_clickMouse(bool right);
void swift_mouseDown(bool right);
void swift_mouseUp(bool right);
void swift_typeText(const char* text);
void swift_pressKey(uint16_t keyCode);
float swift_getVolume();
void swift_setVolume(float vol);
void swift_setMute();
void swift_showDock();
void swift_switchToApp(int32_t pid);
char* swift_getRunningAppsHash();
char* swift_getRunningAppsJSON();
bool swift_isTextInputFocused();
void swift_showQRUI(char* url);
void swift_closeQRUI();
void swift_showCodeUI(char* code);
void swift_closeCodeUI();
void swift_showSuccessUI(char* info);
*/
import "C"
import (
	"unsafe"
)

// swiftIsTextInputFocused is undocumented. Please add documentation.
func swiftIsTextInputFocused() bool {
	return bool(C.swift_isTextInputFocused())
}

// swiftGetBrightness is undocumented. Please add documentation.
func swiftGetBrightness() float32 {
	return float32(C.swift_getBrightness())
}

// swiftShowQRUI is undocumented. Please add documentation.
func swiftShowQRUI(url string) {
	cUrl := C.CString(url)
	defer C.free(unsafe.Pointer(cUrl))
	C.swift_showQRUI(cUrl)
}

// swiftCloseQRUI is undocumented. Please add documentation.
func swiftCloseQRUI() {
	C.swift_closeQRUI()
}

// swiftShowCodeUI is undocumented. Please add documentation.
func swiftShowCodeUI(code string) {
	cCode := C.CString(code)
	defer C.free(unsafe.Pointer(cCode))
	C.swift_showCodeUI(cCode)
}

// swiftCloseCodeUI is undocumented. Please add documentation.
func swiftCloseCodeUI() {
	C.swift_closeCodeUI()
}

// swiftShowSuccessUI is undocumented. Please add documentation.
func swiftShowSuccessUI(info string) {
	cInfo := C.CString(info)
	defer C.free(unsafe.Pointer(cInfo))
	C.swift_showSuccessUI(cInfo)
}

// swiftSetBrightness is undocumented. Please add documentation.
func swiftSetBrightness(val float32) {
	C.swift_setBrightness(C.float(val))
}

// swiftInitMediaRemote is undocumented. Please add documentation.
func swiftInitMediaRemote() {
	C.swift_initMediaRemote()
}

// swiftGetMediaInfoJSON is undocumented. Please add documentation.
func swiftGetMediaInfoJSON() string {
	cStr := C.swift_getMediaInfoJSON()
	if cStr == nil {
		return "{}"
	}
	defer C.free(unsafe.Pointer(cStr))
	return C.GoString(cStr)
}

// swiftSendMediaCommand is undocumented. Please add documentation.
func swiftSendMediaCommand(cmd int) bool {
	return bool(C.swift_sendMediaCommand(C.int32_t(cmd)))
}

// swiftMoveMouse is undocumented. Please add documentation.
func swiftMoveMouse(dx, dy float64) {
	C.swift_moveMouse(C.double(dx), C.double(dy))
}

// swiftScrollMouse is undocumented. Please add documentation.
func swiftScrollMouse(dy, dx int) {
	C.swift_scrollMouse(C.int32_t(dy), C.int32_t(dx))
}

// swiftClickMouse is undocumented. Please add documentation.
func swiftClickMouse(right bool) {
	C.swift_clickMouse(C.bool(right))
}

// swiftMouseDown is undocumented. Please add documentation.
func swiftMouseDown(right bool) {
	C.swift_mouseDown(C.bool(right))
}

// swiftMouseUp is undocumented. Please add documentation.
func swiftMouseUp(right bool) {
	C.swift_mouseUp(C.bool(right))
}

// swiftTypeText is undocumented. Please add documentation.
func swiftTypeText(text string) {
	cStr := C.CString(text)
	defer C.free(unsafe.Pointer(cStr))
	C.swift_typeText(cStr)
}

// swiftPressKey is undocumented. Please add documentation.
func swiftPressKey(keyCode uint16) {
	C.swift_pressKey(C.uint16_t(keyCode))
}

// swiftGetVolume is undocumented. Please add documentation.
func swiftGetVolume() float32 {
	return float32(C.swift_getVolume())
}

// swiftSetVolume is undocumented. Please add documentation.
func swiftSetVolume(val float32) {
	C.swift_setVolume(C.float(val))
}

// swiftSetMute is undocumented. Please add documentation.
func swiftSetMute() {
	C.swift_setMute()
}

// swiftShowDock is undocumented. Please add documentation.
func swiftShowDock() {
	C.swift_showDock()
}

// swiftSwitchToApp is undocumented. Please add documentation.
func swiftSwitchToApp(pid int) {
	C.swift_switchToApp(C.int32_t(pid))
}

// swiftGetRunningAppsHash is undocumented. Please add documentation.
func swiftGetRunningAppsHash() string {
	cStr := C.swift_getRunningAppsHash()
	if cStr == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cStr))
	return C.GoString(cStr)
}

// swiftGetRunningAppsJSON is undocumented. Please add documentation.
func swiftGetRunningAppsJSON() string {
	cStr := C.swift_getRunningAppsJSON()
	if cStr == nil {
		return "[]"
	}
	defer C.free(unsafe.Pointer(cStr))
	return C.GoString(cStr)
}
