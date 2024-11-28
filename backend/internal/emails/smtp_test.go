package emails

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/random"
	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
)

func TestSendMail(t *testing.T) {
	server := smtpmock.New(smtpmock.ConfigurationAttr{})
	if !assert.NoError(t, server.Start()) {
		return
	}
	defer server.Stop()

	t.Setenv("SMTPHOST", "localhost")
	t.Setenv("SMTPPORT", strconv.Itoa(server.PortNumber()))
	fromEmail := "bla@bla.com"
	toEmail := "blou@blou.com"
	message := random.RandString(15)
	if !assert.NoError(t, SendEmail(context.Background(), fromEmail, toEmail, message)) {
		return
	}
	msgs := server.Messages()
	if !assert.Len(t, msgs, 1) {
		return
	}

	assert.Equal(t, message, strings.TrimSuffix(msgs[0].MsgRequest(), "\r\n"))
}
