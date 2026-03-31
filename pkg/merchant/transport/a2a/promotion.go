package a2a

import (
	"context"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
)

func (s *Server) handleListPromotions(ctx context.Context, ac *actionContext) (map[string]any, error) {
	p, ok := s.merchant.(merchant.Promoter)
	if !ok {
		return map[string]any{"promotions": []any{}}, nil
	}
	return map[string]any{"promotions": p.ListPromotions()}, nil
}
