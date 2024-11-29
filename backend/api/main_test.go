package api

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bennyscetbun/xxxyourappyyy/backend/database"
	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/rpc/apiproto"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/testhelpers"
	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/ztrue/tracerr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func preparePSQL(ctx context.Context) (*postgres.PostgresContainer, error) {
	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16.0",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return postgresContainer, tracerr.Wrap(err)
	}
	str, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		return postgresContainer, tracerr.Wrap(err)
	}
	db, err := sql.Open("pgx", str)
	if err != nil {
		return postgresContainer, tracerr.Wrap(err)
	}
	modulePath, err := testhelpers.GetCurrentGoModulePath()
	if err != nil {
		return postgresContainer, err
	}
	if err := os.Chdir(modulePath); err != nil {
		return postgresContainer, tracerr.Wrap(err)
	}
	dbHost, err := postgresContainer.Host(ctx)
	if err != nil {
		return postgresContainer, tracerr.Wrap(err)
	}
	dbPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return postgresContainer, tracerr.Wrap(err)
	}

	os.Setenv("DBNAME", dbName)
	os.Setenv("DBUSER", dbUser)
	os.Setenv("DBPASSWD", dbPassword)
	os.Setenv("DBHOST", dbHost)
	os.Setenv("DBPORT", dbPort.Port())
	if err := database.MigratePSQL(db); err != nil {
		return postgresContainer, tracerr.Wrap(err)
	}
	return postgresContainer, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	flag.Parse()
	postgresContainer, err := preparePSQL(ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Print(tracerr.Sprint(err))
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
		os.Exit(1)
	}
	exitVal := m.Run()

	os.Exit(exitVal)
}

type testServers struct {
	gormDB     *gorm.DB
	smtpServer *smtpmock.Server
}

func server(t *testing.T) (apiproto.ApiClient, *testServers, func(), error) {

	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	db, err := database.OpenPSQL()
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	gormDB, err := gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}

	baseServer, err := CreateServer(gormDB, "./resources")
	if err != nil {
		return nil, nil, nil, err
	}
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.NewClient("0.0.0.0",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}

	smtpServer := smtpmock.New(smtpmock.ConfigurationAttr{
		LogToStdout:       true,
		LogServerActivity: true,
	})
	if err := smtpServer.Start(); err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}

	closer := func() {
		if err := lis.Close(); err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
		if err := smtpServer.Stop(); err != nil {
			log.Printf("error closing smtpServer: %v", err)
		}
	}

	client := apiproto.NewApiClient(conn)
	t.Setenv("SMTPHOST", "127.0.0.1")
	t.Setenv("SMTPPORT", strconv.Itoa(smtpServer.PortNumber()))
	return client, &testServers{gormDB, smtpServer}, closer, nil
}

func AssertErrorInfo(t *testing.T, expected *apiproto.ErrorInfo, err error) bool {
	if assert.Error(t, err) {
		s := status.Convert(err)
		if assert.Greater(t, len(s.Details()), 0, "Status has an empty detail") {
			v, ok := s.Details()[0].(*apiproto.ErrorInfo)
			if assert.Equal(t, true, ok, "first detail is not ErrorInfo") {
				return assert.EqualExportedValues(t, expected, v)
			}
		}
	}
	return false
}

var tokenInMessageRegexp = regexp.MustCompile(`token=([^&]+)\s?`)
var emailInMessageRegexp = regexp.MustCompile(`email=([^&]+)`)

func purgeOneMessage(t *testing.T, testServersInfo *testServers) {
	msgs := testServersInfo.smtpServer.MessagesAndPurge()
	if !assert.Len(t, msgs, 1) {
		t.FailNow()
	}
}

func extractVerifyEmailInfo(t *testing.T, testServersInfo *testServers) (string, string) {
	msgs := testServersInfo.smtpServer.MessagesAndPurge()
	if !assert.Len(t, msgs, 1) {
		t.FailNow()
		return "", ""
	}
	tokens := tokenInMessageRegexp.FindStringSubmatch(msgs[0].MsgRequest())
	if !assert.Len(t, tokens, 2) {
		t.FailNow()
		return "", ""
	}
	emails := emailInMessageRegexp.FindStringSubmatch(msgs[0].MsgRequest())
	if !assert.Len(t, emails, 2) {
		t.FailNow()
		return "", ""
	}
	return strings.TrimSpace(tokens[1]), strings.TrimSpace(emails[1])
}

func verifyLastEmailSignup(t *testing.T, testServersInfo *testServers, ctx context.Context, client apiproto.ApiClient) string {
	tokenFromMail, emailFromMail := extractVerifyEmailInfo(t, testServersInfo)
	resp, err := client.VerifyEmail(ctx, &apiproto.VerifyEmailRequest{
		VerifyId: tokenFromMail,
		Email:    emailFromMail,
	})
	if !assert.NoError(t, err) {
		t.FailNow()
		return ""
	}
	if !(assert.True(t, resp.UserInfo.IsVerified) && assert.Equal(t, emailFromMail, *resp.UserInfo.VerifiedEmail)) {
		t.FailNow()
		return ""
	}
	return emailFromMail
}
