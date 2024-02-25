package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail"
)

// templateFS is a new variable of type embed.FS (embedded
// file system) to hold email templates. This has a
// comment directive "//go:embed <path>" immediately
// above it, which indicates to Go the contents of the
// ./templates directory is where to store the
// embedded file system variable.
// The "go embed" directive can only be used on global
// variables. The <path> should be relative to the
// source code file containing the directive.

//go:embed "templates"
var templateFS embed.FS

// Mailer is a struct containing the mail.Dialer
// instance (used to connect to a SMTP server) and the
// sender info for the emails (name and address of who
// the email is from.)
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New method initializes a new mail.Dialer.
func New(
	host string,
	port int,
	username string,
	password string,
	sender string,
) Mailer {
	// Initialize a new mail.Dailer instance with the
	// given SMTP server settings. Configure a 5 second
	// timeout when an email is sent.
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	// Return a Mailer instance containing the dialer
	// and sender information.
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send method takes the recepient email address as
// the first parameter, the name of the file containing
// the templates, and the dynamic data for the templates
// as an interface{} parameter.
func (m Mailer) Send(
	recepient string,
	templateFile string,
	data interface{},
) error {
	// Use the ParseFS method to parse the required
	// template file from the embedded file system.
	tmpl, err := template.New("email").ParseFS(
		templateFS,
		"templates/"+templateFile,
	)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in
	// the dynamic data and storing the result in a
	// bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Execute the named template "plainBody", passing in
	// the dynamic data and storing the result in a
	// bytes.Buffer variable.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Execute the named template "htmlBody", passing in
	// the dynamic data and storing the result in a
	// bytes.Buffer variable.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Use the mail.NewMessage() function to initialize a
	// new mail.Message instance. Then use SetHeader()
	// method to set the email recepient, sender, and
	// subject headers, the SetBody() method to set the
	// plaintext body, and the AddAlternative() method
	// to set the HTML body. AddAlternative() must be
	// called AFTER SetBody().
	msg := mail.NewMessage()
	msg.SetHeader("To", recepient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Call BuildAndSend() method on the dialer, passing
	// in the message to send. This opens a connection to
	// the SMTP server, sends the message, then closes the
	// connection. If there is a timeout, it will return a
	// "dial tcp: i/o timeout" error.
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil
}
