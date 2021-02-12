package main

import (
	"net/url"
	"reflect"
	"testing"
	"time"
)

func Test_todayAt(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := todayAt(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("todayAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("todayAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCall(t *testing.T) {
	type args struct {
		data url.Values
	}
	tests := []struct {
		name    string
		args    args
		want    Call
		want1   []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := NewCall(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCall() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewCall() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
