import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNotificationStore } from '@/stores/notification-store'
import { getNotice } from '@/lib/api'
import { useStatus } from '@/hooks/use-status'

function hashString(input: string): string {
  let hash = 0
  if (!input) return '0'

  for (let i = 0; i < input.length; i += 1) {
    const chr = input.charCodeAt(i)
    hash = (hash << 5) - hash + chr
    hash |= 0
  }

  return hash.toString(36)
}

/**
 * Generate a unique key for an announcement
 * Prefer backend id, fall back to a content hash so edits register
 */
function getAnnouncementKey(item: Record<string, unknown>): string {
  if (!item) return ''

  if (item.id !== undefined && item.id !== null) {
    return `id:${item.id}`
  }

  const fingerprint = JSON.stringify({
    publishDate: (item?.publishDate as string) || '',
    content: ((item?.content as string) || '').trim(),
    extra: ((item?.extra as string) || '').trim(),
    type: (item?.type as string) || '',
    title: ((item?.title as string) || '').trim(),
    link: ((item?.link as string) || '').trim(),
  })
  return `hash:${hashString(fingerprint)}`
}

/**
 * Hook to manage notifications (Notice + Announcements)
 * Provides unread counts and read status management
 */
export function useNotifications() {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [activeTab, setActiveTab] = useState<'notice' | 'announcements'>(
    'notice'
  )

  // Fetch Notice from API
  const {
    data: noticeResponse,
    isLoading: noticeLoading,
    refetch: refetchNotice,
  } = useQuery({
    queryKey: ['notice'],
    queryFn: getNotice,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })

  // Fetch Announcements from status
  const { status, loading: statusLoading } = useStatus()
  const announcementsEnabled = status?.announcements_enabled ?? false
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const announcements: Record<string, unknown>[] = announcementsEnabled
    ? ((status?.announcements || []) as Record<string, unknown>[]).slice(0, 20)
    : []

  // Notification store
  const {
    lastReadNotice,
    markNoticeRead,
    markAnnouncementsRead,
    isAnnouncementRead,
  } = useNotificationStore()

  // Extract notice content
  const noticeContent = noticeResponse?.success
    ? (noticeResponse.data || '').trim()
    : ''

  // Calculate unread counts
  const unreadCounts = useMemo(() => {
    const noticeUnread =
      noticeContent && noticeContent !== lastReadNotice ? 1 : 0

    const announcementsUnread = announcements.filter(
      (item: Record<string, unknown>) => {
        const key = getAnnouncementKey(item)
        return !isAnnouncementRead(key)
      }
    ).length

    return {
      notice: noticeUnread,
      announcements: announcementsUnread,
      total: noticeUnread + announcementsUnread,
    }
  }, [noticeContent, lastReadNotice, announcements, isAnnouncementRead])

  const markAnnouncements = () => {
    if (announcements.length === 0) return
    const allKeys = announcements.map((item: Record<string, unknown>) =>
      getAnnouncementKey(item)
    )
    markAnnouncementsRead(allKeys)
  }

  const markTabRead = (tab: 'notice' | 'announcements') => {
    if (tab === 'notice' && noticeContent) {
      markNoticeRead(noticeContent)
    }
    if (tab === 'announcements') {
      markAnnouncements()
    }
  }

  const getPreferredTab = (): 'notice' | 'announcements' => {
    if (noticeContent) return 'notice'
    if (announcements.length > 0) return 'announcements'
    return 'notice'
  }

  useEffect(() => {
    if (!noticeContent || noticeContent === lastReadNotice) return
    setActiveTab('notice')
    setDialogOpen(true)
  }, [lastReadNotice, noticeContent])

  useEffect(() => {
    if (noticeContent || unreadCounts.announcements <= 0) return
    setActiveTab('announcements')
    setDialogOpen(true)
  }, [noticeContent, unreadCounts.announcements])

  // Handle dialog open
  const handleOpenDialog = (tab?: 'notice' | 'announcements') => {
    const nextTab = tab || getPreferredTab()
    markTabRead(nextTab)
    setActiveTab(nextTab)
    setDialogOpen(true)
  }

  // Handle tab change - mark announcements as read when switching to that tab
  const handleTabChange = (tab: 'notice' | 'announcements') => {
    setActiveTab(tab)

    if (tab === 'announcements' && announcements.length > 0) {
      markAnnouncements()
    }
  }

  const closeDialog = () => {
    markTabRead(activeTab)
    setDialogOpen(false)
  }

  const handleDialogOpenChange = (open: boolean) => {
    if (open) {
      handleOpenDialog(activeTab)
      return
    }
    closeDialog()
  }

  return {
    // Data
    notice: noticeContent,
    announcements,
    loading: noticeLoading || statusLoading,

    // Unread counts
    unreadCount: unreadCounts.total,
    unreadNoticeCount: unreadCounts.notice,
    unreadAnnouncementsCount: unreadCounts.announcements,

    // Dialog state
    dialogOpen,
    setDialogOpen: handleDialogOpenChange,
    activeTab,
    setActiveTab: handleTabChange,

    // Actions
    openDialog: handleOpenDialog,
    closeDialog,
    refetchNotice,
  }
}
