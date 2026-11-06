package clients

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestProviderConfiguration(t *testing.T) {
	cases := map[string]struct {
		reason  string
		creds   string
		want    map[string]any
		wantErr bool
	}{
		"DefaultTimeout": {
			reason: "Without request_timeout, the provider gets our 60s default rather than its own 10s.",
			creds:  `{"server_url":"http://netbox:8080","api_token":"t"}`,
			want:   map[string]any{"server_url": "http://netbox:8080", "api_token": "t", "request_timeout": int64(60)},
		},
		"TimeoutAsNumber": {
			reason: "request_timeout can be a JSON number.",
			creds:  `{"server_url":"u","api_token":"t","request_timeout":120}`,
			want:   map[string]any{"server_url": "u", "api_token": "t", "request_timeout": int64(120)},
		},
		"TimeoutAsString": {
			reason: "request_timeout can be a numeric string, as Secret values often are.",
			creds:  `{"server_url":"u","api_token":"t","request_timeout":"30"}`,
			want:   map[string]any{"server_url": "u", "api_token": "t", "request_timeout": int64(30)},
		},
		"MissingServerURL": {
			reason:  "server_url is required.",
			creds:   `{"api_token":"t"}`,
			wantErr: true,
		},
		"EmptyAPIToken": {
			reason:  "api_token must not be empty.",
			creds:   `{"server_url":"u","api_token":""}`,
			wantErr: true,
		},
		"NonStringAPIToken": {
			reason:  "api_token must be a string.",
			creds:   `{"server_url":"u","api_token":42}`,
			wantErr: true,
		},
		"FractionalTimeout": {
			reason:  "request_timeout is a whole number of seconds.",
			creds:   `{"server_url":"u","api_token":"t","request_timeout":1.5}`,
			wantErr: true,
		},
		"ZeroTimeout": {
			reason:  "request_timeout must be positive.",
			creds:   `{"server_url":"u","api_token":"t","request_timeout":0}`,
			wantErr: true,
		},
		"NonNumericTimeout": {
			reason:  "A non-numeric request_timeout is rejected, not ignored.",
			creds:   `{"server_url":"u","api_token":"t","request_timeout":"1m"}`,
			wantErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			creds := map[string]any{}
			if err := json.Unmarshal([]byte(tc.creds), &creds); err != nil {
				t.Fatalf("invalid test input: %v", err)
			}
			got, err := providerConfiguration(creds)
			if tc.wantErr {
				if err == nil {
					t.Errorf("%s\nproviderConfiguration(...): want error, got %v", tc.reason, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("%s\nproviderConfiguration(...): unexpected error: %v", tc.reason, err)
			}
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("%s\nproviderConfiguration(...): want %v, got %v", tc.reason, tc.want, got)
			}
		})
	}
}
