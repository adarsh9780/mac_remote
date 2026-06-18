package main

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/energye/systray"
	"macremote/pair"
)

//go:embed assets/*
var embeddedFiles embed.FS

var (
	lastBrowserTitle string
	browserPlaying   = true

	sseClients = make(map[chan string]bool)
	sseMutex   sync.Mutex

	authSessions = make(map[string]*ConnectedDevice)
	authMutex    sync.Mutex

	mDeviceStatus *systray.MenuItem
	mDisconnect   *systray.MenuItem
)

type ConnectedDevice struct {
	Name      string
	IP        string
	Connected time.Time
}

// main is undocumented. Please add documentation.
func main() {
	systray.Run(onReady, onExit)
}

// onReady is undocumented. Please add documentation.
func onReady() {
	// Load the app icon from embedded assets for the menu bar
	iconBytes, err := embeddedFiles.ReadFile("assets/menuicon.png")
	if err != nil {
		log.Printf("Failed to load menu icon: %v\n", err)
	}
	systray.SetTemplateIcon(iconBytes, iconBytes)
	systray.SetTitle("")
	systray.SetTooltip("Mac Remote Control")

	// Required by energye/systray: explicitly show the menu on click
	systray.SetOnClick(func(menu systray.IMenu) {
		if menu != nil {
			menu.ShowMenu()
		}
	})

	systray.CreateMenu()

	mIP := systray.AddMenuItem("Loading IP...", "Copy URL to clipboard")
	mPair := systray.AddMenuItem("Show QR Code", "Show QR code to connect a device")

	systray.AddSeparator()
	mDeviceStatus = systray.AddMenuItem("No devices connected", "")
	mDeviceStatus.Disable()
	mDisconnect = systray.AddMenuItem("Disconnect Device", "Revoke access for the current device")
	mDisconnect.Hide()

	systray.AddSeparator()
	mReqAccess := systray.AddMenuItem("Grant Accessibility Permission", "Opens Accessibility Settings to toggle permissions for system_helper")
	
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit Application")

	// Get and display the local IP address
	ip := getLocalIP()
	url := fmt.Sprintf("http://%s:5050", ip)
	mIP.SetTitle("Copy Link: " + url)

	go startServer(url)
	
	// Register for MediaRemote notifications so Now Playing info is available
	swiftInitMediaRemote()

	mIP.Click(func() {
		exec.Command("sh", "-c", fmt.Sprintf("printf '%%s' '%s' | pbcopy", url)).Run()
	})

	mPair.Click(func() {
		swiftShowQRUI(url)
	})

	mDisconnect.Click(func() {
		authMutex.Lock()
		authSessions = make(map[string]*ConnectedDevice)
		authMutex.Unlock()
		updateDeviceMenu(nil)
	})

	mQuit.Click(func() {
		systray.Quit()
	})

	mReqAccess.Click(func() {
		// Run a quick background keystroke command to trigger the OS TCC permission registration.
		// This forces macOS to register MacRemote in the Accessibility list automatically.
		go func() {
			swiftPressKey(999)
		}()
		// Open the macOS Accessibility settings pane
		exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Run()
	})
}

// onExit is undocumented. Please add documentation.
func onExit() {
	// Terminate the process to stop the background web server
	os.Exit(0)
}

// getLocalIP is undocumented. Please add documentation.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// startServer is undocumented. Please add documentation.
func startServer(url string) {
	fSys, err := fs.Sub(embeddedFiles, "assets")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(fSys)))
	mux.HandleFunc("/api/action", handleAction)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/mouse", handleMouse)
	mux.HandleFunc("/api/events", handleEvents)
	mux.HandleFunc("/api/apps", handleApps)

	mux.HandleFunc("/api/pair/request", handlePairRequest)
	mux.HandleFunc("/api/pair/verify", handlePairVerify)
	mux.HandleFunc("/pair", handlePairLink)

	go statusMonitor()

	log.Printf("Starting web server. Access it from your phone at: %s\n", url)
	if err := http.ListenAndServe("0.0.0.0:5050", authMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

// issueAuthSession is undocumented. Please add documentation.
func issueAuthSession(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	rand.Read(b)
	token := base64.RawURLEncoding.EncodeToString(b)
	
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	userAgent := r.Header.Get("User-Agent")
	name := "Unknown Device"
	if strings.Contains(userAgent, "iPhone") {
		name = "iPhone"
	} else if strings.Contains(userAgent, "Android") {
		name = "Android"
	} else if strings.Contains(userAgent, "iPad") {
		name = "iPad"
	} else if strings.Contains(userAgent, "Mac OS X") {
		name = "Mac"
	} else if strings.Contains(userAgent, "Windows") {
		name = "Windows PC"
	}

	dev := &ConnectedDevice{
		Name:      name,
		IP:        ip,
		Connected: time.Now(),
	}

	authMutex.Lock()
	authSessions[token] = dev
	authMutex.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "macremote_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   31536000,
	})

	updateDeviceMenu(dev)
	swiftCloseCodeUI()
	swiftShowSuccessUI(fmt.Sprintf("%s (%s)", dev.Name, dev.IP))
}

// updateDeviceMenu is undocumented. Please add documentation.
func updateDeviceMenu(dev *ConnectedDevice) {
	if mDeviceStatus == nil || mDisconnect == nil {
		return
	}
	if dev == nil {
		mDeviceStatus.SetTitle("No devices connected")
		mDisconnect.Hide()
	} else {
		mDeviceStatus.SetTitle(fmt.Sprintf("Connected: %s (%s)", dev.Name, dev.IP))
		mDisconnect.Show()
	}
}

// authMiddleware is undocumented. Please add documentation.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("macremote_session")
		valid := false
		hasActiveSession := false
		
		authMutex.Lock()
		if err == nil {
			_, valid = authSessions[cookie.Value]
		}
		hasActiveSession = len(authSessions) > 0
		authMutex.Unlock()

		// If the user has a valid session, let them through.
		if valid {
			if r.URL.Path == "/pair.html" || r.URL.Path == "/busy.html" {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// User is NOT authenticated.
		// If another session is already active, block pairing!
		if hasActiveSession {
			if r.URL.Path == "/busy.html" || strings.HasSuffix(r.URL.Path, ".css") || strings.HasSuffix(r.URL.Path, ".js") {
				next.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, `{"error": "Busy"}`, http.StatusForbidden)
				return
			}
			http.Redirect(w, r, "/busy.html", http.StatusFound)
			return
		}

		// No active sessions exist. Allow access to pairing pages and static assets.
		if strings.HasPrefix(r.URL.Path, "/api/pair/") || r.URL.Path == "/pair" || r.URL.Path == "/pair.html" || strings.HasSuffix(r.URL.Path, ".css") || strings.HasSuffix(r.URL.Path, ".js") {
			next.ServeHTTP(w, r)
			return
		}

		// Not authenticated and hitting a protected route
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		} else {
			http.Redirect(w, r, "/pair.html", http.StatusFound)
		}
	})
}

// handlePairRequest is undocumented. Please add documentation.
func handlePairRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := pair.Generate(60*time.Second, nil)
	swiftCloseQRUI()
	swiftShowCodeUI(session.Code)
	
	w.WriteHeader(http.StatusOK)
}

// handlePairVerify is undocumented. Please add documentation.
func handlePairVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if pair.VerifyCode(req.Code, nil) {
		issueAuthSession(w, r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	} else {
		http.Error(w, `{"error": "Invalid or expired code"}`, http.StatusUnauthorized)
	}
}

// handlePairLink is undocumented. Please add documentation.
func handlePairLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token != "" && pair.VerifyToken(token, nil) {
		issueAuthSession(w, r)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/pair.html?error=1", http.StatusFound)
}

// handleAction is undocumented. Please add documentation.
func handleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actionType := r.URL.Query().Get("type")
	stepsStr := r.URL.Query().Get("steps")
	steps := 1
	if s, err := strconv.Atoi(stepsStr); err == nil && s > 0 {
		steps = s
	}

	var err error

	switch actionType {
	case "switch_app":
		pid, _ := strconv.Atoi(r.URL.Query().Get("pid"))
		swiftSwitchToApp(pid)
	case "show_dock":
		swiftShowDock()
	case "type_chars":
		text := r.URL.Query().Get("text")
		if text != "" {
			swiftTypeText(text)
		}
	case "type_backspace":
		for i := 0; i < steps; i++ {
			swiftPressKey(51) // kVK_Delete (Backspace)
		}
	case "press_enter":
		swiftPressKey(36) // kVK_Return
	case "volume_up":
		curr := swiftGetVolume()
		swiftSetVolume(curr + 0.06 * float32(steps))
	case "volume_down":
		curr := swiftGetVolume()
		swiftSetVolume(curr - 0.06 * float32(steps))
	case "volume_mute":
		swiftSetMute()
	
	case "brightness_up":
		if curr, cErr := getSystemBrightness(); cErr == nil {
			newVal := curr + 0.0625*float64(steps)
			if newVal > 1.0 {
				newVal = 1.0
			}
			swiftSetBrightness(float32(newVal))
		}
	case "brightness_down":
		if curr, cErr := getSystemBrightness(); cErr == nil {
			newVal := curr - 0.0625*float64(steps)
			if newVal < 0.0 {
				newVal = 0.0
			}
			swiftSetBrightness(float32(newVal))
		}

	case "media_prev":
		swiftSendMediaCommand(5)
	case "media_backward":
		swiftPressKey(123)
	case "media_playpause":
		swiftSendMediaCommand(2)
	case "media_forward":
		swiftPressKey(124)
	case "media_next":
		swiftSendMediaCommand(4)
		
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("Action %s failed: %v\n", actionType, err)
		http.Error(w, "Failed", http.StatusInternalServerError)
		return
	}

	// Broadcast updated status to all SSE clients immediately
	go broadcastStatus()

	sendStatusResponse(w)
}

// handleMouse is undocumented. Please add documentation.
func handleMouse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actionType := r.URL.Query().Get("type")
	
	switch actionType {
	case "move":
		dx, _ := strconv.ParseFloat(r.URL.Query().Get("dx"), 64)
		dy, _ := strconv.ParseFloat(r.URL.Query().Get("dy"), 64)
		swiftMoveMouse(dx, dy)
	case "scroll":
		dx, _ := strconv.Atoi(r.URL.Query().Get("dx"))
		dy, _ := strconv.Atoi(r.URL.Query().Get("dy"))
		swiftScrollMouse(dy, dx)
	case "click":
		button := r.URL.Query().Get("button")
		swiftClickMouse(button == "right")
	case "down":
		button := r.URL.Query().Get("button")
		swiftMouseDown(button == "right")
	case "up":
		button := r.URL.Query().Get("button")
		swiftMouseUp(button == "right")
	default:
		http.Error(w, "Unknown mouse action", http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}


// getSystemMediaInfo is undocumented. Please add documentation.
func getSystemMediaInfo() string {
	return swiftGetMediaInfoJSON()
}

// getSystemVolume is undocumented. Please add documentation.
func getSystemVolume() (int, error) {
	return int(swiftGetVolume() * 100), nil
}

// getSystemBrightness is undocumented. Please add documentation.
func getSystemBrightness() (float64, error) {
	return float64(swiftGetBrightness()), nil
}

// getStatusJSON is undocumented. Please add documentation.
func getStatusJSON() string {
	vol := float64(swiftGetVolume())
	brightness, err := getSystemBrightness()
	if err != nil {
		brightness = 0.5
	}
	mediaInfo := getSystemMediaInfo()
	if mediaInfo == "" {
		mediaInfo = `{"title":"","artist":"","playing":false}`
	}
	appsHash := swiftGetRunningAppsHash()
	textFocused := swiftIsTextInputFocused()
	return fmt.Sprintf(`{"volume": %.3f, "brightness": %.3f, "media": %s, "apps_hash": "%s", "text_focused": %t}`, vol, brightness, mediaInfo, appsHash, textFocused)
}

// handleApps is undocumented. Please add documentation.
func handleApps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", swiftGetRunningAppsJSON())
}

// sendStatusResponse is undocumented. Please add documentation.
func sendStatusResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", getStatusJSON())
}

// handleStatus is undocumented. Please add documentation.
func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sendStatusResponse(w)
}

// handleEvents is undocumented. Please add documentation.
func handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan string, 10)
	
	sseMutex.Lock()
	sseClients[messageChan] = true
	sseMutex.Unlock()

	defer func() {
		sseMutex.Lock()
		delete(sseClients, messageChan)
		sseMutex.Unlock()
		close(messageChan)
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send current status immediately so the client doesn't wait for the next change
	initialStatus := getStatusJSON()
	fmt.Fprintf(w, "data: %s\n\n", initialStatus)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

// broadcastStatus is undocumented. Please add documentation.
func broadcastStatus() {
	status := getStatusJSON()
	sseMutex.Lock()
	for clientChan := range sseClients {
		select {
		case clientChan <- status:
		default:
			// drop message if channel is full to prevent blocking
		}
	}
	sseMutex.Unlock()
}

// statusMonitor is undocumented. Please add documentation.
func statusMonitor() {
	var lastStatus string
	for {
		currentStatus := getStatusJSON()
		if currentStatus != lastStatus {
			lastStatus = currentStatus
			broadcastStatus()
		}
		time.Sleep(500 * time.Millisecond)
	}
}
