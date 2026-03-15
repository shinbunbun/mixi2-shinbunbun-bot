package mixi2

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mixigroup/mixi2-application-sdk-go/auth"
	application_apiv1 "github.com/mixigroup/mixi2-application-sdk-go/gen/go/social/mixi/application/service/application_api/v1"
)

type Client struct {
	apiClient     application_apiv1.ApplicationServiceClient
	authenticator auth.Authenticator
	logger        *slog.Logger
}

func NewClient(apiClient application_apiv1.ApplicationServiceClient, authenticator auth.Authenticator) *Client {
	return &Client{
		apiClient:     apiClient,
		authenticator: authenticator,
		logger:        slog.Default(),
	}
}

func (c *Client) CreatePost(ctx context.Context, text string) error {
	authCtx, err := c.authenticator.AuthorizedContext(ctx)
	if err != nil {
		return fmt.Errorf("getting authorized context: %w", err)
	}

	resp, err := c.apiClient.CreatePost(authCtx, &application_apiv1.CreatePostRequest{
		Text: text,
	})
	if err != nil {
		return fmt.Errorf("creating post: %w", err)
	}

	postID := resp.GetPost().GetPostId()
	c.logger.Info("post created", slog.String("post_id", postID))
	return nil
}
