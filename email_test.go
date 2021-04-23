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
	t.Run("Inline file", func(t *testing.T) {
		msg := NewMSG()
		msg = msg.AddInline("testdata/foo.txt", "foo")
		checkError(t, msg.GetError())
		got := msg.GetInlines()[0].data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline Base64", func(t *testing.T) {
		msg := NewMSG()
		msg = msg.AddInlineBase64("Zm9v", "foo", "")
		checkError(t, msg.GetError())
		got := msg.GetInlines()[0].data
		checkByteSlice(t, got, want)
	})
	t.Run("Inline Data", func(t *testing.T) {
		msg := NewMSG()
		msg = msg.AddInlineData([]byte("foo"), "foo", "")
		checkError(t, msg.GetError())
		got := msg.GetInlines()[0].data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment File", func(t *testing.T) {
		msg := NewMSG()
		msg = msg.AddAttachment("testdata/foo.txt", "foo")
		checkError(t, msg.GetError())
		got := msg.GetAttachments()[0].data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment Base64", func(t *testing.T) {
		msg := NewMSG()
		msg = msg.AddAttachmentBase64("Zm9v", "foo")
		checkError(t, msg.GetError())
		got := msg.GetAttachments()[0].data
		checkByteSlice(t, got, want)
	})
	t.Run("Attachment Data", func(t *testing.T) {
		msg := NewMSG()
		msg = msg.AddAttachmentData([]byte("foo"), "foo", "")
		checkError(t, msg.GetError())
		got := msg.GetAttachments()[0].data
		checkByteSlice(t, got, want)
	})
}
