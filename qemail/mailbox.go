package qemail

import (
	"context"
	"fmt"
	"time"

	"gopkg.qsoa.cloud/service/qemail/internal/emailpb"
)

// Mailbox provides operations on a specific email mailbox.
// The mailboxID is an opaque token issued by the runner during GetMailbox.
type Mailbox struct {
	mailboxID uint64
	client    emailpb.QEmailClient
}

// Send composes and sends an email from this mailbox.
func (m *Mailbox) Send(ctx context.Context, msg Message) (string, error) {
	attachments := make([]*emailpb.Attachment, len(msg.Attachments))
	for i, a := range msg.Attachments {
		attachments[i] = &emailpb.Attachment{
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Data:        a.Data,
		}
	}

	headers := make([]*emailpb.Header, len(msg.Headers))
	for i, h := range msg.Headers {
		headers[i] = &emailpb.Header{Key: h.Key, Value: h.Value}
	}

	resp, err := m.client.SendEmail(ctx, &emailpb.SendEmailReq{
		MailboxId:   m.mailboxID,
		To:          msg.To,
		Cc:          msg.Cc,
		Bcc:         msg.Bcc,
		Subject:     msg.Subject,
		TextBody:    msg.TextBody,
		HtmlBody:    msg.HtmlBody,
		Attachments: attachments,
		Headers:     headers,
		FromName:    msg.FromName,
	})
	if err != nil {
		return "", fmt.Errorf("cannot send email: %v", err)
	}

	return resp.MessageId, nil
}

// ListMessages returns a paginated list of message summaries for a folder.
func (m *Mailbox) ListMessages(ctx context.Context, folder string, offset, limit uint32) ([]MessageSummary, uint32, error) {
	resp, err := m.client.ListMessages(ctx, &emailpb.ListMessagesReq{
		MailboxId: m.mailboxID,
		Folder:    folder,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("cannot list messages: %v", err)
	}

	summaries := make([]MessageSummary, len(resp.Messages))
	for i, msg := range resp.Messages {
		summaries[i] = MessageSummary{
			UID:     msg.Uid,
			From:    msg.From,
			To:      msg.To,
			Subject: msg.Subject,
			Date:    time.Unix(msg.Date, 0),
			Seen:    msg.Seen,
			Flagged: msg.Flagged,
			Size:    msg.Size_,
		}
	}

	return summaries, resp.Total, nil
}

// GetMessage retrieves and parses a single message by UID.
func (m *Mailbox) GetMessage(ctx context.Context, uid, folder string) (*FullMessage, error) {
	resp, err := m.client.GetMessage(ctx, &emailpb.GetMessageReq{
		MailboxId: m.mailboxID,
		Uid:       uid,
		Folder:    folder,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get message: %v", err)
	}

	attachments := make([]Attachment, len(resp.Attachments))
	for i, a := range resp.Attachments {
		attachments[i] = Attachment{
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Data:        a.Data,
		}
	}

	return &FullMessage{
		From:        resp.From,
		To:          resp.To,
		Cc:          resp.Cc,
		Subject:     resp.Subject,
		Date:        time.Unix(resp.Date, 0),
		TextBody:    resp.TextBody,
		HtmlBody:    resp.HtmlBody,
		RawHeaders:  resp.RawHeaders,
		Attachments: attachments,
	}, nil
}

// DeleteMessage removes a message from the mailbox.
func (m *Mailbox) DeleteMessage(ctx context.Context, uid, folder string) error {
	_, err := m.client.DeleteMessage(ctx, &emailpb.DeleteMessageReq{
		MailboxId: m.mailboxID,
		Uid:       uid,
		Folder:    folder,
	})
	if err != nil {
		return fmt.Errorf("cannot delete message: %v", err)
	}
	return nil
}

// MoveMessage moves a message to a different folder.
func (m *Mailbox) MoveMessage(ctx context.Context, uid, toFolder string) error {
	_, err := m.client.MoveMessage(ctx, &emailpb.MoveMessageReq{
		MailboxId: m.mailboxID,
		Uid:       uid,
		ToFolder:  toFolder,
	})
	if err != nil {
		return fmt.Errorf("cannot move message: %v", err)
	}
	return nil
}
