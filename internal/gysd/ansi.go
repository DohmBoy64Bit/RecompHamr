package gysd

import "regexp"

// ansiRE matches a CSI-style ANSI escape (ESC [ … final-byte) plus the
// shorter ESC ] OSC sequences pytest, npm, etc. emit. Storing color codes
// verbatim would make evidence-match brittle — the model reads stripped
// output in its tool-result, so it would quote the stripped form, but
// VerifyLog kept the colored form: substring miss every time. Strip on
// store fixes that asymmetry once.
//
// The third + fourth alternatives scrub trailing *incomplete* escape
// sequences — a verify subprocess killed mid-color (Ctrl+C, timeout) often
// emits a partial `\x1b[31` or `\x1b]0;tit` with no final byte. Without the
// trailing-anchor cases those raw ESC bytes survive into VerifyLog and
// later into the user-block on a red-streak yield, where tea.Println dumps
// them straight to the terminal — at best a cosmetic glitch, at worst a
// stray ESC that mangles the prompt re-render.
var ansiRE = regexp.MustCompile(
	"(?:\x1b\\[[0-?]*[ -/]*[@-~])" + // complete CSI
		"|(?:\x1b\\][^\x07\x1b]*(?:\x07|\x1b\\\\))" + // complete OSC
		"|(?:\x1b\\[[0-?]*[ -/]*$)" + // trailing partial CSI (no final byte)
		"|(?:\x1b\\][^\x07\x1b]*$)" + // trailing partial OSC (no terminator)
		"|(?:\x1b$)") // bare trailing ESC

func stripANSI(s string) string {
	if !needsStrip(s) {
		return s
	}
	return ansiRE.ReplaceAllString(s, "")
}

// needsStrip is a fast pre-check so the regex doesn't run on the common
// case (no escape codes at all).
func needsStrip(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			return true
		}
	}
	return false
}
