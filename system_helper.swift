import Foundation
import CoreGraphics
import AppKit

// Dynamic loading of DisplayServices (Brightness)
let displayHandle = dlopen("/System/Library/PrivateFrameworks/DisplayServices.framework/DisplayServices", RTLD_NOW)
// Dynamic loading of MediaRemote (Media Controls)
let mediaHandle = dlopen("/System/Library/PrivateFrameworks/MediaRemote.framework/MediaRemote", RTLD_NOW)

typealias GetBrightnessType = @convention(c) (CGDirectDisplayID, UnsafeMutablePointer<Float>) -> Int32
typealias SetBrightnessType = @convention(c) (CGDirectDisplayID, Float) -> Int32
typealias GetNowPlayingInfoFunc = @convention(c) (DispatchQueue, @escaping (CFDictionary?) -> Void) -> Void
typealias SendCommandFunc = @convention(c) (Int32, AnyObject?) -> Bool
typealias RegisterForNotificationsFunc = @convention(c) (DispatchQueue) -> Void

@_cdecl("swift_getBrightness")
public func swift_getBrightness() -> Float {
    guard let handle = displayHandle,
          let sym = dlsym(handle, "DisplayServicesGetBrightness") else { return 0.5 }
    let GetBrightness = unsafeBitCast(sym, to: GetBrightnessType.self)
    var val: Float = 0.0
    if GetBrightness(CGMainDisplayID(), &val) == 0 {
        return val
    }
    return 0.5
}

@_cdecl("swift_setBrightness")
public func swift_setBrightness(_ val: Float) {
    guard let handle = displayHandle,
          let sym = dlsym(handle, "DisplayServicesSetBrightness") else { return }
    let SetBrightness = unsafeBitCast(sym, to: SetBrightnessType.self)
    _ = SetBrightness(CGMainDisplayID(), val)
}

@_cdecl("swift_initMediaRemote")
public func swift_initMediaRemote() {
    guard let handle = mediaHandle,
          let sym = dlsym(handle, "MRMediaRemoteRegisterForNowPlayingNotifications") else {
        return
    }
    let Register = unsafeBitCast(sym, to: RegisterForNotificationsFunc.self)
    Register(DispatchQueue.main)
}

// Helper: run an AppleScript via Process (subprocess)
// This avoids the main-thread requirement of NSAppleScript
// which deadlocks when systray's NSApplication.run() holds the main thread
private func runScript(_ source: String) -> String? {
    let proc = Process()
    proc.executableURL = URL(fileURLWithPath: "/usr/bin/osascript")
    proc.arguments = ["-e", source]
    
    let pipe = Pipe()
    proc.standardOutput = pipe
    proc.standardError = Pipe() // discard errors
    
    do {
        try proc.run()
        proc.waitUntilExit()
    } catch {
        return nil
    }
    
    let data = pipe.fileHandleForReading.readDataToEndOfFile()
    let str = String(data: data, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
    return str.isEmpty ? nil : str
}

// Helper: build a JSON string from media info
private func mediaJSON(title: String, artist: String, playing: Bool?) -> String {
    var jsonDict: [String: Any] = ["title": title, "artist": artist]
    if let p = playing {
        jsonDict["playing"] = p
    }
    if let data = try? JSONSerialization.data(withJSONObject: jsonDict, options: []),
       let jsonStr = String(data: data, encoding: .utf8) {
        return jsonStr
    }
    return "{}"
}

@_cdecl("swift_getMediaInfoJSON")
public func swift_getMediaInfoJSON() -> UnsafeMutablePointer<CChar>? {
    var mediaPlaying = false
    var nativeTitle = ""
    var nativeArtist = ""
    
    // --- Step 1: Query MediaRemote for play/pause state + native app info ---
    // isPlaying works for browser media too, even when title is unavailable
    if let handle = mediaHandle,
       let symInfo = dlsym(handle, "MRMediaRemoteGetNowPlayingInfo"),
       let symPlay = dlsym(handle, "MRMediaRemoteGetNowPlayingApplicationIsPlaying") {
        
        typealias GetIsPlayingFunc = @convention(c) (DispatchQueue, @escaping (Bool) -> Void) -> Void
        let GetNowPlayingInfo = unsafeBitCast(symInfo, to: GetNowPlayingInfoFunc.self)
        let GetIsPlaying = unsafeBitCast(symPlay, to: GetIsPlayingFunc.self)
        
        let sem = DispatchSemaphore(value: 0)
        
        GetIsPlaying(DispatchQueue.global()) { isPlaying in
            mediaPlaying = isPlaying
            GetNowPlayingInfo(DispatchQueue.global()) { info in
                if let dict = info as? [String: Any] {
                    nativeTitle = dict["kMRMediaRemoteNowPlayingInfoTitle"] as? String ?? ""
                    nativeArtist = dict["kMRMediaRemoteNowPlayingInfoArtist"] as? String ?? ""
                }
                sem.signal()
            }
        }
        
        _ = sem.wait(timeout: .now() + 1.0)
        
        // If native app has title info, return immediately
        if !nativeTitle.isEmpty {
            return strdup(mediaJSON(title: nativeTitle, artist: nativeArtist, playing: mediaPlaying))
        }
    }
    
    // --- Step 2: Fall back to browser tab titles ---
    // We cannot reliably get play/pause state from browsers without JS injection,
    // so we pass `nil` for playing. The frontend will rely on local toggling.
    let browsers: [(name: String, script: String)] = [
        ("Brave Browser", """
            tell application "System Events"
                if not (exists (processes whose name is "Brave Browser")) then return ""
            end tell
            tell application "Brave Browser"
                if (count of windows) > 0 then return title of active tab of front window
            end tell
            return ""
        """),
        ("Google Chrome", """
            tell application "System Events"
                if not (exists (processes whose name is "Google Chrome")) then return ""
            end tell
            tell application "Google Chrome"
                if (count of windows) > 0 then return title of active tab of front window
            end tell
            return ""
        """),
        ("Safari", """
            tell application "System Events"
                if not (exists (processes whose name is "Safari")) then return ""
            end tell
            tell application "Safari"
                if (count of windows) > 0 then return name of current tab of front window
            end tell
            return ""
        """),
    ]
    
    for browser in browsers {
        if let title = runScript(browser.script) {
            return strdup(mediaJSON(title: title, artist: browser.name, playing: nil))
        }
    }
    
    return strdup("{}")
}

@_cdecl("swift_sendMediaCommand")
public func swift_sendMediaCommand(_ cmd: Int32) -> Bool {
    guard let handle = mediaHandle,
          let sym = dlsym(handle, "MRMediaRemoteSendCommand") else { return false }
    let SendCommand = unsafeBitCast(sym, to: SendCommandFunc.self)
    return SendCommand(cmd, nil)
}

@_cdecl("swift_moveMouse")
public func swift_moveMouse(_ dx: Double, _ dy: Double) {
    let currentEvent = CGEvent(source: nil)
    guard let currentLoc = currentEvent?.location else { return }
    let newLoc = CGPoint(x: currentLoc.x + dx, y: currentLoc.y + dy)
    let moveEvent = CGEvent(mouseEventSource: nil, mouseType: .mouseMoved, mouseCursorPosition: newLoc, mouseButton: .left)
    moveEvent?.setIntegerValueField(.mouseEventDeltaX, value: Int64(dx))
    moveEvent?.setIntegerValueField(.mouseEventDeltaY, value: Int64(dy))
    moveEvent?.post(tap: .cghidEventTap)
}

@_cdecl("swift_scrollMouse")
public func swift_scrollMouse(_ dy: Int32, _ dx: Int32) {
    let scrollEvent = CGEvent(scrollWheelEvent2Source: nil, units: .pixel, wheelCount: 2, wheel1: dy, wheel2: dx, wheel3: 0)
    scrollEvent?.post(tap: .cghidEventTap)
}

@_cdecl("swift_clickMouse")
public func swift_clickMouse(_ right: Bool) {
    let currentEvent = CGEvent(source: nil)
    let currentLoc = currentEvent?.location ?? CGPoint.zero
    let button: CGMouseButton = right ? .right : .left
    let typeDown: CGEventType = right ? .rightMouseDown : .leftMouseDown
    let typeUp: CGEventType = right ? .rightMouseUp : .leftMouseUp
    
    let clickDown = CGEvent(mouseEventSource: nil, mouseType: typeDown, mouseCursorPosition: currentLoc, mouseButton: button)
    let clickUp = CGEvent(mouseEventSource: nil, mouseType: typeUp, mouseCursorPosition: currentLoc, mouseButton: button)
    
    clickDown?.post(tap: .cghidEventTap)
    clickUp?.post(tap: .cghidEventTap)
}

@_cdecl("swift_mouseDown")
public func swift_mouseDown(_ right: Bool) {
    let currentEvent = CGEvent(source: nil)
    let currentLoc = currentEvent?.location ?? CGPoint.zero
    let button: CGMouseButton = right ? .right : .left
    let typeDown: CGEventType = right ? .rightMouseDown : .leftMouseDown
    
    // Using mouseDragged to ensure the OS registers it as held if moved
    let clickDown = CGEvent(mouseEventSource: nil, mouseType: typeDown, mouseCursorPosition: currentLoc, mouseButton: button)
    clickDown?.post(tap: .cghidEventTap)
}

@_cdecl("swift_mouseUp")
public func swift_mouseUp(_ right: Bool) {
    let currentEvent = CGEvent(source: nil)
    let currentLoc = currentEvent?.location ?? CGPoint.zero
    let button: CGMouseButton = right ? .right : .left
    let typeUp: CGEventType = right ? .rightMouseUp : .leftMouseUp
    
    let clickUp = CGEvent(mouseEventSource: nil, mouseType: typeUp, mouseCursorPosition: currentLoc, mouseButton: button)
    clickUp?.post(tap: .cghidEventTap)
}

@_cdecl("swift_typeText")
public func swift_typeText(_ cText: UnsafePointer<CChar>) {
    let text = String(cString: cText)
    let str = text as NSString
    var chars = [UniChar](repeating: 0, count: str.length)
    str.getCharacters(&chars, range: NSRange(location: 0, length: str.length))

    let eventDown = CGEvent(keyboardEventSource: nil, virtualKey: 0, keyDown: true)
    eventDown?.keyboardSetUnicodeString(stringLength: str.length, unicodeString: &chars)
    eventDown?.post(tap: .cghidEventTap)

    let eventUp = CGEvent(keyboardEventSource: nil, virtualKey: 0, keyDown: false)
    eventUp?.keyboardSetUnicodeString(stringLength: str.length, unicodeString: &chars)
    eventUp?.post(tap: .cghidEventTap)
}

@_cdecl("swift_pressKey")
public func swift_pressKey(_ keyCode: UInt16) {
    let keyDown = CGEvent(keyboardEventSource: nil, virtualKey: CGKeyCode(keyCode), keyDown: true)
    let keyUp = CGEvent(keyboardEventSource: nil, virtualKey: CGKeyCode(keyCode), keyDown: false)
    keyDown?.post(tap: .cghidEventTap)
    keyUp?.post(tap: .cghidEventTap)
}

import CoreAudio

@_cdecl("swift_getVolume")
public func swift_getVolume() -> Float32 {
    var defaultOutputDeviceID = AudioDeviceID(0)
    var defaultOutputDeviceIDSize = UInt32(MemoryLayout.size(ofValue: defaultOutputDeviceID))
    var getDefaultOutputDevicePropertyAddress = AudioObjectPropertyAddress(
        mSelector: kAudioHardwarePropertyDefaultOutputDevice,
        mScope: kAudioObjectPropertyScopeGlobal,
        mElement: AudioObjectPropertyElement(kAudioObjectPropertyElementMain))
    
    AudioObjectGetPropertyData(
        AudioObjectID(kAudioObjectSystemObject),
        &getDefaultOutputDevicePropertyAddress,
        0,
        nil,
        &defaultOutputDeviceIDSize,
        &defaultOutputDeviceID)
        
    var volume: Float32 = 0.0
    var volumeSize = UInt32(MemoryLayout.size(ofValue: volume))
    var volumePropertyAddress = AudioObjectPropertyAddress(
        mSelector: kAudioDevicePropertyVolumeScalar,
        mScope: kAudioDevicePropertyScopeOutput,
        mElement: kAudioObjectPropertyElementMain)
        
    let status = AudioObjectGetPropertyData(
        defaultOutputDeviceID,
        &volumePropertyAddress,
        0,
        nil,
        &volumeSize,
        &volume)
    
    if status != 0 {
        volumePropertyAddress.mElement = kAudioObjectPropertyElementMain
        AudioObjectGetPropertyData(
            defaultOutputDeviceID,
            &volumePropertyAddress,
            0,
            nil,
            &volumeSize,
            &volume)
    }
        
    return volume
}

@_cdecl("swift_setVolume")
public func swift_setVolume(_ vol: Float32) {
    var defaultOutputDeviceID = AudioDeviceID(0)
    var defaultOutputDeviceIDSize = UInt32(MemoryLayout.size(ofValue: defaultOutputDeviceID))
    var getDefaultOutputDevicePropertyAddress = AudioObjectPropertyAddress(
        mSelector: kAudioHardwarePropertyDefaultOutputDevice,
        mScope: kAudioObjectPropertyScopeGlobal,
        mElement: AudioObjectPropertyElement(kAudioObjectPropertyElementMain))
    
    AudioObjectGetPropertyData(
        AudioObjectID(kAudioObjectSystemObject),
        &getDefaultOutputDevicePropertyAddress,
        0,
        nil,
        &defaultOutputDeviceIDSize,
        &defaultOutputDeviceID)
        
    var volume: Float32 = vol
    let volumeSize = UInt32(MemoryLayout.size(ofValue: volume))
    var volumePropertyAddress = AudioObjectPropertyAddress(
        mSelector: kAudioDevicePropertyVolumeScalar,
        mScope: kAudioDevicePropertyScopeOutput,
        mElement: kAudioObjectPropertyElementMain)
        
    let status = AudioObjectSetPropertyData(
        defaultOutputDeviceID,
        &volumePropertyAddress,
        0,
        nil,
        volumeSize,
        &volume)
        
    if status != 0 {
        volumePropertyAddress.mElement = kAudioObjectPropertyElementMain
        AudioObjectSetPropertyData(
            defaultOutputDeviceID,
            &volumePropertyAddress,
            0,
            nil,
            volumeSize,
            &volume)
    }
}

@_cdecl("swift_setMute")
public func swift_setMute() {
    var defaultOutputDeviceID = AudioDeviceID(0)
    var defaultOutputDeviceIDSize = UInt32(MemoryLayout.size(ofValue: defaultOutputDeviceID))
    var getDefaultOutputDevicePropertyAddress = AudioObjectPropertyAddress(
        mSelector: kAudioHardwarePropertyDefaultOutputDevice,
        mScope: kAudioObjectPropertyScopeGlobal,
        mElement: AudioObjectPropertyElement(kAudioObjectPropertyElementMain))
    
    AudioObjectGetPropertyData(
        AudioObjectID(kAudioObjectSystemObject),
        &getDefaultOutputDevicePropertyAddress,
        0,
        nil,
        &defaultOutputDeviceIDSize,
        &defaultOutputDeviceID)
        
    var mute: UInt32 = 0
    var muteSize = UInt32(MemoryLayout.size(ofValue: mute))
    var mutePropertyAddress = AudioObjectPropertyAddress(
        mSelector: kAudioDevicePropertyMute,
        mScope: kAudioDevicePropertyScopeOutput,
        mElement: kAudioObjectPropertyElementMain)
        
    AudioObjectGetPropertyData(
        defaultOutputDeviceID,
        &mutePropertyAddress,
        0,
        nil,
        &muteSize,
        &mute)
        
    mute = (mute == 0) ? 1 : 0
        
    AudioObjectSetPropertyData(
        defaultOutputDeviceID,
        &mutePropertyAddress,
        0,
        nil,
        muteSize,
        &mute)
}

// --- App Switcher & Dock Access ---

@_cdecl("swift_showDock")
public func swift_showDock() {
    let script = "tell application \"System Events\" to key code 99 using control down"
    var error: NSDictionary?
    if let appleScript = NSAppleScript(source: script) {
        appleScript.executeAndReturnError(&error)
    }
}

@_cdecl("swift_isTextInputFocused")
public func swift_isTextInputFocused() -> Bool {
    let systemWideElement = AXUIElementCreateSystemWide()
    var focusedElement: CFTypeRef?
    
    let result = AXUIElementCopyAttributeValue(systemWideElement, kAXFocusedUIElementAttribute as CFString, &focusedElement)
    if result == .success, let element = focusedElement {
        let axElement = element as! AXUIElement
        var role: CFTypeRef?
        if AXUIElementCopyAttributeValue(axElement, kAXRoleAttribute as CFString, &role) == .success, let roleStr = role as? String {
            return roleStr == "AXTextField" || roleStr == "AXTextArea" || roleStr == "AXComboBox" || roleStr == "AXWebArea"
        }
    }
    return false
}

@_cdecl("swift_switchToApp")
public func swift_switchToApp(_ pid: Int32) {
    if let app = NSRunningApplication(processIdentifier: pid) {
        app.activate(options: .activateIgnoringOtherApps)
    }
}

@_cdecl("swift_getRunningAppsHash")
public func swift_getRunningAppsHash() -> UnsafeMutablePointer<CChar>? {
    let apps = NSWorkspace.shared.runningApplications
    let pids = apps.filter { $0.activationPolicy == .regular }.map { String($0.processIdentifier) }.joined(separator: ",")
    return strdup(pids)
}

@_cdecl("swift_getRunningAppsJSON")
public func swift_getRunningAppsJSON() -> UnsafeMutablePointer<CChar>? {
    let apps = NSWorkspace.shared.runningApplications
    let regularApps = apps.filter { $0.activationPolicy == .regular }
    
    var jsonArray = [String]()
    for app in regularApps {
        let pid = app.processIdentifier
        let name = app.localizedName ?? "Unknown"
        var base64Icon = ""
        
        if let image = app.icon {
            let size = NSSize(width: 64, height: 64)
            let resizedImage = NSImage(size: size)
            resizedImage.lockFocus()
            image.draw(in: NSRect(origin: .zero, size: size),
                       from: NSRect(origin: .zero, size: image.size),
                       operation: .copy,
                       fraction: 1.0)
            resizedImage.unlockFocus()
            
            if let tiffData = resizedImage.tiffRepresentation,
               let bitmap = NSBitmapImageRep(data: tiffData),
               let pngData = bitmap.representation(using: .png, properties: [:]) {
                base64Icon = pngData.base64EncodedString()
            }
        }
        
        let jsonStr = "{\"pid\":\(app.processIdentifier),\"name\":\"\(app.localizedName ?? "Unknown")\",\"icon\":\"\(base64Icon)\"}"
        jsonArray.append(jsonStr)
    }
    
    let result = "[" + jsonArray.joined(separator: ",") + "]"
    return strdup(result)
}

// --- Pairing UI ---

class QRWindowController: NSObject, NSWindowDelegate {
    static let shared = QRWindowController()
    var window: NSWindow?
    
    func show(url: String) {
        if window == nil {
            let win = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 300, height: 320),
                               styleMask: [.titled, .closable],
                               backing: .buffered, defer: false)
            win.title = "MacRemote QR Code"
            win.center()
            win.isReleasedWhenClosed = false
            win.delegate = self
            
            let view = NSView(frame: win.contentView!.bounds)
            win.contentView = view
            
            let label = NSTextField(labelWithString: "Scan to Connect")
            label.font = NSFont.systemFont(ofSize: 20, weight: .semibold)
            label.alignment = .center
            label.frame = NSRect(x: 20, y: 260, width: 260, height: 30)
            view.addSubview(label)
            
            let imgView = NSImageView(frame: NSRect(x: 40, y: 30, width: 220, height: 220))
            if let qrImage = generateQR(text: url) {
                imgView.image = qrImage
            }
            view.addSubview(imgView)
            
            self.window = win
        } else {
            if let view = window?.contentView {
                if let imgView = view.subviews.first(where: { $0 is NSImageView }) as? NSImageView {
                    imgView.image = generateQR(text: url)
                }
            }
        }
        
        window?.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }
    
    func generateQR(text: String) -> NSImage? {
        let data = text.data(using: String.Encoding.ascii)
        if let filter = CIFilter(name: "CIQRCodeGenerator") {
            filter.setValue(data, forKey: "inputMessage")
            filter.setValue("M", forKey: "inputCorrectionLevel")
            if let ciImage = filter.outputImage {
                let transform = CGAffineTransform(scaleX: 10, y: 10)
                let scaledCIImage = ciImage.transformed(by: transform)
                let rep = NSCIImageRep(ciImage: scaledCIImage)
                let nsImage = NSImage(size: rep.size)
                nsImage.addRepresentation(rep)
                return nsImage
            }
        }
        return nil
    }
    
    func close() {
        window?.close()
        window = nil
    }
    
    func windowWillClose(_ notification: Notification) {
        window = nil
    }
}

class CodeWindowController: NSObject, NSWindowDelegate {
    static let shared = CodeWindowController()
    var window: NSWindow?
    var timer: Timer?
    var countdownLabel: NSTextField!
    var remainingSeconds = 60
    
    func show(code: String) {
        if window == nil {
            let win = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 300, height: 200),
                               styleMask: [.titled, .closable],
                               backing: .buffered, defer: false)
            win.title = "Connection Request"
            win.center()
            win.isReleasedWhenClosed = false
            win.delegate = self
            
            let view = NSView(frame: win.contentView!.bounds)
            win.contentView = view
            
            let titleLabel = NSTextField(labelWithString: "Enter code on your phone:")
            titleLabel.font = NSFont.systemFont(ofSize: 14)
            titleLabel.alignment = .center
            titleLabel.frame = NSRect(x: 20, y: 140, width: 260, height: 30)
            view.addSubview(titleLabel)
            
            let codeLabel = NSTextField(labelWithString: code)
            codeLabel.font = NSFont.monospacedDigitSystemFont(ofSize: 48, weight: .bold)
            codeLabel.alignment = .center
            codeLabel.frame = NSRect(x: 20, y: 60, width: 260, height: 70)
            view.addSubview(codeLabel)
            
            countdownLabel = NSTextField(labelWithString: "Expires in 60s")
            countdownLabel.font = NSFont.systemFont(ofSize: 14)
            countdownLabel.textColor = .systemRed
            countdownLabel.alignment = .center
            countdownLabel.frame = NSRect(x: 20, y: 20, width: 260, height: 30)
            view.addSubview(countdownLabel)
            
            self.window = win
        } else {
            // Update existing window
            if let view = window?.contentView {
                if let codeLabel = view.subviews.first(where: { $0 is NSTextField && ($0 as! NSTextField).font?.pointSize == 48 }) as? NSTextField {
                    codeLabel.stringValue = code
                }
            }
        }
        
        remainingSeconds = 60
        countdownLabel.stringValue = "Expires in 60s"
        countdownLabel.textColor = .systemRed
        
        timer?.invalidate()
        timer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] _ in
            guard let self = self else { return }
            self.remainingSeconds -= 1
            if self.remainingSeconds > 0 {
                self.countdownLabel.stringValue = "Expires in \(self.remainingSeconds)s"
            } else {
                self.countdownLabel.stringValue = "Expired"
                self.countdownLabel.textColor = .gray
                self.timer?.invalidate()
            }
        }
        
        window?.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }
    
    func close() {
        timer?.invalidate()
        timer = nil
        window?.close()
        window = nil
    }
    
    func windowWillClose(_ notification: Notification) {
        timer?.invalidate()
        timer = nil
        window = nil
    }
}

class SuccessWindowController: NSObject, NSWindowDelegate {
    static let shared = SuccessWindowController()
    var window: NSWindow?
    
    func show(clientInfo: String) {
        if window == nil {
            let win = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 350, height: 80),
                               styleMask: [.borderless],
                               backing: .buffered, defer: false)
            win.center()
            win.isReleasedWhenClosed = false
            win.delegate = self
            win.backgroundColor = NSColor.windowBackgroundColor.withAlphaComponent(0.9)
            win.isOpaque = false
            win.hasShadow = true
            
            let view = NSView(frame: win.contentView!.bounds)
            win.contentView = view
            
            let label = NSTextField(labelWithString: "Successfully Connected:\n\(clientInfo)")
            label.font = NSFont.systemFont(ofSize: 14, weight: .medium)
            label.alignment = .center
            label.textColor = .systemGreen
            label.frame = NSRect(x: 20, y: 15, width: 310, height: 50)
            view.addSubview(label)
            
            self.window = win
        } else {
            if let view = window?.contentView,
               let label = view.subviews.first(where: { $0 is NSTextField }) as? NSTextField {
                label.stringValue = "Successfully Connected:\n\(clientInfo)"
            }
        }
        
        window?.makeKeyAndOrderFront(nil)
        
        // Auto fade out
        DispatchQueue.main.asyncAfter(deadline: .now() + 3.0) { [weak self] in
            NSAnimationContext.runAnimationGroup({ context in
                context.duration = 0.5
                self?.window?.animator().alphaValue = 0
            }, completionHandler: {
                self?.window?.close()
                self?.window = nil
            })
        }
    }
}

@_cdecl("swift_showQRUI")
public func swift_showQRUI(_ qrURLStr: UnsafePointer<CChar>) {
    let qrURL = String(cString: qrURLStr)
    DispatchQueue.main.async {
        QRWindowController.shared.show(url: qrURL)
    }
}

@_cdecl("swift_closeQRUI")
public func swift_closeQRUI() {
    DispatchQueue.main.async {
        QRWindowController.shared.close()
    }
}

@_cdecl("swift_showCodeUI")
public func swift_showCodeUI(_ codeStr: UnsafePointer<CChar>) {
    let code = String(cString: codeStr)
    DispatchQueue.main.async {
        CodeWindowController.shared.show(code: code)
    }
}

@_cdecl("swift_closeCodeUI")
public func swift_closeCodeUI() {
    DispatchQueue.main.async {
        CodeWindowController.shared.close()
    }
}

@_cdecl("swift_showSuccessUI")
public func swift_showSuccessUI(_ clientInfoStr: UnsafePointer<CChar>) {
    let info = String(cString: clientInfoStr)
    DispatchQueue.main.async {
        SuccessWindowController.shared.show(clientInfo: info)
    }
}
