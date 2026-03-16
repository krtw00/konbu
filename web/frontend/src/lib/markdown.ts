import { marked } from 'marked'
import DOMPurify from 'dompurify'
import markedAlert from 'marked-alert'
import { gfmHeadingId } from 'marked-gfm-heading-id'
import hljs from 'highlight.js/lib/core'
import markdown from 'highlight.js/lib/languages/markdown'
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import go from 'highlight.js/lib/languages/go'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import css from 'highlight.js/lib/languages/css'
import xml from 'highlight.js/lib/languages/xml'
import sql from 'highlight.js/lib/languages/sql'
import rust from 'highlight.js/lib/languages/rust'
import diff from 'highlight.js/lib/languages/diff'
import shell from 'highlight.js/lib/languages/shell'
import powershell from 'highlight.js/lib/languages/powershell'

hljs.registerLanguage('markdown', markdown)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('js', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('ts', typescript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('py', python)
hljs.registerLanguage('go', go)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('sh', bash)
hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('yml', yaml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('rust', rust)
hljs.registerLanguage('diff', diff)
hljs.registerLanguage('shell', shell)
hljs.registerLanguage('powershell', powershell)

marked.use(
  gfmHeadingId(),
  markedAlert(),
  {
    gfm: true,
    breaks: true,
    renderer: {
      code({ text, lang }: { text: string; lang?: string }) {
        const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
        let highlighted: string
        try {
          highlighted = language !== 'plaintext'
            ? hljs.highlight(text, { language }).value
            : text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        } catch {
          highlighted = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        }
        return `<pre><code class="hljs language-${language}">${highlighted}</code></pre>`
      },
      checkbox({ checked }: { checked: boolean }) {
        return `<input type="checkbox" ${checked ? 'checked' : ''} disabled />`
      },
    },
  },
)

export function renderMarkdown(content: string): string {
  // Replace [[memo name]] with clickable links before parsing
  const withLinks = (content || '').replace(
    /\[\[([^\]]+)\]\]/g,
    '<a href="#" data-memo-link="$1" class="memo-link">$1</a>'
  )
  const html = marked.parse(withLinks) as string
  return DOMPurify.sanitize(html, {
    ADD_ATTR: ['data-memo-link'],
    ADD_TAGS: ['input'],
  })
}
