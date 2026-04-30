export function formatSubscriptionDuration(plan, t) {
  const unit = plan?.duration_unit || 'month';
  const value = plan?.duration_value || 1;
  const unitLabels = {
    year: t('年'),
    month: t('个月'),
    day: t('天'),
    hour: t('小时'),
    custom: t('自定义'),
  };
  if (unit === 'custom') {
    const seconds = plan?.custom_seconds || 0;
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('天')}`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('小时')}`;
    return `${seconds} ${t('秒')}`;
  }
  return `${value} ${unitLabels[unit] || unit}`;
}

export function formatSubscriptionResetPeriod(plan, t) {
  const period = plan?.quota_reset_period || 'never';
  if (period === 'never') return t('不重置');
  if (period === 'daily') return t('每天');
  if (period === 'weekly') return t('每周');
  if (period === 'monthly') return t('每月');
  if (period === 'custom') {
    const seconds = Number(plan?.quota_reset_custom_seconds || 0);
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('天')}`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('小时')}`;
    if (seconds >= 60) return `${Math.floor(seconds / 60)} ${t('分钟')}`;
    return `${seconds} ${t('秒')}`;
  }
  return t('不重置');
}

function getSubscriptionDurationSeconds(plan) {
  const unit = plan?.duration_unit || 'month';
  const value = Number(plan?.duration_value || 1);
  if (unit === 'year' || unit === 'month') {
    const start = new Date();
    const end = new Date(start.getTime());
    if (unit === 'year') end.setFullYear(end.getFullYear() + value);
    if (unit === 'month') end.setMonth(end.getMonth() + value);
    return Math.max(0, Math.floor((end.getTime() - start.getTime()) / 1000));
  }
  if (unit === 'day') return value * 86400;
  if (unit === 'hour') return value * 3600;
  if (unit === 'custom') return Number(plan?.custom_seconds || 0);
  return 0;
}

export function getSubscriptionPlanQuotaDisplay(plan) {
  const periodAmount = Number(plan?.total_amount || 0);
  const isCustomReset = plan?.quota_reset_period === 'custom';
  const resetSeconds = isCustomReset
    ? Number(plan?.quota_reset_custom_seconds || 0)
    : 0;
  const durationSeconds = getSubscriptionDurationSeconds(plan);
  const cycleCount =
    periodAmount > 0 && resetSeconds > 0 && durationSeconds > 0
      ? Math.max(1, Math.ceil(durationSeconds / resetSeconds))
      : 1;

  return {
    cycleCount,
    periodAmount,
    totalAmount: periodAmount * cycleCount,
    showPeriodAmount:
      periodAmount > 0 && resetSeconds > 0 && durationSeconds > 0,
  };
}
