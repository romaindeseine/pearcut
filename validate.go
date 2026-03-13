package main

import "fmt"

func validateSlug(exp Experiment) error {
	if exp.Slug == "" {
		return fmt.Errorf("experiment with empty slug")
	}
	return nil
}

func validateStatus(exp Experiment) error {
	switch exp.Status {
	case StatusDraft, StatusRunning, StatusPaused, StatusStopped:
		return nil
	default:
		return fmt.Errorf("experiment %q has invalid status %q", exp.Slug, exp.Status)
	}
}

func validateVariantName(exp Experiment) error {
	for _, v := range exp.Variants {
		if v.Name == "" {
			return fmt.Errorf("experiment %q has variant with empty name", exp.Slug)
		}
	}
	return nil
}

func validateVariantWeight(exp Experiment) error {
	for _, v := range exp.Variants {
		if v.Weight <= 0 {
			return fmt.Errorf("experiment %q variant %q has non-positive weight", exp.Slug, v.Name)
		}
	}
	return nil
}

func validateUniqueVariants(exp Experiment) error {
	names := make(map[string]bool, len(exp.Variants))
	for _, v := range exp.Variants {
		if names[v.Name] {
			return fmt.Errorf("experiment %q has duplicate variant %q", exp.Slug, v.Name)
		}
		names[v.Name] = true
	}
	return nil
}

func validateVariantsNotEmpty(exp Experiment) error {
	if len(exp.Variants) == 0 {
		return fmt.Errorf("experiment %q has no variants", exp.Slug)
	}
	return nil
}

func validateOverrides(exp Experiment) error {
	variantNames := make(map[string]bool, len(exp.Variants))
	for _, v := range exp.Variants {
		variantNames[v.Name] = true
	}
	for userID, variantName := range exp.Overrides {
		if !variantNames[variantName] {
			return fmt.Errorf("experiment %q override for %q references unknown variant %q", exp.Slug, userID, variantName)
		}
	}
	return nil
}

func validateVariants(exp Experiment) error {
	if err := validateVariantsNotEmpty(exp); err != nil {
		return err
	}
	if err := validateVariantName(exp); err != nil {
		return err
	}
	if err := validateVariantWeight(exp); err != nil {
		return err
	}
	if err := validateUniqueVariants(exp); err != nil {
		return err
	}
	return nil
}
