package domain

import participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusCanceled InvitationStatus = "canceled"
	InvitationStatusExpired  InvitationStatus = "expired"
)

func MapProtoInvitationStatusToDomain(protoStatus participationrolesv1.StaffInvitationStatus) InvitationStatus {
	switch protoStatus {
	case participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_PENDING:
		return InvitationStatusPending
	case participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_ACCEPTED:
		return InvitationStatusAccepted
	case participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_DECLINED:
		return InvitationStatusDeclined
	case participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_CANCELED:
		return InvitationStatusCanceled
	case participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_EXPIRED:
		return InvitationStatusExpired
	default:
		return InvitationStatusPending
	}
}

func MapDomainInvitationStatusToProto(status InvitationStatus) participationrolesv1.StaffInvitationStatus {
	switch status {
	case InvitationStatusPending:
		return participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_PENDING
	case InvitationStatusAccepted:
		return participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_ACCEPTED
	case InvitationStatusDeclined:
		return participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_DECLINED
	case InvitationStatusCanceled:
		return participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_CANCELED
	case InvitationStatusExpired:
		return participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_EXPIRED
	default:
		return participationrolesv1.StaffInvitationStatus_STAFF_INVITATION_STATUS_UNSPECIFIED
	}
}
