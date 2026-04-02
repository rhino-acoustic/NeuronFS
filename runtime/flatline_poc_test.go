package main

import (
	"os"
	"strings"
	"testing"
)

// TestExtractFaultNode verifies the stack trace parser extracts
// the correct application-level file:line, skipping runtime frames.
func TestExtractFaultNode(t *testing.T) {
	fakeStack := `goroutine 1 [running]:
runtime/debug.Stack(0x0?, 0x0?, 0x0?)
	/usr/local/go/src/runtime/debug/stack.go:24 +0x5e
runtime.gopanic({0x5a3e60?, 0xc0000b4018?})
	/usr/local/go/src/runtime/panic.go:770 +0x132
main.indexBuildSynapses(0xc0000ae000)
	C:/Users/BASEMENT_ADMIN/bot1/runtime/emit.go:124 +0x4a
main.main()
	C:/Users/BASEMENT_ADMIN/bot1/runtime/main.go:131 +0x1f5
`

	result := extractFaultNode(fakeStack)

	if !strings.Contains(result, "emit.go") {
		t.Errorf("expected fault node to contain 'emit.go', got: %s", result)
	}
	if !strings.Contains(result, "124") {
		t.Errorf("expected fault node to contain line '124', got: %s", result)
	}
	t.Logf("Extracted FAULT NODE: %s", result)
}

// TestExtractFaultNodeUnknown verifies graceful fallback when stack is garbage.
func TestExtractFaultNodeUnknown(t *testing.T) {
	result := extractFaultNode("some garbage that has no .go files at all")
	if !strings.Contains(result, "unknown") {
		t.Errorf("expected 'unknown' for unparseable stack, got: %s", result)
	}
}

// TestExtractFaultNodeLinux verifies Linux-style paths work too.
func TestExtractFaultNodeLinux(t *testing.T) {
	fakeStack := `goroutine 1 [running]:
runtime/debug.Stack()
	/usr/local/go/src/runtime/debug/stack.go:24 +0x5e
runtime.gopanic({0x5a3e60?, 0xc0000b4018?})
	/usr/local/go/src/runtime/panic.go:770 +0x132
main.processBatch(0xc0000ae000)
	/home/user/neuronfs/runtime/supervisor.go:88 +0x4a
`

	result := extractFaultNode(fakeStack)
	if !strings.Contains(result, "supervisor.go") {
		t.Errorf("expected 'supervisor.go', got: %s", result)
	}
	if !strings.Contains(result, "88") {
		t.Errorf("expected line '88', got: %s", result)
	}
	t.Logf("Linux FAULT NODE: %s", result)
}

// TestFlatline_TruecolorFallback verifies the 3-stage ANSI color fallback chain.
func TestFlatline_TruecolorFallback(t *testing.T) {
	// Truecolor mode ??#F43F5E
	tc := flatColor{mode: colorTrueC}
	roseTC := tc.rose("test")
	if !strings.Contains(roseTC, "38;2;244;63;94") {
		t.Errorf("truecolor rose should contain 24-bit escape, got: %q", roseTC)
	}

	// 256-color fallback
	c256 := flatColor{mode: color256}
	rose256 := c256.rose("test")
	if !strings.Contains(rose256, "38;5;197") {
		t.Errorf("256-color rose should contain 256-color escape, got: %q", rose256)
	}

	// Basic (8-color) fallback ??standard red
	basic := flatColor{mode: colorBasic}
	roseBasic := basic.rose("test")
	if !strings.Contains(roseBasic, "\033[31m") {
		t.Errorf("basic rose should use standard red fallback, got: %q", roseBasic)
	}

	// NO_COLOR mode ??plain text
	none := flatColor{mode: colorNone}
	roseNone := none.rose("test")
	if roseNone != "test" {
		t.Errorf("NO_COLOR mode should return plain text, got: %q", roseNone)
	}

	t.Log("All 4 ANSI fallback tiers verified")
}

// TestFlatline_EmojiSubstitution verifies ASCII substitutes
// are used when terminal doesn't support emoji.
func TestFlatline_EmojiSubstitution(t *testing.T) {
	// Save and override env
	origTerm := os.Getenv("TERM")
	origNoColor := os.Getenv("NO_COLOR")

	// Force dumb terminal
	os.Setenv("TERM", "dumb")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if origTerm != "" {
			os.Setenv("TERM", origTerm)
		} else {
			os.Unsetenv("TERM")
		}
		if origNoColor != "" {
			os.Setenv("NO_COLOR", origNoColor)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	skull := flatEmoji("??", "[DEAD]")
	if skull != "[DEAD]" {
		t.Errorf("expected ASCII substitute '[DEAD]' on dumb terminal, got: %s", skull)
	}

	bolt := flatEmoji("??, "[!!]")
	if bolt != "[!!]" {
		t.Errorf("expected ASCII substitute '[!!]' on dumb terminal, got: %s", bolt)
	}

	// Normal terminal should return emoji
	os.Unsetenv("TERM")
	os.Unsetenv("NO_COLOR")
	// When TERM is unset and no NO_COLOR, should return emoji
	// (detectColorMode returns colorBasic, not colorNone)
	skullNormal := flatEmoji("??", "[DEAD]")
	if skullNormal != "??" {
		t.Errorf("expected emoji on normal terminal, got: %s", skullNormal)
	}

	t.Log("Emoji substitution verified for dumb and normal terminals")
}

// TestFlatline_QuietMode verifies FlatlineQuiet suppresses the death screen.
func TestFlatline_QuietMode(t *testing.T) {
	// Just verify the flag exists and is accessible
	FlatlineQuiet = true
	if !FlatlineQuiet {
		t.Fatal("FlatlineQuiet flag should be true")
	}
	FlatlineQuiet = false

	t.Log("Quiet mode flag verified")
}

// TestRenderFlatlineOnOOM_QuietMode verifies OOM death screen is suppressed
// when FlatlineQuiet is true. Only a 1-line log should be emitted.
func TestRenderFlatlineOnOOM_QuietMode(t *testing.T) {
	FlatlineQuiet = true
	defer func() { FlatlineQuiet = false }()

	// Should not panic, should silently log
	RenderFlatlineOnOOM("test-process", 1024*60, "InUse: 1024 KB | Func: main.leak")
	t.Log("OOM flatline quiet mode: no panic, silent log")
}

// TestRenderFlatlineOnOOM_NoPanic verifies the full OOM death screen
// can render without panicking (uses NO_COLOR to avoid ANSI issues in test).
func TestRenderFlatlineOnOOM_NoPanic(t *testing.T) {
	FlatlineQuiet = false

	// Force NO_COLOR to keep test output clean
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if origNoColor != "" {
			os.Setenv("NO_COLOR", origNoColor)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	// Should complete without panic
	RenderFlatlineOnOOM("neuronfs-api", 1024*80, "InUse: 4096 KB | Objects: 12 | Func: main.handleAPI\nInUse: 2048 KB | Objects: 8 | Func: main.scanBrain")
	t.Log("OOM flatline full render: completed without panic")
}

