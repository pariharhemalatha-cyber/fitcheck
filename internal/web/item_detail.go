package web

import (
	"html/template"
	"io"
)

var itemDetailContent = template.Must(template.New("item_detail").Parse(`
<div class="space-y-6">
  <div class="flex items-center gap-3">
    <a href="/closet" class="text-sm text-stone-500 hover:text-stone-900">← Back to closet</a>
  </div>

  <div class="grid md:grid-cols-2 gap-6">
    <div class="bg-white border border-stone-200 rounded-2xl overflow-hidden shadow-sm">
      <div class="aspect-square bg-stone-100 flex items-center justify-center overflow-hidden">
        {{if .Item.ImageURL}}
        <img src="{{.Item.ImageURL}}" alt="{{.Item.Name}}" class="w-full h-full object-cover">
        {{else}}
        <span class="text-6xl">{{if .Item.Emoji}}{{.Item.Emoji}}{{else}}👕{{end}}</span>
        {{end}}
      </div>
      <div class="p-5 border-t border-stone-100">
        <h1 class="text-xl font-bold">{{.Item.Name}}</h1>
        <p class="text-sm text-stone-500 mt-1 capitalize">{{.Item.Category}}</p>
      </div>
    </div>

    <div class="space-y-4">
      <div class="bg-white border border-stone-200 rounded-2xl p-5 shadow-sm">
        <h2 class="font-semibold mb-3">AI tags</h2>
        <dl class="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
          <div>
            <dt class="text-stone-500 text-xs uppercase tracking-wide">Color</dt>
            <dd class="font-medium mt-0.5">{{if .Item.MainColor}}{{.Item.MainColor}}{{else}}—{{end}}</dd>
          </div>
          <div>
            <dt class="text-stone-500 text-xs uppercase tracking-wide">Pattern</dt>
            <dd class="font-medium mt-0.5">{{if .Item.Pattern}}{{.Item.Pattern}}{{else}}—{{end}}</dd>
          </div>
          <div>
            <dt class="text-stone-500 text-xs uppercase tracking-wide">Material</dt>
            <dd class="font-medium mt-0.5">{{if .Item.Material}}{{.Item.Material}}{{else}}—{{end}}</dd>
          </div>
          <div>
            <dt class="text-stone-500 text-xs uppercase tracking-wide">Formality</dt>
            <dd class="font-medium mt-0.5">{{if .Item.Formality}}{{.Item.Formality}}/10{{else}}—{{end}}</dd>
          </div>
          <div class="col-span-2">
            <dt class="text-stone-500 text-xs uppercase tracking-wide">Season</dt>
            <dd class="font-medium mt-0.5">
              {{if .Item.SeasonTags}}
              <div class="flex flex-wrap gap-1.5 mt-1">
                {{range .Item.SeasonTags}}
                <span class="bg-stone-100 text-stone-700 px-2 py-0.5 rounded-full text-xs">{{.}}</span>
                {{end}}
              </div>
              {{else}}—{{end}}
            </dd>
          </div>
          <div class="col-span-2">
            <dt class="text-stone-500 text-xs uppercase tracking-wide">Activities</dt>
            <dd class="font-medium mt-0.5">
              {{if .Item.ActivityTags}}
              <div class="flex flex-wrap gap-1.5 mt-1">
                {{range .Item.ActivityTags}}
                <span class="bg-stone-100 text-stone-700 px-2 py-0.5 rounded-full text-xs">{{.}}</span>
                {{end}}
              </div>
              {{else}}—{{end}}
            </dd>
          </div>
        </dl>
      </div>

      <form action="/closet/items/{{.Item.ID}}" method="POST" class="bg-white border border-stone-200 rounded-2xl p-5 shadow-sm space-y-4">
        <h2 class="font-semibold">Edit tags</h2>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Name</label>
          <input type="text" name="name" value="{{.Item.Name}}"
            class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
        </div>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Category</label>
          <select name="category" class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
            {{range .Categories}}
            <option value="{{.Slug}}" {{if eq $.Item.Category .Slug}}selected{{end}}>{{.Label}}</option>
            {{end}}
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Main color</label>
          <input type="text" name="main_color" value="{{.Item.MainColor}}" placeholder="Navy, beige, white..."
            class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
        </div>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Pattern</label>
          <input type="text" name="pattern" value="{{.Item.Pattern}}" placeholder="Solid, striped, plaid..."
            class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
        </div>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Material</label>
          <input type="text" name="material" value="{{.Item.Material}}" placeholder="Cotton, denim, wool..."
            class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
        </div>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Formality (1–10)</label>
          <input type="number" name="formality" min="1" max="10" value="{{.Item.Formality}}"
            class="w-24 border border-stone-300 rounded-lg px-3 py-2 text-sm">
        </div>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Season tags</label>
          <input type="text" name="season_tags" value="{{.Item.SeasonTagsText}}" placeholder="Spring, summer, fall, winter"
            class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
          <p class="text-xs text-stone-500 mt-1">Comma-separated</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-stone-700 mb-1">Activity tags</label>
          <input type="text" name="activity_tags" value="{{.Item.ActivityTagsText}}" placeholder="Casual, office, travel, dinner..."
            class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
          <p class="text-xs text-stone-500 mt-1">Comma-separated</p>
        </div>

        <button type="submit" class="w-full bg-stone-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-stone-800">
          Save changes
        </button>
      </form>
    </div>
  </div>
</div>`))

type ItemDetail struct {
	ID               string
	Name             string
	Category         string
	ImageURL         string
	Emoji            string
	MainColor        string
	Pattern          string
	Material         string
	Formality        int
	SeasonTags       []string
	ActivityTags     []string
	SeasonTagsText   string
	ActivityTagsText string
}

type ItemDetailPageData struct {
	Item       ItemDetail
	Categories []Category
}

func RenderItemDetail(w io.Writer, page PageData, data ItemDetailPageData) error {
	return renderPage(w, page, itemDetailContent, data)
}
