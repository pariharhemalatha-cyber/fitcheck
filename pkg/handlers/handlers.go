package handlers

import (
	"net/http"
	"strconv"

	"github.com/ashokparihar/fitcheck/pkg/ai"
	"github.com/ashokparihar/fitcheck/pkg/db"
	"github.com/ashokparihar/fitcheck/pkg/service"
	"github.com/ashokparihar/fitcheck/pkg/storage"
	"github.com/ashokparihar/fitcheck/pkg/store"
	"github.com/ashokparihar/fitcheck/pkg/web"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store   store.Store
	svc     *service.Service
	storage storage.Storage
}

func New(s store.Store, uploads storage.Storage, aiClient *ai.Client) *Handler {
	return &Handler{
		store:   s,
		svc:     &service.Service{Store: s, AI: aiClient},
		storage: uploads,
	}
}

func (h *Handler) imgSrc(storagePath string) string {
	return storage.PublicSrc(storagePath, h.storage)
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	_ = web.RenderHome(w, web.PageData{Title: "Home", ActiveNav: ""})
}

func (h *Handler) Closet(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		category = "all"
	}

	items := h.store.ListItems(category)
	closetItems := make([]web.ClosetItem, len(items))
	for i, item := range items {
		closetItems[i] = toClosetItem(item, h.storage)
	}

	page := web.PageData{Title: "My Closet", ActiveNav: "closet"}
	if flash := r.URL.Query().Get("flash"); flash != "" {
		page.Flash = flash
	}

	_ = web.RenderCloset(w, page, web.ClosetPageData{
		ActiveCategory: category,
		Categories:     web.DefaultCategories,
		Items:          closetItems,
	})
}

func (h *Handler) ItemDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if r.Method == http.MethodPost {
		h.updateItem(w, r, id)
		return
	}

	item, err := h.store.GetItem(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	page := web.PageData{Title: item.Name, ActiveNav: "closet"}
	if flash := r.URL.Query().Get("flash"); flash != "" {
		page.Flash = flash
	}

	_ = web.RenderItemDetail(w, page, web.ItemDetailPageData{
		Item:       toItemDetail(item, h.storage),
		Categories: web.DefaultCategories[1:],
	})
}

func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request, id string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	item, err := h.store.GetItem(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	formality, _ := strconv.Atoi(r.FormValue("formality"))
	item.Name = r.FormValue("name")
	item.Category = r.FormValue("category")
	item.MainColor = r.FormValue("main_color")
	item.Pattern = r.FormValue("pattern")
	item.Material = r.FormValue("material")
	item.Formality = formality
	item.SeasonTags = db.ParseCommaList(r.FormValue("season_tags"))
	item.ActivityTags = db.ParseCommaList(r.FormValue("activity_tags"))

	if err := h.store.UpdateItem(item); err != nil {
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/closet/items/"+id+"?flash=Saved", http.StatusSeeOther)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	category := r.FormValue("category")
	if category == "" {
		category = "tshirt"
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo required", http.StatusBadRequest)
		return
	}
	storagePath, err := h.storage.Save(file, header)
	if err != nil {
		http.Error(w, "failed to save upload", http.StatusInternalServerError)
		return
	}

	fsPath, err := h.storage.LocalPath(storagePath)
	if err != nil {
		http.Error(w, "failed to prepare image for analysis", http.StatusInternalServerError)
		return
	}

	attrs, _ := ai.AnalyzeItem(r.Context(), h.svc.AI, fsPath)
	if attrs.Category != "" {
		category = attrs.Category
	}
	if name == "" && attrs.Name != "" {
		name = attrs.Name
	}

	_, err = h.store.AddItemWithAttrs(name, category, storagePath, store.ItemAttrs{
		Name:            name,
		Category:        category,
		MainColor:       attrs.MainColor,
		SecondaryColors: attrs.SecondaryColors,
		Pattern:         attrs.Pattern,
		Material:        attrs.Material,
		Fit:             attrs.Fit,
		Formality:       attrs.Formality,
		SeasonTags:      attrs.SeasonTags,
		RainOK:          attrs.RainOK,
		ActivityTags:    attrs.ActivityTags,
		VibeTags:        attrs.VibeTags,
	})
	if err != nil {
		http.Error(w, "failed to add item", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/closet?flash=Item+added+and+analyzed", http.StatusSeeOther)
}

func (h *Handler) Style(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.saveStyle(w, r)
		return
	}

	profile := h.store.GetProfile()
	page := web.PageData{Title: "My Style", ActiveNav: "style"}
	if flash := r.URL.Query().Get("flash"); flash != "" {
		page.Flash = flash
	}

	_ = web.RenderStyle(w, page, web.StylePageData{
		Profile: web.StyleProfile{
			StylePrimary:    profile.StylePrimary,
			Likes:           profile.Likes,
			Dislikes:        profile.Dislikes,
			ComfortBias:     profile.ComfortBias,
			PhotoLookBias:   profile.PhotoLookBias,
			NoRepeatTopDays: profile.NoRepeatTopDays,
		},
		Styles: web.DefaultStyles,
	})
}

func (h *Handler) saveStyle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	comfort, _ := strconv.Atoi(r.FormValue("comfort_bias"))
	photo, _ := strconv.Atoi(r.FormValue("photo_look_bias"))
	noRepeat, _ := strconv.Atoi(r.FormValue("no_repeat_top_days"))

	if err := h.store.SaveProfile(store.StyleProfile{
		StylePrimary:    r.FormValue("style_primary"),
		Likes:           r.FormValue("likes"),
		Dislikes:        r.FormValue("dislikes"),
		ComfortBias:     comfort,
		PhotoLookBias:   photo,
		NoRepeatTopDays: noRepeat,
	}); err != nil {
		http.Error(w, "failed to save profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/style?flash=Preferences+saved", http.StatusSeeOther)
}

func (h *Handler) Plan(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 1
	}

	_ = web.RenderPlan(w, web.PageData{Title: "Plan", ActiveNav: "plan"}, web.PlanPageData{
		Location:   location,
		StartDate:  r.URL.Query().Get("start_date"),
		Days:       days,
		Activities: r.URL.Query().Get("activities"),
		Formality:  r.URL.Query().Get("formality"),
		LookGoal:   r.URL.Query().Get("look_goal"),
	})
}

func (h *Handler) Outfits(w http.ResponseWriter, r *http.Request) {
	req := h.planFromQuery(r)
	outfits := h.svc.GenerateOutfits(r.Context(), req)

	webOutfits := make([]web.Outfit, len(outfits))
	for i, o := range outfits {
		items := make([]web.OutfitItem, len(o.ItemIDs))
		byID := store.ItemsByID(h.store.ListItems(""))
		for j, id := range o.ItemIDs {
			item := byID[id]
			items[j] = web.OutfitItem{
				Name:     item.Name,
				Emoji:    item.Emoji,
				ImageURL: h.imgSrc(item.StoragePath),
			}
		}
		webOutfits[i] = web.Outfit{
			ID: o.ID, Label: o.Label, Score: o.Score, Why: o.Why, Items: items,
		}
	}

	_ = web.RenderOutfits(w, web.PageData{Title: "Outfits", ActiveNav: "outfits"}, web.OutfitsPageData{
		Location: req.Location, Days: req.Days, Activities: req.Activities, Outfits: webOutfits,
	})
}

func (h *Handler) WearOutfit(w http.ResponseWriter, r *http.Request) {
	outfitID := chi.URLParam(r, "id")
	if err := h.store.LogWear(outfitID, nil, "recommended"); err != nil {
		http.Error(w, "failed to log wear", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/outfits?flash=Outfit+logged", http.StatusSeeOther)
}

func (h *Handler) Trip(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 5
	}
	laundry := r.URL.Query().Get("laundry")
	if laundry == "" {
		laundry = "no"
	}
	luggage := r.URL.Query().Get("luggage")
	if luggage == "" {
		luggage = "carry_on"
	}

	data := web.TripPageData{
		Location: location, StartDate: r.URL.Query().Get("start_date"), Days: days,
		Activities: r.URL.Query().Get("activities"), Laundry: laundry, Luggage: luggage,
	}

	if location != "" {
		packing, dayOutfits := h.svc.GenerateTrip(r.Context(), service.TripRequest{
			PlanRequest: h.planFromQuery(r), Laundry: laundry, Luggage: luggage,
		})
		data.HasPlan = true
		for _, cat := range packing {
			data.PackingCategories = append(data.PackingCategories, web.PackingCategory{Label: cat.Label, Items: cat.Items})
		}
		for _, d := range dayOutfits {
			data.DayOutfits = append(data.DayOutfits, web.DayOutfit{Day: d.Day, Description: d.Description})
		}
	}

	_ = web.RenderTrip(w, web.PageData{Title: "Trip", ActiveNav: "trip"}, data)
}

func (h *Handler) FitCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.runFitCheck(w, r)
		return
	}

	items := h.store.ListItems("")
	closetItems := make([]web.FitCheckClosetItem, len(items))
	for i, item := range items {
		closetItems[i] = web.FitCheckClosetItem{
			ID: item.ID, Name: item.Name, Category: item.Category,
			ImageURL: h.imgSrc(item.StoragePath), Emoji: item.Emoji,
		}
	}

	_ = web.RenderFitCheck(w, web.PageData{Title: "Fit Check", ActiveNav: "fitcheck"}, web.FitCheckPageData{
		ClosetItems: closetItems,
	})
}

func (h *Handler) runFitCheck(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	itemIDs := r.Form["item_ids"]
	if len(itemIDs) == 0 {
		http.Error(w, "select at least one item", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("selfie")
	if err != nil {
		http.Error(w, "selfie required", http.StatusBadRequest)
		return
	}
	storagePath, err := h.storage.Save(file, header)
	if err != nil {
		http.Error(w, "failed to save selfie", http.StatusInternalServerError)
		return
	}

	fsPath, err := h.storage.LocalPath(storagePath)
	if err != nil {
		http.Error(w, "failed to prepare selfie", http.StatusInternalServerError)
		return
	}

	allItems := store.ToOutfitItems(h.store.ListItems(""))
	result, _ := ai.AnalyzeFitCheck(r.Context(), h.svc.AI, fsPath, itemIDs, allItems)

	byID := store.ItemsByID(h.store.ListItems(""))
	swaps := make([]web.SwapSuggestion, len(result.SuggestedSwaps))
	for i, s := range result.SuggestedSwaps {
		from, to := s.FromItemID, s.ToItemID
		if item, ok := byID[s.FromItemID]; ok {
			from = item.Name
		}
		if item, ok := byID[s.ToItemID]; ok {
			to = item.Name
		}
		swaps[i] = web.SwapSuggestion{From: from, To: to, Reason: s.Reason}
	}

	critique := []string{result.Critique}
	if result.Critique == "" {
		critique = nil
	}

	_ = h.store.SaveFitCheck(storagePath, itemIDs, result.Score, result.Critique)

	fitResult := web.FitCheckResult{
		Score: result.Score, Critique: critique, Swaps: swaps,
		SelfieURL: h.imgSrc(storagePath),
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		_ = web.RenderFitCheckResult(w, fitResult)
		return
	}

	items := h.store.ListItems("")
	closetItems := make([]web.FitCheckClosetItem, len(items))
	for i, item := range items {
		closetItems[i] = web.FitCheckClosetItem{
			ID: item.ID, Name: item.Name, Category: item.Category,
			ImageURL: h.imgSrc(item.StoragePath), Emoji: item.Emoji,
		}
	}
	_ = web.RenderFitCheck(w, web.PageData{Title: "Fit Check", ActiveNav: "fitcheck"}, web.FitCheckPageData{
		ClosetItems: closetItems, Result: &fitResult,
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) planFromQuery(r *http.Request) service.PlanRequest {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 1
	}
	formality := r.URL.Query().Get("formality")
	if formality == "" {
		formality = "casual"
	}
	lookGoal := r.URL.Query().Get("look_goal")
	if lookGoal == "" {
		lookGoal = "balanced"
	}
	return service.PlanRequest{
		Location: r.URL.Query().Get("location"), StartDate: r.URL.Query().Get("start_date"),
		Days: days, Activities: r.URL.Query().Get("activities"),
		Formality: formality, LookGoal: lookGoal,
	}
}

func toClosetItem(item store.Item, s storage.Storage) web.ClosetItem {
	return web.ClosetItem{
		ID: item.ID, Name: item.Name, Category: item.Category,
		MainColor: item.MainColor, Emoji: item.Emoji,
		ImageURL: storage.PublicSrc(item.StoragePath, s),
	}
}

func toItemDetail(item store.Item, s storage.Storage) web.ItemDetail {
	return web.ItemDetail{
		ID: item.ID, Name: item.Name, Category: item.Category,
		ImageURL: storage.PublicSrc(item.StoragePath, s), Emoji: item.Emoji,
		MainColor: item.MainColor, Pattern: item.Pattern, Material: item.Material,
		Formality: item.Formality, SeasonTags: item.SeasonTags, ActivityTags: item.ActivityTags,
		SeasonTagsText: db.JoinCommaList(item.SeasonTags), ActivityTagsText: db.JoinCommaList(item.ActivityTags),
	}
}
