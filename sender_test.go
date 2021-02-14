package main

import (
	"reflect"
	"testing"
)

func TestTwillioSender_SendMessage(t *testing.T) {
	type fields struct {
		endpoint string
		user     string
		token    string
		from     string
	}
	type args struct {
		msgTo   string
		msgBody string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TwillioSender{
				endpoint: tt.fields.endpoint,
				user:     tt.fields.user,
				token:    tt.fields.token,
				from:     tt.fields.from,
			}
			if err := s.SendMessage(tt.args.msgTo, tt.args.msgBody); (err != nil) != tt.wantErr {
				t.Errorf("TwillioSender.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTwillioSender_SendMessageOnboarding(t *testing.T) {
	type fields struct {
		endpoint string
		user     string
		token    string
		from     string
	}
	type args struct {
		toPhone string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TwillioSender{
				endpoint: tt.fields.endpoint,
				user:     tt.fields.user,
				token:    tt.fields.token,
				from:     tt.fields.from,
			}
			if err := s.SendMessageOnboarding(tt.args.toPhone); (err != nil) != tt.wantErr {
				t.Errorf("TwillioSender.SendMessageOnboarding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTwillioSender_SendMessageReject(t *testing.T) {
	type fields struct {
		endpoint string
		user     string
		token    string
		from     string
	}
	type args struct {
		toPhone string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TwillioSender{
				endpoint: tt.fields.endpoint,
				user:     tt.fields.user,
				token:    tt.fields.token,
				from:     tt.fields.from,
			}
			if err := s.SendMessageReject(tt.args.toPhone); (err != nil) != tt.wantErr {
				t.Errorf("TwillioSender.SendMessageReject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTwillioSender_SendMessageDelete(t *testing.T) {
	type fields struct {
		endpoint string
		user     string
		token    string
		from     string
	}
	type args struct {
		toPhone string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TwillioSender{
				endpoint: tt.fields.endpoint,
				user:     tt.fields.user,
				token:    tt.fields.token,
				from:     tt.fields.from,
			}
			if err := s.SendMessageDelete(tt.args.toPhone); (err != nil) != tt.wantErr {
				t.Errorf("TwillioSender.SendMessageDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTwillioSender(t *testing.T) {
	type args struct {
		endpoint string
		user     string
		token    string
		from     string
	}
	tests := []struct {
		name string
		args args
		want *TwillioSender
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTwillioSender(tt.args.endpoint, tt.args.user, tt.args.token, tt.args.from); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTwillioSender() = %v, want %v", got, tt.want)
			}
		})
	}
}
