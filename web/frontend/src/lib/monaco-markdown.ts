import type * as Monaco from 'monaco-editor'

type MonacoEditor = Monaco.editor.IStandaloneCodeEditor

export function registerMarkdownFeatures(monaco: typeof Monaco, editor: MonacoEditor) {
  // Snippets as completion items
  const snippets = [
    { label: 'table', detail: 'Table', insertText: '| ${1:Header} | ${2:Header} |\n| --- | --- |\n| ${3:Cell} | ${4:Cell} |' },
    { label: 'code', detail: 'Code block', insertText: '```${1:language}\n${2}\n```' },
    { label: 'task', detail: 'Task list', insertText: '- [ ] ${1:Task}\n- [ ] ${2:Task}' },
    { label: 'link', detail: '[text](url)', insertText: '[${1:text}](${2:url})' },
    { label: 'image', detail: '![alt](url)', insertText: '![${1:alt}](${2:url})' },
    { label: 'bold', detail: '**bold**', insertText: '**${1:text}**' },
    { label: 'italic', detail: '*italic*', insertText: '*${1:text}*' },
    { label: 'strikethrough', detail: '~~text~~', insertText: '~~${1:text}~~' },
    { label: 'heading1', detail: '# Heading 1', insertText: '# ${1:Heading}' },
    { label: 'heading2', detail: '## Heading 2', insertText: '## ${1:Heading}' },
    { label: 'heading3', detail: '### Heading 3', insertText: '### ${1:Heading}' },
    { label: 'blockquote', detail: '> Quote', insertText: '> ${1:quote}' },
    { label: 'hr', detail: 'Horizontal rule', insertText: '\n---\n' },
    { label: 'footnote', detail: 'Footnote', insertText: '[^${1:id}]: ${2:text}' },
    { label: 'details', detail: 'Collapsible section', insertText: '<details>\n<summary>${1:Summary}</summary>\n\n${2:Content}\n\n</details>' },
    { label: 'alert-note', detail: '> [!NOTE]', insertText: '> [!NOTE]\n> ${1:text}' },
    { label: 'alert-tip', detail: '> [!TIP]', insertText: '> [!TIP]\n> ${1:text}' },
    { label: 'alert-important', detail: '> [!IMPORTANT]', insertText: '> [!IMPORTANT]\n> ${1:text}' },
    { label: 'alert-warning', detail: '> [!WARNING]', insertText: '> [!WARNING]\n> ${1:text}' },
    { label: 'alert-caution', detail: '> [!CAUTION]', insertText: '> [!CAUTION]\n> ${1:text}' },
    { label: 'math-inline', detail: 'Inline math $...$', insertText: '$$${1:expression}$$' },
    { label: 'math-block', detail: 'Math block', insertText: '$$\n${1:expression}\n$$' },
  ]

  monaco.languages.registerCompletionItemProvider('markdown', {
    triggerCharacters: ['/', '[', '!', '`', '#', '-', '>', '$'],
    provideCompletionItems: (model, position) => {
      const word = model.getWordUntilPosition(position)
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      }
      return {
        suggestions: snippets.map((s) => ({
          label: s.label,
          kind: monaco.languages.CompletionItemKind.Snippet,
          detail: s.detail,
          insertText: s.insertText,
          insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          range,
        })),
      }
    },
  })

  // Auto-continue lists on Enter
  editor.onDidChangeModelContent((e) => {
    const model = editor.getModel()
    if (!model) return
    const pos = editor.getPosition()
    if (!pos) return

    // Only handle Enter key (single newline insert)
    const change = e.changes[0]
    if (!change || change.text !== '\n') return

    const prevLineNum = pos.lineNumber - 1
    if (prevLineNum < 1) return
    const prevLine = model.getLineContent(prevLineNum)

    // Empty list item → remove list marker
    const emptyUl = prevLine.match(/^(\s*)([-*+])\s$/)
    const emptyOl = prevLine.match(/^(\s*)\d+\.\s$/)
    const emptyTask = prevLine.match(/^(\s*)([-*+])\s\[[ x]\]\s$/)
    const emptyBq = prevLine.match(/^(\s*>+)\s$/)
    if (emptyUl || emptyOl || emptyTask || emptyBq) {
      // Use setTimeout to avoid recursive content change
      setTimeout(() => {
        editor.executeEdits('clear-empty-list', [{
          range: { startLineNumber: prevLineNum, startColumn: 1, endLineNumber: pos.lineNumber, endColumn: model.getLineContent(pos.lineNumber).length + 1 },
          text: '',
        }])
      }, 0)
      return
    }

    // Task list
    const taskMatch = prevLine.match(/^(\s*)([-*+])\s\[[ x]\]\s(.+)/)
    if (taskMatch) {
      setTimeout(() => {
        const cur = editor.getPosition()
        if (!cur) return
        editor.executeEdits('auto-task', [{
          range: { startLineNumber: cur.lineNumber, startColumn: 1, endLineNumber: cur.lineNumber, endColumn: cur.column },
          text: `${taskMatch[1]}${taskMatch[2]} [ ] `,
        }])
      }, 0)
      return
    }

    // Unordered list
    const ulMatch = prevLine.match(/^(\s*)([-*+])\s(.+)/)
    if (ulMatch) {
      setTimeout(() => {
        const cur = editor.getPosition()
        if (!cur) return
        editor.executeEdits('auto-list', [{
          range: { startLineNumber: cur.lineNumber, startColumn: 1, endLineNumber: cur.lineNumber, endColumn: cur.column },
          text: `${ulMatch[1]}${ulMatch[2]} `,
        }])
      }, 0)
      return
    }

    // Ordered list
    const olMatch = prevLine.match(/^(\s*)(\d+)\.\s(.+)/)
    if (olMatch) {
      const nextNum = parseInt(olMatch[2]) + 1
      setTimeout(() => {
        const cur = editor.getPosition()
        if (!cur) return
        editor.executeEdits('auto-ol', [{
          range: { startLineNumber: cur.lineNumber, startColumn: 1, endLineNumber: cur.lineNumber, endColumn: cur.column },
          text: `${olMatch[1]}${nextNum}. `,
        }])
      }, 0)
      return
    }

    // Blockquote continuation
    const bqMatch = prevLine.match(/^(\s*>+)\s(.+)/)
    if (bqMatch) {
      setTimeout(() => {
        const cur = editor.getPosition()
        if (!cur) return
        editor.executeEdits('auto-bq', [{
          range: { startLineNumber: cur.lineNumber, startColumn: 1, endLineNumber: cur.lineNumber, endColumn: cur.column },
          text: `${bqMatch[1]} `,
        }])
      }, 0)
    }
  })

  // Keyboard shortcuts
  editor.addAction({
    id: 'markdown-bold',
    label: 'Bold',
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyB],
    run: (ed) => wrapSelection(ed as MonacoEditor, '**'),
  })
  editor.addAction({
    id: 'markdown-italic',
    label: 'Italic',
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyI],
    run: (ed) => wrapSelection(ed as MonacoEditor, '*'),
  })
  editor.addAction({
    id: 'markdown-strikethrough',
    label: 'Strikethrough',
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyX],
    run: (ed) => wrapSelection(ed as MonacoEditor, '~~'),
  })
  editor.addAction({
    id: 'markdown-code-inline',
    label: 'Inline code',
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyE],
    run: (ed) => wrapSelection(ed as MonacoEditor, '`'),
  })
}

function wrapSelection(editor: Monaco.editor.IStandaloneCodeEditor, wrapper: string) {
  const selection = editor.getSelection()
  if (!selection) return
  const model = editor.getModel()
  if (!model) return
  const text = model.getValueInRange(selection)
  if (text.startsWith(wrapper) && text.endsWith(wrapper)) {
    editor.executeEdits('unwrap', [{
      range: selection,
      text: text.slice(wrapper.length, -wrapper.length),
    }])
  } else {
    editor.executeEdits('wrap', [{
      range: selection,
      text: `${wrapper}${text}${wrapper}`,
    }])
  }
}
