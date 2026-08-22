package max

import "context"

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
