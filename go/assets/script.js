
document.addEventListener('DOMContentLoaded', () => {
    // Media & System Actions
    function sendAction(actionType) {
        if (navigator.vibrate) navigator.vibrate(15);
        fetch(`/api/action?type=${actionType}`, { method: 'POST' })
            .catch(err => console.error('Error sending action:', err));
    }

    // Attach click handlers to all data-action buttons
    document.querySelectorAll('.row-btn[data-action]').forEach(btn => {
        btn.addEventListener('touchstart', (e) => {
            e.preventDefault();
            btn.classList.add('active-hold');
            const action = btn.getAttribute('data-action');
            if (action) {
                sendAction(action);
                if (action === 'media_playpause') {
                    const mainPlayBtn = document.getElementById('mainPlayBtn');
                    const playSvg = document.getElementById('playSvg');
                    const playLbl = document.getElementById('playLbl');
                    if (mainPlayBtn && playSvg && playLbl) {
                        if (mainPlayBtn.classList.contains('playing-state')) {
                            playLbl.innerText = "Play";
                            playSvg.innerHTML = `<path d="M8 5v14l11-7z"/>`;
                            mainPlayBtn.classList.remove('playing-state');
                        } else {
                            playLbl.innerText = "Pause";
                            playSvg.innerHTML = `<path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>`;
                            mainPlayBtn.classList.add('playing-state');
                        }
                    }
                }
            }
        });
        btn.addEventListener('touchend', (e) => {
            e.preventDefault();
            btn.classList.remove('active-hold');
        });
        // Desktop fallback
        btn.addEventListener('mousedown', (e) => {
            btn.classList.add('active-hold');
            const action = btn.getAttribute('data-action');
            if (action) sendAction(action);
        });
        btn.addEventListener('mouseup', () => btn.classList.remove('active-hold'));
        btn.addEventListener('mouseleave', () => btn.classList.remove('active-hold'));
    });

    // Real-Time Server-Sent Events (SSE) for Status
    const cassetteText = document.getElementById('cassetteText');
    const marqueeWrapper = document.getElementById('marqueeWrapper');
    const brightBadge = document.getElementById('brightBadge');
    const volBadge = document.getElementById('volBadge');
    const playSvg = document.getElementById('playSvg');
    const playLbl = document.getElementById('playLbl');

    function setupSSE() {
        const eventSource = new EventSource('/api/events');
        
        eventSource.onmessage = (e) => {
            try {
                const data = JSON.parse(e.data);
                
                if (volBadge) volBadge.innerText = Math.round(data.volume * 100);
                if (brightBadge) brightBadge.innerText = Math.round(data.brightness * 100);

                // Update Cassette Marquee
                if (data.media && data.media.title && data.media.title !== "") {
                    cassetteText.innerText = `NOW PLAYING: ${data.media.title}  ●  SYSTEM STATE: Online  ●  `;
                    marqueeWrapper.classList.add('playing');
                } else {
                    cassetteText.innerText = `SYSTEM STATE: Online  ●  NO MEDIA PLAYING`;
                    marqueeWrapper.classList.remove('playing');
                }

                // Play/Pause button sync (only if server provides 'playing' state)
                const mainPlayBtn = document.getElementById('mainPlayBtn');
                if (mainPlayBtn && playSvg && playLbl) {
                    if (data.media && data.media.playing !== undefined && data.media.playing !== null) {
                        if (data.media.playing) {
                            playLbl.innerText = "Pause";
                            playSvg.innerHTML = `<path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>`;
                            mainPlayBtn.classList.add('playing-state');
                        } else {
                            playLbl.innerText = "Play";
                            playSvg.innerHTML = `<path d="M8 5v14l11-7z"/>`;
                            mainPlayBtn.classList.remove('playing-state');
                        }
                    }
                }

                if (window.checkAppsHash && data.apps_hash) {
                    window.checkAppsHash(data.apps_hash);
                }

                if (window.handleTextFocusChange && data.text_focused !== undefined) {
                    window.handleTextFocusChange(data.text_focused);
                }
            } catch (err) {
                console.error('Error parsing SSE data:', err);
            }
        };

        eventSource.onerror = (err) => {
            console.error('SSE Error:', err);
            eventSource.close();
            // Attempt to reconnect after 3 seconds
            setTimeout(setupSSE, 3000);
        };
    }

    setupSSE();

    // Settings Panel & Media Toggle
    const settingsToggleBtn = document.getElementById('settings-toggle-btn');
    const settingsPanel = document.getElementById('settingsPanel');
    const cassetteBtn = document.getElementById('cassette-btn');
    const mediaRowContainer = document.getElementById('media-row-container');

    if (settingsToggleBtn && settingsPanel) {
        settingsToggleBtn.addEventListener('click', () => {
            if (navigator.vibrate) navigator.vibrate(10);
            settingsPanel.classList.toggle('open');
        });
    }

    if (cassetteBtn && mediaRowContainer) {
        cassetteBtn.addEventListener('click', () => {
            if (navigator.vibrate) navigator.vibrate(10);
            if (mediaRowContainer.style.display === 'none') {
                mediaRowContainer.style.display = 'grid';
            } else {
                mediaRowContainer.style.display = 'none';
            }
        });
    }

    // Sensitivity & Scroll Speed settings
    let baseSensitivity = 2.5;
    let baseScrollSpeed = 2.0;
    const sensitivitySlider = document.getElementById('sensitivity-slider');
    const sensitivityVal = document.getElementById('sensitivity-val');
    const scrollSpeedSlider = document.getElementById('scroll-speed-slider');
    const scrollSpeedVal = document.getElementById('scroll-speed-val');

    if (sensitivitySlider && sensitivityVal) {
        sensitivitySlider.addEventListener('input', (e) => {
            baseSensitivity = parseFloat(e.target.value);
            sensitivityVal.innerText = baseSensitivity.toFixed(1);
        });
    }
    if (scrollSpeedSlider && scrollSpeedVal) {
        scrollSpeedSlider.addEventListener('input', (e) => {
            baseScrollSpeed = parseFloat(e.target.value);
            scrollSpeedVal.innerText = baseScrollSpeed.toFixed(1);
        });
    }

    // Mouse Action Sender
    function sendMouseAction(type, data = {}) {
        const qs = new URLSearchParams({ type, ...data }).toString();
        fetch(`/api/mouse?${qs}`, {
            method: 'POST'
        }).catch(err => console.error('Error sending mouse action:', err));
    }

    // Continuous Scroll Buttons
    const scrollUpBtn = document.getElementById('scroll-up-btn');
    const scrollDownBtn = document.getElementById('scroll-down-btn');
    let scrollInterval = null;

    function startContinuousScroll(direction) {
        if (navigator.vibrate) navigator.vibrate(15);
        sendMouseAction('scroll', { dx: 0, dy: direction * baseScrollSpeed * 10 });
        scrollInterval = setInterval(() => {
            sendMouseAction('scroll', { dx: 0, dy: direction * baseScrollSpeed * 10 });
        }, 30);
    }

    function stopContinuousScroll() {
        if (scrollInterval) clearInterval(scrollInterval);
        scrollInterval = null;
    }

    function bindScrollBtn(btn, direction) {
        if (!btn) return;
        btn.addEventListener('touchstart', (e) => { e.preventDefault(); e.stopPropagation(); startContinuousScroll(direction); });
        btn.addEventListener('touchend', (e) => { e.preventDefault(); e.stopPropagation(); stopContinuousScroll(); });
        btn.addEventListener('mousedown', (e) => { e.stopPropagation(); startContinuousScroll(direction); });
        btn.addEventListener('mouseup', (e) => { e.stopPropagation(); stopContinuousScroll(); });
        btn.addEventListener('mouseleave', (e) => { e.stopPropagation(); stopContinuousScroll(); });
    }

    bindScrollBtn(scrollUpBtn, 1);
    bindScrollBtn(scrollDownBtn, -1);

    // Trackpad Multi-Touch & Gesture Logic
    const trackpadCanvas = document.getElementById('trackpad-canvas');
    const tpLeftClick = document.getElementById('tp-left-click');
    const tpRightClick = document.getElementById('tp-right-click');

    let lastX = 0;
    let lastY = 0;
    let pendingDx = 0;
    let pendingDy = 0;
    let isMoving = false;
    let touchMode = 0; // 0 = none, 1 = move, 2 = scroll
    let isMouseDown = false;
    let hasDragged = false;
    let singleTapTimeout = null;
    let lastTapTime = 0;
    const DOUBLE_TAP_DELAY = 300;

    // Movement Loop for high-frequency updates
    setInterval(() => {
        if (isMoving && (pendingDx !== 0 || pendingDy !== 0)) {
            const actionType = touchMode === 2 ? 'scroll' : 'move';
            // Scroll usually requires higher multiplier on Mac
            const multiplier = actionType === 'scroll' ? 2.5 : 1.0;
            sendMouseAction(actionType, { dx: pendingDx * multiplier, dy: pendingDy * multiplier });
            pendingDx = 0;
            pendingDy = 0;
        }
    }, 16);

    if (trackpadCanvas) {
        trackpadCanvas.addEventListener('touchstart', (e) => {
            e.preventDefault();
            const now = Date.now();
            touchMode = e.touches.length;
            isMoving = true;
            hasDragged = false;

            lastX = (touchMode === 1) ? e.touches[0].clientX : (e.touches[0].clientX + e.touches[1].clientX) / 2;
            lastY = (touchMode === 1) ? e.touches[0].clientY : (e.touches[0].clientY + e.touches[1].clientY) / 2;

            if (touchMode === 1) {
                // Double tap to drag detection
                if (now - lastTapTime < DOUBLE_TAP_DELAY) {
                    clearTimeout(singleTapTimeout);
                    isMouseDown = true;
                    if (navigator.vibrate) navigator.vibrate(30);
                    sendMouseAction('down', { button: 'left' });
                    // Provide visual feedback for drag mode
                    trackpadCanvas.style.boxShadow = "inset 0 0 20px rgba(57, 255, 20, 0.2)";
                }
            } else {
                // Multi-touch cancels drag
                if (isMouseDown) {
                    isMouseDown = false;
                    sendMouseAction('up', { button: 'left' });
                    trackpadCanvas.style.boxShadow = "inset 0 4px 12px rgba(0,0,0,0.8)";
                }
            }
        });

        trackpadCanvas.addEventListener('touchmove', (e) => {
            e.preventDefault();
            if (isMoving) {
                let currentX = 0, currentY = 0;
                
                if (touchMode === 1) {
                    currentX = e.touches[0].clientX;
                    currentY = e.touches[0].clientY;
                } else if (touchMode === 2) {
                    currentX = (e.touches[0].clientX + e.touches[1].clientX) / 2;
                    currentY = (e.touches[0].clientY + e.touches[1].clientY) / 2;
                }

                if (touchMode === 1 || touchMode === 2) {
                    const dx = currentX - lastX;
                    const dy = currentY - lastY;
                    const sensitivity = touchMode === 2 ? baseScrollSpeed : baseSensitivity; 

                    if (Math.abs(dx) > 1 || Math.abs(dy) > 1) {
                        hasDragged = true;
                    }

                    pendingDx += dx * sensitivity;
                    pendingDy += dy * sensitivity;

                    lastX = currentX;
                    lastY = currentY;
                }
            }
        });

        trackpadCanvas.addEventListener('touchend', (e) => {
            e.preventDefault();
            const now = Date.now();
            isMoving = false;
            pendingDx = 0;
            pendingDy = 0;

            if (touchMode === 1 && !isMouseDown && !hasDragged) {
                // Wait to see if it's a double tap before firing single click
                singleTapTimeout = setTimeout(() => {
                    sendMouseAction('click');
                    if (navigator.vibrate) navigator.vibrate(10);
                }, DOUBLE_TAP_DELAY);
                lastTapTime = now;
            } else if (touchMode === 2 && !hasDragged) {
                sendMouseAction('rightclick');
                if (navigator.vibrate) navigator.vibrate([10, 30, 10]);
                lastTapTime = 0;
            } else if (isMouseDown) {
                isMouseDown = false;
                sendMouseAction('up', { button: 'left' });
                trackpadCanvas.style.boxShadow = "inset 0 4px 12px rgba(0,0,0,0.8)";
            }

            touchMode = e.touches.length;
            if (touchMode > 0) {
                isMoving = true;
                lastX = (touchMode === 1) ? e.touches[0].clientX : (e.touches[0].clientX + e.touches[1].clientX) / 2;
                lastY = (touchMode === 1) ? e.touches[0].clientY : (e.touches[0].clientY + e.touches[1].clientY) / 2;
            }
        });
    }

    // Physical Button Clicks (Click and hold support)
    function setupClickPad(btn, buttonType) {
        if (!btn) return;
        btn.addEventListener('touchstart', (e) => {
            e.preventDefault();
            e.stopPropagation();
            if (navigator.vibrate) navigator.vibrate(20);
            btn.classList.add('active-hold');
            sendMouseAction('down', { button: buttonType });
        });
        btn.addEventListener('touchend', (e) => {
            e.preventDefault();
            e.stopPropagation();
            btn.classList.remove('active-hold');
            sendMouseAction('up', { button: buttonType });
        });
        // Desktop support
        btn.addEventListener('mousedown', (e) => {
            e.stopPropagation();
            btn.classList.add('active-hold');
            sendMouseAction('down', { button: buttonType });
        });
        btn.addEventListener('mouseup', (e) => {
            e.stopPropagation();
            btn.classList.remove('active-hold');
            sendMouseAction('up', { button: buttonType });
        });
        btn.addEventListener('mouseleave', (e) => {
            e.stopPropagation();
            btn.classList.remove('active-hold');
            sendMouseAction('up', { button: buttonType });
        });
    }

    setupClickPad(tpLeftClick, 'left');
    setupClickPad(tpRightClick, 'right');

    // --- Typing Support ---
    const keyboardToggleBtn = document.getElementById('keyboard-toggle-btn');
    const typingModalOverlay = document.getElementById('typingModalOverlay');
    const typeInput = document.getElementById('typeInput');

    if (keyboardToggleBtn && typingModalOverlay && typeInput) {
        let lastValue = "";

        keyboardToggleBtn.addEventListener('click', () => {
            if (navigator.vibrate) navigator.vibrate(10);
            typingModalOverlay.classList.add('open');
            typeInput.value = "";
            lastValue = "";
            setTimeout(() => typeInput.focus(), 100);
        });

        typingModalOverlay.addEventListener('click', (e) => {
            if (e.target === typingModalOverlay) {
                typingModalOverlay.classList.remove('open');
                typeInput.blur();
            }
        });

        typeInput.addEventListener('input', (e) => {
            const currentValue = typeInput.value;
            
            if (currentValue.length > lastValue.length) {
                const addedChars = currentValue.slice(lastValue.length);
                if (addedChars === '\n') {
                    fetch('/api/action?type=press_enter', { method: 'POST' });
                    // Optional: remove the \n from the box so it doesn't pile up
                    typeInput.value = lastValue;
                } else {
                    fetch(`/api/action?type=type_chars&text=${encodeURIComponent(addedChars)}`, { method: 'POST' });
                }
            } else if (currentValue.length < lastValue.length) {
                const deletedCount = lastValue.length - currentValue.length;
                fetch(`/api/action?type=type_backspace&steps=${deletedCount}`, { method: 'POST' });
            }
            
            lastValue = currentValue;
        });

        const typingEnterBtn = document.getElementById('typingEnterBtn');
        if (typingEnterBtn) {
            typingEnterBtn.addEventListener('click', () => {
                if (navigator.vibrate) navigator.vibrate(10);
                fetch('/api/action?type=press_enter', { method: 'POST' });
                typeInput.value = "";
                lastValue = "";
                typeInput.focus();
            });
        }
    }

    // --- App Switcher & Dock Access ---
    const appsToggleBtn = document.getElementById('apps-toggle-btn');
    const appSwitcherModalOverlay = document.getElementById('appSwitcherModalOverlay');
    const showDockBtn = document.getElementById('showDockBtn');
    const appGrid = document.getElementById('appGrid');
    
    let currentAppsHash = "";
    
    function renderAppGrid(apps) {
        if (!appGrid) return;
        appGrid.innerHTML = '';
        apps.forEach(app => {
            const el = document.createElement('div');
            el.className = 'app-item';
            el.innerHTML = `
                <img src="data:image/png;base64,${app.icon}" alt="${app.name}">
                <span>${app.name}</span>
            `;
            el.addEventListener('click', () => {
                if (navigator.vibrate) navigator.vibrate(10);
                fetch(`/api/action?type=switch_app&pid=${app.pid}`, { method: 'POST' });
                appSwitcherModalOverlay.classList.remove('open');
            });
            appGrid.appendChild(el);
        });
    }

    function fetchAndRenderApps() {
        fetch('/api/apps')
            .then(res => res.json())
            .then(apps => renderAppGrid(apps))
            .catch(err => console.error('Failed to load apps:', err));
    }

    if (appsToggleBtn && appSwitcherModalOverlay) {
        appsToggleBtn.addEventListener('click', () => {
            if (navigator.vibrate) navigator.vibrate(10);
            appSwitcherModalOverlay.classList.add('open');
            fetchAndRenderApps(); // Always fetch fresh on open just in case
        });

        appSwitcherModalOverlay.addEventListener('click', (e) => {
            if (e.target === appSwitcherModalOverlay) {
                appSwitcherModalOverlay.classList.remove('open');
            }
        });
    }

    if (showDockBtn) {
        showDockBtn.addEventListener('click', () => {
            if (navigator.vibrate) navigator.vibrate(10);
            fetch('/api/action?type=show_dock', { method: 'POST' });
            appSwitcherModalOverlay.classList.remove('open');
        });
    }

    // Export a function or hook to let the SSE message know the hash changed
    window.checkAppsHash = function(newHash) {
        if (newHash && newHash !== currentAppsHash) {
            currentAppsHash = newHash;
            // Only fetch if the modal is currently open
            if (appSwitcherModalOverlay && appSwitcherModalOverlay.classList.contains('open')) {
                fetchAndRenderApps();
            }
        }
    };

    let lastTextFocused = false;
    window.handleTextFocusChange = function(isFocused) {
        if (isFocused !== lastTextFocused) {
            lastTextFocused = isFocused;
            if (isFocused && typingModalOverlay && !typingModalOverlay.classList.contains('open')) {
                // Auto-open keyboard
                if (navigator.vibrate) navigator.vibrate(10);
                typingModalOverlay.classList.add('open');
                typeInput.value = "";
                lastValue = "";
                setTimeout(() => typeInput.focus(), 100);
            }
        }
    };
});
