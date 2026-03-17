import { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Download, Upload, Trash2, ArrowUp, ArrowDown, MoreVertical, Plus } from 'lucide-react'
import type { TableColumn, MemoRow } from '@/types/api'

interface TableEditorProps {
  memoId: string
  columns: TableColumn[]
  onColumnsChange: (cols: TableColumn[]) => void
}

const LIMIT = 100

function genColId(): string {
  return 'col_' + Math.random().toString(36).slice(2, 10)
}

export function TableEditor({ memoId, columns, onColumnsChange }: TableEditorProps) {
  const { t } = useTranslation()
  const [rows, setRows] = useState<MemoRow[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [sortCol, setSortCol] = useState('')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc')
  const [editingCell, setEditingCell] = useState<{ rowId: string; colId: string } | null>(null)
  const [editValue, setEditValue] = useState('')
  const [newRowData, setNewRowData] = useState<Record<string, string>>({})
  const [newColName, setNewColName] = useState('')
  const saveTimers = useRef<Record<string, ReturnType<typeof setTimeout>>>({})
  const inputRef = useRef<HTMLInputElement>(null)

  const loadRows = useCallback(async (off = 0) => {
    const r = await api.listMemoRows(memoId, {
      limit: LIMIT,
      offset: off,
      sort: sortCol || undefined,
      order: sortCol ? sortOrder : undefined,
    })
    setRows(r.data || [])
    setTotal(r.total || 0)
    setOffset(off)
  }, [memoId, sortCol, sortOrder])

  useEffect(() => {
    loadRows(0)
  }, [loadRows])

  useEffect(() => {
    if (editingCell && inputRef.current) {
      inputRef.current.focus()
    }
  }, [editingCell])

  function handleCellClick(rowId: string, colId: string, currentValue: string) {
    // Save current cell first
    commitEdit()
    setEditingCell({ rowId, colId })
    setEditValue(currentValue || '')
  }

  function commitEdit() {
    if (!editingCell) return
    const { rowId, colId } = editingCell
    const row = rows.find(r => r.id === rowId)
    if (!row) return
    if (row.row_data[colId] === editValue) return

    const newData = { ...row.row_data, [colId]: editValue }
    setRows(prev => prev.map(r => r.id === rowId ? { ...r, row_data: newData } : r))

    if (saveTimers.current[rowId]) clearTimeout(saveTimers.current[rowId])
    saveTimers.current[rowId] = setTimeout(async () => {
      await api.updateMemoRow(memoId, rowId, newData)
    }, 1000)
  }

  function handleCellBlur() {
    // Delay to allow click on another cell to fire first
    setTimeout(() => {
      commitEdit()
      setEditingCell(prev => {
        // Only clear if no new cell was selected
        return prev
      })
    }, 100)
  }

  function handleCellChange(value: string) {
    setEditValue(value)
    if (!editingCell) return
    const { rowId, colId } = editingCell

    // Update local state immediately for responsiveness
    const newData = { ...rows.find(r => r.id === rowId)?.row_data, [colId]: value }
    setRows(prev => prev.map(r => r.id === rowId ? { ...r, row_data: { ...r.row_data, [colId]: value } } : r))

    // Debounced save
    if (saveTimers.current[rowId]) clearTimeout(saveTimers.current[rowId])
    saveTimers.current[rowId] = setTimeout(async () => {
      await api.updateMemoRow(memoId, rowId, newData as Record<string, string>)
    }, 1000)
  }

  function handleCellKeyDown(e: React.KeyboardEvent, rowId: string, colId: string) {
    if (e.key === 'Tab') {
      e.preventDefault()
      commitEdit()
      const colIdx = columns.findIndex(c => c.id === colId)
      if (colIdx < columns.length - 1) {
        const nextCol = columns[colIdx + 1]
        const row = rows.find(r => r.id === rowId)
        setEditingCell({ rowId, colId: nextCol.id })
        setEditValue(row?.row_data[nextCol.id] || '')
      }
    } else if (e.key === 'Enter') {
      e.preventDefault()
      commitEdit()
      const rowIdx = rows.findIndex(r => r.id === rowId)
      if (rowIdx < rows.length - 1) {
        const nextRow = rows[rowIdx + 1]
        setEditingCell({ rowId: nextRow.id, colId })
        setEditValue(nextRow.row_data[colId] || '')
      }
    } else if (e.key === 'Escape') {
      setEditingCell(null)
    }
  }

  async function handleNewRowInput(colId: string, value: string) {
    const updated = { ...newRowData, [colId]: value }
    setNewRowData(updated)
  }

  async function handleNewRowBlur() {
    const hasValue = Object.values(newRowData).some(v => v.trim())
    if (!hasValue) return

    await api.createMemoRow(memoId, newRowData)
    setNewRowData({})
    loadRows(offset)
  }

  async function handleNewRowKeyDown(e: React.KeyboardEvent, colId: string) {
    if (e.key === 'Enter') {
      e.preventDefault()
      await handleNewRowBlur()
    } else if (e.key === 'Tab') {
      const colIdx = columns.findIndex(c => c.id === colId)
      if (colIdx >= columns.length - 1) {
        e.preventDefault()
        await handleNewRowBlur()
      }
    }
  }

  async function handleNewColumn() {
    const name = newColName.trim()
    if (!name) return
    const newCol: TableColumn = { id: genColId(), name }
    const updated = [...columns, newCol]
    onColumnsChange(updated)
    setNewColName('')
  }

  function handleNewColKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleNewColumn()
    }
  }

  async function handleDeleteRow(rowId: string) {
    await api.deleteMemoRow(memoId, rowId)
    setRows(prev => prev.filter(r => r.id !== rowId))
    setTotal(prev => prev - 1)
  }

  async function handleDeleteColumn(colId: string) {
    const updated = columns.filter(c => c.id !== colId)
    onColumnsChange(updated)
  }

  function handleRenameColumn(colId: string) {
    const col = columns.find(c => c.id === colId)
    if (!col) return
    const name = prompt(t('table.renameColumn'), col.name)
    if (name === null || name.trim() === '' || name === col.name) return
    const updated = columns.map(c => c.id === colId ? { ...c, name: name.trim() } : c)
    onColumnsChange(updated)
  }

  function handleSort(colId: string) {
    if (sortCol === colId) {
      setSortOrder(prev => prev === 'asc' ? 'desc' : 'asc')
    } else {
      setSortCol(colId)
      setSortOrder('asc')
    }
  }

  async function handleExport() {
    const csv = await api.exportMemoRowsCSV(memoId)
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'export.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  async function handleImport() {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.csv'
    input.onchange = async () => {
      const file = input.files?.[0]
      if (!file) return
      const text = await file.text()
      const lines = text.split('\n').map(l => l.trim()).filter(Boolean)
      if (lines.length < 2) return

      const headers = parseCSVLine(lines[0])

      // Auto-create columns if none exist
      let cols = [...columns]
      if (cols.length === 0) {
        cols = headers.map(h => ({ id: genColId(), name: h }))
        onColumnsChange(cols)
      }

      const rowsData: Record<string, string>[] = []
      for (let i = 1; i < lines.length; i++) {
        const values = parseCSVLine(lines[i])
        const row: Record<string, string> = {}
        for (let j = 0; j < cols.length && j < values.length; j++) {
          row[cols[j].id] = values[j]
        }
        rowsData.push(row)
      }

      if (rowsData.length > 0) {
        await api.batchCreateMemoRows(memoId, rowsData)
        loadRows(0)
      }
    }
    input.click()
  }

  const hasMore = offset + LIMIT < total

  return (
    <div className="flex flex-col gap-2 h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={handleImport}>
          <Upload size={14} className="mr-1" /> {t('table.import')}
        </Button>
        <Button variant="outline" size="sm" onClick={handleExport}>
          <Download size={14} className="mr-1" /> {t('table.export')}
        </Button>
        <span className="text-xs text-muted-foreground ml-auto">
          {t('table.rowCount', { count: total })}
        </span>
      </div>

      {/* Table */}
      <div className="flex-1 overflow-auto border border-border rounded-lg">
        <table className="w-full border-collapse text-sm">
          <thead className="sticky top-0 z-10 bg-muted">
            <tr>
              <th className="w-8 border-b border-border px-1" />
              {columns.map(col => (
                <th
                  key={col.id}
                  className="border-b border-r border-border px-2 py-1.5 text-left font-medium cursor-pointer hover:bg-accent/50 select-none"
                  onClick={() => handleSort(col.id)}
                >
                  <div className="flex items-center gap-1">
                    <span className="truncate">{col.name}</span>
                    {sortCol === col.id && (
                      sortOrder === 'asc' ? <ArrowUp size={12} /> : <ArrowDown size={12} />
                    )}
                    <DropdownMenu>
                      <DropdownMenuTrigger onClick={e => e.stopPropagation()} className="ml-auto opacity-0 group-hover:opacity-100">
                        <MoreVertical size={12} className="text-muted-foreground" />
                      </DropdownMenuTrigger>
                      <DropdownMenuContent>
                        <DropdownMenuItem onClick={() => handleRenameColumn(col.id)}>
                          {t('table.renameColumn')}
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-destructive" onClick={() => handleDeleteColumn(col.id)}>
                          {t('common.delete')}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </th>
              ))}
              {/* New column header */}
              <th className="border-b border-border px-2 py-1.5 min-w-24">
                <input
                  value={newColName}
                  onChange={e => setNewColName(e.target.value)}
                  onKeyDown={handleNewColKeyDown}
                  onBlur={handleNewColumn}
                  placeholder={t('table.newColumn')}
                  className="w-full bg-transparent text-xs text-muted-foreground placeholder:text-muted-foreground/50 focus:outline-none"
                />
              </th>
            </tr>
          </thead>
          <tbody>
            {rows.map(row => (
              <tr key={row.id} className="group hover:bg-accent/30">
                <td className="border-b border-border px-1 text-center">
                  <button
                    onClick={() => handleDeleteRow(row.id)}
                    className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive"
                  >
                    <Trash2 size={12} />
                  </button>
                </td>
                {columns.map(col => (
                  <td
                    key={col.id}
                    className="border-b border-r border-border px-2 py-1 cursor-text"
                    onClick={() => handleCellClick(row.id, col.id, row.row_data[col.id] || '')}
                  >
                    {editingCell?.rowId === row.id && editingCell?.colId === col.id ? (
                      <input
                        ref={inputRef}
                        value={editValue}
                        onChange={e => handleCellChange(e.target.value)}
                        onBlur={handleCellBlur}
                        onKeyDown={e => handleCellKeyDown(e, row.id, col.id)}
                        className="w-full bg-transparent focus:outline-none"
                      />
                    ) : (
                      <span className="truncate block min-h-[1.25rem]">{row.row_data[col.id] || ''}</span>
                    )}
                  </td>
                ))}
                <td className="border-b border-border" />
              </tr>
            ))}
            {/* New row */}
            <tr className="bg-muted/30">
              <td className="border-b border-border px-1 text-center text-muted-foreground">
                <Plus size={12} />
              </td>
              {columns.map(col => (
                <td key={col.id} className="border-b border-r border-border px-2 py-1">
                  <input
                    value={newRowData[col.id] || ''}
                    onChange={e => handleNewRowInput(col.id, e.target.value)}
                    onBlur={handleNewRowBlur}
                    onKeyDown={e => handleNewRowKeyDown(e, col.id)}
                    placeholder="..."
                    className="w-full bg-transparent text-sm placeholder:text-muted-foreground/30 focus:outline-none"
                  />
                </td>
              ))}
              <td className="border-b border-border" />
            </tr>
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {hasMore && (
        <div className="flex justify-center">
          <Button variant="outline" size="sm" onClick={() => loadRows(offset + LIMIT)}>
            {t('table.loadMore')}
          </Button>
        </div>
      )}
    </div>
  )
}

function parseCSVLine(line: string): string[] {
  const result: string[] = []
  let current = ''
  let inQuotes = false
  for (let i = 0; i < line.length; i++) {
    const ch = line[i]
    if (inQuotes) {
      if (ch === '"' && line[i + 1] === '"') {
        current += '"'
        i++
      } else if (ch === '"') {
        inQuotes = false
      } else {
        current += ch
      }
    } else {
      if (ch === '"') {
        inQuotes = true
      } else if (ch === ',') {
        result.push(current)
        current = ''
      } else {
        current += ch
      }
    }
  }
  result.push(current)
  return result
}
