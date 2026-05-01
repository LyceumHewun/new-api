import type { TFunction } from 'i18next'
import dayjs from '@/lib/dayjs'
import type { SubscriptionPlan } from '../types'

export function formatDuration(
  plan: Partial<SubscriptionPlan>,
  t: TFunction
): string {
  const unit = plan?.duration_unit || 'month'
  const value = plan?.duration_value || 1
  const unitLabels: Record<string, string> = {
    year: t('years'),
    month: t('months'),
    day: t('days'),
    hour: t('hours'),
    custom: t('Custom (seconds)'),
  }
  if (unit === 'custom') {
    const seconds = plan?.custom_seconds || 0
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('days')}`
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('hours')}`
    return `${seconds} ${t('seconds')}`
  }
  return `${value} ${unitLabels[unit] || unit}`
}

export function formatResetPeriod(
  plan: Partial<SubscriptionPlan>,
  t: TFunction
): string {
  const period = plan?.quota_reset_period || 'never'
  if (period === 'daily') return t('Daily')
  if (period === 'weekly') return t('Weekly')
  if (period === 'monthly') return t('Monthly')
  if (period === 'custom') {
    const seconds = Number(plan?.quota_reset_custom_seconds || 0)
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('days')}`
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('hours')}`
    if (seconds >= 60) return `${Math.floor(seconds / 60)} ${t('minutes')}`
    return `${seconds} ${t('seconds')}`
  }
  return t('No Reset')
}

function getDurationSeconds(plan: Partial<SubscriptionPlan>): number {
  const unit = plan?.duration_unit || 'month'
  const value = Number(plan?.duration_value || 1)
  if (unit === 'year' || unit === 'month') {
    const start = new Date()
    const end = new Date(start.getTime())
    if (unit === 'year') end.setFullYear(end.getFullYear() + value)
    if (unit === 'month') end.setMonth(end.getMonth() + value)
    return Math.max(0, Math.floor((end.getTime() - start.getTime()) / 1000))
  }
  if (unit === 'day') return value * 86400
  if (unit === 'hour') return value * 3600
  if (unit === 'custom') return Number(plan?.custom_seconds || 0)
  return 0
}

export function getSubscriptionPlanQuotaDisplay(
  plan: Partial<SubscriptionPlan>
) {
  const periodAmount = Number(plan?.total_amount || 0)
  const resetSeconds =
    plan?.quota_reset_period === 'custom'
      ? Number(plan?.quota_reset_custom_seconds || 0)
      : 0
  const durationSeconds = getDurationSeconds(plan)
  const cycleCount =
    periodAmount > 0 && resetSeconds > 0 && durationSeconds > 0
      ? Math.max(1, Math.ceil(durationSeconds / resetSeconds))
      : 1

  return {
    cycleCount,
    periodAmount,
    totalAmount: periodAmount * cycleCount,
    showPeriodAmount:
      periodAmount > 0 && resetSeconds > 0 && durationSeconds > 0,
  }
}

export function formatTimestamp(ts: number): string {
  if (!ts) return '-'
  return dayjs(ts * 1000).format('YYYY-MM-DD HH:mm:ss')
}
