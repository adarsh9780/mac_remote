result = `/usr/bin/osascript -e 'button returned of (display dialog "MacRemote is not signed by an Apple Developer ($99/yr). Do you want to trust and install it anyway?" buttons {"No", "Yes"} default button "Yes" cancel button "No")'`.strip
if result != "Yes"
  exit 1
end
puts "User trusted"
