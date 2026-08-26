package web

import (
	"bytes"
	"html/template"
	"io"
)

var homeContent = template.Must(template.New("home").Parse(`
<div class="space-y-10">
  <section class="text-center space-y-4 pt-8">
    <h1 class="text-4xl font-bold tracking-tight">What should I wear?</h1>
    <p class="text-stone-500 max-w-lg mx-auto">Your personal outfit decision engine. Upload your clothes, tell us where you're going, and get outfits made only from what you own.</p>
  </section>

  <section class="bg-white border border-stone-200 rounded-2xl p-6 shadow-sm max-w-xl mx-auto">
    <h2 class="font-semibold text-lg mb-4">✨ Style Me</h2>
    <form action="/plan" method="GET" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-stone-700 mb-1">Where are you going?</label>
        <input name="location" type="text" placeholder="Toronto, Chicago, friend's birthday tonight..."
          class="w-full border border-stone-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-stone-400">
      </div>
      <button type="submit" class="w-full bg-stone-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-stone-800 transition">
        Plan my outfit →
      </button>
    </form>
  </section>

  <section class="grid grid-cols-2 md:grid-cols-4 gap-4">
    <a href="/closet" class="bg-white border border-stone-200 rounded-xl p-5 hover:border-stone-400 transition group">
      <div class="text-2xl mb-2">👕</div>
      <div class="font-medium group-hover:underline">My Closet</div>
      <div class="text-xs text-stone-500 mt-1">Upload & tag clothes</div>
    </a>
    <a href="/style" class="bg-white border border-stone-200 rounded-xl p-5 hover:border-stone-400 transition group">
      <div class="text-2xl mb-2">👤</div>
      <div class="font-medium group-hover:underline">My Style</div>
      <div class="text-xs text-stone-500 mt-1">Set preferences</div>
    </a>
    <a href="/plan" class="bg-white border border-stone-200 rounded-xl p-5 hover:border-stone-400 transition group">
      <div class="text-2xl mb-2">🌎</div>
      <div class="font-medium group-hover:underline">Plan</div>
      <div class="text-xs text-stone-500 mt-1">Occasion & weather</div>
    </a>
    <a href="/trip" class="bg-white border border-stone-200 rounded-xl p-5 hover:border-stone-400 transition group">
      <div class="text-2xl mb-2">🧳</div>
      <div class="font-medium group-hover:underline">Trip</div>
      <div class="text-xs text-stone-500 mt-1">Packing & day outfits</div>
    </a>
  </section>
</div>`))

func RenderHome(w io.Writer, data PageData) error {
	var buf bytes.Buffer
	if err := homeContent.Execute(&buf, nil); err != nil {
		return err
	}
	data.Content = template.HTML(buf.String())
	return RenderLayout(w, data)
}
