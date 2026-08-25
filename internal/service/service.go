package service

import (
	"context"
	"strings"
	"time"

	"github.com/ashokparihar/fitcheck/internal/ai"
	"github.com/ashokparihar/fitcheck/internal/outfit"
	"github.com/ashokparihar/fitcheck/internal/store"
	"github.com/ashokparihar/fitcheck/internal/weather"
	"github.com/google/uuid"
)

// Service orchestrates AI, weather, and outfit logic over the store.
type Service struct {
	Store     store.Store
	AI        *ai.Client
	UploadDir string
}

type PlanRequest struct {
	Location   string
	StartDate  string
	Days       int
	Activities string
	Formality  string
	LookGoal   string
}

type TripRequest struct {
	PlanRequest
	Laundry string
	Luggage string
}

func (s *Service) GenerateOutfits(ctx context.Context, req PlanRequest) []store.Outfit {
	items := s.Store.ListItems("")
	if len(items) == 0 {
		return nil
	}

	forecast := s.fetchWeather(ctx, req)
	plan := toOutfitPlan(req)
	profile := ToOutfitProfile(s.Store.GetProfile())
	outfitItems := store.ToOutfitItems(items)

	var generated []ai.GeneratedOutfit
	if s.AI != nil {
		generated, _ = ai.GenerateOutfits(ctx, s.AI, profile, outfitItems, forecast, plan)
	}
	if len(generated) == 0 {
		generated = ruleOutfits(profile, outfitItems, forecast, plan)
	}

	return s.persistOutfits(ctx, items, generated)
}

func (s *Service) GenerateTrip(ctx context.Context, req TripRequest) ([]store.PackingCategory, []store.DayOutfit) {
	items := s.Store.ListItems("")
	if len(items) == 0 {
		return nil, nil
	}

	forecast := s.fetchWeather(ctx, req.PlanRequest)
	outfitItems := store.ToOutfitItems(items)

	laundry := req.Laundry == "yes"
	solution := outfit.SolvePacking(outfitItems, req.Days, laundry, req.Luggage, forecast)

	allItems := store.ItemsByID(items)
	packing := make([]store.PackingCategory, len(solution.Packing.Categories))
	for i, cat := range solution.Packing.Categories {
		var names []string
		for _, id := range cat.ItemIDs {
			if item, ok := allItems[id]; ok {
				names = append(names, item.Name)
			}
		}
		packing[i] = store.PackingCategory{Label: cat.Label, Items: names}
	}

	dayOutfits := make([]store.DayOutfit, len(solution.DayOutfits))
	for i, d := range solution.DayOutfits {
		desc := describeOutfit(allItems, d.ItemIDs)
		dayOutfits[i] = store.DayOutfit{
			Day:         d.Day,
			Label:       d.Label,
			Description: desc,
			ItemIDs:     d.ItemIDs,
		}
	}

	// Persist day outfits
	for _, o := range solution.DayOutfits {
		if len(o.ItemIDs) == 0 {
			continue
		}
		_, _ = s.Store.SaveOutfitSet(ctx, store.Outfit{
			ID:      uuid.New().String(),
			Label:   o.Label,
			ItemIDs: o.ItemIDs,
			Why:     describeOutfit(allItems, o.ItemIDs),
		})
	}

	return packing, dayOutfits
}

func (s *Service) fetchWeather(ctx context.Context, req PlanRequest) []weather.DailyForecast {
	days := req.Days
	if days < 1 {
		days = 1
	}

	geo, err := weather.Geocode(ctx, req.Location)
	if err != nil {
		return nil
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		start = time.Now()
	}

	forecast, err := weather.GetForecast(ctx, geo.Lat, geo.Lng, start, days)
	if err != nil {
		return nil
	}
	return forecast
}

func (s *Service) persistOutfits(ctx context.Context, all []store.Item, generated []ai.GeneratedOutfit) []store.Outfit {
	var outfits []store.Outfit
	for _, g := range generated {
		o := store.Outfit{
			ID:      uuid.New().String(),
			Label:   g.Label,
			Score:   g.Score,
			Why:     g.Why,
			ItemIDs: g.ItemIDs,
			Items:   store.OutfitItemsFromIDs(all, g.ItemIDs),
		}
		saved, err := s.Store.SaveOutfitSet(ctx, o)
		if err == nil {
			o.ID = saved.ID
		}
		outfits = append(outfits, o)
	}
	return outfits
}

func ruleOutfits(profile outfit.Profile, items []outfit.Item, forecast []weather.DailyForecast, plan outfit.Plan) []ai.GeneratedOutfit {
	built := outfit.BuildOutfits(items, forecast, plan, profile)
	out := make([]ai.GeneratedOutfit, len(built))
	for i, o := range built {
		out[i] = ai.GeneratedOutfit{
			ItemIDs: o.ItemIDs,
			Label:   o.Label,
			Why:     o.Why,
			Score:   o.Score,
		}
	}
	return out
}

func toOutfitPlan(req PlanRequest) outfit.Plan {
	activities := splitActivities(req.Activities)
	formality := req.Formality
	if formality == "" {
		formality = "casual"
	}
	lookGoal := req.LookGoal
	if lookGoal == "" {
		lookGoal = "balanced"
	}
	days := req.Days
	if days < 1 {
		days = 1
	}
	return outfit.Plan{
		Location:   req.Location,
		Days:       days,
		Activities: activities,
		Formality:  formality,
		LookGoal:   lookGoal,
	}
}

func ToOutfitProfile(p store.StyleProfile) outfit.Profile {
	return store.ToOutfitProfile(p)
}

func splitActivities(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func describeOutfit(byID map[string]store.Item, ids []string) string {
	var parts []string
	for _, id := range ids {
		if item, ok := byID[id]; ok {
			parts = append(parts, item.Name)
		}
	}
	if len(parts) == 0 {
		return "Outfit from your closet"
	}
	return strings.Join(parts, " + ")
}
