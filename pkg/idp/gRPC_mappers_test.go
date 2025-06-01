// Copyright 2025 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package idp

import (
	"encoding/json"
	v0Idps "github.com/canonical/identity-platform-api/v0/idps"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
)

//go:generate mockgen -build_flags=--mod=mod -package idp -destination ./mock_logger.go -source=../../internal/logging/interfaces.go

func TestGrpcPbMapper_FromConfigurations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	requestedClaims := `{"user_info": {"given_name": {"essential": true}}}`
	rc, _ := json.Marshal(requestedClaims)

	tests := []struct {
		name  string
		input []*Configuration
		want  []*v0Idps.Idp
	}{
		{
			name: "Successful mapping from Configurations to Idps",
			input: []*Configuration{
				{
					ID:              "microsoft_af675f353bd7451588e2b8032e315f6f",
					Provider:        "microsoft",
					Label:           "microsoft",
					ClientID:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
					ClientSecret:    "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
					IssuerURL:       "",
					AuthURL:         "",
					TokenURL:        "",
					Tenant:          "e1574293-28de-4e94-87d5-b61c76fc14e1",
					SubjectSource:   "",
					TeamId:          "",
					PrivateKeyId:    "",
					PrivateKey:      "",
					Scope:           []string{"profile", "email", "address", "phone"},
					Mapper:          "file:///etc/config/kratos/microsoft_schema.jsonnet",
					RequestedClaims: rc,
				},
			},
			want: []*v0Idps.Idp{
				{
					Id:                "microsoft_af675f353bd7451588e2b8032e315f6f",
					Provider:          "microsoft",
					Label:             strPtr("microsoft"),
					ClientId:          "af675f35-3bd7-4515-88e2-b8032e315f6f",
					ClientSecret:      strPtr("3y38Q~aslkdhaskjhd~W0xWDB.123u98asd"),
					IssuerUrl:         strPtr(""),
					AuthUrl:           strPtr(""),
					TokenUrl:          strPtr(""),
					MicrosoftTenant:   strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
					SubjectSource:     strPtr(""),
					AppleTeamId:       strPtr(""),
					ApplePrivateKeyId: strPtr(""),
					ApplePrivateKey:   strPtr(""),
					Scope:             []string{"profile", "email", "address", "phone"},
					MapperUrl:         strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
					RequestedClaims:   &requestedClaims,
				},
			},
		},
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			m := NewGrpcMapper(mockLogger)

			got, err := m.FromConfigurations(test.input)
			if err != nil {
				t.Errorf("FromConfigurations() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("FromConfigurations() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcPbMapper_FromCreateIdpBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	requestedClaims := `{"user_info": {"given_name": {"essential": true}}}`
	rc, _ := json.Marshal(requestedClaims)

	tests := []struct {
		name  string
		input *v0Idps.CreateIdpBody
		want  *Configuration
	}{
		{
			name: "Successful mapping from CreateIdpBody to Configuration",
			input: &v0Idps.CreateIdpBody{
				Id:              "microsoft_af675f353bd7451588e2b8032e315f6f",
				Provider:        "microsoft",
				Label:           strPtr("microsoft"),
				ClientId:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
				ClientSecret:    strPtr("3y38Q~aslkdhaskjhd~W0xWDB.123u98asd"),
				MicrosoftTenant: strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
				Scope:           []string{"profile", "email", "address", "phone"},
				MapperUrl:       strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
				RequestedClaims: &requestedClaims,
			},
			want: &Configuration{
				ID:              "microsoft_af675f353bd7451588e2b8032e315f6f",
				Provider:        "microsoft",
				Label:           "microsoft",
				ClientID:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
				ClientSecret:    "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
				IssuerURL:       "",
				AuthURL:         "",
				TokenURL:        "",
				Tenant:          "e1574293-28de-4e94-87d5-b61c76fc14e1",
				SubjectSource:   "",
				TeamId:          "",
				PrivateKeyId:    "",
				PrivateKey:      "",
				Scope:           []string{"profile", "email", "address", "phone"},
				Mapper:          "file:///etc/config/kratos/microsoft_schema.jsonnet",
				RequestedClaims: rc,
			},
		},
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			m := NewGrpcMapper(mockLogger)

			got, err := m.ToCreateIdpBody(test.input)
			if err != nil {
				t.Errorf("FromCreateIdpBody() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("FromCreateIdpBody() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGrpcPbMapper_FromUpdateIdpBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	requestedClaims := `{"user_info": {"given_name": {"essential": true}}}`
	rc, _ := json.Marshal(requestedClaims)

	tests := []struct {
		name  string
		input *v0Idps.UpdateIdpBody
		want  *Configuration
	}{
		{
			name: "Successful mapping from UpdateIdpBody to Configuration",
			input: &v0Idps.UpdateIdpBody{
				Id:              "microsoft_af675f353bd7451588e2b8032e315f6f",
				Provider:        "microsoft",
				Label:           strPtr("microsoft"),
				ClientId:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
				ClientSecret:    strPtr("3y38Q~aslkdhaskjhd~W0xWDB.123u98asd"),
				MicrosoftTenant: strPtr("e1574293-28de-4e94-87d5-b61c76fc14e1"),
				Scope:           []string{"profile", "email", "address", "phone"},
				MapperUrl:       strPtr("file:///etc/config/kratos/microsoft_schema.jsonnet"),
				RequestedClaims: &requestedClaims,
			},
			want: &Configuration{
				ID:              "microsoft_af675f353bd7451588e2b8032e315f6f",
				Provider:        "microsoft",
				Label:           "microsoft",
				ClientID:        "af675f35-3bd7-4515-88e2-b8032e315f6f",
				ClientSecret:    "3y38Q~aslkdhaskjhd~W0xWDB.123u98asd",
				IssuerURL:       "",
				AuthURL:         "",
				TokenURL:        "",
				Tenant:          "e1574293-28de-4e94-87d5-b61c76fc14e1",
				SubjectSource:   "",
				TeamId:          "",
				PrivateKeyId:    "",
				PrivateKey:      "",
				Scope:           []string{"profile", "email", "address", "phone"},
				Mapper:          "file:///etc/config/kratos/microsoft_schema.jsonnet",
				RequestedClaims: rc,
			},
		},
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger := NewMockLoggerInterface(ctrl)

			m := NewGrpcMapper(mockLogger)

			got, err := m.ToUpdateIdpBody(test.input)
			if err != nil {
				t.Errorf("FromUpdateIdpBody() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("FromUpdateIdpBody() got = %v, want %v", got, test.want)
			}
		})
	}
}
