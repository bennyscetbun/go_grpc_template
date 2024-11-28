package emails

import (
	"context"
	"net/smtp"

	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/environment"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/logger"
	"github.com/ztrue/tracerr"
)

func SendEmail(ctx context.Context, from, to, content string) error {
	smtpHost, err := environment.GetenvString("SMTPHOST", "")
	if err != nil {
		return tracerr.Wrap(err)
	}
	if smtpHost == "" {
		logger.Warningln("SMTPHOST not set")
		if environment.IsDebug() {
			logger.Println("EMAIL:", from, to, content)
		}
		return nil
	}
	smtpPort, err := environment.GetenvString("SMTPPORT", "587")
	if err != nil {
		return tracerr.Wrap(err)
	}
	smtpUser, err := environment.GetenvString("SMTPUSER", "")
	if err != nil {
		return tracerr.Wrap(err)
	}
	smtpPasswd, err := environment.GetenvString("SMTPPASSWORD", "")
	if err != nil {
		return tracerr.Wrap(err)
	}
	var auth smtp.Auth
	if smtpUser == "" || smtpPasswd == "" {
		logger.Warningln("SMTPUSER OR SMTPPASSWORD not set, NO PLAIN AUTH")
	} else {
		auth = smtp.PlainAuth("", smtpUser, smtpPasswd, smtpHost)
	}
	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(content)); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}
