package main

type Sender interface {
	Send(to, body string) error
}

type NotificationService struct {
	emailSender Sender
}

func NewNotificationService(sender Sender) *NotificationService {
	return &NotificationService{sender}
}

func (s *NotificationService) Notify(email, msg string) error {
	if len(email) == 0 || len(msg) == 0 {
		return ErrWrongInput
	}
	return s.emailSender.Send(email, msg)
}
