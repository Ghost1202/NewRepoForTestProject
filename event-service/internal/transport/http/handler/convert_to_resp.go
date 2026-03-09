package handler

import "github.com/turtlepavlo/event-service/internal/domain"

type ResponseConverter struct{}

func NewResponseConverter() *ResponseConverter {
	return &ResponseConverter{}
}

func (c *ResponseConverter) ToRespAddPromo(promoIDs []int64) AddPromoResp {
	return AddPromoResp{
		PromoIDs: promoIDs,
	}
}

func (c *ResponseConverter) ToRespAddEarly(earlyIDs []int64) AddEarlyResp {
	return AddEarlyResp{
		EarlyIDs: earlyIDs,
	}
}

func (c *ResponseConverter) ToRespAddBundle(bundleIDs []int64) AddBundleResp {
	return AddBundleResp{
		BundleIDs: bundleIDs,
	}
}

func (c *ResponseConverter) ToRespGetPromo(promos []domain.Promo) GetPromoRes {
	return GetPromoRes{
		Promos: c.ToRespPromos(promos),
	}
}

func (c *ResponseConverter) ToRespGetEarly(early []domain.Early) GetEarlyRes {
	return GetEarlyRes{
		Early: c.ToRespEarly(early),
	}
}

func (c *ResponseConverter) ToRespGetBundle(bundles []domain.Bundle) GetBundleRes {
	return GetBundleRes{
		Bundles: c.ToRespBundles(bundles),
	}
}

func (c *ResponseConverter) ToRespPromos(domains []domain.Promo) []Promo {
	if len(domains) == 0 {
		return []Promo{}
	}
	result := make([]Promo, len(domains))
	for i := range domains {
		result[i] = Promo{
			Code:   domains[i].Code,
			Type:   domains[i].Type,
			Value:  domains[i].Value,
			Sector: domains[i].Sector,
		}
	}
	return result
}

func (c *ResponseConverter) ToRespEarly(domains []domain.Early) []Early {
	if len(domains) == 0 {
		return []Early{}
	}
	result := make([]Early, len(domains))
	for i := range domains {
		result[i] = Early{
			Code:       domains[i].Code,
			Type:       domains[i].Type,
			Value:      domains[i].Value,
			Sector:     domains[i].Sector,
			ValidUntil: domains[i].ValidUntil,
		}
	}
	return result
}

func (c *ResponseConverter) ToRespBundles(domains []domain.Bundle) []Bundle {
	if len(domains) == 0 {
		return []Bundle{}
	}
	result := make([]Bundle, len(domains))
	for i := range domains {
		result[i] = Bundle{
			Code:     domains[i].Code,
			Sector:   domains[i].Sector,
			BuyCount: domains[i].BuyCount,
			GetCount: domains[i].GetCount,
		}
	}
	return result
}
