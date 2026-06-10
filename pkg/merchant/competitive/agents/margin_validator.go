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

	// Calculate cost: use ActualCost if set, otherwise estimate from percentage
	costPrice := a.config.ActualCost
	if costPrice == 0 {
		costPrice = ourPrice * a.config.CostPercent / 100
	}
	finalPrice := rec.TargetPrice

	warnings := []string{}

	// Hard floor: never sell below cost
	if a.config.HardFloor && finalPrice < costPrice {
		// WINNING STRATEGY: Instead of complete rejection, offer the BEST price we can
		// Set price to cost + minimum margin to stay competitive while protecting margin
		bestPossiblePrice := costPrice * (100 + a.config.MinMarginPercent) / 100
		bestDiscount := ourPrice - bestPossiblePrice

		// If even the best possible price is still higher than our current price, keep current
		if bestPossiblePrice >= ourPrice {
			return models.ValidationResult{
				ProductID:       rec.ProductID,
				Approved:        false,
				FinalPrice:      ourPrice,
				FinalDiscount:   0,
				Margin:          (ourPrice - costPrice) * 100 / ourPrice,
				MarginDollars:   ourPrice - costPrice,
				Warnings:        []string{"Cannot compete: market price below our cost"},
				Rejected:        true,
				RejectionReason: fmt.Sprintf("Target price $%.2f is below cost $%.2f, and current price already optimal", float64(finalPrice)/100, float64(costPrice)/100),
			}, nil
		}

		// Apply the best possible competitive price
		margin := (bestPossiblePrice - costPrice) * 100 / bestPossiblePrice
		return models.ValidationResult{
			ProductID:     rec.ProductID,
			Approved:      true, // APPROVE the adjusted price
			FinalPrice:    bestPossiblePrice,
			FinalDiscount: bestDiscount,
			Margin:        margin,
			MarginDollars: bestPossiblePrice - costPrice,
			Warnings: []string{
				fmt.Sprintf("⚠️ Prix cible $%.2f < coût $%.2f : ajusté à $%.2f (coût + %d%% marge)",
					float64(finalPrice)/100,
					float64(costPrice)/100,
					float64(bestPossiblePrice)/100,
					a.config.MinMarginPercent),
				"Prix compétitif impossible sans perte - prix minimum appliqué",
			},
			Rejected: false,
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
		// WINNING STRATEGY: Accept lower margin to guarantee victory
		// Only adjust UP if we're still below cost (hard floor)
		if finalPrice < costPrice && a.config.HardFloor {
			// Below cost = selling at loss → REJECT completely
			return models.ValidationResult{
				ProductID:       rec.ProductID,
				Approved:        false,
				FinalPrice:      ourPrice,
				FinalDiscount:   0,
				Margin:          100,
				MarginDollars:   ourPrice - costPrice,
				Warnings:        []string{fmt.Sprintf("REJECTED: Price $%.2f below cost $%.2f", float64(finalPrice)/100, float64(costPrice)/100)},
				Rejected:        true,
				RejectionReason: fmt.Sprintf("Cannot win without selling at loss (target $%.2f < cost $%.2f)", float64(finalPrice)/100, float64(costPrice)/100),
			}, nil
		}

		// Above cost but below target margin → ACCEPT to WIN
		warnings = append(warnings,
			fmt.Sprintf("⚠️ Marge réduite: %d%% (cible: %d%%) pour GAGNER", margin, a.config.MinMarginPercent))
		warnings = append(warnings,
			fmt.Sprintf("Prix $%.2f accepté pour battre concurrence", float64(finalPrice)/100))

		return models.ValidationResult{
			ProductID:     rec.ProductID,
			Approved:      true,
			FinalPrice:    finalPrice, // KEEP recommended price to WIN
			FinalDiscount: rec.DiscountAmount,
			Margin:        margin, // Lower than target, but still positive
			MarginDollars: marginDollars,
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
