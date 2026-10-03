package mail

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
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

func TestInlineContentID(t *testing.T) {
	const contentID = "image@example.com"
	const body = `<img src="cid:image@example.com">`
	tests := []struct {
		name string
		file File
	}{
		{"Data", File{Data: []byte("foo")}},
		{"Base64", File{B64Data: "Zm9v"}},
		{"File", File{FilePath: "testdata/foo.txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := tt.file
			file.Name = "foo.txt"
			file.ContentID = contentID
			file.Inline = true
			msg := NewMSG().SetBody(TextHTML, body).Attach(&file)
			if msg.Error != nil {
				t.Fatal(msg.Error)
			}

			parsed, err := mail.ReadMessage(strings.NewReader(msg.GetMessage()))
			if err != nil {
				t.Fatal(err)
			}
			mediaType, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
			if err != nil {
				t.Fatal(err)
			}
			if mediaType != "multipart/related" {
				t.Fatalf("got content type %q, want multipart/related", mediaType)
			}
			reader := multipart.NewReader(parsed.Body, params["boundary"])
			html, err := reader.NextPart()
			if err != nil {
				t.Fatal(err)
			}
			gotBody, err := ioutil.ReadAll(html)
			if err != nil {
				t.Fatal(err)
			}
			if string(gotBody) != body {
				t.Fatalf("got body %q, want %q", gotBody, body)
			}
			attachment, err := reader.NextPart()
			if err != nil {
				t.Fatal(err)
			}
			if got := attachment.Header.Get("Content-ID"); got != "<"+contentID+">" {
				t.Errorf("got Content-ID %q, want <%s>", got, contentID)
			}
			if attachment.FileName() != file.Name {
				t.Errorf("got filename %q, want %q", attachment.FileName(), file.Name)
			}
			data, err := ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, attachment))
			if err != nil {
				t.Fatal(err)
			}
			checkByteSlice(t, data, []byte("foo"))
		})
	}
}
