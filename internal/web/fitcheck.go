package web

import (
	"html/template"
	"io"
)

var fitCheckContent = template.Must(template.New("fitcheck").Parse(`
<style>.htmx-indicator{display:none}.htmx-request .htmx-indicator{display:block}</style>
<div class="space-y-6">
  <div>
    <h1 class="text-2xl font-bold">Fit Check</h1>
    <p class="text-stone-500 text-sm mt-1">Upload a mirror selfie wearing your outfit. We'll score the look and suggest swaps from your closet.</p>
  </div>

  <div class="grid md:grid-cols-2 gap-6">
    <div class="bg-white border border-stone-200 rounded-2xl p-6 shadow-sm space-y-5">
      <h2 class="font-semibold">Upload selfie</h2>
      <form action="/fitcheck" method="POST" enctype="multipart/form-data" class="space-y-4"
        hx-post="/fitcheck" hx-target="#fit-check-result" hx-swap="innerHTML" hx-indicator="#fit-check-loading">
        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Mirror selfie</label>
          <input type="file" name="selfie" accept="image/*" required
            class="w-full text-sm text-stone-600 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:bg-stone-100 file:text-stone-700">
        </div>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-2">Outfit items from closet</label>
          {{if .ClosetItems}}
          <div class="grid grid-cols-3 gap-2 max-h-64 overflow-y-auto pr-1">
            {{range .ClosetItems}}
            <label class="cursor-pointer group">
              <input type="checkbox" name="item_ids" value="{{.ID}}" class="sr-only peer">
              <div class="border border-stone-200 rounded-xl overflow-hidden peer-checked:ring-2 peer-checked:ring-stone-900 peer-checked:border-stone-900 hover:border-stone-400 transition">
                <div class="aspect-square bg-stone-100 flex items-center justify-center overflow-hidden">
                  {{if .ImageURL}}
                  <img src="/uploads/{{.ImageURL}}" alt="{{.Name}}" class="w-full h-full object-cover">
                  {{else}}
                  <span class="text-2xl">{{if .Emoji}}{{.Emoji}}{{else}}👕{{end}}</span>
                  {{end}}
                </div>
                <div class="p-1.5 text-center">
                  <div class="text-xs font-medium truncate">{{.Name}}</div>
                  <div class="text-[10px] text-stone-500">{{.Category}}</div>
                </div>
              </div>
            </label>
            {{end}}
          </div>
          {{else}}
          <div class="text-sm text-stone-500 bg-stone-50 border border-dashed border-stone-300 rounded-xl px-4 py-6 text-center">
            No items in your closet yet.
            <a href="/closet" class="block mt-2 text-stone-900 font-medium hover:underline">Add clothes first →</a>
          </div>
          {{end}}
        </div>

        <button type="submit" class="w-full bg-stone-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-stone-800 disabled:opacity-50"
          {{if not .ClosetItems}}disabled{{end}}>
          Check my fit →
        </button>
      </form>
      <div id="fit-check-loading" class="htmx-indicator text-sm text-stone-500 text-center">Analyzing your outfit…</div>
    </div>

    <div id="fit-check-result">
      {{if .Result}}
      {{template "fitCheckResult" .Result}}
      {{else}}
      <div class="bg-white border border-dashed border-stone-300 rounded-2xl p-8 text-center h-full flex flex-col items-center justify-center min-h-[280px]">
        <div class="text-4xl mb-3">📸</div>
        <p class="text-stone-500 text-sm max-w-xs">Upload a selfie and select what you're wearing to get your fit score.</p>
      </div>
      {{end}}
    </div>
  </div>
</div>`))

var fitCheckResultPartial = template.Must(fitCheckContent.New("fitCheckResult").Parse(`
<div class="bg-white border border-stone-200 rounded-2xl p-6 shadow-sm space-y-5 h-full">
  {{if .SelfieURL}}
  <div class="aspect-[3/4] max-h-48 bg-stone-100 rounded-xl overflow-hidden mx-auto">
    <img src="/uploads/{{.SelfieURL}}" alt="Your selfie" class="w-full h-full object-cover">
  </div>
  {{end}}

  <div class="text-center">
    <div class="text-xs uppercase tracking-wide text-stone-500 font-medium mb-1">Fit score</div>
    <div class="text-5xl font-bold tracking-tight">{{printf "%.1f" .Score}}<span class="text-2xl text-stone-400 font-normal">/10</span></div>
  </div>

  {{if .Critique}}
  <div>
    <h3 class="text-sm font-semibold text-stone-700 mb-2">Critique</h3>
    <ul class="space-y-2">
      {{range .Critique}}
      <li class="flex gap-2 text-sm text-stone-600">
        <span class="text-stone-400 shrink-0">•</span>
        <span>{{.}}</span>
      </li>
      {{end}}
    </ul>
  </div>
  {{end}}

  {{if .Swaps}}
  <div>
    <h3 class="text-sm font-semibold text-stone-700 mb-2">Suggested swaps</h3>
    <div class="space-y-2">
      {{range .Swaps}}
      <div class="bg-stone-50 border border-stone-200 rounded-xl px-4 py-3 text-sm">
        <div class="font-medium">{{.From}} → {{.To}}</div>
        {{if .Reason}}<div class="text-stone-500 text-xs mt-1">{{.Reason}}</div>{{end}}
      </div>
      {{end}}
    </div>
  </div>
  {{end}}

  <form action="/fitcheck" method="GET">
    <button type="submit" class="w-full border border-stone-300 rounded-lg py-2 text-sm hover:bg-stone-50">Check another outfit</button>
  </form>
</div>`))

type FitCheckClosetItem struct {
	ID       string
	Name     string
	Category string
	ImageURL string
	Emoji    string
}

type SwapSuggestion struct {
	From   string
	To     string
	Reason string
}

type FitCheckResult struct {
	Score     float64
	Critique  []string
	Swaps     []SwapSuggestion
	SelfieURL string
}

type FitCheckPageData struct {
	ClosetItems []FitCheckClosetItem
	Result      *FitCheckResult
}

func RenderFitCheck(w io.Writer, page PageData, data FitCheckPageData) error {
	return renderPage(w, page, fitCheckContent, data)
}

func RenderFitCheckResult(w io.Writer, result FitCheckResult) error {
	return fitCheckResultPartial.ExecuteTemplate(w, "fitCheckResult", result)
}
