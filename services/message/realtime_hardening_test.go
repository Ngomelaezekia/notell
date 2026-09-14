package main

import "testing"

func TestValidMessageBody(t *testing.T) {
    if err := validMessageBody("hello"); err != nil { t.Fatal(err) }
    if validMessageBody("   ") == nil { t.Fatal("expected empty body rejection") }
    if validMessageBody(string(make([]byte, messageMaxBytes+1))) == nil { t.Fatal("expected oversized body rejection") }
}
