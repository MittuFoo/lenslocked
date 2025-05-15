package main

import (
	"os"

	gomail "github.com/go-mail/mail/v2"
)

const (
	host     = "sandbox.smtp.mailtrap.io"
	port     = 25
	username = "252d33aa8af817"
	password = "95a74fbe24c009"
)

func main() {
  from := "test@lenslocked.com"
  to := "jon@calhoun.io"
  subject := "This is a test email"
  plaintext := "This is the body of the email"
  html := `<h1>Hello there buddy!</h1><p>This is the email</p><p>Hope you enjoy it</p>`

  dialer := gomail.NewDialer(host, port, username, password)
  msg := gomail.NewMessage()
  msg.SetHeader("To", to)
  msg.SetHeader("From", from)
  msg.SetHeader("Subject", subject)
  msg.SetBody("text/plain", plaintext)
  msg.AddAlternative("text/html", html)
  msg.WriteTo(os.Stdout)

  err := dialer.DialAndSend(msg)
    if err != nil {
	// TODO: Handle the error correctly
	panic(err)
    }
}