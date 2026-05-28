package agents

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// MarginValidatorAgent validates pricing decisions against margin constraints.
type MarginValidatorAgent struct {
	config models.MarginConfig
}

// NewMarginValidatorAgent creates a new margin validator agent.
func NewMarginValidatorAgent(config models.MarginConfig) *MarginValidatorAgent {
	return &MarginValidatorAgent{
		config: config,
	}
}

// Validate validates a pricing recommendation.
func (a *MarginValidatorAgent) Validate(
	rec models.PricingRecommendation,
	ourPrice int,
) (models.ValidationResult, error) {

	// Calculate cost and final price
	costPrice := ourPrice * a.config.CostPercent / 100
	finalPrice := rec.TargetPrice

	warnings := []string{}

	// Hard floor: never sell below cost
	if a.config.HardFloor && finalPrice < costPrice {
		return models.ValidationResult{
			ProductID:       rec.ProductID,
			Approved:        false,
			FinalPrice:      ourPrice,
			FinalDiscount:   0,
			Margin:          100,
			MarginDollars:   ourPrice - costPrice,
			Warnings:        []string{"REJECTED: Price below cost"},
			Rejected:        true,
			RejectionReason: fmt.Sprintf("Target price $%.2f is below cost $%.2f", float64(finalPrice)/100, float64(costPrice)/100),
		}, nil
	}

	// Calculate margin at target price
	margin := 0
	marginDollars := 0
	if finalPrice > 0 {
		marginDollars = finalPrice - costPrice
		margin = marginDollars * 100 / finalPrice
	}

	// Check minimum margin
	if margin < a.config.MinMarginPercent {
		// Adjust to meet minimum margin
		// margin% = (price - cost) / price * 100
		// price = cost / (1 - margin%/100)
		minAcceptablePrice := costPrice * 100 / (100 - a.config.MinMarginPercent)

		adjustedDiscount := ourPrice - minAcceptablePrice
		adjustedMargin := a.config.MinMarginPercent
		adjustedMarginDollars := minAcceptablePrice - costPrice

		warnings = append(warnings,
			fmt.Sprintf("Adjusted: margin was %d%%, increased to minimum %d%%", margin, a.config.MinMarginPercent))
		warnings = append(warnings,
			fmt.Sprintf("Price adjusted from $%.2f to $%.2f", float64(finalPrice)/100, float64(minAcceptablePrice)/100))

		return models.ValidationResult{
			ProductID:     rec.ProductID,
			Approved:      true,
			FinalPrice:    minAcceptablePrice,
			FinalDiscount: adjustedDiscount,
			Margin:        adjustedMargin,
			MarginDollars: adjustedMarginDollars,
			Warnings:      warnings,
			Rejected:      false,
		}, nil
	}

	// Approved as-is
	return models.ValidationResult{
		ProductID:     rec.ProductID,
		Approved:      true,
		FinalPrice:    finalPrice,
		FinalDiscount: rec.DiscountAmount,
		Margin:        margin,
		MarginDollars: marginDollars,
		Warnings:      warnings,
		Rejected:      false,
	}, nil
}
