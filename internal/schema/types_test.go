package schema

import (
	"strings"
	"testing"
)

func TestSchema_Validate(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid schema",
			schema: Schema{
				Version: "1",
				Groups: []Group{
					{Name: "app", Description: "App Settings"},
				},
				Vars: []Var{
					{Name: "PORT", Type: TypePort},
					{Name: "APP_ENV", Type: TypeEnum, AllowedValues: []string{"dev", "prod"}, Group: "app"},
					{Name: "REGEX_VAR", Type: TypeString, Pattern: "^[0-9]+$"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing or invalid version",
			schema: Schema{
				Version: "2",
			},
			wantErr: true,
			errMsg:  "invalid schema version",
		},
		{
			name: "missing var name",
			schema: Schema{
				Version: "1",
				Vars: []Var{
					{Type: TypeString},
				},
			},
			wantErr: true,
			errMsg:  "var has an empty name",
		},
		{
			name: "unknown type",
			schema: Schema{
				Version: "1",
				Vars: []Var{
					{Name: "DATABASE_URL", Type: "uri"},
				},
			},
			wantErr: true,
			errMsg:  "var 'DATABASE_URL': unknown type 'uri', supported types are: string, integer, float, boolean, url, email, port, enum, filepath, duration",
		},
		{
			name: "enum with no allowed_values",
			schema: Schema{
				Version: "1",
				Vars: []Var{
					{Name: "MODE", Type: TypeEnum},
				},
			},
			wantErr: true,
			errMsg:  "var 'MODE': type is 'enum' but allowed_values is empty",
		},
		{
			name: "duplicate var names",
			schema: Schema{
				Version: "1",
				Vars: []Var{
					{Name: "DUP_VAR", Type: TypeString},
					{Name: "DUP_VAR", Type: TypeInteger},
				},
			},
			wantErr: true,
			errMsg:  "duplicate var name found: 'DUP_VAR'",
		},
		{
			name: "invalid regex pattern",
			schema: Schema{
				Version: "1",
				Vars: []Var{
					{Name: "BAD_REGEX", Type: TypeString, Pattern: "["},
				},
			},
			wantErr: true,
			errMsg:  "var 'BAD_REGEX': invalid regex pattern",
		},
		{
			name: "undefined group",
			schema: Schema{
				Version: "1",
				Vars: []Var{
					{Name: "GROUP_VAR", Type: TypeString, Group: "missing"},
				},
			},
			wantErr: true,
			errMsg:  "var 'GROUP_VAR': references undefined group 'missing'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, expected message to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}
