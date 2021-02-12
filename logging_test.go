package main

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_logRequest(t *testing.T) {
	type args struct {
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
			logRequest(tt.args.r)
		})
	}
}

func Test_middlewareLog(t *testing.T) {
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
			if got := middlewareLog(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("middlewareLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
