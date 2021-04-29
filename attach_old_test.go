package mail

import (
	"testing"
)

// TODO: Remove this file before launch v3

func TestAttachmentsOLD(t *testing.T) {
	want := []byte("foo")
	t.Run("Inline File", func(t *testing.T) {
		msg := NewMSG()
		msg.AddInline("testdata/foo.txt", "foo")
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline Base64", func(t *testing.T) {
		msg := NewMSG()
		msg.AddInlineBase64("Zm9v", "foo", "")
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline Data", func(t *testing.T) {
		msg := NewMSG()
		msg.AddInlineData([]byte("foo"), "foo", "")
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment File", func(t *testing.T) {
		msg := NewMSG()
		msg.AddAttachment("testdata/foo.txt", "foo")
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment Base64", func(t *testing.T) {
		msg := NewMSG()
		msg.AddAttachmentBase64("Zm9v", "foo")
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment Data", func(t *testing.T) {
		msg := NewMSG()
		msg.AddAttachmentData([]byte("foo"), "foo", "")
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
}
