#!/usr/bin/env python3
"""Build a standalone HTML viewer for backend-architecture-guide.md.

Self-contained: pulls marked + mermaid + highlight.js from CDNs, renders
the Markdown client-side, and converts ```mermaid fences to live diagrams.
Run from documentation/docs/:  python3 _build_html.py
"""
from pathlib import Path

HERE = Path(__file__).parent
MD_PATH = HERE / "backend-architecture-guide.md"
OUT = HERE / "backend-architecture-guide.html"

TOP = """<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Backend Architecture Guide</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/styles/github.min.css" media="(prefers-color-scheme: light)">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/styles/github-dark.min.css" media="(prefers-color-scheme: dark)">
<style>
  :root {
    --bg: #ffffff; --fg: #1f2328; --muted: #57606a;
    --accent: #0969da; --border: #d0d7de; --code-bg: #f6f8fa;
    --quote-bg: #f6f8fa; --table-stripe: #f6f8fa;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #0d1117; --fg: #e6edf3; --muted: #8d96a0;
      --accent: #58a6ff; --border: #30363d; --code-bg: #161b22;
      --quote-bg: #161b22; --table-stripe: #161b22;
    }
  }
  html, body { background: var(--bg); color: var(--fg); }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", Arial, sans-serif;
    line-height: 1.65; max-width: 980px; margin: 0 auto; padding: 32px 28px 96px;
  }
  h1, h2, h3, h4 { line-height: 1.25; margin-top: 2em; margin-bottom: 0.6em; }
  h1 { font-size: 2em; border-bottom: 1px solid var(--border); padding-bottom: 0.3em; }
  h2 { font-size: 1.5em; border-bottom: 1px solid var(--border); padding-bottom: 0.3em; }
  h3 { font-size: 1.2em; }
  a { color: var(--accent); text-decoration: none; }
  a:hover { text-decoration: underline; }
  code { background: var(--code-bg); padding: 2px 6px; border-radius: 6px;
         font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
         font-size: 0.9em; }
  pre { background: var(--code-bg); padding: 16px; border-radius: 8px; overflow-x: auto; }
  pre code { background: transparent; padding: 0; font-size: 0.875em; }
  blockquote { border-left: 4px solid var(--border); margin: 0; padding: 4px 16px;
               color: var(--muted); background: var(--quote-bg); border-radius: 0 6px 6px 0; }
  table { border-collapse: collapse; margin: 1em 0; display: block; overflow-x: auto; }
  th, td { border: 1px solid var(--border); padding: 8px 12px; }
  th { background: var(--code-bg); }
  tbody tr:nth-child(even) { background: var(--table-stripe); }
  details { margin: 1em 0; padding: 8px 14px; border: 1px solid var(--border); border-radius: 6px; }
  details[open] { background: var(--quote-bg); }
  summary { cursor: pointer; font-weight: 600; }
  hr { border: none; border-top: 1px solid var(--border); margin: 2em 0; }
  .mermaid { background: var(--bg); text-align: center; margin: 1.2em 0; }
  /* Anchor link gutter on heading hover */
  h1:hover .anchor, h2:hover .anchor, h3:hover .anchor, h4:hover .anchor { opacity: 1; }
  .anchor { opacity: 0; margin-left: 6px; font-size: 0.7em; vertical-align: middle; }
  /* TOC nav (auto-collapsed on small screens) */
  #toc-toggle { position: fixed; top: 16px; right: 16px; z-index: 10;
                background: var(--code-bg); border: 1px solid var(--border);
                padding: 6px 10px; border-radius: 6px; cursor: pointer; font-size: 0.85em; }
</style>
</head>
<body>
<button id="toc-toggle" onclick="window.scrollTo({top:0,behavior:'smooth'})">↑ Top</button>
<article id="content">Loading…</article>

<script id="markdown-source" type="text/plain">
"""

BOTTOM = """
</script>

<script src="https://cdn.jsdelivr.net/npm/marked@12.0.2/marked.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/mermaid@10.9.0/dist/mermaid.min.js"></script>
<script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.9.0/build/highlight.min.js"></script>
<script>
  // Slugify for heading anchors (matches GitHub-ish behavior)
  function slug(text) {
    return String(text).toLowerCase()
      .replace(/[^\\w\\s-]/g, '')
      .trim().replace(/\\s+/g, '-');
  }

  const renderer = new marked.Renderer();
  const origCode = renderer.code.bind(renderer);
  renderer.code = function(code, lang, escaped) {
    if ((lang || '').toLowerCase() === 'mermaid') {
      return '<div class="mermaid">' + code + '</div>';
    }
    if (lang && hljs.getLanguage(lang)) {
      try { return '<pre><code class="hljs language-' + lang + '">'
        + hljs.highlight(code, { language: lang, ignoreIllegals: true }).value
        + '</code></pre>'; } catch(e) {}
    }
    return origCode(code, lang, escaped);
  };
  renderer.heading = function(text, level, raw) {
    const id = slug(raw);
    return '<h' + level + ' id="' + id + '">' + text
      + ' <a class="anchor" href="#' + id + '">#</a></h' + level + '>';
  };

  marked.setOptions({ renderer, gfm: true, breaks: false });

  const src = document.getElementById('markdown-source').textContent;
  document.getElementById('content').innerHTML = marked.parse(src);

  mermaid.initialize({
    startOnLoad: false,
    theme: window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'default',
    securityLevel: 'loose',
    flowchart: { htmlLabels: true, curve: 'basis' },
    sequence: { actorMargin: 50 }
  });
  mermaid.run({ querySelector: '.mermaid' });

  // Smooth-scroll anchor jumps
  document.querySelectorAll('a[href^="#"]').forEach(a => {
    a.addEventListener('click', e => {
      const id = a.getAttribute('href').slice(1);
      const el = document.getElementById(id);
      if (el) { e.preventDefault(); el.scrollIntoView({ behavior: 'smooth', block: 'start' });
                history.replaceState(null, '', '#' + id); }
    });
  });
</script>
</body>
</html>
"""

md = MD_PATH.read_text(encoding="utf-8")
if "</script>" in md:
    raise SystemExit("Markdown contains a literal </script> — adjust embedding strategy.")

OUT.write_text(TOP + md + BOTTOM, encoding="utf-8")
print(f"wrote {OUT} ({OUT.stat().st_size:,} bytes)")
