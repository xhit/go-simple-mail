package mail

import (
	"bytes"
	"testing"
)

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("got error: %v", err)
	}
}

func checkByteSlice(t *testing.T, got, want []byte) {
	if !bytes.Equal(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func TestAttachments(t *testing.T) {
	want := []byte("foo")
	t.Run("Inline File", func(t *testing.T) {
		msg := NewMSG()
		msg.Attach(&File{FilePath: "testdata/foo.txt", Name: "foo", Inline: true})
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline Base64", func(t *testing.T) {
		msg := NewMSG()
		msg.Attach(&File{B64Data: "Zm9v", Name: "foo", Inline: true})
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline Data", func(t *testing.T) {
		msg := NewMSG()
		msg.Attach(&File{Data: []byte("foo"), Name: "foo", Inline: true})
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment File", func(t *testing.T) {
		msg := NewMSG()
		msg.Attach(&File{FilePath: "testdata/foo.txt", Name: "foo"})
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment Base64", func(t *testing.T) {
		msg := NewMSG()
		msg.Attach(&File{B64Data: "Zm9v", Name: "foo"})
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment Data", func(t *testing.T) {
		msg := NewMSG()
		msg.Attach(&File{Data: []byte("foo"), Name: "foo"})
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})

	// DEPRECATED. TODO: Remove before launch v3
	t.Run("Inline File Deprecated", func(t *testing.T) {
		msg := NewMSG()
		msg.AddInline("testdata/foo.txt", "foo")
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline Base64 Deprecated", func(t *testing.T) {
		msg := NewMSG()
		msg.AddInlineBase64("Zm9v", "foo", "")
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline Data Deprecated", func(t *testing.T) {
		msg := NewMSG()
		msg.AddInlineData([]byte("foo"), "foo", "")
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment File Deprecated", func(t *testing.T) {
		msg := NewMSG()
		msg.AddAttachment("testdata/foo.txt", "foo")
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment Base64 Deprecated", func(t *testing.T) {
		msg := NewMSG()
		msg.AddAttachmentBase64("Zm9v", "foo")
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment Data Deprecated", func(t *testing.T) {
		msg := NewMSG()
		msg.AddAttachmentData([]byte("foo"), "foo", "")
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline File not name Deprecated", func(t *testing.T) {
		msg := NewMSG()
		msg.AddInline("testdata/foo.txt")
		checkError(t, msg.Error)
		got := msg.inlines[0].Data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment File not name Deprecated", func(t *testing.T) {
		msg := NewMSG()
		msg.AddAttachment("testdata/foo.txt")
		checkError(t, msg.Error)
		got := msg.attachments[0].Data
		checkByteSlice(t, got, want)
	})
}
