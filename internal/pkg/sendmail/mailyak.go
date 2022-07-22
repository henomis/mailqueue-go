package sendmail

import (
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/domodwyer/mailyak/v3"
	"github.com/henomis/mailqueue-go/internal/pkg/email"
)

const (
	filePrefix = "file://"
	//ErrNilClient not-initialized client pointer
	ErrNilClient = "Invalid smtp client"
)

//MailYakClient sends email using MailYak package
type MailYakClient struct {
	options *Options

	client *mailyak.MailYak
}

//NewMailYakClient returns a new MailYak smtp client
func NewMailYakClient(options *Options) *MailYakClient {

	auth := newLoginAuth(options.Username, options.Password)
	c := mailyak.New(options.Server, auth)

	c.From(options.From)
	c.FromName(options.FromName)
	c.ReplyTo(options.ReplyTo)

	return &MailYakClient{
		client:  c,
		options: options,
	}

}

//Send implements sending email
func (s *MailYakClient) Send(email *email.Email) error {

	if s.client == nil {
		return errors.New(ErrNilClient)
	}

	s.client.To(email.To)
	s.client.Subject(email.Subject)
	s.client.Cc(email.Cc)
	s.client.Bcc(email.Bcc)

	s.client.ClearAttachments()

	for _, v := range email.Attachments {

		if strings.HasPrefix(v.Data, filePrefix) {
			path := strings.TrimPrefix(v.Data, filePrefix)

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			s.client.AttachWithMimeType(v.Name, f, v.Mime)

		} else {
			s.client.AttachWithMimeType(v.Name, base64.NewDecoder(base64.StdEncoding, strings.NewReader(v.Data)), v.Mime)
		}

	}

	s.client.HTML().Set(email.Data)

	return s.client.Send()

}

//Attempts return attempts
func (s *MailYakClient) Attempts() int {
	a, _ := strconv.Atoi(s.options.Attempts)
	return a
}
