import { useState } from 'react';
import './index.css';

function Home() {
  return (
    <div>
      <section className="hero">
        <div className="container">
          <h1>Mac Remote</h1>
          <p>Control your Mac from your phone without installing an app.</p>
          <div className="actions">
            <div className="code-block">
              <code>curl -sL https://raw.githubusercontent.com/adarsh9780/mac_remote/main/install.sh | bash</code>
            </div>
            <a href="https://github.com/adarsh9780/mac_remote" className="btn btn-secondary">View on GitHub</a>
          </div>
          
          <div className="hero-gallery">
            <img src="screenshots/media-controls.png" alt="Media Controls" />
            <img src="screenshots/app-switcher.png" alt="App Switcher" />
            <img src="screenshots/keyboard.png" alt="Keyboard" />
          </div>
        </div>
      </section>

      <section className="section glass-panel" style={{ margin: '40px auto', maxWidth: '1000px', borderRadius: '0' }}>
        <div className="container">
          <h2 className="section-title">Features</h2>
          <div className="feature-grid">
            <div className="feature-card">
              <img src="screenshots/brightness-vol.png" alt="Brightness and Volume controls" />
              <h3>Brightness & Volume</h3>
              <p>Control brightness and volume with a live readout of the current value.</p>
            </div>
            <div className="feature-card">
              <img src="screenshots/media-controls.png" alt="Media Controls" />
              <h3>Media Controls</h3>
              <p>Rewind, previous, play/pause, next, fast-forward, plus a scrolling "Now Playing" readout.</p>
            </div>
            <div className="feature-card">
              <img src="screenshots/trackpad-settings.png" alt="Trackpad Settings" />
              <h3>Trackpad</h3>
              <p>Move the cursor with adjustable pointer sensitivity and scroll speed.</p>
            </div>
            <div className="feature-card">
              <img src="screenshots/keyboard.png" alt="Keyboard" />
              <h3>Native Keyboard</h3>
              <p>Type into text fields using your phone's native keyboard. Includes a "Send Enter" action.</p>
            </div>
            <div className="feature-card">
              <img src="screenshots/app-switcher.png" alt="App Switcher" />
              <h3>Remote App Switcher</h3>
              <p>See what's running and launch or switch to an app directly from your phone.</p>
            </div>
            <div className="feature-card" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
              <h3>Zero Install on Client</h3>
              <p>Works from Safari, Chrome, or any mobile browser.</p>
              <h3 style={{ marginTop: '20px' }}>Two-Way Synchronization</h3>
              <p>If you change volume or media directly on your Mac, the web UI instantly updates.</p>
            </div>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container text-content glass-panel">
          <h2>How it works</h2>
          <p>
            The Swift menu bar app starts the Go server and shows status controls (QR code pairing, quit) from the menu bar icon. The Go server serves the control UI and talks directly to the Swift object via Cgo to perform the native macOS actions.
          </p>
          <div className="code-block" style={{ margin: '20px 0', overflowX: 'auto', whiteSpace: 'pre' }}>
{`Phone browser  ──HTTP──▶  Go server (menu bar app, :5050)  ──Cgo──▶  Swift object
                                                            (CoreAudio, brightness,
                                                             NSWorkspace, CGEvent)`}
          </div>
          
          <h3>Security & Pairing</h3>
          <p>
            Mac Remote is designed to operate on your local area network (LAN). It uses the following measures to restrict access:
          </p>
          <ul>
            <li><strong>QR-code & One-time-code pairing:</strong> An on-screen 6-digit one-time code is required before a new device gets control.</li>
            <li><strong>Brute-force protection:</strong> Automatic lockout after 5 failed OTP attempts.</li>
            <li><strong>Device Management:</strong> MacRemote enforces exactly one connected device at a time. You can instantly disconnect the active user from the Mac menu bar to securely allow a new pairing.</li>
          </ul>
          
          <h3>System Requirements</h3>
          <ul>
            <li><strong>macOS</strong> 13.0 or later</li>
            <li><strong>Go</strong> 1.21+ (only if building from source)</li>
            <li><strong>Xcode Command Line Tools</strong> (only if building from source)</li>
          </ul>
        </div>
      </section>
    </div>
  );
}

function PrivacyPolicy() {
  return (
    <div className="container text-content" style={{ padding: '60px 20px' }}>
      <h1>Privacy Policy</h1>
      
      <h2>What this app does</h2>
      <p>Mac Remote is a local network tool that allows you to control your own Mac from your own phone.</p>

      <h2>Data collection</h2>
      <p>This application collects absolutely no personal data. It contains no analytics, no telemetry, and no crash reporting that phones home. Mac Remote makes zero network calls outside of your local area network (LAN). Everything stays on your devices.</p>

      <h2>Network access</h2>
      <p>The local server runs entirely on your Mac, listening on your local IP address. It is only reachable by devices connected to the exact same Wi-Fi or LAN network. There is no cloud component and no third party ever sees your data or interactions.</p>

      <h2>Third parties</h2>
      <p>None. We do not use any third-party SDKs, ad networks, or external services.</p>

      <h2>Open source note</h2>
      <p>Mac Remote is entirely open source. You do not have to trust this policy blindly. You can <a href="https://github.com/adarsh9780/mac_remote">read the source code yourself on GitHub</a> to verify our privacy commitments.</p>

      <h2>Changes to this policy</h2>
      <p>Any changes will be reflected in this document and noted in the project's changelog on GitHub.</p>

      <h2>Contact</h2>
      <p>For questions or reports, please email <a href="mailto:adarshmaurya7@gmail.com">adarshmaurya7@gmail.com</a> or open an issue on our GitHub repository.</p>
    </div>
  );
}

function TermsOfService() {
  return (
    <div className="container text-content" style={{ padding: '60px 20px' }}>
      <h1>Terms of Service</h1>

      <h2>License</h2>
      <p>Mac Remote is released under the <a href="https://github.com/adarsh9780/mac_remote/blob/main/LICENSE">MIT License</a>.</p>

      <h2>No warranty</h2>
      <p>This software is provided "as is", without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no event shall the authors or copyright holders be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the software or the use or other dealings in the software. You use it entirely at your own risk.</p>

      <h2>Acceptable use</h2>
      <p>This application is intended strictly for controlling a Mac that you own or have explicit permission to control. It is not intended for, and must not be used for, unauthorized access to someone else's machine.</p>

      <h2>Unsigned binary disclaimer</h2>
      <p>Because Mac Remote is a free and open-source project, the pre-built binary does not currently possess an Apple Developer certificate signature. You can bypass the Gatekeeper warning upon installation, or alternatively, you are encouraged to verify the code and <a href="https://github.com/adarsh9780/mac_remote#method-2-build-from-source">build from source</a> if you prefer not to trust the pre-built binary.</p>

      <h2>Limitation of liability</h2>
      <p>We shall not be held liable for any damages, data loss, or other issues arising directly or indirectly from the use or inability to use this application.</p>

      <h2>Contact</h2>
      <p>If you have any questions, please reach out to <a href="mailto:adarshmaurya7@gmail.com">adarshmaurya7@gmail.com</a>.</p>
    </div>
  );
}

function App() {
  const [currentPage, setCurrentPage] = useState('home');

  return (
    <>
      <nav>
        <div className="container nav-container">
          <div className="nav-brand" onClick={() => setCurrentPage('home')}>
            Mac Remote
          </div>
          <div className="nav-links">
            <button 
              className={currentPage === 'home' ? 'active' : ''} 
              onClick={() => setCurrentPage('home')}
            >
              Home
            </button>
            <button 
              className={currentPage === 'privacy' ? 'active' : ''} 
              onClick={() => setCurrentPage('privacy')}
            >
              Privacy Policy
            </button>
            <button 
              className={currentPage === 'tos' ? 'active' : ''} 
              onClick={() => setCurrentPage('tos')}
            >
              Terms of Service
            </button>
          </div>
        </div>
      </nav>

      <main>
        {currentPage === 'home' && <Home />}
        {currentPage === 'privacy' && <PrivacyPolicy />}
        {currentPage === 'tos' && <TermsOfService />}
      </main>

      <footer>
        <div className="container">
          <div>
            <a href="https://github.com/adarsh9780/mac_remote/blob/main/SECURITY.md">Security Policy</a>
            <a href="https://github.com/adarsh9780/mac_remote">GitHub</a>
            <a href="https://www.linkedin.com/in/adarshmaurya/">LinkedIn</a>
          </div>
          <div className="footer-bottom">
            <p>Built by one person. Released under the MIT License.</p>
          </div>
        </div>
      </footer>
    </>
  );
}

export default App;
