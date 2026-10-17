import { mount } from 'svelte'
import './app.css'
import App from './App.svelte'
import { applyTheme } from './lib/state.svelte'

applyTheme()
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme)

export default mount(App, { target: document.getElementById('app')! })
