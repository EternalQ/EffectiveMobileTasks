package main

type Sender interface {
	Send(to, body string) error
}

type NotificationService struct {
	emailSender Sender
}

func NewNotificationSercice(sender Sender) *NotificationService {
	return &NotificationService{sender}
}

func (s *NotificationService) Notify(email, msg string) error {
	return s.emailSender.Send(email, msg)
}
