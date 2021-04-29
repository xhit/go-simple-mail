package mail

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"mime"
	"path/filepath"
)

// TODO: Remove this file before launch v3

// AddAttachment. DEPRECATED. Use Attach method. Allows you to add an attachment to the email message.
// You can optionally provide a different name for the file.
func (email *Email) AddAttachment(file string, name ...string) *Email {
	if email.Error != nil {
		return email
	}

	if len(name) > 1 {
		email.Error = errors.New("Mail Error: Attach can only have a file and an optional name")
		return email
	}

	email.Error = email.attach(file, false, name[0], "")

	return email
}

// AddAttachmentData. DEPRECATED. Use Attach method. Allows you to add an in-memory attachment to the email message.
func (email *Email) AddAttachmentData(data []byte, filename, mimeType string) *Email {
	if email.Error != nil {
		return email
	}

	email.attachDataOLD(data, false, filename, mimeType)

	return email
}

// AddAttachmentBase64. DEPRECATED. Use Attach method. Allows you to add an attachment in base64 to the email message.
// You need provide a name for the file.
func (email *Email) AddAttachmentBase64(b64File, name string) *Email {
	if email.Error != nil {
		return email
	}

	if len(name) < 1 || len(b64File) < 1 {
		email.Error = errors.New("Mail Error: Attach Base64 need have a base64 string and name")
		return email
	}

	email.Error = email.attachB64OLD(b64File, false, name, "")

	return email
}

// AddInline. DEPRECATED. Use Attach method. Allows you to add an inline attachment to the email message.
// You can optionally provide a different name for the file.
func (email *Email) AddInline(file string, name ...string) *Email {
	if email.Error != nil {
		return email
	}

	if len(name) > 1 {
		email.Error = errors.New("Mail Error: Inline can only have a file and an optional name")
		return email
	}

	email.Error = email.attach(file, true, name[0], "")

	return email
}

// AddInlineData. DEPRECATED. Use Attach method. Allows you to add an inline in-memory attachment to the email message.
func (email *Email) AddInlineData(data []byte, filename, mimeType string) *Email {
	if email.Error != nil {
		return email
	}

	email.attachDataOLD(data, true, filename, mimeType)

	return email
}

// AddInlineBase64. DEPRECATED. Use Attach method. Allows you to add an inline in-memory base64 encoded attachment to the email message.
// You need provide a name for the file. If mimeType is an empty string, attachment mime type will be deduced
// from the file name extension and defaults to application/octet-stream.
func (email *Email) AddInlineBase64(b64File, name, mimeType string) *Email {
	if email.Error != nil {
		return email
	}

	if len(name) < 1 || len(b64File) < 1 {
		email.Error = errors.New("Mail Error: Attach Base64 need have a base64 string and name")
		return email
	}

	email.Error = email.attachB64OLD(b64File, true, name, mimeType)

	return email
}

// attach does the low level attaching of the files
func (email *Email) attach(f string, inline bool, name, mimeType string) error {
	// Get the file data
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return errors.New("Mail Error: Failed to add file with following error: " + err.Error())
	}

	// if no alternative name was provided, get the filename
	if len(name) == 0 {
		_, name = filepath.Split(f)
	}

	email.attachDataOLD(data, inline, name, mimeType)

	return nil
}

// attachDataOLD does the low level attaching of the in-memory data
func (email *Email) attachDataOLD(data []byte, inline bool, name, mimeType string) {
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(name))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	if inline {
		email.inlines = append(email.inlines, &File{
			Name:     name,
			MimeType: mimeType,
			Data:     data,
		})
	} else {
		email.attachments = append(email.attachments, &File{
			Name:     name,
			MimeType: mimeType,
			Data:     data,
		})
	}
}

// attachB64OLD does the low level attaching of the files but decoding base64 instead have a filepath
func (email *Email) attachB64OLD(b64File string, inline bool, name, mimeType string) error {

	// decode the string
	dec, err := base64.StdEncoding.DecodeString(b64File)
	if err != nil {
		return errors.New("Mail Error: Failed to decode base64 attachment with following error: " + err.Error())
	}

	email.attachDataOLD(dec, inline, name, mimeType)

	return nil
}
