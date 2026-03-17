import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { SetupPage } from '@/pages/SetupPage'

const { registerMock, checkAuthMock } = vi.hoisted(() => ({
  registerMock: vi.fn(),
  checkAuthMock: vi.fn(),
}))

vi.mock('@/lib/api', () => ({
  api: {
    register: registerMock,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    checkAuth: checkAuthMock,
  }),
}))

function deferred() {
  let resolve!: () => void
  const promise = new Promise<void>((r) => {
    resolve = r
  })
  return { promise, resolve }
}

describe('SetupPage', () => {
  beforeEach(() => {
    registerMock.mockReset()
    checkAuthMock.mockReset()
  })

  it('prevents duplicate setup submissions while registration is pending', async () => {
    const pending = deferred()
    registerMock.mockReturnValueOnce(pending.promise)
    checkAuthMock.mockResolvedValueOnce(undefined)

    render(<SetupPage />)

    fireEvent.change(screen.getByPlaceholderText('Name'), { target: { value: 'Owner' } })
    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 'owner@example.com' } })
    fireEvent.change(screen.getByPlaceholderText('Password'), { target: { value: 'password123' } })
    fireEvent.change(screen.getByPlaceholderText(/confirm password/i), { target: { value: 'password123' } })

    const submit = screen.getByRole('button', { name: /create account/i })
    const form = submit.closest('form')
    expect(form).not.toBeNull()

    fireEvent.submit(form!)
    fireEvent.submit(form!)

    expect(registerMock).toHaveBeenCalledTimes(1)

    pending.resolve()

    await waitFor(() => {
      expect(checkAuthMock).toHaveBeenCalledTimes(1)
    })
  })
})
