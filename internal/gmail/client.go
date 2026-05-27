package gmail

import (
	"context"
	"errors"
	"strings"

	gmailv1 "google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
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
	items, _, err := c.ListMessagesPage(ctx, maxResults, query, "")
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) ListMessagesPage(ctx context.Context, maxResults int64, query, pageToken string) ([]Message, string, error) {
	call := c.service.Users.Messages.List(c.userID).Context(ctx).MaxResults(maxResults)
	if query != "" {
		call = call.Q(query)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	listResponse, err := call.Do()
	if err != nil {
		return nil, "", err
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
			return nil, "", err
		}

		result = append(result, Message{
			ID:       messageResponse.Id,
			ThreadID: messageResponse.ThreadId,
			From:     headerValue(messageResponse, "From"),
			Subject:  headerValue(messageResponse, "Subject"),
			Snippet:  messageResponse.Snippet,
		})
	}

	return result, listResponse.NextPageToken, nil
}

func (c *Client) EnsureLabel(ctx context.Context, labelName string) (string, error) {
	labelsResponse, err := c.service.Users.Labels.List(c.userID).Context(ctx).Do()
	if err != nil {
		return "", err
	}

	for _, label := range labelsResponse.Labels {
		if label == nil {
			continue
		}
		if strings.EqualFold(label.Name, labelName) {
			return label.Id, nil
		}
	}

	created, err := c.service.Users.Labels.Create(c.userID, &gmailv1.Label{
		Name:                  labelName,
		LabelListVisibility:   "labelShow",
		MessageListVisibility: "show",
	}).Context(ctx).Do()
	if err != nil {
		return "", err
	}

	return created.Id, nil
}

func (c *Client) ApplyLabelToMessage(ctx context.Context, messageID, labelID string, markRead bool) error {
	removeLabelIDs := []string{"INBOX"}
	if markRead {
		removeLabelIDs = append(removeLabelIDs, "UNREAD")
	}

	_, err := c.service.Users.Messages.Modify(c.userID, messageID, &gmailv1.ModifyMessageRequest{
		AddLabelIds:    []string{labelID},
		RemoveLabelIds: removeLabelIDs,
	}).Context(ctx).Do()
	return err
}

func IsPermanentError(err error) bool {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.Code >= 400 && apiErr.Code < 500 && apiErr.Code != 429
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
