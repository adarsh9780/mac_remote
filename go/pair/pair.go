package pair

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"math/big"
	"sync"
	"time"
)

type Session struct {
	Code      string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	Consumed  bool
	Attempts  int
}

var (
	activeSession *Session
	sessionMutex  sync.Mutex
)

// Generate creates a new PairingSession, invalidating any previous one.
// The `clock` parameter allows injecting a fake time for testing.
func Generate(ttl time.Duration, clock func() time.Time) *Session {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if clock == nil {
		clock = time.Now
	}

	now := clock()

	activeSession = &Session{
		Code:      generateCode(),
		Token:     generateToken(),
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		Consumed:  false,
		Attempts:  0,
	}

	// Return a copy so the caller doesn't modify the global state directly
	cpy := *activeSession
	return &cpy
}

// GetActiveIfValid returns the active session if it is unexpired and unconsumed.
// Otherwise returns nil.
func GetActiveIfValid(clock func() time.Time) *Session {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if clock == nil {
		clock = time.Now
	}

	if activeSession == nil {
		return nil
	}

	if activeSession.Consumed {
		return nil
	}

	if clock().After(activeSession.ExpiresAt) {
		return nil
	}

	cpy := *activeSession
	return &cpy
}

// VerifyCode verifies a 6-digit code.
func VerifyCode(code string, clock func() time.Time) bool {
	return verify(func(s *Session) bool {
		return subtle.ConstantTimeCompare([]byte(s.Code), []byte(code)) == 1
	}, clock)
}

// VerifyToken verifies a URL token.
func VerifyToken(token string, clock func() time.Time) bool {
	return verify(func(s *Session) bool {
		return subtle.ConstantTimeCompare([]byte(s.Token), []byte(token)) == 1
	}, clock)
}

// verify is undocumented. Please add documentation.
func verify(match func(*Session) bool, clock func() time.Time) bool {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if clock == nil {
		clock = time.Now
	}

	if activeSession == nil {
		return false
	}

	// Always increment attempts first
	activeSession.Attempts++

	// Prevent brute force (cap at 5 attempts)
	if activeSession.Attempts > 5 {
		return false
	}

	// Check consumed
	if activeSession.Consumed {
		return false
	}

	// Check expiry
	if clock().After(activeSession.ExpiresAt) {
		return false
	}

	// Check match
	if !match(activeSession) {
		return false
	}

	// Mark consumed
	activeSession.Consumed = true
	return true
}

// generateCode is undocumented. Please add documentation.
func generateCode() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	// Format with zero padding up to 6 digits
	return string([]byte{
		byte('0' + (n.Int64()/100000)%10),
		byte('0' + (n.Int64()/10000)%10),
		byte('0' + (n.Int64()/1000)%10),
		byte('0' + (n.Int64()/100)%10),
		byte('0' + (n.Int64()/10)%10),
		byte('0' + (n.Int64()/1)%10),
	})
}

// generateToken is undocumented. Please add documentation.
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// GetActiveForTest returns the currently active session for testing purposes.
func GetActiveForTest() *Session {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	return activeSession
}

// ClearForTest clears the active session.
func ClearForTest() {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	activeSession = nil
}
