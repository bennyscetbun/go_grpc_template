package apihelpers

import (
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/database/dbmodels"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/generated/rpc/apiproto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserDbModelToProto(user *dbmodels.User) *apiproto.UserInfo {
	return &apiproto.UserInfo{
		UserId:        user.ID,
		Username:      user.Username,
		VerifiedEmail: user.VerifiedEmail,
		NewEmail:      user.NewEmail,
		IsVerified:    user.IsVerified,
		CreatedAt:     timestamppb.New(user.CreatedAt),
		UpdatedAt:     timestamppb.New(user.UpdatedAt),
	}
}