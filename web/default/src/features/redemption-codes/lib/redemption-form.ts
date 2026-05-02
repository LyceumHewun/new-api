import { z } from 'zod'
import type { TFunction } from 'i18next'
import { parseQuotaFromDollars, quotaUnitsToDollars } from '@/lib/format'
import {
  REDEMPTION_VALIDATION,
  getRedemptionFormErrorMessages,
} from '../constants'
import { type RedemptionFormData, type Redemption } from '../types'

function isValidRemainCount(value: number | string) {
  const count = Number(value)
  return (
    value !== '' &&
    value !== '-' &&
    Number.isInteger(count) &&
    count >= REDEMPTION_VALIDATION.REMAIN_COUNT_MIN
  )
}

// ============================================================================
// Form Schema (use getRedemptionFormSchema(t) in components for i18n messages)
// ============================================================================

export function getRedemptionFormSchema(t: TFunction) {
  const msg = getRedemptionFormErrorMessages(t)
  return z.object({
    name: z
      .string()
      .min(REDEMPTION_VALIDATION.NAME_MIN_LENGTH, msg.NAME_LENGTH_INVALID)
      .max(REDEMPTION_VALIDATION.NAME_MAX_LENGTH, msg.NAME_LENGTH_INVALID),
    quota_dollars: z.number().min(0, t('Quota must be a positive number')),
    expired_time: z.date().optional(),
    count: z
      .number()
      .min(REDEMPTION_VALIDATION.COUNT_MIN, msg.COUNT_INVALID)
      .max(REDEMPTION_VALIDATION.COUNT_MAX, msg.COUNT_INVALID)
      .optional(),
    remain_count: z
      .union([z.number(), z.string()])
      .refine(isValidRemainCount, msg.REMAIN_COUNT_INVALID),
    disable_invite_rebate: z.boolean(),
  })
}

export type RedemptionFormValues = {
  name: string
  quota_dollars: number
  expired_time?: Date
  count?: number
  remain_count: number | string
  disable_invite_rebate: boolean
}

// ============================================================================
// Form Defaults
// ============================================================================

export const REDEMPTION_FORM_DEFAULT_VALUES: RedemptionFormValues = {
  name: '',
  quota_dollars: 10,
  expired_time: undefined,
  count: 1,
  remain_count: 1,
  disable_invite_rebate: false,
}

// ============================================================================
// Form Data Transformation
// ============================================================================

/**
 * Transform form data to API payload
 */
export function transformFormDataToPayload(
  data: RedemptionFormValues
): RedemptionFormData {
  return {
    name: data.name,
    quota: parseQuotaFromDollars(data.quota_dollars),
    expired_time: data.expired_time
      ? Math.floor(data.expired_time.getTime() / 1000)
      : 0,
    remain_count: Number(data.remain_count),
    disable_invite_rebate: data.disable_invite_rebate,
    count: data.count || 1,
  }
}

/**
 * Transform redemption data to form defaults
 */
export function transformRedemptionToFormDefaults(
  redemption: Redemption
): RedemptionFormValues {
  return {
    name: redemption.name,
    quota_dollars: quotaUnitsToDollars(redemption.quota),
    expired_time:
      redemption.expired_time > 0
        ? new Date(redemption.expired_time * 1000)
        : undefined,
    count: 1,
    remain_count: redemption.remain_count,
    disable_invite_rebate: redemption.disable_invite_rebate,
  }
}
