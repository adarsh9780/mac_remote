build:
	# Create app bundle directories
	mkdir -p MacRemote.app/Contents/MacOS
	mkdir -p MacRemote.app/Contents/Resources
	
	# Copy Info.plist and AppIcon.icns into the bundle
	cp packaging/Info.plist MacRemote.app/Contents/Info.plist
	cp packaging/AppIcon.icns MacRemote.app/Contents/Resources/AppIcon.icns
	
	# Compile Swift code into an object file (no executable)
	swiftc -parse-as-library -O -c system_helper.swift -o go/system_helper.o
	
	# Build the unified Go binary (statically linking the Swift object file)
	cd go && go build -o ../MacRemote.app/Contents/MacOS/MacRemote main.go swift_bridge.go
	
	# Clean up old standalone binaries (if they exist)
	rm -f MacRemote.app/Contents/Resources/system_helper
	rm -f MacRemote.app/Contents/Resources/system_helper.swift
	
	# Ad-Hoc sign the app bundle so macOS tracks accessibility permissions perfectly
	codesign --force --deep --sign - MacRemote.app

test:
	# Run Go tests in the go directory
	cd go && go test -v -race ./...

run: build
	open MacRemote.app

clean:
	rm -f system_helper.o
