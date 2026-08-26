package web

import (
	"html/template"
	"io"
)

type PageData struct {
	Title       string
	ActiveNav   string
	Content     template.HTML
	Flash       string
	SupabaseURL string
	SupabaseKey string
}

var layoutTmpl = template.Must(template.New("layout").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}} · FitCheck</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700;1,9..40,400&display=swap" rel="stylesheet">
  <style>
    body { font-family: 'DM Sans', system-ui, sans-serif; }
  </style>
</head>
<body class="bg-stone-50 text-stone-900 min-h-screen">
  <nav class="border-b border-stone-200 bg-white/80 backdrop-blur sticky top-0 z-50">
    <div class="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
      <a href="/" class="font-semibold text-lg tracking-tight">FitCheck</a>
      <div class="flex gap-1 text-sm">
        <a href="/closet" class="px-3 py-1.5 rounded-lg {{if eq .ActiveNav "closet"}}bg-stone-900 text-white{{else}}text-stone-600 hover:bg-stone-100{{end}}">Closet</a>
        <a href="/style" class="px-3 py-1.5 rounded-lg {{if eq .ActiveNav "style"}}bg-stone-900 text-white{{else}}text-stone-600 hover:bg-stone-100{{end}}">My Style</a>
        <a href="/plan" class="px-3 py-1.5 rounded-lg {{if eq .ActiveNav "plan"}}bg-stone-900 text-white{{else}}text-stone-600 hover:bg-stone-100{{end}}">Plan</a>
        <a href="/outfits" class="px-3 py-1.5 rounded-lg {{if eq .ActiveNav "outfits"}}bg-stone-900 text-white{{else}}text-stone-600 hover:bg-stone-100{{end}}">Outfits</a>
        <a href="/fitcheck" class="px-3 py-1.5 rounded-lg {{if eq .ActiveNav "fitcheck"}}bg-stone-900 text-white{{else}}text-stone-600 hover:bg-stone-100{{end}}">Fit Check</a>
        <a href="/trip" class="px-3 py-1.5 rounded-lg {{if eq .ActiveNav "trip"}}bg-stone-900 text-white{{else}}text-stone-600 hover:bg-stone-100{{end}}">Trip</a>
      </div>
    </div>
  </nav>
  {{if .Flash}}
  <div class="max-w-5xl mx-auto px-4 pt-4">
    <div class="bg-emerald-50 border border-emerald-200 text-emerald-800 text-sm px-4 py-2 rounded-lg">{{.Flash}}</div>
  </div>
  {{end}}
  <main class="max-w-5xl mx-auto px-4 py-8">
    {{.Content}}
  </main>
</body>
</html>`))

func RenderLayout(w io.Writer, data PageData) error {
	return layoutTmpl.Execute(w, data)
}
