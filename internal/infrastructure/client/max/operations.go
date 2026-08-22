package max

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"math/rand/v2"
	"time"
)

func (c *Client) SendText(ctx context.Context, chatID int64, text string) (string, error) {
	cid := time.Now().UnixMilli()*1000 + rand.Int64N(1000)

	payload, err := c.call(
		ctx, OpSendMessage, sendMessagePayload{
			ChatID: chatID,
			Message: outboundBody{
				Text:     text,
				CID:      cid,
				Elements: []any{},
			},
			Notify: true,
		},
	)
	if err != nil {
		return "", fmt.Errorf("send to chat %d: %w", chatID, err)
	}

	var resp sentMessageResponse

	if err = json.Unmarshal(payload, &resp); err != nil || resp.Message == nil {
		c.log.Debug("send response carries no message id", "chat_id", chatID)

		return "", nil
	}

	return resp.Message.ID, nil
}

func (c *Client) contacts(ctx context.Context, ids []int64) ([]contactEntry, error) {
	resp, err := c.callAs[contactsResponse](ctx, OpContactGet, contactGetPayload{ContactIDs: ids})
	if err != nil {
		return nil, err
	}

	if len(resp.Contacts) > 0 {
		return resp.Contacts, nil
	}

	return resp.Users, nil
}
