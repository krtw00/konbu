// CodeMirror 6 — Markdown editor setup
import {EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, drawSelection, rectangularSelection} from 'https://esm.sh/@codemirror/view@6?bundle'
import {EditorState} from 'https://esm.sh/@codemirror/state@6?bundle'
import {defaultKeymap, history, historyKeymap, indentWithTab} from 'https://esm.sh/@codemirror/commands@6?bundle'
import {markdown} from 'https://esm.sh/@codemirror/lang-markdown@6?bundle'
import {syntaxHighlighting, defaultHighlightStyle, bracketMatching} from 'https://esm.sh/@codemirror/language@6?bundle'
import {closeBrackets, closeBracketsKeymap} from 'https://esm.sh/@codemirror/autocomplete@6?bundle'
import {highlightSelectionMatches, searchKeymap} from 'https://esm.sh/@codemirror/search@6?bundle'

// Theme that adapts to CSS variables
const cmTheme = EditorView.theme({
  '&': {height: '100%', fontSize: '14px'},
  '.cm-scroller': {overflow: 'auto', fontFamily: "'JetBrains Mono','Fira Code','Cascadia Code',Consolas,monospace"},
  '.cm-content': {padding: '24px 32px', caretColor: 'var(--accent)'},
  '.cm-gutters': {background: 'var(--bg2)', color: 'var(--text3)', border: 'none', paddingLeft: '8px'},
  '.cm-activeLineGutter': {background: 'var(--bg3)'},
  '.cm-activeLine': {background: 'var(--accent-bg)'},
  '.cm-selectionBackground, ::selection': {background: 'var(--accent-bg) !important'},
  '.cm-cursor': {borderLeftColor: 'var(--accent)'},
  '.cm-line': {lineHeight: '1.7'},
  // Markdown syntax colors
  '.cm-header-1': {fontSize: '1.4em', fontWeight: '700', color: 'var(--text)'},
  '.cm-header-2': {fontSize: '1.2em', fontWeight: '600', color: 'var(--text)'},
  '.cm-header-3': {fontSize: '1.1em', fontWeight: '600', color: 'var(--text2)'},
  '.cm-strong': {fontWeight: '700'},
  '.cm-emphasis': {fontStyle: 'italic'},
  '.cm-url': {color: 'var(--accent)', textDecoration: 'underline'},
  '.cm-link': {color: 'var(--accent)'},
  '.cm-meta': {color: 'var(--text3)'},
  '.cm-comment': {color: 'var(--text3)'},
  '.cm-monospace': {fontFamily: 'inherit'},
}, {dark: false})

// Expose factory to global scope for app.js
window.createCM = function(parent, content, onChange) {
  const state = EditorState.create({
    doc: content || '',
    extensions: [
      lineNumbers(),
      highlightActiveLine(),
      highlightActiveLineGutter(),
      drawSelection(),
      rectangularSelection(),
      history(),
      bracketMatching(),
      closeBrackets(),
      highlightSelectionMatches(),
      markdown(),
      syntaxHighlighting(defaultHighlightStyle),
      cmTheme,
      keymap.of([
        ...defaultKeymap,
        ...historyKeymap,
        ...closeBracketsKeymap,
        ...searchKeymap,
        indentWithTab,
      ]),
      EditorView.updateListener.of(update => {
        if (update.docChanged && onChange) onChange(update.state.doc.toString())
      }),
      EditorView.lineWrapping,
    ],
  })
  return new EditorView({state, parent})
}
