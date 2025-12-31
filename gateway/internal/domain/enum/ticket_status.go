package enum

type TicketStatus string

const (
	TicketStatusPending  TicketStatus = "pending"
	TicketStatusAnswered TicketStatus = "answered"
	TicketStatusClosed   TicketStatus = "closed"
)

type TicketService string

const (
	TicketServiceOCR      TicketService = "ocr"
	TicketServiceFaceAuth TicketService = "face_auth"
	TicketServiceLiveness TicketService = "liveness"
	TicketServiceOther    TicketService = "other"
)

// ServiceDisplayName returns the Persian display name for the service
func (s TicketService) DisplayName() string {
	switch s {
	case TicketServiceOCR:
		return "OCR مدارک"
	case TicketServiceFaceAuth:
		return "احراز هویت چهره"
	case TicketServiceLiveness:
		return "تشخیص زنده بودن"
	case TicketServiceOther:
		return "سایر"
	default:
		return string(s)
	}
}

type MessageSender string

const (
	MessageSenderUser  MessageSender = "user"
	MessageSenderAdmin MessageSender = "admin"
)
