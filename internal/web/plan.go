package web

import (
	"html/template"
	"io"
)

var planContent = template.Must(template.New("plan").Parse(`
<div class="max-w-xl space-y-6">
  <div>
    <h1 class="text-2xl font-bold">Plan</h1>
    <p class="text-stone-500 text-sm mt-1">Where are you going and what are you doing?</p>
  </div>

  <form action="/outfits" method="GET" class="bg-white border border-stone-200 rounded-2xl p-6 space-y-4 shadow-sm">
    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Location</label>
      <input type="text" name="location" value="{{.Location}}" placeholder="Toronto, Chicago..."
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
      <input type="text" name="activities" value="{{.Activities}}" placeholder="Walking, sightseeing, restaurants..."
        class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
    </div>

    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Formality</label>
      <select name="formality" class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
        <option value="casual" {{if eq .Formality "casual"}}selected{{end}}>Casual</option>
        <option value="smart_casual" {{if eq .Formality "smart_casual"}}selected{{end}}>Smart casual</option>
        <option value="formal" {{if eq .Formality "formal"}}selected{{end}}>Formal</option>
      </select>
    </div>

    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Look goal</label>
      <select name="look_goal" class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
        <option value="comfort" {{if eq .LookGoal "comfort"}}selected{{end}}>Comfortable</option>
        <option value="photos" {{if eq .LookGoal "photos"}}selected{{end}}>Look good in photos</option>
        <option value="balanced" {{if eq .LookGoal "balanced"}}selected{{end}}>Balanced</option>
      </select>
    </div>

    <button type="submit" class="w-full bg-stone-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-stone-800">
      Generate outfits →
    </button>
  </form>
</div>`))

type PlanPageData struct {
	Location   string
	StartDate  string
	Days       int
	Activities string
	Formality  string
	LookGoal   string
}

func RenderPlan(w io.Writer, page PageData, data PlanPageData) error {
	return renderPage(w, page, planContent, data)
}
