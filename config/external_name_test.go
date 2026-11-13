package config

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestPrimaryMACAddressIsNotFoundDiagnostic(t *testing.T) {
	isNotFound := primaryMACAddressExternalName().IsNotFoundDiagnosticFn
	diag := func(severity tfprotov6.DiagnosticSeverity, summary, detail string) []*tfprotov6.Diagnostic {
		return []*tfprotov6.Diagnostic{{Severity: severity, Summary: summary, Detail: detail}}
	}
	cases := map[string]struct {
		diags []*tfprotov6.Diagnostic
		want  bool
	}{
		"NotCreatedYet": {diags: diag(tfprotov6.DiagnosticSeverityError, "Invalid id", `expected an interface id, got ""`), want: true},
		"InvalidID":     {diags: diag(tfprotov6.DiagnosticSeverityError, "Invalid id", `expected an interface id, got "eth0"`), want: false},
		"OtherError":    {diags: diag(tfprotov6.DiagnosticSeverityError, "Error reading interface", "500"), want: false},
		"Warning":       {diags: diag(tfprotov6.DiagnosticSeverityWarning, "Invalid id", `expected an interface id, got ""`), want: false},
		"NoDiagnostics": {diags: nil, want: false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := isNotFound(tc.diags); got != tc.want {
				t.Errorf("IsNotFoundDiagnosticFn() = %v, want %v", got, tc.want)
			}
		})
	}
}
