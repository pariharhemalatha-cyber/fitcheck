package web

import (
	"html/template"
	"io"
)

var tripContent = template.Must(template.New("trip").Parse(`
<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-bold">Trip</h1>
    <p class="text-stone-500 text-sm mt-1">Packing plan and day-by-day outfits.</p>
  </div>

  <form action="/trip" method="GET" class="bg-white border border-stone-200 rounded-2xl p-6 space-y-4 shadow-sm max-w-xl">
    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Destination</label>
      <input type="text" name="location" value="{{.Location}}" placeholder="Chicago..."
        class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm" required>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-sm font-medium text-stone-700 mb-1">Start date</label>
        <input type="date" name="start_date" value="{{.StartDate}}"
          class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
      </div>
      <div>
        <label class="block text-sm font-medium text-stone-700 mb-1">Days</label>
        <input type="number" name="days" min="1" max="30" value="{{.Days}}"
          class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
      </div>
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Activities</label>
      <input type="text" name="activities" value="{{.Activities}}" placeholder="Walking, sightseeing, dinner..."
        class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-sm font-medium text-stone-700 mb-1">Laundry available?</label>
        <select name="laundry" class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
          <option value="no" {{if eq .Laundry "no"}}selected{{end}}>No</option>
          <option value="yes" {{if eq .Laundry "yes"}}selected{{end}}>Yes</option>
        </select>
      </div>
      <div>
        <label class="block text-sm font-medium text-stone-700 mb-1">Luggage</label>
        <select name="luggage" class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
          <option value="carry_on" {{if eq .Luggage "carry_on"}}selected{{end}}>Carry-on</option>
          <option value="checked" {{if eq .Luggage "checked"}}selected{{end}}>Checked bag</option>
          <option value="unlimited" {{if eq .Luggage "unlimited"}}selected{{end}}>No limit</option>
        </select>
      </div>
    </div>
    <button type="submit" class="w-full bg-stone-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-stone-800">
      Generate trip plan →
    </button>
  </form>

  {{if .HasPlan}}
  <div class="grid md:grid-cols-2 gap-6">
    <div class="bg-white border border-stone-200 rounded-2xl p-5 shadow-sm">
      <h2 class="font-semibold mb-4">🧳 Packing Plan</h2>
      {{range .PackingCategories}}
      <div class="mb-3">
        <div class="text-sm font-medium text-stone-700">{{.Label}}</div>
        <ul class="text-sm text-stone-600 mt-1 space-y-0.5">
          {{range .Items}}<li>· {{.}}</li>{{end}}
        </ul>
      </div>
      {{end}}
    </div>

    <div class="bg-white border border-stone-200 rounded-2xl p-5 shadow-sm">
      <h2 class="font-semibold mb-4">📅 Day-by-day</h2>
      <div class="space-y-3">
        {{range .DayOutfits}}
        <div class="border-b border-stone-100 pb-3 last:border-0">
          <div class="text-xs font-medium text-stone-500 mb-1">Day {{.Day}}</div>
          <div class="text-sm">{{.Description}}</div>
        </div>
        {{end}}
      </div>
    </div>
  </div>
  {{end}}
</div>`))

type PackingCategory struct {
	Label string
	Items []string
}

type DayOutfit struct {
	Day         int
	Description string
}

type TripPageData struct {
	Location          string
	StartDate         string
	Days              int
	Activities        string
	Laundry           string
	Luggage           string
	HasPlan           bool
	PackingCategories []PackingCategory
	DayOutfits        []DayOutfit
}

func RenderTrip(w io.Writer, page PageData, data TripPageData) error {
	return renderPage(w, page, tripContent, data)
}
