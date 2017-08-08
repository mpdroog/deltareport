package queue

import (
	"fmt"
	"gopkg.in/gomail.v1"
	"deltareport/config"
)

func MailSend(conf config.Queue, subject, text string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("Message-ID", fmt.Sprintf("<%s@%s>", RandText(32), config.Hostname))
	msg.SetHeader("X-Mailer", "dnsproxy")
	msg.SetHeader("X-Priority", "3")

	msg.SetHeader("From", fmt.Sprintf("%s <%s>", conf.FromName, conf.From))
	//msg.SetHeader("Reply-To", fmt.Sprintf("Support <%s>", conf.From))

	msg.SetHeader("To", conf.To...)
	msg.SetHeader("Subject", conf.Subject + subject)
	msg.SetBody("text/plain", text)

	mailer := gomail.NewMailer(conf.Host, conf.User, conf.Pass, conf.Port)
	return mailer.Send(msg)
}