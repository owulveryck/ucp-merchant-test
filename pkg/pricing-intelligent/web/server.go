// Package main provides a web interface for intelligent pricing system.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	pricing "github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent/agents"
	"github.com/owulveryck/ucp-merchant-test/pkg/pricing-intelligent/models"
)

// Server handles web requests for pricing calculations.
type Server struct {
	orchestrator *pricing.Orchestrator
	customerData *MockCustomerDataSource
	competitorData *MockCompetitorDataSource
}

// MockCustomerDataSource provides mock customer data.
type MockCustomerDataSource struct {
	customers map[string]models.CustomerProfile
}

func NewMockCustomerData() *MockCustomerDataSource {
	return &MockCustomerDataSource{
		customers: map[string]models.CustomerProfile{
			"VIP001": {
				CustomerID:      "VIP001",
				TotalSpent:      120000, // $1200
				PurchaseCount:   15,
				AveragePurchase: 8000,
				LastPurchaseDays: 10,
				IsVIP:           true,
			},
			"STD001": {
				CustomerID:      "STD001",
				TotalSpent:      20000, // $200
				PurchaseCount:   2,
				AveragePurchase: 10000,
				LastPurchaseDays: 45,
				IsVIP:           false,
			},
			"NEW001": {
				CustomerID:      "NEW001",
				TotalSpent:      0,
				PurchaseCount:   0,
				AveragePurchase: 0,
				LastPurchaseDays: 999,
				IsVIP:           false,
			},
		},
	}
}

func (m *MockCustomerDataSource) GetCustomerProfile(customerID string) (models.CustomerProfile, error) {
	if profile, ok := m.customers[customerID]; ok {
		return profile, nil
	}
	// Return new customer profile
	return models.CustomerProfile{
		CustomerID:      customerID,
		TotalSpent:      0,
		PurchaseCount:   0,
		AveragePurchase: 0,
		LastPurchaseDays: 999,
		IsVIP:           false,
	}, nil
}

// MockCompetitorDataSource provides mock competitor prices.
type MockCompetitorDataSource struct {
	prices map[string][]int
}

func NewMockCompetitorData() *MockCompetitorDataSource {
	return &MockCompetitorDataSource{
		prices: map[string][]int{
			"PROD001": {6200, 6500, 5900}, // $62, $65, $59
			"PROD002": {7000, 7200},       // $70, $72
			"PROD003": {},                 // No competitors
		},
	}
}

func (m *MockCompetitorDataSource) GetCompetitorPrices(productID string) ([]int, error) {
	if prices, ok := m.prices[productID]; ok {
		return prices, nil
	}
	return []int{}, nil
}

// PricingAPIRequest represents API request.
type PricingAPIRequest struct {
	CustomerID string `json:"customer_id"`
	ProductID  string `json:"product_id"`
	BasePrice  int    `json:"base_price"`
	CostPrice  int    `json:"cost_price"`
}

// PricingAPIResponse represents API response.
type PricingAPIResponse struct {
	Success bool               `json:"success"`
	Result  *models.PricingResult `json:"result,omitempty"`
	Error   string             `json:"error,omitempty"`
}

func NewServer() *Server {
	customerData := NewMockCustomerData()
	competitorData := NewMockCompetitorData()

	loyaltyAgent := agents.NewLoyaltyAgent(customerData, agents.DefaultVIPThreshold)
	compAgent := agents.NewCompetitivenessAgent(competitorData)
	orchestrator := pricing.NewOrchestrator(loyaltyAgent, compAgent, 10)

	return &Server{
		orchestrator:   orchestrator,
		customerData:   customerData,
		competitorData: competitorData,
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlTemplate)
}

func (s *Server) handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PricingAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(PricingAPIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Calculate optimal price
	request := models.PricingRequest{
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		BasePrice:  req.BasePrice,
		CostPrice:  req.CostPrice,
	}

	result, err := s.orchestrator.CalculateOptimalPrice(request)
	if err != nil {
		json.NewEncoder(w).Encode(PricingAPIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PricingAPIResponse{
		Success: true,
		Result:  &result,
	})
}

func (s *Server) handleGetCustomers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.customerData.customers)
}

func (s *Server) handleGetCompetitors(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "product_id required",
		})
		return
	}

	prices, _ := s.competitorData.GetCompetitorPrices(productID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"product_id":        productID,
		"competitor_prices": prices,
	})
}

func main() {
	server := NewServer()

	http.HandleFunc("/", server.handleHome)
	http.HandleFunc("/api/calculate", server.handleCalculate)
	http.HandleFunc("/api/customers", server.handleGetCustomers)
	http.HandleFunc("/api/competitors", server.handleGetCompetitors)

	port := 8183
	fmt.Printf("\n🚀 Serveur démarré sur http://localhost:%d\n\n", port)
	fmt.Println("📋 Clients disponibles:")
	fmt.Println("   - VIP001  : Client VIP Premium ($1200 dépensé)")
	fmt.Println("   - STD001  : Client Standard ($200 dépensé)")
	fmt.Println("   - NEW001  : Nouveau client")
	fmt.Println("\n📦 Produits disponibles:")
	fmt.Println("   - PROD001 : Marché compétitif (3 concurrents)")
	fmt.Println("   - PROD002 : Leader de marché (2 concurrents)")
	fmt.Println("   - PROD003 : Monopole (aucun concurrent)")
	fmt.Println("\n👉 Ouvrez http://localhost:8183 dans votre navigateur\n")

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Système Multi-Agents de Pricing Intelligent</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
            min-height: 100vh;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            color: white;
            text-align: center;
            margin-bottom: 10px;
            font-size: 2.5em;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }
        .subtitle {
            color: rgba(255,255,255,0.9);
            text-align: center;
            margin-bottom: 30px;
            font-size: 1.2em;
        }
        .card {
            background: white;
            border-radius: 15px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #333;
        }
        select, input {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        select:focus, input:focus {
            outline: none;
            border-color: #667eea;
        }
        .info-box {
            background: #f5f5f5;
            padding: 15px;
            border-radius: 8px;
            margin-top: 10px;
            font-size: 14px;
            color: #666;
        }
        button {
            width: 100%;
            padding: 15px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 18px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
        }
        button:active {
            transform: translateY(0);
        }
        #result {
            display: none;
        }
        .result-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
        }
        .result-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }
        .metric {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #667eea;
        }
        .metric-label {
            font-size: 12px;
            color: #666;
            text-transform: uppercase;
            margin-bottom: 5px;
        }
        .metric-value {
            font-size: 24px;
            font-weight: 700;
            color: #333;
        }
        .reasoning {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin-top: 20px;
        }
        .reasoning h3 {
            margin-bottom: 15px;
            color: #333;
        }
        .reasoning-item {
            padding: 8px 0;
            border-bottom: 1px solid #e0e0e0;
        }
        .reasoning-item:last-child {
            border-bottom: none;
        }
        .badge {
            display: inline-block;
            padding: 5px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
            margin-left: 10px;
        }
        .badge-vip {
            background: #ffd700;
            color: #333;
        }
        .badge-competitive {
            background: #4CAF50;
            color: white;
        }
        .loading {
            text-align: center;
            padding: 20px;
            display: none;
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #667eea;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 0 auto;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎯 Système Multi-Agents de Pricing Intelligent</h1>
        <p class="subtitle">Le bon prix, au bon client, selon le marché</p>

        <div class="card">
            <form id="pricingForm">
                <div class="form-group">
                    <label>👤 Client</label>
                    <select id="customerId" required>
                        <option value="">Sélectionner un client</option>
                        <option value="VIP001">VIP001 - Client VIP Premium ($1200 dépensé)</option>
                        <option value="STD001">STD001 - Client Standard ($200 dépensé)</option>
                        <option value="NEW001">NEW001 - Nouveau client</option>
                    </select>
                </div>

                <div class="form-group">
                    <label>📦 Produit</label>
                    <select id="productId" required>
                        <option value="">Sélectionner un produit</option>
                        <option value="PROD001">PROD001 - Marché compétitif (3 concurrents)</option>
                        <option value="PROD002">PROD002 - Leader de marché (2 concurrents)</option>
                        <option value="PROD003">PROD003 - Monopole (aucun concurrent)</option>
                    </select>
                </div>

                <div class="form-group">
                    <label>💵 Prix de base ($)</label>
                    <input type="number" id="basePrice" value="60" min="1" required>
                    <div class="info-box">Prix actuel du produit en dollars</div>
                </div>

                <div class="form-group">
                    <label>💰 Prix de coût ($)</label>
                    <input type="number" id="costPrice" value="50" min="1" required>
                    <div class="info-box">Prix d'achat/production en dollars</div>
                </div>

                <button type="submit">🚀 Calculer le prix optimal</button>
            </form>
        </div>

        <div class="loading" id="loading">
            <div class="spinner"></div>
            <p style="margin-top: 15px; color: white;">Calcul en cours...</p>
        </div>

        <div class="card" id="result">
            <div class="result-header">
                <h2>📊 Résultat de l'analyse</h2>
            </div>

            <div class="result-grid">
                <div class="metric">
                    <div class="metric-label">Prix de base</div>
                    <div class="metric-value" id="basePriceResult">$0.00</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Prix final</div>
                    <div class="metric-value" style="color: #667eea;" id="finalPriceResult">$0.00</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Réduction</div>
                    <div class="metric-value" style="color: #4CAF50;" id="discountResult">$0.00</div>
                </div>
                <div class="metric">
                    <div class="metric-label">Marge</div>
                    <div class="metric-value" style="color: #FF9800;" id="marginResult">0%</div>
                </div>
            </div>

            <div id="badges"></div>

            <div class="reasoning">
                <h3>💡 Raisonnement détaillé</h3>
                <div id="reasoningList"></div>
            </div>
        </div>
    </div>

    <script>
        document.getElementById('pricingForm').addEventListener('submit', async (e) => {
            e.preventDefault();

            const customerId = document.getElementById('customerId').value;
            const productId = document.getElementById('productId').value;
            const basePrice = parseInt(document.getElementById('basePrice').value) * 100; // Convert to cents
            const costPrice = parseInt(document.getElementById('costPrice').value) * 100; // Convert to cents

            document.getElementById('loading').style.display = 'block';
            document.getElementById('result').style.display = 'none';

            try {
                const response = await fetch('/api/calculate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        customer_id: customerId,
                        product_id: productId,
                        base_price: basePrice,
                        cost_price: costPrice
                    })
                });

                const data = await response.json();

                document.getElementById('loading').style.display = 'none';

                if (data.success) {
                    displayResult(data.result);
                } else {
                    alert('Erreur: ' + data.error);
                }
            } catch (error) {
                document.getElementById('loading').style.display = 'none';
                alert('Erreur de connexion: ' + error.message);
            }
        });

        function displayResult(result) {
            document.getElementById('basePriceResult').textContent = '$' + (result.BasePrice / 100).toFixed(2);
            document.getElementById('finalPriceResult').textContent = '$' + (result.FinalPrice / 100).toFixed(2);
            document.getElementById('discountResult').textContent = '$' + (result.Discount / 100).toFixed(2) + ' (' + result.DiscountPercent + '%)';
            document.getElementById('marginResult').textContent = result.Margin + '%';

            // Badges
            let badges = '';
            if (result.IsVIP) {
                badges += '<span class="badge badge-vip">💎 Client VIP</span>';
            }
            if (result.IsCompetitive) {
                badges += '<span class="badge badge-competitive">✓ Prix compétitif</span>';
            }
            document.getElementById('badges').innerHTML = badges;

            // Reasoning
            const reasoningList = document.getElementById('reasoningList');
            reasoningList.innerHTML = result.Reasoning.map(r =>
                '<div class="reasoning-item">' + r + '</div>'
            ).join('');

            document.getElementById('result').style.display = 'block';
            document.getElementById('result').scrollIntoView({ behavior: 'smooth' });
        }
    </script>
</body>
</html>
`
