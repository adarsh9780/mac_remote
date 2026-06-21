# Security Policy

## Scope

Mac Remote is a local-network tool. The server only accepts connections
from devices on the same Wi-Fi/LAN as the Mac it's running on — it is
never exposed to the public internet. Pairing requires a one-time
passcode or QR code shown on the Mac's own screen, and only one device
can be connected at a time.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| older   | :x:                |

Only the latest release is actively supported. Please update before
reporting an issue.

## Reporting a Vulnerability

If you find a security issue, please **do not open a public GitHub
issue**. Instead, email me directly at adarshmaurya7@gmail.com or use GitHub's
private vulnerability reporting (Security tab → "Report a
vulnerability") so it can be addressed before details are public.

I'll do my best to respond within a few days. This is a side project
maintained by one person, so please bear with me on turnaround time.

## Known Limitations

- Traffic between the phone and Mac is plain HTTP, not encrypted —
  appropriate for a trusted home LAN, not for use on a network you
  don't control.
- The app currently ships unsigned (no Apple Developer certificate
  yet). You can build from source if you'd rather not run the
  pre-built binary.
