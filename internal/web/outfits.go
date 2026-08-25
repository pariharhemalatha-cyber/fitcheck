package web

import (
	"html/template"
	"io"
)

var outfitsContent = template.Must(template.New("outfits").Parse(`
<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-bold">Outfits</h1>
    {{if .Location}}
    <p class="text-stone-500 text-sm mt-1">{{.Location}} · {{.Days}} day(s) · {{.Activities}}</p>
    {{else}}
    <p class="text-stone-500 text-sm mt-1">Outfit recommendations from your closet.</p>
    {{end}}
  </div>

  {{if .Outfits}}
  <div class="space-y-4">
    {{range .Outfits}}
    <div class="bg-white border border-stone-200 rounded-2xl p-5 shadow-sm">
      <div class="flex items-start justify-between mb-4">
        <div>
          <h2 class="font-semibold">{{.Label}}</h2>
          {{if .Score}}<span class="text-xs text-stone-500">Score: {{.Score}}/10</span>{{end}}
        </div>
        <form action="/outfits/{{.ID}}/wear" method="POST">
          <button type="submit" class="text-xs bg-stone-100 hover:bg-stone-200 px-3 py-1.5 rounded-lg font-medium">Wear this</button>
        </form>
      </div>
      <div class="flex gap-3 flex-wrap mb-4">
        {{range .Items}}
        <div class="text-center">
          <div class="w-20 h-20 bg-stone-100 rounded-xl flex items-center justify-center overflow-hidden">
            {{if .ImageURL}}
            <img src="/uploads/{{.ImageURL}}" alt="{{.Name}}" class="w-full h-full object-cover">
            {{else}}
            <span class="text-2xl">{{if .Emoji}}{{.Emoji}}{{else}}👕{{end}}</span>
            {{end}}
          </div>
          <div class="text-xs text-stone-600 mt-1 max-w-[80px] truncate">{{.Name}}</div>
        </div>
        {{end}}
      </div>
      {{if .Why}}<p class="text-sm text-stone-600 bg-stone-50 rounded-lg px-3 py-2">{{.Why}}</p>{{end}}
      <div class="flex gap-2 mt-3">
        <a href="/outfits/{{.ID}}/variant?type=more_casual" class="text-xs border border-stone-300 px-3 py-1 rounded-lg hover:bg-stone-50">More casual</a>
        <a href="/outfits/{{.ID}}/variant?type=different_shoes" class="text-xs border border-stone-300 px-3 py-1 rounded-lg hover:bg-stone-50">Different shoes</a>
        <a href="/outfits/{{.ID}}/variant?type=try_another" class="text-xs border border-stone-300 px-3 py-1 rounded-lg hover:bg-stone-50">Try another</a>
      </div>
    </div>
    {{end}}
  </div>
  {{else}}
  <div class="text-center py-16 bg-white border border-dashed border-stone-300 rounded-2xl">
    <div class="text-4xl mb-3">✨</div>
    <p class="text-stone-500 text-sm">No outfits yet. Add clothes to your closet and create a plan.</p>
    <div class="flex gap-3 justify-center mt-4">
      <a href="/closet" class="border border-stone-300 px-4 py-2 rounded-lg text-sm hover:bg-stone-50">Go to Closet</a>
      <a href="/plan" class="bg-stone-900 text-white px-4 py-2 rounded-lg text-sm hover:bg-stone-800">Create a plan</a>
    </div>
  </div>
  {{end}}
</div>`))

type OutfitItem struct {
	Name     string
	ImageURL string
	Emoji    string
}

type Outfit struct {
	ID     string
	Label  string
	Score  float64
	Why    string
	Items  []OutfitItem
}

type OutfitsPageData struct {
	Location   string
	Days       int
	Activities string
	Outfits    []Outfit
}

func RenderOutfits(w io.Writer, page PageData, data OutfitsPageData) error {
	return renderPage(w, page, outfitsContent, data)
}
