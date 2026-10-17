import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

// Interface embarquée dans l'exécutable par Wails (frontend/dist).
export default defineConfig({
  plugins: [svelte(), tailwindcss()],
})
