import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { LoginPage } from '@/pages/LoginPage'

const { registerMock, loginMock, checkAuthMock } = vi.hoisted(() => ({
  registerMock: vi.fn(),
  loginMock: vi.fn(),
  checkAuthMock: vi.fn(),
}))

vi.mock('@/lib/api', () => ({
  api: {
    register: registerMock,
    login: loginMock,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    checkAuth: checkAuthMock,
    openRegistration: true,
    googleAuth: false,
  }),
}))

function deferred() {
  let resolve!: () => void
  const promise = new Promise<void>((r) => {
    resolve = r
  })
  return { promise, resolve }
}

describe('LoginPage', () => {
  beforeEach(() => {
    registerMock.mockReset()
    loginMock.mockReset()
    checkAuthMock.mockReset()
  })

  it('prevents duplicate register submissions while a request is in flight', async () => {
    const pending = deferred()
    registerMock.mockReturnValueOnce(pending.promise)
    loginMock.mockResolvedValueOnce(undefined)
    checkAuthMock.mockResolvedValueOnce(undefined)

    render(<LoginPage />)

    fireEvent.click(screen.getByRole('button', { name: /create account/i }))
    fireEvent.change(screen.getByPlaceholderText('Name'), { target: { value: 'Test User' } })
    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 'test@example.com' } })
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
      expect(loginMock).toHaveBeenCalledTimes(1)
      expect(checkAuthMock).toHaveBeenCalledTimes(1)
    })
  })
})
