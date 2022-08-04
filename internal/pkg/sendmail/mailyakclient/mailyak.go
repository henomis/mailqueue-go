package mailyakclient

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/domodwyer/mailyak/v3"
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
)

const (
	filePrefix = "file://"
)

type MailYakClientOptions struct {
	Server   string
	Username string
	Password string

	From     string
	FromName string
	ReplyTo  string
	Attempts string
}

type MailYakClient struct {
	mailYakClientOptions *MailYakClientOptions
	mailyakInstance      *mailyak.MailYak
}

func New(mailYakClientOptions *MailYakClientOptions) *MailYakClient {

	auth := newLoginAuth(mailYakClientOptions.Username, mailYakClientOptions.Password)
	mailyakInstance := mailyak.New(mailYakClientOptions.Server, auth)

	mailyakInstance.From(mailYakClientOptions.From)
	mailyakInstance.FromName(mailYakClientOptions.FromName)
	mailyakInstance.ReplyTo(mailYakClientOptions.ReplyTo)

	return &MailYakClient{
		mailyakInstance:      mailyakInstance,
		mailYakClientOptions: mailYakClientOptions,
	}

}

func (s *MailYakClient) Send(email *storagemodel.Email) error {

	if s.mailyakInstance == nil {
		return fmt.Errorf("MailYak instance is nil")
	}

	if len(email.To) > 0 {
		s.mailyakInstance.To(email.To)
	}

	if len(email.Subject) > 0 {
		s.mailyakInstance.Subject(email.Subject)
	}

	if len(email.Cc) > 0 {
		s.mailyakInstance.Cc(email.Cc)
	}

	if len(email.Bcc) > 0 {
		s.mailyakInstance.Bcc(email.Bcc)
	}

	s.mailyakInstance.ClearAttachments()

	for _, v := range email.Attachments {

		if strings.HasPrefix(v.Data, filePrefix) {
			path := strings.TrimPrefix(v.Data, filePrefix)

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			s.mailyakInstance.AttachWithMimeType(v.Name, f, v.Mime)

		} else {
			s.mailyakInstance.AttachWithMimeType(v.Name, base64.NewDecoder(base64.StdEncoding, strings.NewReader(v.Data)), v.Mime)
		}

	}

	if len(email.HTML) > 0 {
		s.mailyakInstance.HTML().Set(email.HTML)
	} else {
		s.mailyakInstance.HTML().Set(email.Data)
	}

	return s.mailyakInstance.Send()

}

func (s *MailYakClient) Attempts() int {
	attemptsAsInt, _ := strconv.Atoi(s.mailYakClientOptions.Attempts)
	return attemptsAsInt
}
