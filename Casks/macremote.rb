cask "macremote" do
  version "0.1.0"
  sha256 "0019dfc4b32d63c1392aa264aed2253c1e0c2fb09216f8e2cc269bbfb8bb49b5"

  url "https://github.com/adarsh9780/mac_remote/releases/download/v#{version}/MacRemote.dmg"
  name "MacRemote"
  desc "Control your Mac from your phone without installing an app"
  homepage "https://github.com/adarsh9780/mac_remote"

  app "MacRemote.app"

  preflight do
    # Prompt the user for explicit trust since we are not paying the $99/yr Apple Developer fee
    prompt_script = <<~EOS
      button returned of (display dialog "MacRemote is not signed by an Apple Developer to avoid the $99/yr fee. Do you want to trust and install it anyway?" buttons {"No", "Yes"} default button "Yes" cancel button "No")
    EOS
    
    result = system_command "/usr/bin/osascript", args: ["-e", prompt_script], print_stdout: true
    
    if result.exit_status != 0 || result.stdout.strip != "Yes"
      # If they click "No", the osascript throws an error (User canceled) or returns No
      odie "Installation aborted by the user. You chose not to trust MacRemote."
    end
  end

  postflight do
    # Clear the quarantine attribute so the app can run without Gatekeeper blocking it
    system_command "/usr/bin/xattr", args: ["-cr", "#{appdir}/MacRemote.app"]
  end

  zap trash: [
    "~/.gemini/antigravity" # Generic placeholder, normally would be ~/Library/Preferences/com.macremote.plist etc
  ]
end
