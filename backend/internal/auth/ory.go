package auth

import (
	"context"
	"fmt"
	"strings"

	ory "github.com/ory/client-go"
)

type Service struct {
	client *ory.APIClient
}

func NewService(client *ory.APIClient) *Service {
	return &Service{
		client: client,
	}
}

func NewClient(oryURL string) *ory.APIClient {
	cfg := ory.NewConfiguration()
	cfg.Servers = []ory.ServerConfiguration{
		{
			URL: oryURL,
		},
	}
	return ory.NewAPIClient(cfg)
}

func (s *Service) ValidateSession(ctx context.Context, credentials SessionCredentials) (*ory.Session, error) {
	request := s.client.FrontendAPI.ToSession(ctx)

	switch {
	case strings.TrimSpace(credentials.SessionToken) != "":
		request = request.XSessionToken(credentials.SessionToken)
	case strings.TrimSpace(credentials.CookieHeader) != "":
		request = request.Cookie(credentials.CookieHeader)
	default:
		return nil, fmt.Errorf("missing session credentials")
	}

	session, _, err := request.Execute()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) GetUserID(session *ory.Session) (string, error) {
	if session == nil || session.Identity == nil {
		return "", fmt.Errorf("invalid session")
	}
	return session.Identity.Id, nil
}
