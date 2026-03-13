package main

import "testing"

func validExperiment() Experiment {
	return Experiment{
		Slug:   "test-exp",
		Status: StatusRunning,
		Variants: []Variant{
			{Name: "control", Weight: 50},
			{Name: "treatment", Weight: 50},
		},
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name    string
		exp     Experiment
		wantErr bool
	}{
		{
			name: "valid slug",
			exp:  validExperiment(),
		},
		{
			name:    "empty slug",
			exp:     Experiment{Slug: "", Status: StatusRunning, Variants: []Variant{{Name: "a", Weight: 1}}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSlug(tt.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSlug() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  ExperimentStatus
		wantErr bool
	}{
		{"draft", StatusDraft, false},
		{"running", StatusRunning, false},
		{"paused", StatusPaused, false},
		{"stopped", StatusStopped, false},
		{"invalid", ExperimentStatus("invalid"), true},
		{"empty", ExperimentStatus(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := validExperiment()
			exp.Status = tt.status
			err := validateStatus(exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVariantsNotEmpty(t *testing.T) {
	tests := []struct {
		name    string
		exp     Experiment
		wantErr bool
	}{
		{
			name: "has variants",
			exp:  validExperiment(),
		},
		{
			name:    "no variants",
			exp:     Experiment{Slug: "test", Status: StatusRunning, Variants: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVariantsNotEmpty(tt.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVariantsNotEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVariantName(t *testing.T) {
	tests := []struct {
		name    string
		exp     Experiment
		wantErr bool
	}{
		{
			name: "valid names",
			exp:  validExperiment(),
		},
		{
			name: "empty variant name",
			exp: Experiment{
				Slug:     "test",
				Status:   StatusRunning,
				Variants: []Variant{{Name: "", Weight: 50}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVariantName(tt.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVariantName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVariantWeight(t *testing.T) {
	tests := []struct {
		name    string
		exp     Experiment
		wantErr bool
	}{
		{
			name: "positive weights",
			exp:  validExperiment(),
		},
		{
			name: "zero weight",
			exp: Experiment{
				Slug:     "test",
				Status:   StatusRunning,
				Variants: []Variant{{Name: "a", Weight: 0}},
			},
			wantErr: true,
		},
		{
			name: "negative weight",
			exp: Experiment{
				Slug:     "test",
				Status:   StatusRunning,
				Variants: []Variant{{Name: "a", Weight: -1}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVariantWeight(tt.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVariantWeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUniqueVariants(t *testing.T) {
	tests := []struct {
		name    string
		exp     Experiment
		wantErr bool
	}{
		{
			name: "unique variants",
			exp:  validExperiment(),
		},
		{
			name: "duplicate variant names",
			exp: Experiment{
				Slug:   "test",
				Status: StatusRunning,
				Variants: []Variant{
					{Name: "control", Weight: 50},
					{Name: "control", Weight: 50},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUniqueVariants(tt.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUniqueVariants() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOverrides(t *testing.T) {
	tests := []struct {
		name    string
		exp     Experiment
		wantErr bool
	}{
		{
			name: "no overrides",
			exp:  validExperiment(),
		},
		{
			name: "valid override",
			exp: Experiment{
				Slug:      "test",
				Status:    StatusRunning,
				Variants:  []Variant{{Name: "control", Weight: 50}, {Name: "treatment", Weight: 50}},
				Overrides: map[string]string{"user-1": "control"},
			},
		},
		{
			name: "override references unknown variant",
			exp: Experiment{
				Slug:      "test",
				Status:    StatusRunning,
				Variants:  []Variant{{Name: "control", Weight: 50}},
				Overrides: map[string]string{"user-1": "nonexistent"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOverrides(tt.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOverrides() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExperimentValidate(t *testing.T) {
	tests := []struct {
		name    string
		exp     Experiment
		wantErr bool
	}{
		{
			name: "valid experiment",
			exp:  validExperiment(),
		},
		{
			name:    "empty slug",
			exp:     Experiment{Status: StatusRunning, Variants: []Variant{{Name: "a", Weight: 1}}},
			wantErr: true,
		},
		{
			name:    "invalid status",
			exp:     Experiment{Slug: "test", Status: "bad", Variants: []Variant{{Name: "a", Weight: 1}}},
			wantErr: true,
		},
		{
			name:    "no variants",
			exp:     Experiment{Slug: "test", Status: StatusRunning},
			wantErr: true,
		},
		{
			name: "invalid override",
			exp: Experiment{
				Slug:      "test",
				Status:    StatusRunning,
				Variants:  []Variant{{Name: "a", Weight: 1}},
				Overrides: map[string]string{"u1": "nonexistent"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.exp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
