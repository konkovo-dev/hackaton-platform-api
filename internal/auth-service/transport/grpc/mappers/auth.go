package mappers

import (
	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func AuthOutToRegisterResponse(out *auth.AuthOut) *authv1.RegisterResponse {
	return &authv1.RegisterResponse{
		AccessToken:      out.AccessToken,
		RefreshToken:     out.RefreshToken,
		AccessExpiresAt:  timestamppb.New(out.AccessExpiresAt),
		RefreshExpiresAt: timestamppb.New(out.RefreshExpiresAt),
	}
}

func AuthOutToLoginResponse(out *auth.AuthOut) *authv1.LoginResponse {
	return &authv1.LoginResponse{
		AccessToken:      out.AccessToken,
		RefreshToken:     out.RefreshToken,
		AccessExpiresAt:  timestamppb.New(out.AccessExpiresAt),
		RefreshExpiresAt: timestamppb.New(out.RefreshExpiresAt),
	}
}

func AuthOutToRefreshResponse(out *auth.AuthOut) *authv1.RefreshResponse {
	return &authv1.RefreshResponse{
		AccessToken:      out.AccessToken,
		RefreshToken:     out.RefreshToken,
		AccessExpiresAt:  timestamppb.New(out.AccessExpiresAt),
		RefreshExpiresAt: timestamppb.New(out.RefreshExpiresAt),
	}
}

func IntrospectTokenOutToResponse(out *auth.IntrospectTokenOut) *authv1.IntrospectTokenResponse {
	return &authv1.IntrospectTokenResponse{
		Active:    out.Active,
		UserId:    out.UserID.String(),
		ExpiresAt: timestamppb.New(out.ExpiresAt),
	}
}
