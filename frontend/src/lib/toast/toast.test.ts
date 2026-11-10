import { beforeEach, describe, expect, it } from 'vite-plus/test'
import { configureToast, dismiss, toast, toasts } from './toast'

describe('toast core', () => {
  beforeEach(() => {
    dismiss()
    configureToast({ duration: 3000, max: 5, newestOnTop: true })
  })

  it('shows a toast with the default duration and returns its id', () => {
    const id = toast.success('Saved')
    expect(toasts).toHaveLength(1)
    expect(toasts[0]).toMatchObject({ id, type: 'success', message: 'Saved', duration: 3000 })
  })

  it('puts the newest on top and keeps at most `max`', () => {
    configureToast({ max: 2 })
    toast.info('a')
    toast.info('b')
    toast.info('c')
    expect(toasts.map((t) => t.message)).toEqual(['c', 'b'])
  })

  it('appends at the bottom when newestOnTop is off', () => {
    configureToast({ newestOnTop: false, max: 2 })
    toast.info('a')
    toast.info('b')
    toast.info('c')
    expect(toasts.map((t) => t.message)).toEqual(['b', 'c'])
  })

  it('replaces a toast with the same id instead of stacking it', () => {
    toast.error('first', { id: 'net' })
    toast.error('second', { id: 'net' })
    expect(toasts).toHaveLength(1)
    expect(toasts[0].message).toBe('second')
  })

  it('dismisses one or all; duration 0 means sticky', () => {
    const a = toast.warning('a', { duration: 0 })
    toast.info('b')
    expect(toasts.find((t) => t.id === a)?.duration).toBe(0)
    dismiss(a)
    expect(toasts.map((t) => t.message)).toEqual(['b'])
    dismiss()
    expect(toasts).toHaveLength(0)
  })
})
