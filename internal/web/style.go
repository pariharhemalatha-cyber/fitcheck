package web

import (
	"html/template"
	"io"
)

var styleContent = template.Must(template.New("style").Parse(`
<div class="max-w-xl space-y-6">
  <div>
    <h1 class="text-2xl font-bold">My Style</h1>
    <p class="text-stone-500 text-sm mt-1">Tell us how you like to dress. We'll use this every time we recommend outfits.</p>
  </div>

  <form action="/style" method="POST" class="bg-white border border-stone-200 rounded-2xl p-6 space-y-5 shadow-sm">
    <div>
      <label class="block text-sm font-medium text-stone-700 mb-2">Primary style</label>
      <div class="flex flex-wrap gap-2">
        {{range .Styles}}
        <label class="cursor-pointer">
          <input type="radio" name="style_primary" value="{{.}}" class="sr-only peer" {{if eq $.Profile.StylePrimary .}}checked{{end}}>
          <span class="block px-3 py-1.5 rounded-full text-sm border border-stone-300 peer-checked:bg-stone-900 peer-checked:text-white peer-checked:border-stone-900 hover:border-stone-500 transition">{{.}}</span>
        </label>
        {{end}}
      </div>
    </div>

    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">I like</label>
      <textarea name="likes" rows="3" placeholder="Dark colors, oversized shirts, sneakers, neutral combinations..."
        class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm resize-none">{{.Profile.Likes}}</textarea>
    </div>

    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">I don't like</label>
      <textarea name="dislikes" rows="3" placeholder="Bright colors, skinny pants, too many patterns..."
        class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm resize-none">{{.Profile.Dislikes}}</textarea>
    </div>

    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Comfort vs. photo-ready (1–10)</label>
      <div class="flex gap-4">
        <div class="flex-1">
          <label class="text-xs text-stone-500">Comfort</label>
          <input type="range" name="comfort_bias" min="1" max="10" value="{{.Profile.ComfortBias}}" class="w-full">
        </div>
        <div class="flex-1">
          <label class="text-xs text-stone-500">Look good in photos</label>
          <input type="range" name="photo_look_bias" min="1" max="10" value="{{.Profile.PhotoLookBias}}" class="w-full">
        </div>
      </div>
    </div>

    <div>
      <label class="block text-sm font-medium text-stone-700 mb-1">Don't repeat same top within (days)</label>
      <input type="number" name="no_repeat_top_days" min="1" max="14" value="{{.Profile.NoRepeatTopDays}}"
        class="w-24 border border-stone-300 rounded-lg px-3 py-2 text-sm">
    </div>

    <button type="submit" class="w-full bg-stone-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-stone-800">
      Save preferences
    </button>
  </form>
</div>`))

type StyleProfile struct {
	StylePrimary    string
	Likes           string
	Dislikes        string
	ComfortBias     int
	PhotoLookBias   int
	NoRepeatTopDays int
}

type StylePageData struct {
	Profile StyleProfile
	Styles  []string
}

var DefaultStyles = []string{"Casual", "Smart casual", "Streetwear", "Minimal", "Formal", "Athletic"}

func RenderStyle(w io.Writer, page PageData, data StylePageData) error {
	return renderPage(w, page, styleContent, data)
}
