package main_test

import (
	"errors"
	"testing"

	main "github.com/EternalQ/eff-mobile-tasks/task5"
	"github.com/stretchr/testify/mock"
)

type MockSender struct {
	mock.Mock
}

func (m *MockSender) Send(to, body string) error {
	args := m.Called(to, body)
	return args.Error(0)
}

func TestNotificationService_Notify(t *testing.T) {
	sender := new(MockSender)
	sender.On("Send", "email", "body").Return(nil)
	sender.On("Send", "", "body").Return(errors.New("wrong email"))
	sender.On("Send", "email", "").Return(errors.New("empty body"))
	sender.On("Send", "", "").Return(errors.New("empty params"))

	tests := []struct {
		name string

		email   string
		msg     string
		wantErr bool
	}{
		{
			name:    "normal",
			email:   "email",
			msg:     "body",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			msg:     "body",
			wantErr: true,
		},
		{
			name:    "empty body",
			email:   "email",
			msg:     "",
			wantErr: true,
		},
		{
			name:    "empty params",
			email:   "",
			msg:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := main.NewNotificationSercice(sender)
			gotErr := s.Notify(tt.email, tt.msg)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Notify() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Notify() succeeded unexpectedly")
			}
		})
	}
}
