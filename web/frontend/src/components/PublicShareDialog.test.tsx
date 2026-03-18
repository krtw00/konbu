import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { PublicShareDialog } from '@/components/PublicShareDialog'

const { getPublicShareMock, createPublicShareMock, deletePublicShareMock, appURLMock } = vi.hoisted(() => ({
  getPublicShareMock: vi.fn(),
  createPublicShareMock: vi.fn(),
  deletePublicShareMock: vi.fn(),
  appURLMock: vi.fn((path: string) => `https://public.example.com${path}`),
}))

vi.mock('@/lib/api', () => ({
  api: {
    getPublicShare: getPublicShareMock,
    createPublicShare: createPublicShareMock,
    deletePublicShare: deletePublicShareMock,
  },
}))

vi.mock('@/lib/runtime', () => ({
  appURL: appURLMock,
}))

describe('PublicShareDialog', () => {
  const writeText = vi.fn()
  const openMock = vi.fn()

  beforeEach(() => {
    getPublicShareMock.mockReset()
    createPublicShareMock.mockReset()
    deletePublicShareMock.mockReset()
    appURLMock.mockClear()
    writeText.mockReset()
    openMock.mockReset()

    Object.assign(navigator, {
      clipboard: {
        writeText,
      },
    })
    vi.stubGlobal('open', openMock)
  })

  it('loads and displays an existing public URL', async () => {
    getPublicShareMock.mockResolvedValueOnce({
      data: { token: 'memo-token' },
    })

    render(<PublicShareDialog resourceType="memo" resourceId="memo-1" />)

    await userEvent.click(screen.getByRole('button', { name: /share|publish/i }))

    expect(getPublicShareMock).toHaveBeenCalledWith('memo', 'memo-1')

    await screen.findByText('https://public.example.com/public/memo-token')
    expect(appURLMock).toHaveBeenCalledWith('/public/memo-token')
  })

  it('creates a public link and supports copy/open actions', async () => {
    getPublicShareMock.mockResolvedValueOnce({ data: null })
    createPublicShareMock.mockResolvedValueOnce({
      data: { token: 'event-token' },
    })

    render(<PublicShareDialog resourceType="event" resourceId="event-1" />)

    await userEvent.click(screen.getByRole('button', { name: /share|publish/i }))
    await screen.findByText(/No share link yet.|This item is not published yet./i)

    await userEvent.click(screen.getByRole('button', { name: /create share link|start publishing/i }))

    await screen.findByText('https://public.example.com/public/event-token')
    expect(createPublicShareMock).toHaveBeenCalledWith('event', 'event-1')

    await userEvent.click(screen.getByRole('button', { name: /copy/i }))
    expect(writeText).toHaveBeenCalledWith('https://public.example.com/public/event-token')

    await userEvent.click(screen.getByRole('button', { name: /open/i }))
    expect(openMock).toHaveBeenCalledWith('https://public.example.com/public/event-token', '_blank', 'noopener,noreferrer')
  })

  it('shows a load error when the publish status cannot be fetched', async () => {
    getPublicShareMock.mockRejectedValueOnce(new Error('load failed'))

    render(<PublicShareDialog resourceType="calendar" resourceId="calendar-1" />)

    await userEvent.click(screen.getByRole('button', { name: /share|publish/i }))

    await waitFor(() => {
      expect(screen.getByText('load failed')).toBeInTheDocument()
    })
  })
})
