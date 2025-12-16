package constraints_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pedantigo "github.com/SmrutAI/pedantigo"
)

// checkValidationErrorWithField verifies validation errors and optionally checks the first error's field.
func checkValidationErrorWithField(t *testing.T, err error, wantErr bool, wantField string) {
	t.Helper()

	if wantErr {
		require.Error(t, err, "expected validation error")
		var ve *pedantigo.ValidationError
		require.ErrorAs(t, err, &ve, "expected ValidationError, got %T", err)
		require.NotEmpty(t, ve.Errors, "expected errors list")
		if wantField != "" {
			assert.Equal(t, wantField, ve.Errors[0].Field)
		}
	} else {
		require.NoError(t, err, "expected no errors")
	}
}

// ============================================================================
// required_if Tests
// ============================================================================

// TestRequiredIf tests RequiredIf validation.
func TestRequiredIf(t *testing.T) {
	type FormStringCondition struct {
		Country string `json:"country"`
		State   string `json:"state" pedantigo:"required_if=Country:US"`
	}

	type FormBoolCondition struct {
		IsPremium      bool   `json:"is_premium"`
		PremiumFeature string `json:"premium_feature" pedantigo:"required_if=IsPremium:true"`
	}

	type FormIntCondition struct {
		Status       int    `json:"status"`
		TrackingCode string `json:"tracking_code" pedantigo:"required_if=Status:2"`
	}

	type FormMultiple struct {
		Country  string `json:"country"`
		Domestic bool   `json:"domestic"`
		State    string `json:"state" pedantigo:"required_if=Country:US"`
		TaxID    string `json:"tax_id" pedantigo:"required_if=Domestic:true"`
	}

	tests := []struct {
		name      string
		form      interface{}
		validator interface{}
		wantErr   bool
		wantField string
	}{
		{
			name: "string condition - met with field",
			form: &FormStringCondition{Country: "US", State: "CA"},
		},
		{
			name:      "string condition - met without field",
			form:      &FormStringCondition{Country: "US", State: ""},
			wantErr:   true,
			wantField: "State",
		},
		{
			name: "string condition - not met without field",
			form: &FormStringCondition{Country: "CA", State: ""},
		},
		{
			name: "string condition - not met with field",
			form: &FormStringCondition{Country: "Canada", State: "Ontario"},
		},
		{
			name: "boolean condition - true with field",
			form: &FormBoolCondition{IsPremium: true, PremiumFeature: "advanced_analytics"},
		},
		{
			name:      "boolean condition - true without field",
			form:      &FormBoolCondition{IsPremium: true, PremiumFeature: ""},
			wantErr:   true,
			wantField: "PremiumFeature",
		},
		{
			name: "boolean condition - false",
			form: &FormBoolCondition{IsPremium: false, PremiumFeature: ""},
		},
		{
			name: "integer condition - match with field",
			form: &FormIntCondition{Status: 2, TrackingCode: "TRACK123"},
		},
		{
			name:      "integer condition - match without field",
			form:      &FormIntCondition{Status: 2, TrackingCode: ""},
			wantErr:   true,
			wantField: "TrackingCode",
		},
		{
			name: "integer condition - no match",
			form: &FormIntCondition{Status: 0, TrackingCode: ""},
		},
		{
			name: "multiple conditions - all satisfied",
			form: &FormMultiple{Country: "US", Domestic: true, State: "CA", TaxID: "123456"},
		},
		{
			name:    "multiple conditions - State missing",
			form:    &FormMultiple{Country: "US", Domestic: true, State: "", TaxID: "123456"},
			wantErr: true,
		},
		{
			name:    "multiple conditions - TaxID missing",
			form:    &FormMultiple{Country: "US", Domestic: true, State: "CA", TaxID: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch form := tt.form.(type) {
			case *FormStringCondition:
				validator := pedantigo.New[FormStringCondition]()
				err := validator.Validate(form)
				if tt.wantErr {
					require.Error(t, err, "expected validation error")
					var ve *pedantigo.ValidationError
					require.ErrorAs(t, err, &ve, "expected ValidationError, got %T", err)
					require.NotEmpty(t, ve.Errors, "expected errors list")
					if tt.wantField != "" {
						assert.Equal(t, tt.wantField, ve.Errors[0].Field)
					}
				} else {
					require.NoError(t, err, "expected no errors")
				}
			case *FormBoolCondition:
				validator := pedantigo.New[FormBoolCondition]()
				err := validator.Validate(form)
				if tt.wantErr {
					require.Error(t, err, "expected validation error")
					var ve *pedantigo.ValidationError
					require.ErrorAs(t, err, &ve, "expected ValidationError, got %T", err)
					require.NotEmpty(t, ve.Errors, "expected errors list")
					if tt.wantField != "" {
						assert.Equal(t, tt.wantField, ve.Errors[0].Field)
					}
				} else {
					require.NoError(t, err, "expected no errors")
				}
			case *FormIntCondition:
				validator := pedantigo.New[FormIntCondition]()
				err := validator.Validate(form)
				if tt.wantErr {
					require.Error(t, err, "expected validation error")
					var ve *pedantigo.ValidationError
					require.ErrorAs(t, err, &ve, "expected ValidationError, got %T", err)
					require.NotEmpty(t, ve.Errors, "expected errors list")
					if tt.wantField != "" {
						assert.Equal(t, tt.wantField, ve.Errors[0].Field)
					}
				} else {
					require.NoError(t, err, "expected no errors")
				}
			case *FormMultiple:
				validator := pedantigo.New[FormMultiple]()
				err := validator.Validate(form)
				if tt.wantErr {
					require.Error(t, err, "expected validation error")
				} else {
					require.NoError(t, err, "expected no errors")
				}
			}
		})
	}
}

// ============================================================================
// required_unless Tests
// ============================================================================

// TestRequiredUnless tests RequiredUnless validation.
func TestRequiredUnless(t *testing.T) {
	type FormStringCondition struct {
		Status   string `json:"status"`
		Password string `json:"password" pedantigo:"required_unless=Status:guest"`
	}

	type FormBoolCondition struct {
		Automated   bool   `json:"automated"`
		CaptchaCode string `json:"captcha_code" pedantigo:"required_unless=Automated:true"`
	}

	type FormMultiple struct {
		UserType     string `json:"user_type"`
		IsBot        bool   `json:"is_bot"`
		Email        string `json:"email" pedantigo:"required_unless=UserType:anonymous"`
		Verification string `json:"verification" pedantigo:"required_unless=IsBot:true"`
	}

	tests := []struct {
		name      string
		form      interface{}
		wantErr   bool
		wantField string
	}{
		{
			name: "string condition - not met with field",
			form: &FormStringCondition{Status: "active", Password: "securepass123"},
		},
		{
			name:      "string condition - not met without field",
			form:      &FormStringCondition{Status: "active", Password: ""},
			wantErr:   true,
			wantField: "Password",
		},
		{
			name: "string condition - met without field",
			form: &FormStringCondition{Status: "guest", Password: ""},
		},
		{
			name: "string condition - met with field",
			form: &FormStringCondition{Status: "guest", Password: "anypassword"},
		},
		{
			name: "boolean condition - false with field",
			form: &FormBoolCondition{Automated: false, CaptchaCode: "abc123xyz"},
		},
		{
			name:    "boolean condition - false without field",
			form:    &FormBoolCondition{Automated: false, CaptchaCode: ""},
			wantErr: true,
		},
		{
			name: "boolean condition - true",
			form: &FormBoolCondition{Automated: true, CaptchaCode: ""},
		},
		{
			name: "multiple conditions - all satisfied",
			form: &FormMultiple{UserType: "user", IsBot: false, Email: "user@example.com", Verification: "code123"},
		},
		{
			name:    "multiple conditions - Email missing",
			form:    &FormMultiple{UserType: "user", IsBot: false, Email: "", Verification: "code123"},
			wantErr: true,
		},
		{
			name:    "multiple conditions - Verification missing",
			form:    &FormMultiple{UserType: "user", IsBot: false, Email: "user@example.com", Verification: ""},
			wantErr: true,
		},
		{
			name: "multiple conditions - both exceptions",
			form: &FormMultiple{UserType: "anonymous", IsBot: true, Email: "", Verification: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch form := tt.form.(type) {
			case *FormStringCondition:
				validator := pedantigo.New[FormStringCondition]()
				err := validator.Validate(form)
				if tt.wantErr {
					require.Error(t, err, "expected validation error")
					var ve *pedantigo.ValidationError
					require.ErrorAs(t, err, &ve, "expected ValidationError, got %T", err)
					require.NotEmpty(t, ve.Errors, "expected errors list")
					if tt.wantField != "" {
						assert.Equal(t, tt.wantField, ve.Errors[0].Field)
					}
				} else {
					require.NoError(t, err, "expected no errors")
				}
			case *FormBoolCondition:
				validator := pedantigo.New[FormBoolCondition]()
				err := validator.Validate(form)
				if tt.wantErr {
					require.Error(t, err, "expected validation error")
				} else {
					require.NoError(t, err, "expected no errors")
				}
			case *FormMultiple:
				validator := pedantigo.New[FormMultiple]()
				err := validator.Validate(form)
				if tt.wantErr {
					require.Error(t, err, "expected validation error")
				} else {
					require.NoError(t, err, "expected no errors")
				}
			}
		})
	}
}

// ============================================================================
// required_with Tests
// ============================================================================

// TestRequiredWith tests RequiredWith validation.
func TestRequiredWith(t *testing.T) {
	type FormStringCondition struct {
		Method string `json:"method"`
		Token  string `json:"token" pedantigo:"required_with=Method"`
	}

	type FormIntCondition struct {
		Quantity  int    `json:"quantity"`
		Warehouse string `json:"warehouse" pedantigo:"required_with=Quantity"`
	}

	type FormBoolCondition struct {
		Enabled bool   `json:"enabled"`
		Config  string `json:"config" pedantigo:"required_with=Enabled"`
	}

	type FormMultiple struct {
		GuestName     string `json:"guest_name"`
		Phone         string `json:"phone"`
		Address       string `json:"address" pedantigo:"required_with=GuestName"`
		EmergencyName string `json:"emergency_name" pedantigo:"required_with=Phone"`
	}

	tests := []struct {
		name      string
		form      interface{}
		wantErr   bool
		wantField string
	}{
		{
			name: "string field - present with dependency",
			form: &FormStringCondition{Method: "credit_card", Token: "tok_123"},
		},
		{
			name:      "string field - present without dependency",
			form:      &FormStringCondition{Method: "credit_card", Token: ""},
			wantErr:   true,
			wantField: "Token",
		},
		{
			name: "string field - absent without dependency",
			form: &FormStringCondition{Method: "", Token: ""},
		},
		{
			name: "string field - absent with optional dependency",
			form: &FormStringCondition{Method: "", Token: "tok_123"},
		},
		{
			name: "integer field - nonzero with dependency",
			form: &FormIntCondition{Quantity: 5, Warehouse: "main"},
		},
		{
			name:    "integer field - nonzero without dependency",
			form:    &FormIntCondition{Quantity: 5, Warehouse: ""},
			wantErr: true,
		},
		{
			name: "integer field - zero",
			form: &FormIntCondition{Quantity: 0, Warehouse: ""},
		},
		{
			name: "boolean field - true with dependency",
			form: &FormBoolCondition{Enabled: true, Config: "settings"},
		},
		{
			name:    "boolean field - true without dependency",
			form:    &FormBoolCondition{Enabled: true, Config: ""},
			wantErr: true,
		},
		{
			name: "boolean field - false",
			form: &FormBoolCondition{Enabled: false, Config: ""},
		},
		{
			name: "multiple fields - all satisfied",
			form: &FormMultiple{
				GuestName:     "John Doe",
				Phone:         "555-1234",
				Address:       "123 Main St",
				EmergencyName: "Jane Doe",
			},
		},
		{
			name: "multiple fields - Address missing",
			form: &FormMultiple{
				GuestName:     "John Doe",
				Phone:         "555-1234",
				Address:       "",
				EmergencyName: "Jane Doe",
			},
			wantErr: true,
		},
		{
			name: "multiple fields - EmergencyName missing",
			form: &FormMultiple{
				GuestName:     "John Doe",
				Phone:         "555-1234",
				Address:       "123 Main St",
				EmergencyName: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch form := tt.form.(type) {
			case *FormStringCondition:
				validator := pedantigo.New[FormStringCondition]()
				err := validator.Validate(form)
				checkValidationErrorWithField(t, err, tt.wantErr, tt.wantField)
			case *FormIntCondition:
				validator := pedantigo.New[FormIntCondition]()
				err := validator.Validate(form)
				checkValidationErrorWithField(t, err, tt.wantErr, "")
			case *FormBoolCondition:
				validator := pedantigo.New[FormBoolCondition]()
				err := validator.Validate(form)
				checkValidationErrorWithField(t, err, tt.wantErr, "")
			case *FormMultiple:
				validator := pedantigo.New[FormMultiple]()
				err := validator.Validate(form)
				checkValidationErrorWithField(t, err, tt.wantErr, "")
			}
		})
	}
}

// ============================================================================
// required_without Tests
// ============================================================================

// TestRequiredWithout tests RequiredWithout validation.
func TestRequiredWithout(t *testing.T) {
	type FormStringCondition struct {
		DefaultAddress string `json:"default_address"`
		CustomAddress  string `json:"custom_address" pedantigo:"required_without=DefaultAddress"`
	}

	type FormIntCondition struct {
		FixedAmount    int    `json:"fixed_amount"`
		PercentageCode string `json:"percentage_code" pedantigo:"required_without=FixedAmount"`
	}

	type FormBoolCondition struct {
		UseDefault bool   `json:"use_default"`
		CustomRule string `json:"custom_rule" pedantigo:"required_without=UseDefault"`
	}

	type FormMultiple struct {
		WarehouseLocation string `json:"warehouse_location"`
		ShippingLabel     string `json:"shipping_label"`
		StorageBox        string `json:"storage_box" pedantigo:"required_without=WarehouseLocation"`
		ShippingTracking  string `json:"shipping_tracking" pedantigo:"required_without=ShippingLabel"`
	}

	tests := []struct {
		name      string
		form      interface{}
		wantErr   bool
		wantField string
	}{
		{
			name: "string field - absent with dependency",
			form: &FormStringCondition{DefaultAddress: "", CustomAddress: "123 Oak St"},
		},
		{
			name:      "string field - absent without dependency",
			form:      &FormStringCondition{DefaultAddress: "", CustomAddress: ""},
			wantErr:   true,
			wantField: "CustomAddress",
		},
		{
			name: "string field - present without dependency",
			form: &FormStringCondition{DefaultAddress: "456 Elm St", CustomAddress: ""},
		},
		{
			name: "string field - present with optional dependency",
			form: &FormStringCondition{DefaultAddress: "456 Elm St", CustomAddress: "123 Oak St"},
		},
		{
			name: "integer field - zero with dependency",
			form: &FormIntCondition{FixedAmount: 0, PercentageCode: "SAVE10"},
		},
		{
			name:    "integer field - zero without dependency",
			form:    &FormIntCondition{FixedAmount: 0, PercentageCode: ""},
			wantErr: true,
		},
		{
			name: "integer field - nonzero",
			form: &FormIntCondition{FixedAmount: 50, PercentageCode: ""},
		},
		{
			name: "boolean field - false with dependency",
			form: &FormBoolCondition{UseDefault: false, CustomRule: "notify_all"},
		},
		{
			name:    "boolean field - false without dependency",
			form:    &FormBoolCondition{UseDefault: false, CustomRule: ""},
			wantErr: true,
		},
		{
			name: "boolean field - true",
			form: &FormBoolCondition{UseDefault: true, CustomRule: ""},
		},
		{
			name: "multiple fields - all satisfied",
			form: &FormMultiple{
				WarehouseLocation: "WH-A1",
				ShippingLabel:     "LABEL-123",
				StorageBox:        "",
				ShippingTracking:  "",
			},
		},
		{
			name: "multiple fields - StorageBox missing",
			form: &FormMultiple{
				WarehouseLocation: "",
				ShippingLabel:     "LABEL-123",
				StorageBox:        "",
				ShippingTracking:  "TRACK-456",
			},
			wantErr: true,
		},
		{
			name: "multiple fields - ShippingTracking missing",
			form: &FormMultiple{
				WarehouseLocation: "WH-A1",
				ShippingLabel:     "",
				StorageBox:        "BOX-789",
				ShippingTracking:  "",
			},
			wantErr: true,
		},
		{
			name: "multiple fields - alternates provided",
			form: &FormMultiple{
				WarehouseLocation: "",
				ShippingLabel:     "",
				StorageBox:        "BOX-789",
				ShippingTracking:  "TRACK-456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch form := tt.form.(type) {
			case *FormStringCondition:
				validator := pedantigo.New[FormStringCondition]()
				err := validator.Validate(form)
				checkValidationErrorWithField(t, err, tt.wantErr, tt.wantField)
			case *FormIntCondition:
				validator := pedantigo.New[FormIntCondition]()
				err := validator.Validate(form)
				checkValidationErrorWithField(t, err, tt.wantErr, "")
			case *FormBoolCondition:
				validator := pedantigo.New[FormBoolCondition]()
				err := validator.Validate(form)
				checkValidationErrorWithField(t, err, tt.wantErr, "")
			case *FormMultiple:
				validator := pedantigo.New[FormMultiple]()
				err := validator.Validate(form)
				checkValidationErrorWithField(t, err, tt.wantErr, "")
			}
		})
	}
}

// ============================================================================
// Cross-Constraint Integration Tests
// ============================================================================

// TestCrossFieldConstraints tests CrossFieldConstraints validation.
func TestCrossFieldConstraints(t *testing.T) {
	t.Run("complex scenario - valid business account", func(t *testing.T) {
		type UserProfile struct {
			AccountType      string `json:"account_type"`
			IsVerified       bool   `json:"is_verified"`
			BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
			TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
			VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
			BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
			NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
		}
		validator := pedantigo.New[UserProfile]()
		err := validator.Validate(&UserProfile{
			AccountType:      "business",
			IsVerified:       true,
			BusinessName:     "Acme Corp",
			TaxID:            "12-3456789",
			VerificationDoc:  "doc_123",
			BackupEmail:      "backup@acme.com",
			NotificationPref: "email",
		})
		require.NoError(t, err, "expected no errors")
	})

	t.Run("complex scenario - business without BusinessName", func(t *testing.T) {
		type UserProfile struct {
			AccountType      string `json:"account_type"`
			IsVerified       bool   `json:"is_verified"`
			BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
			TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
			VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
			BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
			NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
		}
		validator := pedantigo.New[UserProfile]()
		err := validator.Validate(&UserProfile{
			AccountType:      "business",
			IsVerified:       false,
			BusinessName:     "",
			TaxID:            "12-3456789",
			VerificationDoc:  "",
			BackupEmail:      "backup@example.com",
			NotificationPref: "email",
		})
		require.Error(t, err, "expected validation error for BusinessName")
	})

	t.Run("complex scenario - valid government account", func(t *testing.T) {
		type UserProfile struct {
			AccountType      string `json:"account_type"`
			IsVerified       bool   `json:"is_verified"`
			BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
			TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
			VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
			BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
			NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
		}
		validator := pedantigo.New[UserProfile]()
		err := validator.Validate(&UserProfile{
			AccountType:      "government",
			IsVerified:       true,
			BusinessName:     "",
			TaxID:            "",
			VerificationDoc:  "gov_doc_456",
			BackupEmail:      "",
			NotificationPref: "",
		})
		require.NoError(t, err, "expected no errors")
	})

	t.Run("complex scenario - missing NotificationPref for BackupEmail", func(t *testing.T) {
		type UserProfile struct {
			AccountType      string `json:"account_type"`
			IsVerified       bool   `json:"is_verified"`
			BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
			TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
			VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
			BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
			NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
		}
		validator := pedantigo.New[UserProfile]()
		err := validator.Validate(&UserProfile{
			AccountType:      "personal",
			IsVerified:       false,
			BusinessName:     "",
			TaxID:            "",
			VerificationDoc:  "",
			BackupEmail:      "backup@example.com",
			NotificationPref: "",
		})
		require.Error(t, err, "expected validation error for NotificationPref")
	})

	t.Run("field index resolution", func(t *testing.T) {
		type Form struct {
			Field1 string `json:"field1"`
			Field2 string `json:"field2"`
			Field3 string `json:"field3" pedantigo:"required_if=Field1:trigger"`
			Field4 string `json:"field4" pedantigo:"required_unless=Field2:skip"`
		}
		validator := pedantigo.New[Form]()
		require.NotNil(t, validator, "validator creation failed")
		_ = validator.Validate(&Form{
			Field1: "trigger",
			Field2: "active",
			Field3: "value",
			Field4: "value",
		})
	})

	t.Run("zero value distinction", func(t *testing.T) {
		type Form struct {
			TriggerField string `json:"trigger_field"`
			TargetField  string `json:"target_field" pedantigo:"required_with=TriggerField"`
		}
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{
			TriggerField: "value",
			TargetField:  "",
		})
		require.Error(t, err, "expected validation error for zero value with required_with")
	})

	t.Run("unexported fields handling - valid", func(t *testing.T) {
		type Form struct {
			privateField string
			PublicField  string `json:"public_field"`
			Conditional  string `json:"conditional" pedantigo:"required_if=PublicField:trigger"`
		}
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{
			privateField: "ignored",
			PublicField:  "trigger",
			Conditional:  "value",
		})
		assert.NoError(t, err, "expected validation to work")
	})

	t.Run("unexported fields handling - invalid", func(t *testing.T) {
		type Form struct {
			privateField string
			PublicField  string `json:"public_field"`
			Conditional  string `json:"conditional" pedantigo:"required_if=PublicField:trigger"`
		}
		validator := pedantigo.New[Form]()
		err := validator.Validate(&Form{
			privateField: "ignored",
			PublicField:  "trigger",
			Conditional:  "",
		})
		require.Error(t, err, "expected validation error when Conditional missing")
	})

	t.Run("reflect value handling", func(t *testing.T) {
		type Form struct {
			Status string `json:"status"`
			Detail string `json:"detail" pedantigo:"required_if=Status:complete"`
		}
		validator := pedantigo.New[Form]()
		form := Form{
			Status: "complete",
			Detail: "",
		}
		err := validator.Validate(&form)
		require.Error(t, err, "expected validation error")
	})
}
