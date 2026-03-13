import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './i18n'
import './index.css'
import App from './App.tsx'
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

window.MonacoEnvironment = {
  getWorker: () => new editorWorker(),
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
