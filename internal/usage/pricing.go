package usage

type ModelPricing struct {
	InputPer1M  float64
	OutputPer1M float64
	CachedPer1M float64
}

var modelPricing = map[string]ModelPricing{
	"claude-3-5-haiku-20241022":    {InputPer1M: 0.80, OutputPer1M: 4.00, CachedPer1M: 0.08},
	"claude-3-7-sonnet-20250219":   {InputPer1M: 3.00, OutputPer1M: 15.00, CachedPer1M: 0.30},
	"claude-haiku-4-5-20251001":    {InputPer1M: 0.80, OutputPer1M: 4.00, CachedPer1M: 0.08},
	"claude-opus-4-1-20250805":     {InputPer1M: 15.00, OutputPer1M: 75.00, CachedPer1M: 1.50},
	"claude-opus-4-20250514":       {InputPer1M: 15.00, OutputPer1M: 75.00, CachedPer1M: 1.50},
	"claude-opus-4-5-20251101":     {InputPer1M: 15.00, OutputPer1M: 75.00, CachedPer1M: 1.50},
	"claude-opus-4-6":              {InputPer1M: 15.00, OutputPer1M: 75.00, CachedPer1M: 1.50},
	"claude-sonnet-4-20250514":     {InputPer1M: 3.00, OutputPer1M: 15.00, CachedPer1M: 0.30},
	"claude-sonnet-4-5-20250929":   {InputPer1M: 3.00, OutputPer1M: 15.00, CachedPer1M: 0.30},
	"claude-sonnet-4-6":            {InputPer1M: 3.00, OutputPer1M: 15.00, CachedPer1M: 0.30},
	"gpt-4o":                       {InputPer1M: 2.50, OutputPer1M: 10.00, CachedPer1M: 1.25},
	"gpt-4o-mini":                  {InputPer1M: 0.15, OutputPer1M: 0.60, CachedPer1M: 0.075},
	"gpt-4-turbo":                  {InputPer1M: 10.00, OutputPer1M: 30.00},
	"gpt-4":                        {InputPer1M: 30.00, OutputPer1M: 60.00},
	"gpt-3.5-turbo":                {InputPer1M: 0.50, OutputPer1M: 1.50},
	"o1":                           {InputPer1M: 15.00, OutputPer1M: 60.00, CachedPer1M: 7.50},
	"o1-mini":                      {InputPer1M: 1.10, OutputPer1M: 4.40, CachedPer1M: 0.55},
	"o1-pro":                       {InputPer1M: 150.00, OutputPer1M: 600.00},
	"o3":                           {InputPer1M: 10.00, OutputPer1M: 40.00, CachedPer1M: 2.50},
	"o3-mini":                      {InputPer1M: 1.10, OutputPer1M: 4.40, CachedPer1M: 0.275},
	"o4-mini":                      {InputPer1M: 1.10, OutputPer1M: 4.40, CachedPer1M: 0.275},
	"gemini-2.0-flash":             {InputPer1M: 0.10, OutputPer1M: 0.40},
	"gemini-2.0-flash-lite":        {InputPer1M: 0.075, OutputPer1M: 0.30},
	"gemini-2.5-flash":             {InputPer1M: 0.15, OutputPer1M: 0.60},
	"gemini-2.5-pro":               {InputPer1M: 1.25, OutputPer1M: 10.00},
	"gemini-1.5-flash":             {InputPer1M: 0.075, OutputPer1M: 0.30},
	"gemini-1.5-pro":               {InputPer1M: 1.25, OutputPer1M: 5.00},
}

func CalculateCostUSD(model string, inputTokens, outputTokens, cachedTokens int64) float64 {
	pricing, ok := modelPricing[model]
	if !ok {
		pricing = matchPricingByPrefix(model)
		if pricing.InputPer1M == 0 && pricing.OutputPer1M == 0 {
			return 0
		}
	}
	inputCost := float64(inputTokens) * pricing.InputPer1M / 1_000_000
	outputCost := float64(outputTokens) * pricing.OutputPer1M / 1_000_000
	cachedCost := float64(cachedTokens) * pricing.CachedPer1M / 1_000_000
	return inputCost + outputCost + cachedCost
}

func matchPricingByPrefix(model string) ModelPricing {
	prefixMap := map[string]ModelPricing{
		"claude-3-5-haiku":  {InputPer1M: 0.80, OutputPer1M: 4.00, CachedPer1M: 0.08},
		"claude-3-7-sonnet": {InputPer1M: 3.00, OutputPer1M: 15.00, CachedPer1M: 0.30},
		"claude-haiku-4":    {InputPer1M: 0.80, OutputPer1M: 4.00, CachedPer1M: 0.08},
		"claude-opus-4":     {InputPer1M: 15.00, OutputPer1M: 75.00, CachedPer1M: 1.50},
		"claude-sonnet-4":   {InputPer1M: 3.00, OutputPer1M: 15.00, CachedPer1M: 0.30},
		"gpt-4o-mini":       {InputPer1M: 0.15, OutputPer1M: 0.60, CachedPer1M: 0.075},
		"gpt-4o":            {InputPer1M: 2.50, OutputPer1M: 10.00, CachedPer1M: 1.25},
		"gpt-4-turbo":       {InputPer1M: 10.00, OutputPer1M: 30.00},
		"gpt-4":             {InputPer1M: 30.00, OutputPer1M: 60.00},
		"gpt-3.5":           {InputPer1M: 0.50, OutputPer1M: 1.50},
		"o1-mini":           {InputPer1M: 1.10, OutputPer1M: 4.40, CachedPer1M: 0.55},
		"o1-pro":            {InputPer1M: 150.00, OutputPer1M: 600.00},
		"o1":                {InputPer1M: 15.00, OutputPer1M: 60.00, CachedPer1M: 7.50},
		"o3-mini":           {InputPer1M: 1.10, OutputPer1M: 4.40, CachedPer1M: 0.275},
		"o3":                {InputPer1M: 10.00, OutputPer1M: 40.00, CachedPer1M: 2.50},
		"o4-mini":           {InputPer1M: 1.10, OutputPer1M: 4.40, CachedPer1M: 0.275},
		"gemini-2.5-pro":    {InputPer1M: 1.25, OutputPer1M: 10.00},
		"gemini-2.5-flash":  {InputPer1M: 0.15, OutputPer1M: 0.60},
		"gemini-2.0-flash":  {InputPer1M: 0.10, OutputPer1M: 0.40},
		"gemini-1.5-pro":    {InputPer1M: 1.25, OutputPer1M: 5.00},
		"gemini-1.5-flash":  {InputPer1M: 0.075, OutputPer1M: 0.30},
	}
	for prefix, pricing := range prefixMap {
		if len(model) >= len(prefix) && model[:len(prefix)] == prefix {
			return pricing
		}
	}
	return ModelPricing{}
}

func GetModelPricing(model string) (ModelPricing, bool) {
	pricing, ok := modelPricing[model]
	if ok {
		return pricing, true
	}
	pricing = matchPricingByPrefix(model)
	if pricing.InputPer1M > 0 || pricing.OutputPer1M > 0 {
		return pricing, true
	}
	return ModelPricing{}, false
}
