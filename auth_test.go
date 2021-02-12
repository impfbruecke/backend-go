package main

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/sessions"
)

func Test_authenticateUser(t *testing.T) {
	type args struct {
		user string
		pass string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := authenticateUser(tt.args.user, tt.args.pass); got != tt.want {
				t.Errorf("authenticateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHandler(tt.args.w, tt.args.r)
		})
	}
}

func Test_loginHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginHandler(tt.args.w, tt.args.r)
		})
	}
}

func Test_middlewareAuth(t *testing.T) {
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := middlewareAuth(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("middlewareAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_forbiddenHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			forbiddenHandler(tt.args.w, tt.args.r)
		})
	}
}

func Test_getUser(t *testing.T) {
	type args struct {
		s *sessions.Session
	}
	tests := []struct {
		name string
		args args
		want User
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUser(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
