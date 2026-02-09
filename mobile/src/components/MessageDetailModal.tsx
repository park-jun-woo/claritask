import React from 'react';
import {View, StyleSheet, ScrollView, Modal} from 'react-native';
import {
  Text,
  useTheme,
  IconButton,
  Chip,
  Divider,
} from 'react-native-paper';
import MarkdownRenderer from './MarkdownRenderer';
import type {Message} from '../types';

interface MessageDetailModalProps {
  message: Message | null;
  visible: boolean;
  onDismiss: () => void;
}

export default function MessageDetailModal({
  message,
  visible,
  onDismiss,
}: MessageDetailModalProps) {
  const theme = useTheme();

  if (!message) return null;

  const statusColor = getStatusColor(message.status);
  const statusLabel = getStatusLabel(message.status);

  return (
    <Modal
      visible={visible}
      animationType="slide"
      presentationStyle="pageSheet"
      onRequestClose={onDismiss}>
      <View style={[styles.container, {backgroundColor: theme.colors.background}]}>
        {/* Header */}
        <View style={[styles.header, {borderBottomColor: theme.colors.outline}]}>
          <View style={styles.headerLeft}>
            <IconButton icon="close" size={24} onPress={onDismiss} />
            <Text variant="titleMedium" style={styles.headerTitle}>
              Message #{message.id}
            </Text>
          </View>
          <View style={styles.headerChips}>
            <Chip
              compact
              textStyle={styles.chipText}
              style={[styles.chip, {backgroundColor: statusColor + '20'}]}>
              <Text style={{color: statusColor, fontSize: 12, fontWeight: '600'}}>
                {statusLabel}
              </Text>
            </Chip>
            <Chip
              compact
              icon={sourceIcon(message.source)}
              textStyle={styles.chipText}
              style={styles.chip}>
              {message.source}
            </Chip>
          </View>
        </View>

        {/* Body */}
        <ScrollView style={styles.body} contentContainerStyle={styles.bodyContent}>
          {/* Timestamp */}
          <Text style={[styles.meta, {color: theme.colors.onSurfaceVariant}]}>
            {formatDateTime(message.created_at)}
            {message.completed_at && ` → ${formatDateTime(message.completed_at)}`}
          </Text>

          {/* Content */}
          <View style={styles.section}>
            <Text
              variant="labelMedium"
              style={[styles.sectionLabel, {color: theme.colors.onSurfaceVariant}]}>
              Content
            </Text>
            <Text style={[styles.contentText, {color: theme.colors.onSurface}]}>
              {message.content}
            </Text>
          </View>

          {/* Result */}
          {message.result ? (
            <>
              <Divider style={styles.divider} />
              <View style={styles.section}>
                <Text
                  variant="labelMedium"
                  style={[styles.sectionLabel, {color: theme.colors.onSurfaceVariant}]}>
                  Result
                </Text>
                <View
                  style={[
                    styles.resultBox,
                    {backgroundColor: theme.colors.surfaceVariant},
                  ]}>
                  <MarkdownRenderer content={message.result} />
                </View>
              </View>
            </>
          ) : null}

          {/* Error */}
          {message.error ? (
            <>
              <Divider style={styles.divider} />
              <View style={styles.section}>
                <Text
                  variant="labelMedium"
                  style={[styles.sectionLabel, {color: theme.colors.error}]}>
                  Error
                </Text>
                <View
                  style={[
                    styles.errorBox,
                    {backgroundColor: theme.colors.error + '15'},
                  ]}>
                  <Text style={[styles.errorText, {color: theme.colors.error}]}>
                    {message.error}
                  </Text>
                </View>
              </View>
            </>
          ) : null}
        </ScrollView>
      </View>
    </Modal>
  );
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'processing':
      return '#f59e0b';
    case 'done':
      return '#22c55e';
    case 'failed':
      return '#ef4444';
    case 'pending':
      return '#9ca3af';
    default:
      return '#9ca3af';
  }
}

function getStatusLabel(status: string): string {
  switch (status) {
    case 'processing':
      return '처리 중';
    case 'done':
      return '완료';
    case 'failed':
      return '실패';
    case 'pending':
      return '대기 중';
    default:
      return status;
  }
}

function sourceIcon(source: string): string {
  switch (source) {
    case 'telegram':
      return 'send';
    case 'cli':
      return 'console';
    case 'gui':
      return 'monitor';
    case 'schedule':
      return 'clock-outline';
    case 'mobile':
      return 'cellphone';
    default:
      return 'message-outline';
  }
}

function formatDateTime(ts: string): string {
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

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    borderBottomWidth: 1,
    paddingRight: 12,
  },
  headerLeft: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  headerTitle: {
    fontWeight: '600',
  },
  headerChips: {
    flexDirection: 'row',
    gap: 6,
  },
  chip: {
    height: 28,
  },
  chipText: {
    fontSize: 12,
  },
  body: {
    flex: 1,
  },
  bodyContent: {
    padding: 16,
  },
  meta: {
    fontSize: 12,
    marginBottom: 16,
  },
  section: {
    marginBottom: 8,
  },
  sectionLabel: {
    marginBottom: 6,
    fontWeight: '600',
  },
  contentText: {
    fontSize: 14,
    lineHeight: 22,
  },
  divider: {
    marginVertical: 16,
  },
  resultBox: {
    borderRadius: 8,
    padding: 12,
  },
  errorBox: {
    borderRadius: 8,
    padding: 12,
  },
  errorText: {
    fontSize: 14,
    fontFamily: 'monospace',
  },
});
