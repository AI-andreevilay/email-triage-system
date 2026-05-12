package gmail

import (
	"context"
	"strings"

	gmailv1 "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Client struct {
	service *gmailv1.Service
	userID  string
}

func NewClient(ctx context.Context, credentialsFile, tokenFile, userID string) (*Client, error) {
	httpClient, err := AuthenticatedHTTPClient(ctx, credentialsFile, tokenFile)
	if err != nil {
		return nil, err
	}

	service, err := gmailv1.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	if userID == "" {
		userID = "me"
	}

	return &Client{
		service: service,
		userID:  userID,
	}, nil
}

func (c *Client) ListMessages(ctx context.Context, maxResults int64, query string) ([]Message, error) {
	call := c.service.Users.Messages.List(c.userID).Context(ctx).MaxResults(maxResults)
	if query != "" {
		call = call.Q(query)
	}

	listResponse, err := call.Do()
	if err != nil {
		return nil, err
	}

	result := make([]Message, 0, len(listResponse.Messages))
	for _, item := range listResponse.Messages {
		if item == nil || item.Id == "" {
			continue
		}

		messageResponse, err := c.service.Users.Messages.Get(c.userID, item.Id).
			Context(ctx).
			Format("metadata").
			MetadataHeaders("From", "Subject").
			Do()
		if err != nil {
			return nil, err
		}

		result = append(result, Message{
			ID:       messageResponse.Id,
			ThreadID: messageResponse.ThreadId,
			From:     headerValue(messageResponse, "From"),
			Subject:  headerValue(messageResponse, "Subject"),
			Snippet:  messageResponse.Snippet,
		})
	}

	return result, nil
}

func headerValue(message *gmailv1.Message, key string) string {
	if message == nil || message.Payload == nil {
		return ""
	}
	for _, header := range message.Payload.Headers {
		if header == nil {
			continue
		}
		if strings.EqualFold(header.Name, key) {
			return header.Value
		}
	}
	return ""
}
