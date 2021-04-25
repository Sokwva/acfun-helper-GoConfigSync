package main

import (
	"testing"
)

var testCookies string = ""

func Test_userAuth(t *testing.T) {
	type args struct {
		auhtInfo string
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
			if got := userAuth(tt.args.auhtInfo); got != tt.want {
				t.Errorf("userAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}
