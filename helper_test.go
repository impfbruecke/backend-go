package main

import (
	"html/template"
	"net/http"
	"reflect"
	"testing"
)

func Test_parseTemplates(t *testing.T) {
	tests := []struct {
		name string
		want *template.Template
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTemplates(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_contextString(t *testing.T) {
	type args struct {
		key contextKey
		r   *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contextString(tt.args.key, tt.args.r); got != tt.want {
				t.Errorf("contextString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genOTP(t *testing.T) {
	type args struct {
		phone  string
		callID int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genOTP(tt.args.phone, tt.args.callID); got != tt.want {
				t.Errorf("genOTP() = %v, want %v", got, tt.want)
			}
		})
	}
}
