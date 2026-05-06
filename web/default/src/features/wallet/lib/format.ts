import { formatLocalCurrencyAmount } from '@/lib/currency'
import { DEFAULT_DISCOUNT_RATE, PAYMENT_TYPES } from '../constants'

// ============================================================================
// Wallet-specific Formatting Functions
// ============================================================================

/**
 * Format Creem price with currency symbol (USD/EUR)
 */
export function formatCreemPrice(
  price: number,
  currency: 'USD' | 'EUR'
): string {
  const symbol = currency === 'EUR' ? '€' : '$'
  return `${symbol}${price.toFixed(2)}`
}

/**
 * Format large quota numbers with K/M suffix
 */
export function formatQuotaShort(quota: number): string {
  if (quota >= 1000000) {
    return `${(quota / 1000000).toFixed(1)}M`
  }
  if (quota >= 1000) {
    return `${(quota / 1000).toFixed(1)}K`
  }
  return quota.toString()
}

/**
 * Format currency amount that is already in local currency.
 * This is used for payment amounts that have been calculated via priceRatio.
 */
export function formatCurrency(amount: number | string): string {
  const numeric =
    typeof amount === 'number' ? amount : Number.parseFloat(String(amount))
  return formatLocalCurrencyAmount(Number.isFinite(numeric) ? numeric : null, {
    digitsLarge: 2,
    digitsSmall: 2,
    abbreviate: false,
  })
}

/**
 * Format Epay payment amounts. Epay historically charges in CNY.
 */
export function formatEpayCurrency(amount: number | string): string {
  const numeric =
    typeof amount === 'number' ? amount : Number.parseFloat(String(amount))
  if (!Number.isFinite(numeric)) return '-'

  const sign = numeric < 0 ? '-' : ''
  const formattedAmount = Math.abs(numeric).toFixed(2).replace(/\.?0+$/, '')
  return `${sign}¥${formattedAmount}`
}

export function isEpayPaymentType(paymentType?: string): boolean {
  switch (paymentType) {
    case PAYMENT_TYPES.STRIPE:
    case PAYMENT_TYPES.CREEM:
    case PAYMENT_TYPES.WAFFO:
    case PAYMENT_TYPES.WAFFO_PANCAKE:
      return false
    default:
      return true
  }
}

/**
 * Format payment amounts by gateway. Epay methods are CNY; other gateways keep
 * the existing system currency display.
 */
export function formatPaymentCurrency(
  amount: number | string,
  paymentType?: string
): string {
  return isEpayPaymentType(paymentType)
    ? formatEpayCurrency(amount)
    : formatCurrency(amount)
}

/**
 * Get discount label for display (e.g., "20% OFF")
 */
export function getDiscountLabel(discount: number): string {
  if (discount >= DEFAULT_DISCOUNT_RATE) {
    return ''
  }
  const off = Math.round((1 - discount) * 100)
  return `${off}% OFF`
}

/**
 * Calculate pricing details for a preset amount
 */
export function calculatePresetPricing(
  presetValue: number,
  priceRatio: number,
  discount: number,
  usdExchangeRate: number = 1
) {
  const originalPrice = presetValue * priceRatio
  const actualPrice = originalPrice * discount
  const savedAmount = originalPrice - actualPrice
  const hasDiscount = discount < 1.0
  const displayValue = presetValue * usdExchangeRate

  return {
    displayValue,
    originalPrice,
    actualPrice,
    savedAmount,
    hasDiscount,
  }
}
