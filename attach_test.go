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
}
