import { useEffect } from 'react'
import * as z from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import { Textarea } from '@/components/ui/textarea'

const inviteRebateGroupDialogSchema = z.object({
  groupName: z.string().min(1, 'Group name is required'),
  countLimit: z
    .number()
    .min(-1, 'Must be ≥ -1')
    .max(2147483647, 'Must be ≤ 2,147,483,647'),
  chainRatios: z.string().superRefine((value, ctx) => {
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
  }),
})

type InviteRebateGroupDialogFormValues = z.infer<
  typeof inviteRebateGroupDialogSchema
>

export type InviteRebateGroupEntryData = {
  groupName: string
  countLimit: number
  chainRatios: number[]
}

type InviteRebateGroupDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (data: InviteRebateGroupEntryData) => void
  editData?: InviteRebateGroupEntryData | null
}

export function InviteRebateGroupDialog({
  open,
  onOpenChange,
  onSave,
  editData,
}: InviteRebateGroupDialogProps) {
  const { t } = useTranslation()
  const isEditMode = !!editData

  const form = useForm<InviteRebateGroupDialogFormValues>({
    resolver: zodResolver(inviteRebateGroupDialogSchema),
    defaultValues: {
      groupName: '',
      countLimit: -1,
      chainRatios: '[]',
    },
  })

  useEffect(() => {
    if (editData) {
      form.reset({
        groupName: editData.groupName,
        countLimit: editData.countLimit,
        chainRatios: JSON.stringify(editData.chainRatios),
      })
      return
    }

    form.reset({
      groupName: '',
      countLimit: -1,
      chainRatios: '[]',
    })
  }, [editData, form, open])

  const handleSubmit = (values: InviteRebateGroupDialogFormValues) => {
    onSave({
      groupName: values.groupName,
      countLimit: values.countLimit,
      chainRatios: JSON.parse(values.chainRatios || '[]') as number[],
    })
    form.reset()
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[500px]'>
        <DialogHeader>
          <DialogTitle>
            {isEditMode
              ? t('Edit group invite rebate')
              : t('Add group invite rebate')}
          </DialogTitle>
          <DialogDescription>
            {t('Configure invite rebate rules for a specific user group.')}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className='space-y-4'
          >
            <FormField
              control={form.control}
              name='groupName'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Group Name')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder={t('e.g., default, vip, premium')}
                      {...field}
                      disabled={isEditMode}
                    />
                  </FormControl>
                  <FormDescription>
                    {isEditMode
                      ? t('Group name cannot be changed when editing.')
                      : t('Unique identifier for this group.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='countLimit'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Invite Rebate Count Limit')}</FormLabel>
                  <FormControl>
                    <div className='flex items-center gap-2'>
                      <Input
                        type='number'
                        min={-1}
                        max={2147483647}
                        step={1}
                        {...field}
                        onChange={(e) =>
                          field.onChange(parseInt(e.target.value) || 0)
                        }
                      />
                      <span className='text-muted-foreground text-sm'>
                        {t('times')}
                      </span>
                    </div>
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
              name='chainRatios'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Invite Rebate Chain Ratios')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={4}
                      placeholder='[0.3, 0.15, 0.05]'
                      className='font-mono text-sm'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'JSON array of rebate ratios from direct inviter upward'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                onClick={() => onOpenChange(false)}
              >
                {t('Cancel')}
              </Button>
              <Button type='submit'>
                {isEditMode ? t('Update') : t('Add')}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
