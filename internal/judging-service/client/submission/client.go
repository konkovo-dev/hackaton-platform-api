package submission

import (
	"context"
	"fmt"

	submissionv1 "github.com/belikoooova/hackaton-platform-api/api/submission/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn              *grpc.ClientConn
	submissionService submissionv1.SubmissionServiceClient
	serviceToken      string
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.SubmissionServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial submission service: %w", err)
	}

	submissionService := submissionv1.NewSubmissionServiceClient(conn)

	return &Client{
		conn:              conn,
		submissionService: submissionService,
		serviceToken:      cfg.ServiceToken,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) ListFinalSubmissions(ctx context.Context, hackathonID string) ([]*submissionv1.Submission, error) {
	req := &submissionv1.ListSubmissionsRequest{
		HackathonId: hackathonID,
		Query:       nil,
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.submissionService.ListSubmissions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}

	finalSubmissions := make([]*submissionv1.Submission, 0)
	for _, sub := range resp.Submissions {
		if sub.IsFinal {
			finalSubmissions = append(finalSubmissions, sub)
		}
	}

	return finalSubmissions, nil
}

func (c *Client) GetSubmission(ctx context.Context, hackathonID, submissionID string) (*submissionv1.Submission, error) {
	req := &submissionv1.GetSubmissionRequest{
		HackathonId:  hackathonID,
		SubmissionId: submissionID,
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", c.serviceToken)
	if incomingMD, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := incomingMD.Get("authorization"); len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.submissionService.GetSubmission(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	return resp.Submission, nil
}
