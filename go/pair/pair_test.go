package pair

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestGeneratePairingSession_HasSixDigitCode(t *testing.T) {
	ClearForTest()
	s := Generate(60*time.Second, nil)
	if len(s.Code) != 6 {
		t.Fatalf("Expected 6 digits, got %d", len(s.Code))
	}
	matched, _ := regexp.MatchString(`^[0-9]{6}$`, s.Code)
	if !matched {
		t.Fatalf("Expected numeric code, got %s", s.Code)
	}
}

func TestGeneratePairingSession_TokenIsURLSafeAndLong(t *testing.T) {
	ClearForTest()
	s := Generate(60*time.Second, nil)
	if len(s.Token) < 22 {
		t.Fatalf("Expected token len >= 22, got %d", len(s.Token))
	}
	if strings.ContainsAny(s.Token, "+/=") {
		t.Fatalf("Expected URL-safe base64 token, got %s", s.Token)
	}
}

func TestPairingSession_ValidWithinTTL(t *testing.T) {
	ClearForTest()
	fakeNow := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return fakeNow }
	
	s := Generate(60*time.Second, clock)

	// Advance clock by 59s
	fakeNow = fakeNow.Add(59 * time.Second)
	if !VerifyCode(s.Code, clock) {
		t.Fatal("Expected valid at 59s")
	}
}

func TestPairingSession_ExpiredAfterTTL(t *testing.T) {
	ClearForTest()
	fakeNow := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return fakeNow }
	
	s := Generate(60*time.Second, clock)

	// Advance clock by 61s
	fakeNow = fakeNow.Add(61 * time.Second)
	if VerifyCode(s.Code, clock) {
		t.Fatal("Expected invalid after 61s")
	}
}

func TestPairingSession_CannotBeReusedAfterConsumed(t *testing.T) {
	ClearForTest()
	clock := func() time.Time { return time.Now() }
	s := Generate(60*time.Second, clock)

	if !VerifyCode(s.Code, clock) {
		t.Fatal("Expected valid on first try")
	}
	if VerifyCode(s.Code, clock) {
		t.Fatal("Expected invalid on second try")
	}
}

func TestGeneratePairingSession_InvalidatesPreviousSession(t *testing.T) {
	ClearForTest()
	clock := func() time.Time { return time.Now() }
	
	s1 := Generate(60*time.Second, clock)
	Generate(60*time.Second, clock) // s2
	
	if VerifyCode(s1.Code, clock) {
		t.Fatal("Expected s1 to be invalidated by s2")
	}
}

func TestVerifyCode_LocksOutAfterFiveWrongAttempts(t *testing.T) {
	ClearForTest()
	clock := func() time.Time { return time.Now() }
	s := Generate(60*time.Second, clock)

	for i := 0; i < 5; i++ {
		if VerifyCode("000000", clock) {
			t.Fatal("Expected wrong code to fail")
		}
	}
	
	// 6th attempt with correct code should fail
	if VerifyCode(s.Code, clock) {
		t.Fatal("Expected lockout after 5 failed attempts")
	}
}
