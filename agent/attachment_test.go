package agent

import (
	"strings"
	"testing"
)

func TestMessageWithAttachmentFallback(t *testing.T) {
	got := MessageWithAttachmentFallback("Analyze this", []Attachment{{
		Path:     "/tmp/image.png",
		MIMEType: "image/png",
		Name:     "image.png",
	}})

	for _, want := range []string{"Analyze this", "image.png: /tmp/image.png", "image/png"} {
		if !strings.Contains(got, want) {
			t.Fatalf("fallback prompt missing %q: %s", want, got)
		}
	}
}
