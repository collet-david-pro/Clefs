// Serveur local du site de démonstration (captures de la documentation).
import http from 'node:http'
import fs from 'node:fs'
import path from 'node:path'

const root = path.join(import.meta.dirname, 'site')
const types = { '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css', '.woff2': 'font/woff2', '.svg': 'image/svg+xml', '.png': 'image/png' }
http.createServer((req, res) => {
  let p = decodeURIComponent(new URL(req.url, 'http://x').pathname)
  if (p === '/') p = '/index.html'
  const f = path.join(root, p)
  if (!f.startsWith(root) || !fs.existsSync(f)) { res.writeHead(404); return res.end() }
  res.writeHead(200, { 'Content-Type': types[path.extname(f)] || 'application/octet-stream' })
  fs.createReadStream(f).pipe(res)
}).listen(4729)
setTimeout(() => process.exit(0), 10 * 60 * 1000)
