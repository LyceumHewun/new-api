import { useMemo, useState } from 'react'
import { Pencil, Plus, Search, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { safeJsonParseWithValidation } from '../utils/json-parser'
import { isObjectRecord } from '../utils/json-validators'
import {
  InviteRebateGroupDialog,
  type InviteRebateGroupEntryData,
} from './invite-rebate-group-dialog'

type InviteRebateGroupVisualEditorProps = {
  value: string
  onChange: (value: string) => void
}

type InviteRebateGroupEntry = InviteRebateGroupEntryData

export function InviteRebateGroupVisualEditor({
  value,
  onChange,
}: InviteRebateGroupVisualEditorProps) {
  const { t } = useTranslation()
  const [searchText, setSearchText] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editData, setEditData] = useState<InviteRebateGroupEntry | null>(null)

  const groupSettings = useMemo(() => {
    if (!value || value.trim() === '') return []

    const parsed = safeJsonParseWithValidation<Record<string, unknown>>(value, {
      fallback: {},
      validator: isObjectRecord,
      validatorMessage: 'Invite rebate group settings must be a JSON object',
      context: 'invite rebate group settings',
    })

    return Object.entries(parsed)
      .map(([groupName, groupSetting]) => {
        if (!isObjectRecord(groupSetting)) return null

        const countLimit = groupSetting.count_limit
        const chainRatios = groupSetting.chain_ratios
        if (
          typeof countLimit === 'number' &&
          Array.isArray(chainRatios) &&
          chainRatios.every((ratio) => typeof ratio === 'number')
        ) {
          return {
            groupName,
            countLimit,
            chainRatios,
          }
        }
        return null
      })
      .filter((item): item is InviteRebateGroupEntry => item !== null)
  }, [value])

  const filteredGroupSettings = useMemo(() => {
    if (!searchText) return groupSettings
    const lowerSearch = searchText.toLowerCase()
    return groupSettings.filter((setting) =>
      setting.groupName.toLowerCase().includes(lowerSearch)
    )
  }, [groupSettings, searchText])

  const handleSave = (data: InviteRebateGroupEntryData) => {
    const parsed = safeJsonParseWithValidation<Record<string, unknown>>(value, {
      fallback: {},
      validator: isObjectRecord,
      silent: true,
    })

    if (editData && editData.groupName !== data.groupName) {
      delete parsed[editData.groupName]
    }

    parsed[data.groupName] = {
      count_limit: data.countLimit,
      chain_ratios: data.chainRatios,
    }

    onChange(JSON.stringify(parsed, null, 2))
  }

  const handleDelete = (groupName: string) => {
    const parsed = safeJsonParseWithValidation<Record<string, unknown>>(value, {
      fallback: {},
      validator: isObjectRecord,
      silent: true,
    })

    delete parsed[groupName]

    onChange(JSON.stringify(parsed, null, 2))
  }

  const handleEdit = (setting: InviteRebateGroupEntry) => {
    setEditData(setting)
    setDialogOpen(true)
  }

  const handleAdd = () => {
    setEditData(null)
    setDialogOpen(true)
  }

  return (
    <div className='space-y-4'>
      <div className='flex items-center gap-4'>
        <div className='relative flex-1'>
          <Search className='text-muted-foreground absolute top-2.5 left-2.5 h-4 w-4' />
          <Input
            placeholder={t('Search group names...')}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            className='pl-9'
          />
        </div>
        <Button type='button' onClick={handleAdd}>
          <Plus className='mr-2 h-4 w-4' />
          {t('Add group')}
        </Button>
      </div>

      {filteredGroupSettings.length === 0 ? (
        <div className='text-muted-foreground rounded-lg border border-dashed p-8 text-center'>
          {searchText
            ? t('No groups match your search')
            : t(
                'No group-based invite rebate rules configured. Click "Add group" to get started.'
              )}
        </div>
      ) : (
        <div className='rounded-md border'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Group Name')}</TableHead>
                <TableHead className='text-right'>
                  {t('Invite Rebate Count Limit')}
                </TableHead>
                <TableHead>{t('Invite Rebate Chain Ratios')}</TableHead>
                <TableHead className='text-right'>{t('Actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredGroupSettings.map((setting) => (
                <TableRow key={setting.groupName}>
                  <TableCell className='font-medium'>
                    {setting.groupName}
                  </TableCell>
                  <TableCell className='text-right'>
                    <span className='font-mono'>
                      {setting.countLimit === -1
                        ? t('Unlimited')
                        : setting.countLimit.toLocaleString()}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className='font-mono text-sm'>
                      {JSON.stringify(setting.chainRatios)}
                    </span>
                  </TableCell>
                  <TableCell className='text-right'>
                    <div className='flex justify-end gap-2'>
                      <Button
                        type='button'
                        variant='ghost'
                        size='sm'
                        onClick={() => handleEdit(setting)}
                      >
                        <Pencil className='h-4 w-4' />
                      </Button>
                      <Button
                        type='button'
                        variant='ghost'
                        size='sm'
                        onClick={() => handleDelete(setting.groupName)}
                      >
                        <Trash2 className='h-4 w-4' />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <InviteRebateGroupDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onSave={handleSave}
        editData={editData}
      />
    </div>
  )
}
