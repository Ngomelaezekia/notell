package main

import "testing"

func TestValidTransition(t *testing.T) {
    cases := []struct{ from, to string; want bool }{
        {PaymentPending, PaymentProcessing, true},
        {PaymentPending, PaymentSucceeded, false},
        {PaymentProcessing, PaymentSucceeded, true},
        {PaymentProcessing, PaymentFailed, true},
        {PaymentSucceeded, PaymentFailed, false},
        {PaymentFailed, PaymentProcessing, false},
    }
    for _, tc := range cases {
        if got := validTransition(tc.from, tc.to); got != tc.want { t.Fatalf("transition %s -> %s: got %v want %v", tc.from, tc.to, got, tc.want) }
    }
}

func TestNewID(t *testing.T) {
    a, b := newID("pay"), newID("pay")
    if a == b || len(a) < 20 || len(b) < 20 { t.Fatal("expected unique non-empty ids") }
}
