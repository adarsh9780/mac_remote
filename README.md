# Mac Remote

Control your Mac from your phone without installing an app.

Mac Remote is a menu bar app for macOS that runs a local web server. Open your phone's browser to your Mac's IP address and you get a control surface: brightness, volume, media playback, a trackpad, a keyboard, and an app switcher. It runs over your Wi-Fi network with nothing to download on the phone.

## Why this exists

Remote control apps have existed for over a decade. Almost all of them require installing a companion app on the phone in addition to the server on the computer. That means two app stores, two installs, and an account or pairing step before you can do anything.

Mac Remote skips the phone side install. The server renders its own UI as a web page. Any phone, tablet, or laptop with a browser on the same network can use it immediately.

## Features

- **Two-Way Synchronization**: If you change the volume, brightness, or media playback directly on your Mac, the web UI instantly updates to reflect the new values.
- **Brightness & volume control**, with the current value shown as a number.  
  <img src="screenshots/brightness-vol.png" width="260" alt="Brightness and Volume controls" />

- **Media controls** (rewind, previous, play/pause, next, fast-forward) plus a scrolling "Now Playing" readout.  
  <img src="screenshots/media-controls.png" width="260" alt="Expanded Media Controls" />

- **Trackpad** for cursor movement, with adjustable pointer sensitivity and scroll speed.  
  <img src="screenshots/trackpad-settings.png" width="260" alt="Trackpad and Sensitivity Settings" />

- **Native keyboard typing**, when the Mac's cursor is in a text field, your phone's keyboard types into it. Includes a "Send Enter" action so Return behaves correctly instead of inserting a literal newline.  
  <img src="screenshots/keyboard.png" width="260" alt="Virtual Keyboard" />

- **Remote app switcher / Dock view**, see what's running and launch or switch to an app from your phone.  
  <img src="screenshots/app-switcher.png" width="260" alt="App Switcher" />

- **Zero install on the client.** Works from Safari, Chrome, or any mobile browser.
- **Minimal footprint.** The app is about 9MB. It uses a Go server for the web UI and logic, with a statically-linked Swift object for the macOS specific calls (brightness, volume, window management). There is no Electron or bundled browser runtime.

## Drawbacks & Security Risks

While Mac Remote is functional, it has several limitations and risks you should be aware of before running it:

- **Unencrypted Transport (HTTP):** Traffic between your phone and the Mac is currently sent over standard HTTP. If someone is monitoring your local network traffic, they could intercept your keystrokes, trackpad movements, or capture your session token.
- **Accessibility Permissions:** The application relies heavily on macOS Accessibility APIs to move the mouse and inject keystrokes. If the application crashes, macOS can sometimes get confused, requiring you to manually toggle the permission off and on again in System Settings.
- **No Wake-on-LAN:** The Mac must be awake to receive inputs. If your Mac goes to sleep, the web interface will disconnect and cannot wake the computer.

## How it works

```
Phone browser  ──HTTP──▶  Go server (menu bar app, :5050)  ──Cgo──▶  Swift object
                                                                    (CoreAudio, brightness,
                                                                     NSWorkspace, CGEvent)
```

The Swift menu bar app starts the Go server and shows status controls (QR code pairing, quit) from the menu bar icon. The Go server serves the control UI and talks directly to the Swift object via Cgo to perform the native macOS actions.

## Comparison

| | **Mac Remote** | Remote Mouse | Unified Remote | Astropad Workbench |
|---|---|---|---|---|
| Install required on phone | **None** | Yes (App Store) | Yes (App Store) | Yes |
| Install required on Mac | Single ~9MB menu bar app | App + license | Server app | App |
| Open source | **Yes** | No | No | No |
| Cost | Free | Free + in-app purchases | Free + paid full version | Subscription ($10/mo or $50/yr) |
| Native brightness/volume control | **Yes, with live readout** | Volume only, via phone hardware buttons | Yes | No (screen-mirroring model) |
| Trackpad | Yes, adjustable sensitivity/scroll | Yes, plus gyro mode | Yes | Mouse input via streamed display |
| Remote app switcher / Dock | **Yes** | No | File manager only | Full screen mirroring instead |
| Typing | Native phone keyboard | Native + voice dictation | Native | Voice/keyboard over a streamed session |

Astropad Workbench solves a different problem (full screen mirroring for remote desktop control) rather than a lightweight native controls panel. It is included since it is a recent entrant in this space.

## Security

Mac Remote is designed to operate on your local area network (LAN). It uses the following measures to restrict access:

- **QR-code & One-time-code pairing**: An on-screen 6-digit one-time code is required before a new device gets control.
- **Brute-force protection**: Automatic lockout after 5 failed OTP attempts.
- **Device Management**: A list of currently connected devices is visible in the Mac menu bar. You can instantly revoke access for any of them.

## Getting Started

### Requirements
- **macOS** 13.0 or later
- **Go** 1.21+
- **Xcode Command Line Tools** (for the Swift compiler)

```bash
# Clone
git clone https://github.com/adarsh9780/mac_remote.git
cd mac_remote

# Build the unified application
make build

# Run the application
open MacRemote.app
```

> **Accessibility Permissions**
> MacRemote requires Accessibility permissions to control the mouse, keyboard, and system UI. Upon running the app for the first time, click "Grant Accessibility Permission" from the menu bar to open System Settings, and ensure MacRemote is toggled ON.

Once running, click the menu bar icon and choose "Show QR Code", scan it with your phone, and enter the connection request code displayed on your Mac screen.

## Roadmap

- [x] One-time-code pairing with expiry
- [x] QR-code pairing
- [x] Per-device session list with revoke
- [x] Textured/dotted trackpad surface
- [ ] Connection limit (cap on simultaneous connected devices)
- [ ] HTTPS / encrypted local transport

## Contributing

Issues and pull requests are welcome.

## License

MIT License
