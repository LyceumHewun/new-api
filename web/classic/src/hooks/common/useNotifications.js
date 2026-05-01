/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { useState, useEffect } from 'react';
import { API } from '../../helpers';

export const useNotifications = (statusState) => {
  const [noticeVisible, setNoticeVisible] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const [defaultTab, setDefaultTab] = useState('inApp');
  const [hasNoticeContent, setHasNoticeContent] = useState(false);

  const announcements = statusState?.status?.announcements || [];

  // Helper functions
  const getAnnouncementKey = (a) =>
    `${a?.publishDate || ''}-${(a?.content || '').slice(0, 30)}`;

  const calculateUnreadCount = () => {
    if (!announcements.length) return 0;
    let readKeys = [];
    try {
      readKeys = JSON.parse(localStorage.getItem('notice_read_keys')) || [];
    } catch (_) {
      readKeys = [];
    }
    const readSet = new Set(readKeys);
    return announcements.filter((a) => !readSet.has(getAnnouncementKey(a)))
      .length;
  };

  const getUnreadKeys = () => {
    if (!announcements.length) return [];
    let readKeys = [];
    try {
      readKeys = JSON.parse(localStorage.getItem('notice_read_keys')) || [];
    } catch (_) {
      readKeys = [];
    }
    const readSet = new Set(readKeys);
    return announcements
      .filter((a) => !readSet.has(getAnnouncementKey(a)))
      .map(getAnnouncementKey);
  };

  // Effects
  useEffect(() => {
    const nextUnreadCount = calculateUnreadCount();
    setUnreadCount(nextUnreadCount);
    if (nextUnreadCount > 0 && !hasNoticeContent) {
      setDefaultTab('system');
      setNoticeVisible(true);
    }
  }, [announcements, hasNoticeContent]);

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      try {
        const res = await API.get('/api/notice');
        const { success, data } = res.data;
        if (success && data && data.trim() !== '') {
          setHasNoticeContent(true);
          setDefaultTab('inApp');
          setNoticeVisible(true);
        }
      } catch (error) {
        console.error('获取公告失败:', error);
      }
    };

    checkNoticeAndShow();
  }, []);

  // Actions
  const handleNoticeOpen = () => {
    setNoticeVisible(true);
  };

  const handleNoticeClose = () => {
    setNoticeVisible(false);
    if (announcements.length) {
      let readKeys = [];
      try {
        readKeys = JSON.parse(localStorage.getItem('notice_read_keys')) || [];
      } catch (_) {
        readKeys = [];
      }
      const mergedKeys = Array.from(
        new Set([...readKeys, ...announcements.map(getAnnouncementKey)]),
      );
      localStorage.setItem('notice_read_keys', JSON.stringify(mergedKeys));
    }
    setUnreadCount(0);
  };

  return {
    noticeVisible,
    unreadCount,
    defaultTab,
    announcements,
    handleNoticeOpen,
    handleNoticeClose,
    getUnreadKeys,
  };
};
