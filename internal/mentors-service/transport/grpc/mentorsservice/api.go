package mentorsservice

import (
	"context"
	"errors"
	"log/slog"

	mentorsv1 "github.com/belikoooova/hackaton-platform-api/api/mentors/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/usecase/mentors"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type API struct {
	mentorsv1.UnimplementedMentorsServiceServer
	service *mentors.Service
	logger  *slog.Logger
}

var _ mentorsv1.MentorsServiceServer = (*API)(nil)

func NewAPI(service *mentors.Service, logger *slog.Logger) *API {
	return &API{
		service: service,
		logger:  logger,
	}
}

func (a *API) GetMyTickets(ctx context.Context, req *mentorsv1.GetMyTicketsRequest) (*mentorsv1.GetMyTicketsResponse, error) {
	limit := int32(20)
	offset := int32(0)

	if req.Query != nil {
		if req.Query.Limit > 0 {
			limit = req.Query.Limit
		}
		if req.Query.Offset > 0 {
			offset = req.Query.Offset
		}
	}

	output, err := a.service.GetMyTickets(ctx, mentors.GetMyTicketsIn{
		HackathonID: req.HackathonId,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "GetMyTickets")
	}

	tickets := make([]*mentorsv1.Ticket, 0, len(output.Tickets))
	for _, ticket := range output.Tickets {
		tickets = append(tickets, mapTicketToProto(ticket))
	}

	return &mentorsv1.GetMyTicketsResponse{
		Tickets: tickets,
		HasMore: output.HasMore,
	}, nil
}

func (a *API) GetTicketMessages(ctx context.Context, req *mentorsv1.GetTicketMessagesRequest) (*mentorsv1.GetTicketMessagesResponse, error) {
	limit := int32(50)
	offset := int32(0)

	if req.Query != nil {
		if req.Query.GetLimit() > 0 {
			limit = req.Query.GetLimit()
		}
		if req.Query.GetOffset() > 0 {
			offset = req.Query.GetOffset()
		}
	}

	output, err := a.service.GetTicketMessages(ctx, mentors.GetTicketMessagesIn{
		HackathonID: req.HackathonId,
		TicketID:    req.TicketId,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "GetTicketMessages")
	}

	messages := make([]*mentorsv1.Message, 0, len(output.Messages))
	for _, message := range output.Messages {
		messages = append(messages, mapMessageToProto(message))
	}

	return &mentorsv1.GetTicketMessagesResponse{
		Messages: messages,
		HasMore:  output.HasMore,
	}, nil
}

func (a *API) ListAssignedTickets(ctx context.Context, req *mentorsv1.ListAssignedTicketsRequest) (*mentorsv1.ListAssignedTicketsResponse, error) {
	limit := int32(20)
	offset := int32(0)

	if req.Query != nil {
		if req.Query.GetLimit() > 0 {
			limit = req.Query.GetLimit()
		}
		if req.Query.GetOffset() > 0 {
			offset = req.Query.GetOffset()
		}
	}

	output, err := a.service.ListAssignedTickets(ctx, mentors.ListAssignedTicketsIn{
		HackathonID: req.HackathonId,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "ListAssignedTickets")
	}

	tickets := make([]*mentorsv1.Ticket, 0, len(output.Tickets))
	for _, ticket := range output.Tickets {
		tickets = append(tickets, mapTicketToProto(ticket))
	}

	return &mentorsv1.ListAssignedTicketsResponse{
		Tickets: tickets,
		HasMore: output.HasMore,
	}, nil
}

func (a *API) ListAllTickets(ctx context.Context, req *mentorsv1.ListAllTicketsRequest) (*mentorsv1.ListAllTicketsResponse, error) {
	limit := int32(20)
	offset := int32(0)

	if req.Query != nil {
		if req.Query.GetLimit() > 0 {
			limit = req.Query.GetLimit()
		}
		if req.Query.GetOffset() > 0 {
			offset = req.Query.GetOffset()
		}
	}

	output, err := a.service.ListAllTickets(ctx, mentors.ListAllTicketsIn{
		HackathonID: req.HackathonId,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "ListAllTickets")
	}

	tickets := make([]*mentorsv1.Ticket, 0, len(output.Tickets))
	for _, ticket := range output.Tickets {
		tickets = append(tickets, mapTicketToProto(ticket))
	}

	return &mentorsv1.ListAllTicketsResponse{
		Tickets: tickets,
		HasMore: output.HasMore,
	}, nil
}

func (a *API) SendMessage(ctx context.Context, req *mentorsv1.SendMessageRequest) (*mentorsv1.SendMessageResponse, error) {
	idempotencyKey := ""
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	output, err := a.service.SendMessage(ctx, mentors.SendMessageIn{
		HackathonID:     req.HackathonId,
		Text:            req.Text,
		ClientMessageID: req.ClientMessageId,
		IdempotencyKey:  idempotencyKey,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "SendMessage")
	}

	return &mentorsv1.SendMessageResponse{
		MessageId: output.MessageID,
		TicketId:  output.TicketID,
	}, nil
}

func (a *API) ReplyInTicket(ctx context.Context, req *mentorsv1.ReplyInTicketRequest) (*mentorsv1.ReplyInTicketResponse, error) {
	idempotencyKey := ""
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	output, err := a.service.ReplyInTicket(ctx, mentors.ReplyInTicketIn{
		HackathonID:    req.HackathonId,
		TicketID:       req.TicketId,
		Text:           req.Text,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "ReplyInTicket")
	}

	return &mentorsv1.ReplyInTicketResponse{
		MessageId: output.MessageID,
	}, nil
}

func (a *API) CloseTicket(ctx context.Context, req *mentorsv1.CloseTicketRequest) (*mentorsv1.CloseTicketResponse, error) {
	idempotencyKey := ""
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	_, err := a.service.CloseTicket(ctx, mentors.CloseTicketIn{
		HackathonID:    req.HackathonId,
		TicketID:       req.TicketId,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "CloseTicket")
	}

	return &mentorsv1.CloseTicketResponse{}, nil
}

func (a *API) handleError(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	var policyErr *policy.PolicyError
	if errors.As(err, &policyErr) {
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return policyErr.ToGRPCError()
	}

	switch {
	case errors.Is(err, mentors.ErrUnauthorized):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, mentors.ErrForbidden):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, mentors.ErrNotFound):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, mentors.ErrConflict):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, mentors.ErrInvalidInput):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		a.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}

func mapTicketToProto(ticket *entity.Ticket) *mentorsv1.Ticket {
	proto := &mentorsv1.Ticket{
		TicketId:    ticket.ID.String(),
		HackathonId: ticket.HackathonID.String(),
		Status:      mapTicketStatusToProto(ticket.Status),
		OwnerKind:   mapOwnerKindToProto(ticket.OwnerKind),
		OwnerId:     ticket.OwnerID.String(),
		CreatedAt:   timestamppb.New(ticket.CreatedAt),
		UpdatedAt:   timestamppb.New(ticket.UpdatedAt),
	}

	if ticket.AssignedMentorUserID != nil {
		proto.AssignedMentorUserId = ticket.AssignedMentorUserID.String()
	}

	if ticket.ClosedAt != nil {
		proto.ClosedAt = timestamppb.New(*ticket.ClosedAt)
	}

	return proto
}

func mapMessageToProto(message *entity.Message) *mentorsv1.Message {
	return &mentorsv1.Message{
		MessageId:    message.ID.String(),
		TicketId:     message.TicketID.String(),
		AuthorUserId: message.AuthorUserID.String(),
		AuthorRole:   mapAuthorRoleToProto(message.AuthorRole),
		Text:         message.Text,
		CreatedAt:    timestamppb.New(message.CreatedAt),
	}
}

func mapTicketStatusToProto(status string) mentorsv1.TicketStatus {
	switch status {
	case domain.TicketStatusOpen:
		return mentorsv1.TicketStatus_TICKET_STATUS_OPEN
	case domain.TicketStatusClosed:
		return mentorsv1.TicketStatus_TICKET_STATUS_CLOSED
	default:
		return mentorsv1.TicketStatus_TICKET_STATUS_UNSPECIFIED
	}
}

func mapOwnerKindToProto(kind string) mentorsv1.OwnerKind {
	switch kind {
	case domain.OwnerKindUser:
		return mentorsv1.OwnerKind_OWNER_KIND_USER
	case domain.OwnerKindTeam:
		return mentorsv1.OwnerKind_OWNER_KIND_TEAM
	default:
		return mentorsv1.OwnerKind_OWNER_KIND_UNSPECIFIED
	}
}

func mapAuthorRoleToProto(role string) mentorsv1.AuthorRole {
	switch role {
	case domain.AuthorRoleParticipant:
		return mentorsv1.AuthorRole_AUTHOR_ROLE_PARTICIPANT
	case domain.AuthorRoleMentor:
		return mentorsv1.AuthorRole_AUTHOR_ROLE_MENTOR
	case domain.AuthorRoleOrganizer:
		return mentorsv1.AuthorRole_AUTHOR_ROLE_ORGANIZER
	case domain.AuthorRoleSystem:
		return mentorsv1.AuthorRole_AUTHOR_ROLE_SYSTEM
	default:
		return mentorsv1.AuthorRole_AUTHOR_ROLE_UNSPECIFIED
	}
}

func (a *API) ClaimTicket(ctx context.Context, req *mentorsv1.ClaimTicketRequest) (*mentorsv1.ClaimTicketResponse, error) {
	idempotencyKey := ""
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	out, err := a.service.ClaimTicket(ctx, mentors.ClaimTicketIn{
		HackathonID:    req.HackathonId,
		TicketID:       req.TicketId,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "ClaimTicket")
	}

	return &mentorsv1.ClaimTicketResponse{
		TicketId:             out.TicketID,
		AssignedMentorUserId: out.AssignedMentorUserID,
		AssignedAt:           timestamppb.New(out.AssignedAt),
	}, nil
}

func (a *API) AssignTicket(ctx context.Context, req *mentorsv1.AssignTicketRequest) (*mentorsv1.AssignTicketResponse, error) {
	idempotencyKey := ""
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	out, err := a.service.AssignTicket(ctx, mentors.AssignTicketIn{
		HackathonID:    req.HackathonId,
		TicketID:       req.TicketId,
		MentorUserID:   req.MentorUserId,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "AssignTicket")
	}

	return &mentorsv1.AssignTicketResponse{
		TicketId:             out.TicketID,
		AssignedMentorUserId: out.AssignedMentorUserID,
		AssignedAt:           timestamppb.New(out.AssignedAt),
	}, nil
}

func (a *API) GetRealtimeToken(ctx context.Context, req *mentorsv1.GetRealtimeTokenRequest) (*mentorsv1.GetRealtimeTokenResponse, error) {
	out, err := a.service.GetRealtimeToken(ctx, mentors.GetRealtimeTokenIn{
		HackathonID: req.HackathonId,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "GetRealtimeToken")
	}

	return &mentorsv1.GetRealtimeTokenResponse{
		Token:     out.Token,
		ExpiresAt: out.ExpiresAt,
	}, nil
}
