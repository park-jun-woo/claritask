import React, {useState, useRef, useEffect, useCallback, useMemo} from 'react';
import {
  View,
  StyleSheet,
  FlatList,
  TextInput,
  KeyboardAvoidingView,
  Platform,
  RefreshControl,
} from 'react-native';
import {Text, useTheme, IconButton} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import ChatBubble from '../components/ChatBubble';
import MessageDetailModal from '../components/MessageDetailModal';
import ProjectSelector from '../components/ProjectSelector';
import {useMessages, useMessage, useSendMessage, useProjects} from '../hooks/useClaribot';

interface PendingMessage {
  id: string;
  content: string;
  created_at: string;
}

export default function MessagesScreen() {
  const theme = useTheme();
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);
  const [input, setInput] = useState('');
  const [pendingMessages, setPendingMessages] = useState<PendingMessage[]>([]);
  const [selectedMessageId, setSelectedMessageId] = useState<number | undefined>();
  const [detailVisible, setDetailVisible] = useState(false);
  const flatListRef = useRef<FlatList>(null);
  const isInitialLoad = useRef(true);

  const isGlobal = selectedProjectId === null;
  const {data: messagesData, refetch, isRefetching} = useMessages(
    isGlobal,
    isGlobal ? undefined : selectedProjectId,
  );
  const {data: projectsData} = useProjects(true);
  const {data: messageDetail} = useMessage(selectedMessageId);
  const sendMessage = useSendMessage();

  const projects = useMemo(() => {
    const d = projectsData?.data;
    if (!d) return [];
    if (Array.isArray(d)) return d;
    if (d.items && Array.isArray(d.items)) return d.items;
    return [];
  }, [projectsData]);

  const messageItems = useMemo(() => parseItems(messagesData?.data), [messagesData]);

  // Remove pending messages that appeared in actual messages
  const actualContents = useMemo(
    () => new Set(messageItems.map((m: any) => m.content || m.Content || '')),
    [messageItems],
  );

  useEffect(() => {
    const filtered = pendingMessages.filter(pm => !actualContents.has(pm.content));
    if (filtered.length < pendingMessages.length) {
      setPendingMessages(filtered);
    }
  }, [actualContents, pendingMessages]);

  // Merge pending with actual, sort by time
  const sortedMessages = useMemo(() => {
    const merged = [
      ...messageItems,
      ...pendingMessages
        .filter(pm => !actualContents.has(pm.content))
        .map(pm => ({
          id: pm.id,
          content: pm.content,
          status: 'pending' as const,
          source: 'mobile' as const,
          result: '',
          error: '',
          created_at: pm.created_at,
          completed_at: null,
          project_id: selectedProjectId,
        })),
    ];
    return merged.sort((a, b) => {
      const ta = new Date(a.created_at || (a as any).CreatedAt || 0).getTime();
      const tb = new Date(b.created_at || (b as any).CreatedAt || 0).getTime();
      return ta - tb;
    });
  }, [messageItems, pendingMessages, actualContents, selectedProjectId]);

  // Group messages by date
  const groupedData = useMemo(() => groupByDate(sortedMessages), [sortedMessages]);

  // Flatten for FlatList: [{type:'separator', date}, {type:'message', msg}, ...]
  const flatData = useMemo(() => {
    const result: Array<
      | {type: 'separator'; key: string; date: string}
      | {type: 'message'; key: string; msg: any}
    > = [];
    for (const group of groupedData) {
      result.push({type: 'separator', key: `sep-${group.date}`, date: group.date});
      for (const msg of group.messages) {
        const id = msg.id || (msg as any).ID;
        result.push({type: 'message', key: `msg-${id}`, msg});
      }
    }
    return result;
  }, [groupedData]);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    if (flatData.length > 0) {
      const delay = isInitialLoad.current ? 100 : 300;
      const timer = setTimeout(() => {
        flatListRef.current?.scrollToEnd({animated: !isInitialLoad.current});
        isInitialLoad.current = false;
      }, delay);
      return () => clearTimeout(timer);
    }
  }, [flatData.length]);

  const handleSend = useCallback(() => {
    const trimmed = input.trim();
    if (!trimmed) return;

    const tempId = `pending-${Date.now()}`;
    const now = new Date().toISOString();

    setPendingMessages(prev => [...prev, {id: tempId, content: trimmed, created_at: now}]);
    setInput('');

    sendMessage.mutate(
      {content: trimmed, projectId: selectedProjectId || undefined},
      {
        onError: () => {
          setPendingMessages(prev => prev.filter(m => m.id !== tempId));
        },
      },
    );
  }, [input, selectedProjectId, sendMessage]);

  const handleDetailClick = useCallback((id: number) => {
    setSelectedMessageId(id);
    setDetailVisible(true);
  }, []);

  const handleDismissDetail = useCallback(() => {
    setDetailVisible(false);
  }, []);

  const renderItem = useCallback(
    ({item}: {item: (typeof flatData)[number]}) => {
      if (item.type === 'separator') {
        return <DateSeparator date={item.date} />;
      }

      const msg = item.msg;
      const id = msg.id || msg.ID;
      const content = msg.content || msg.Content || '';
      const status = msg.status || msg.Status || 'pending';
      const source = msg.source || msg.Source || 'mobile';
      const result = msg.result || msg.Result || '';
      const error = msg.error || msg.Error || '';
      const createdAt = msg.created_at || msg.CreatedAt || '';
      const completedAt = msg.completed_at || msg.CompletedAt || '';

      return (
        <View>
          {/* User message bubble */}
          <ChatBubble
            type="user"
            content={content}
            source={source}
            time={formatTime(createdAt)}
          />

          {/* Bot response bubble */}
          <ChatBubble
            type="bot"
            content={
              error
                ? error.slice(0, 120) + (error.length > 120 ? '...' : '')
                : result
                  ? getResponseSummary(result)
                  : statusMessage(status)
            }
            result={result || undefined}
            status={status}
            time={formatTime(completedAt || createdAt)}
            onDetailClick={typeof id === 'number' ? () => handleDetailClick(id) : undefined}
          />
        </View>
      );
    },
    [handleDetailClick],
  );

  const selectedMsg = messageDetail?.data || null;

  return (
    <KeyboardAvoidingView
      style={[styles.container, {backgroundColor: theme.colors.background}]}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      keyboardVerticalOffset={Platform.OS === 'ios' ? 90 : 0}>
      {/* Header */}
      <View style={[styles.header, {borderBottomColor: theme.colors.outline}]}>
        <View style={styles.headerRow}>
          <MaterialCommunityIcons
            name="message-text-outline"
            size={20}
            color={theme.colors.onSurface}
          />
          <Text variant="titleMedium" style={styles.headerTitle}>
            Messages
          </Text>
        </View>
        <View style={styles.selectorWrap}>
          <ProjectSelector
            projects={projects}
            selectedId={selectedProjectId}
            onSelect={setSelectedProjectId}
          />
        </View>
      </View>

      {/* Chat Messages */}
      <FlatList
        ref={flatListRef}
        data={flatData}
        keyExtractor={item => item.key}
        renderItem={renderItem}
        contentContainerStyle={styles.listContent}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl refreshing={isRefetching} onRefresh={refetch} />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <MaterialCommunityIcons
              name="message-text-outline"
              size={48}
              color={theme.colors.onSurfaceVariant}
              style={{opacity: 0.3}}
            />
            <Text style={[styles.emptyText, {color: theme.colors.onSurfaceVariant}]}>
              메시지가 없습니다
            </Text>
          </View>
        }
      />

      {/* Input Area */}
      <View style={[styles.inputArea, {borderTopColor: theme.colors.outline}]}>
        <TextInput
          style={[
            styles.textInput,
            {
              backgroundColor: theme.colors.surfaceVariant,
              color: theme.colors.onSurface,
              borderColor: theme.colors.outline,
            },
          ]}
          placeholder="메시지를 입력하세요..."
          placeholderTextColor={theme.colors.onSurfaceVariant}
          value={input}
          onChangeText={setInput}
          multiline
          maxLength={2000}
          returnKeyType="default"
        />
        <IconButton
          icon="send"
          mode="contained"
          size={22}
          onPress={handleSend}
          disabled={!input.trim()}
          style={styles.sendButton}
        />
      </View>

      {/* Message Detail Modal */}
      <MessageDetailModal
        message={selectedMsg}
        visible={detailVisible}
        onDismiss={handleDismissDetail}
      />
    </KeyboardAvoidingView>
  );
}

// --- Sub-components ---

function DateSeparator({date}: {date: string}) {
  const theme = useTheme();
  return (
    <View style={styles.dateSeparator}>
      <View style={[styles.dateLine, {backgroundColor: theme.colors.outline}]} />
      <Text style={[styles.dateText, {color: theme.colors.onSurfaceVariant}]}>
        {date}
      </Text>
      <View style={[styles.dateLine, {backgroundColor: theme.colors.outline}]} />
    </View>
  );
}

// --- Utilities ---

function formatTime(ts: string): string {
  if (!ts) return '';
  try {
    const d = new Date(ts);
    return d.toLocaleString('ko-KR', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return ts;
  }
}

function formatDate(ts: string): string {
  if (!ts) return '';
  try {
    const d = new Date(ts);
    return d.toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  } catch {
    return ts;
  }
}

function parseItems(data: any): any[] {
  if (!data) return [];
  if (Array.isArray(data)) return data;
  if (data.items && Array.isArray(data.items)) return data.items;
  return [];
}

function getResponseSummary(result: string): string {
  if (!result) return '';
  const plain = result.replace(/[#*`>\-[\]()!]/g, '').trim();
  const first = plain.split('\n').find(l => l.trim())?.trim() || '';
  return first.length > 100 ? first.slice(0, 100) + '...' : first;
}

function statusMessage(status: string): string {
  switch (status) {
    case 'pending':
      return '대기 중...';
    case 'processing':
      return '처리 중...';
    case 'done':
      return '완료';
    case 'failed':
      return '실패';
    default:
      return status;
  }
}

function groupByDate(messages: any[]): {date: string; messages: any[]}[] {
  const groups: Map<string, any[]> = new Map();
  for (const msg of messages) {
    const ts = msg.created_at || msg.CreatedAt || '';
    const dateKey = formatDate(ts) || 'Unknown';
    if (!groups.has(dateKey)) {
      groups.set(dateKey, []);
    }
    groups.get(dateKey)!.push(msg);
  }
  return Array.from(groups.entries()).map(([date, msgs]) => ({date, messages: msgs}));
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  header: {
    paddingHorizontal: 16,
    paddingTop: 12,
    paddingBottom: 8,
    borderBottomWidth: 1,
    gap: 8,
  },
  headerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  headerTitle: {
    fontWeight: '600',
  },
  selectorWrap: {
    marginBottom: 4,
  },
  listContent: {
    paddingHorizontal: 16,
    paddingVertical: 12,
    flexGrow: 1,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 80,
    gap: 8,
  },
  emptyText: {
    fontSize: 14,
  },
  inputArea: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderTopWidth: 1,
    gap: 4,
  },
  textInput: {
    flex: 1,
    borderRadius: 20,
    borderWidth: 1,
    paddingHorizontal: 16,
    paddingVertical: 10,
    fontSize: 14,
    maxHeight: 100,
    minHeight: 44,
  },
  sendButton: {
    marginBottom: 2,
  },
  dateSeparator: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 16,
    gap: 12,
  },
  dateLine: {
    flex: 1,
    height: StyleSheet.hairlineWidth,
  },
  dateText: {
    fontSize: 11,
    fontWeight: '500',
  },
});
