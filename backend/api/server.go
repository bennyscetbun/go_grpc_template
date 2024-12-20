package api

import (
	"html/template"
	"path/filepath"

	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/database/dbqueries"
	"github.com/bennyscetbun/xxxyourappyyy/backend/generated/rpc/apiproto"
	"github.com/ztrue/tracerr"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type GRPCServer struct {
	apiproto.UnimplementedApiServer
	DB                    *gorm.DB
	DBQueries             *dbqueries.Query
	ResourceDirectoryPath string
	Templates             *template.Template
	TokenSecret           []byte
}

var _ apiproto.ApiServer = (*GRPCServer)(nil)

func CreateServer(gormDB *gorm.DB, ResourceDirectoryPath string, tokenSecret []byte) (*grpc.Server, error) {
	tmpl, err := template.ParseGlob(filepath.Join(ResourceDirectoryPath, "templates", "*.tmpl.html"))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	server := &GRPCServer{
		DB:                    gormDB,
		DBQueries:             dbqueries.Use(gormDB),
		ResourceDirectoryPath: ResourceDirectoryPath,
		Templates:             tmpl,
		TokenSecret:           tokenSecret,
	}
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.TimeOutInterceptor,
			server.AuthInterceptor,
		),
	)
	apiproto.RegisterApiServer(grpcServer, server)

	return grpcServer, nil
}
