package qemail

import (
	"context"
	"flag"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gopkg.qsoa.cloud/service/qemail/internal/emailpb"
)

var (
	sockAddr       = flag.String("q_email_sock", "", "Email socket")
	emailClient    emailpb.QEmailClient
	emailClientMtx sync.Mutex
)

// GetMailbox verifies that the email address exists and belongs to the
// service's project, then returns a Mailbox handle for sending and reading.
func GetMailbox(address string) (*Mailbox, error) {
	client, err := getEmailClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.ResolveMailbox(context.Background(), &emailpb.ResolveMailboxReq{
		Address: address,
	})
	if err != nil {
		return nil, fmt.Errorf("mailbox %s: %v", address, err)
	}

	return &Mailbox{mailboxID: resp.MailboxId, client: client}, nil
}

func getEmailClient() (emailpb.QEmailClient, error) {
	emailClientMtx.Lock()
	defer emailClientMtx.Unlock()

	if emailClient == nil {
		flag.Parse()

		cc, err := grpc.NewClient(*sockAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("cannot connect to email server: %v", err)
		}

		emailClient = emailpb.NewQEmailClient(cc)
	}

	return emailClient, nil
}
