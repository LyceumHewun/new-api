import { useState } from 'react'
import * as z from 'zod'
import type { Resolver } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Code2, Palette } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { FormDirtyIndicator } from '../components/form-dirty-indicator'
import { FormNavigationGuard } from '../components/form-navigation-guard'
import { SettingsSection } from '../components/settings-section'
import { useSettingsForm } from '../hooks/use-settings-form'
import { useUpdateOption } from '../hooks/use-update-option'
import { InviteRebateGroupVisualEditor } from './invite-rebate-group-visual-editor'

const validateRatioArray = (value: string, ctx: z.RefinementCtx) => {
  try {
    const parsed = JSON.parse(value || '[]')
    if (!Array.isArray(parsed)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Must be a JSON array',
      })
      return
    }
    if (parsed.some((ratio) => typeof ratio !== 'number' || ratio < 0)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Ratios must be non-negative numbers',
      })
    }
  } catch {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'Invalid JSON',
    })
  }
}

const validateInviteRebateGroupSettings = (
  value: string,
  ctx: z.RefinementCtx
) => {
  try {
    const parsed = JSON.parse(value || '{}')
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Must be a JSON object',
      })
      return
    }
    for (const [groupName, groupSetting] of Object.entries(parsed)) {
      if (!groupName) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: 'Group name is required',
        })
        return
      }
      if (
        !groupSetting ||
        typeof groupSetting !== 'object' ||
        Array.isArray(groupSetting)
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: 'Each group setting must be an object',
        })
        return
      }
      const record = groupSetting as Record<string, unknown>
      if (
        typeof record.count_limit !== 'number' ||
        record.count_limit < -1 ||
        record.count_limit > 2147483647
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: 'count_limit must be ≥ -1',
        })
        return
      }
      if (
        !Array.isArray(record.chain_ratios) ||
        record.chain_ratios.some(
          (ratio) => typeof ratio !== 'number' || ratio < 0
        )
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: 'chain_ratios must be a JSON array of non-negative numbers',
        })
        return
      }
    }
  } catch {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: 'Invalid JSON',
    })
  }
}

const quotaSchema = z.object({
  QuotaForNewUser: z.coerce.number().min(0),
  PreConsumedQuota: z.coerce.number().min(0),
  QuotaForInviter: z.coerce.number().min(0),
  QuotaForInvitee: z.coerce.number().min(0),
  invite_rebate_setting: z.object({
    count_limit: z.coerce.number().min(-1),
    chain_ratios: z.string().superRefine(validateRatioArray),
    group_settings: z.string().superRefine(validateInviteRebateGroupSettings),
  }),
  TopUpLink: z.string().url().optional().or(z.literal('')),
  general_setting: z.object({
    docs_link: z.string().url().optional().or(z.literal('')),
  }),
  quota_setting: z.object({
    enable_free_model_pre_consume: z.boolean(),
  }),
})

type QuotaFormValues = z.infer<typeof quotaSchema>

type QuotaSettingsSectionProps = {
  defaultValues: QuotaFormValues
}

export function QuotaSettingsSection({
  defaultValues,
}: QuotaSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const [useGroupVisualEditor, setUseGroupVisualEditor] = useState(true)

  const { form, handleSubmit, isDirty, isSubmitting } =
    useSettingsForm<QuotaFormValues>({
      resolver: zodResolver(quotaSchema) as Resolver<
        QuotaFormValues,
        unknown,
        QuotaFormValues
      >,
      defaultValues,
      onSubmit: async (_data, changedFields) => {
        for (const [key, value] of Object.entries(changedFields)) {
          await updateOption.mutateAsync({
            key,
            value: value as string | number | boolean,
          })
        }
      },
    })

  return (
    <SettingsSection
      title={t('Quota Settings')}
      description={t('Configure user quota allocation and rewards')}
    >
      <FormNavigationGuard when={isDirty} />

      <Form {...form}>
        <form onSubmit={handleSubmit} className='space-y-6'>
          <FormDirtyIndicator isDirty={isDirty} />
          <FormField
            control={form.control}
            name='QuotaForNewUser'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('New User Quota')}</FormLabel>
                <FormControl>
                  <Input
                    type='number'
                    value={field.value as number}
                    onChange={(e) => field.onChange(e.target.valueAsNumber)}
                    name={field.name}
                    onBlur={field.onBlur}
                    ref={field.ref}
                  />
                </FormControl>
                <FormDescription>
                  {t('Initial quota given to new users')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='PreConsumedQuota'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Pre-Consumed Quota')}</FormLabel>
                <FormControl>
                  <Input
                    type='number'
                    value={field.value as number}
                    onChange={(e) => field.onChange(e.target.valueAsNumber)}
                    name={field.name}
                    onBlur={field.onBlur}
                    ref={field.ref}
                  />
                </FormControl>
                <FormDescription>
                  {t('Quota consumed before charging users')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='QuotaForInviter'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Inviter Reward')}</FormLabel>
                <FormControl>
                  <Input
                    type='number'
                    value={field.value as number}
                    onChange={(e) => field.onChange(e.target.valueAsNumber)}
                    name={field.name}
                    onBlur={field.onBlur}
                    ref={field.ref}
                  />
                </FormControl>
                <FormDescription>
                  {t('Quota given to users who invite others')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='QuotaForInvitee'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Invitee Reward')}</FormLabel>
                <FormControl>
                  <Input
                    type='number'
                    value={field.value as number}
                    onChange={(e) => field.onChange(e.target.valueAsNumber)}
                    name={field.name}
                    onBlur={field.onBlur}
                    ref={field.ref}
                  />
                </FormControl>
                <FormDescription>
                  {t('Quota given to invited users')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='invite_rebate_setting.count_limit'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Invite Rebate Count Limit')}</FormLabel>
                <FormControl>
                  <Input
                    type='number'
                    value={field.value as number}
                    onChange={(e) => field.onChange(e.target.valueAsNumber)}
                    name={field.name}
                    onBlur={field.onBlur}
                    ref={field.ref}
                  />
                </FormControl>
                <FormDescription>
                  {t('-1 means unlimited, 0 disables recharge rebates')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='invite_rebate_setting.chain_ratios'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Invite Rebate Chain Ratios')}</FormLabel>
                <FormControl>
                  <Textarea rows={4} placeholder='[0.3,0.2,0.1]' {...field} />
                </FormControl>
                <FormDescription>
                  {t('JSON array of rebate ratios from direct inviter upward')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='invite_rebate_setting.group_settings'
            render={({ field }) => (
              <FormItem>
                <div className='flex items-center justify-between'>
                  <FormLabel>
                    {t('Group-based invite rebate chain ratios')}
                  </FormLabel>
                  <Button
                    type='button'
                    variant='outline'
                    size='sm'
                    onClick={() =>
                      setUseGroupVisualEditor(!useGroupVisualEditor)
                    }
                  >
                    {useGroupVisualEditor ? (
                      <>
                        <Code2 className='mr-2 h-4 w-4' />
                        {t('JSON Mode')}
                      </>
                    ) : (
                      <>
                        <Palette className='mr-2 h-4 w-4' />
                        {t('Visual Mode')}
                      </>
                    )}
                  </Button>
                </div>
                <FormControl>
                  {useGroupVisualEditor ? (
                    <InviteRebateGroupVisualEditor
                      value={field.value || ''}
                      onChange={field.onChange}
                    />
                  ) : (
                    <Textarea
                      rows={8}
                      placeholder={`{\n  "vip": {\n    "count_limit": -1,\n    "chain_ratios": [0.3, 0.15]\n  }\n}`}
                      className='font-mono text-sm'
                      {...field}
                    />
                  )}
                </FormControl>
                {!useGroupVisualEditor && (
                  <FormDescription>
                    <div className='space-y-1 text-xs'>
                      <p className='font-semibold'>{t('Format:')}</p>
                      <ul className='list-inside list-disc space-y-0.5 pl-2'>
                        <li>
                          {t('JSON object:')}{' '}
                          {`{"groupName": {"count_limit": -1, "chain_ratios": [0.3, 0.15]}}`}
                        </li>
                        <li>
                          {t('Rules apply by beneficiary inviter group.')}
                        </li>
                        <li>
                          {t(
                            'Missing group falls back to global invite rebate settings.'
                          )}
                        </li>
                        <li>
                          {t(
                            'Configured group without a ratio for the current level gives no rebate for that level.'
                          )}
                        </li>
                      </ul>
                    </div>
                  </FormDescription>
                )}
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='quota_setting.enable_free_model_pre_consume'
            render={({ field }) => (
              <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                <div className='space-y-0.5'>
                  <FormLabel className='text-base'>
                    {t('Pre-Consume for Free Models')}
                  </FormLabel>
                  <FormDescription>
                    {t(
                      'When enabled, zero-cost models also pre-consume quota before final settlement.'
                    )}
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                    disabled={updateOption.isPending}
                  />
                </FormControl>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='TopUpLink'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Top-Up Link')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder={t('https://example.com/topup')}
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t('External link for users to purchase quota')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='general_setting.docs_link'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Documentation Link')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder={t('https://docs.example.com')}
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t('Link to your documentation site')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <Button
            type='submit'
            disabled={updateOption.isPending || isSubmitting}
          >
            {updateOption.isPending ? t('Saving...') : t('Save Changes')}
          </Button>
        </form>
      </Form>
    </SettingsSection>
  )
}
