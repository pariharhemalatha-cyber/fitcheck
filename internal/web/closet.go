package web

import (
	"bytes"
	"html/template"
	"io"
)

func renderPage(w io.Writer, data PageData, contentTmpl *template.Template, contentData any) error {
	var buf bytes.Buffer
	if err := contentTmpl.Execute(&buf, contentData); err != nil {
		return err
	}
	data.Content = template.HTML(buf.String())
	return RenderLayout(w, data)
}

var closetContent = template.Must(template.New("closet").Parse(`
<div class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold">My Closet</h1>
      <p class="text-stone-500 text-sm mt-1">Upload photos of your clothes. AI will tag them automatically.</p>
    </div>
    <button onclick="document.getElementById('upload-modal').classList.remove('hidden')"
      class="bg-stone-900 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-stone-800">
      + Add item
    </button>
  </div>

  <div class="flex gap-2 flex-wrap">
    {{range .Categories}}
    <a href="/closet?category={{.Slug}}"
      class="px-3 py-1 rounded-full text-sm border {{if eq $.ActiveCategory .Slug}}bg-stone-900 text-white border-stone-900{{else}}border-stone-300 text-stone-600 hover:border-stone-500{{end}}">
      {{.Emoji}} {{.Label}}
    </a>
    {{end}}
  </div>

  {{if .Items}}
  <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
    {{range .Items}}
    <a href="/closet/items/{{.ID}}" class="bg-white border border-stone-200 rounded-xl overflow-hidden hover:border-stone-400 transition group">
      <div class="aspect-square bg-stone-100 flex items-center justify-center overflow-hidden">
        {{if .ImageURL}}
        <img src="/uploads/{{.ImageURL}}" alt="{{.Name}}" class="w-full h-full object-cover group-hover:scale-105 transition duration-200">
        {{else}}
        <span class="text-4xl">{{if .Emoji}}{{.Emoji}}{{else}}👕{{end}}</span>
        {{end}}
      </div>
      <div class="p-3">
        <div class="font-medium text-sm truncate">{{.Name}}</div>
        <div class="text-xs text-stone-500">{{.Category}} · {{.MainColor}}</div>
      </div>
    </a>
    {{end}}
  </div>
  {{else}}
  <div class="text-center py-16 bg-white border border-dashed border-stone-300 rounded-2xl">
    <div class="text-4xl mb-3">👕</div>
    <p class="text-stone-500 text-sm">No items yet. Add your first piece of clothing.</p>
    <button onclick="document.getElementById('upload-modal').classList.remove('hidden')"
      class="mt-4 bg-stone-900 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-stone-800">
      Upload first item
    </button>
  </div>
  {{end}}
</div>

<div id="upload-modal" class="hidden fixed inset-0 bg-black/40 flex items-center justify-center z-50">
  <div class="bg-white rounded-2xl p-6 w-full max-w-md mx-4 shadow-xl">
    <h2 class="font-semibold text-lg mb-4">Add clothing item</h2>
    <form action="/closet/items" method="POST" enctype="multipart/form-data" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-stone-700 mb-1">Photo</label>
        <input type="file" name="photo" accept="image/*" required
          class="w-full text-sm text-stone-600 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:bg-stone-100 file:text-stone-700">
      </div>
      <div>
        <label class="block text-sm font-medium text-stone-700 mb-1">Category</label>
        <select name="category" class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
          <option value="tshirt">T-shirt</option>
          <option value="shirt">Shirt</option>
          <option value="pants">Pants</option>
          <option value="shorts">Shorts</option>
          <option value="jacket">Jacket</option>
          <option value="shoes">Shoes</option>
          <option value="accessory">Accessory</option>
        </select>
      </div>
      <div>
        <label class="block text-sm font-medium text-stone-700 mb-1">Name (optional)</label>
        <input type="text" name="name" placeholder="White oversized tee"
          class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm">
      </div>
      <div class="flex gap-3 pt-2">
        <button type="button" onclick="document.getElementById('upload-modal').classList.remove('hidden'); this.closest('#upload-modal').classList.add('hidden')"
          class="flex-1 border border-stone-300 rounded-lg py-2 text-sm hover:bg-stone-50">Cancel</button>
        <button type="submit" class="flex-1 bg-stone-900 text-white rounded-lg py-2 text-sm font-medium hover:bg-stone-800">Upload</button>
      </div>
    </form>
  </div>
</div>`))

type Category struct {
	Slug  string
	Label string
	Emoji string
}

type ClosetItem struct {
	ID        string
	Name      string
	Category  string
	MainColor string
	ImageURL  string
	Emoji     string
}

type ClosetPageData struct {
	ActiveCategory string
	Categories     []Category
	Items          []ClosetItem
}

var DefaultCategories = []Category{
	{Slug: "all", Label: "All", Emoji: "🗂️"},
	{Slug: "tshirt", Label: "T-shirts", Emoji: "👕"},
	{Slug: "shirt", Label: "Shirts", Emoji: "👔"},
	{Slug: "pants", Label: "Pants", Emoji: "👖"},
	{Slug: "shorts", Label: "Shorts", Emoji: "🩳"},
	{Slug: "jacket", Label: "Jackets", Emoji: "🧥"},
	{Slug: "shoes", Label: "Shoes", Emoji: "👟"},
	{Slug: "accessory", Label: "Accessories", Emoji: "🧢"},
}

func RenderCloset(w io.Writer, page PageData, data ClosetPageData) error {
	return renderPage(w, page, closetContent, data)
}
