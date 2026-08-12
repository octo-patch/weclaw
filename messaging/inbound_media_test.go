package messaging

import (
	"testing"

	"github.com/fastclaw-ai/weclaw/ilink"
)

func TestExtractVideo(t *testing.T) {
	want := &ilink.VideoItem{VideoSize: 42}
	msg := ilink.WeixinMessage{ItemList: []ilink.MessageItem{
		{Type: ilink.ItemTypeText, TextItem: &ilink.TextItem{Text: "caption"}},
		{Type: ilink.ItemTypeVideo, VideoItem: want},
	}}
	if got := extractVideo(msg); got != want {
		t.Fatalf("extractVideo() = %#v, want %#v", got, want)
	}
}

func TestMimeTypeFromImageExt(t *testing.T) {
	tests := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".bin":  "application/octet-stream",
	}
	for ext, want := range tests {
		if got := mimeTypeFromImageExt(ext); got != want {
			t.Fatalf("mimeTypeFromImageExt(%q) = %q, want %q", ext, got, want)
		}
	}
}
