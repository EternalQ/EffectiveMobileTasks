package main_test

import (
	"testing"

	main "github.com/EternalQ/eff-mobile-tasks/task5"
	"github.com/stretchr/testify/assert"
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
	tests := []struct {
		name string

		email   string
		msg     string
		wantErr error
	}{
		{
			name:    "normal",
			email:   "email",
			msg:     "body",
			wantErr: nil,
		},
		{
			name:    "empty email",
			email:   "",
			msg:     "body",
			wantErr: main.ErrWrongInput,
		},
		{
			name:    "empty body",
			email:   "email",
			msg:     "",
			wantErr: main.ErrWrongInput,
		},
		{
			name:    "empty params",
			email:   "",
			msg:     "",
			wantErr: main.ErrWrongInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := new(MockSender)
			if tt.wantErr == nil {
				sender.On("Send", tt.email, tt.msg).Return(tt.wantErr)
			}

			s := main.NewNotificationService(sender)
			gotErr := s.Notify(tt.email, tt.msg)

			assert.Equal(t, tt.wantErr, gotErr)

			sender.AssertExpectations(t)
		})
	}
}
