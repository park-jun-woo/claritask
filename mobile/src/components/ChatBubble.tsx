import React from 'react';
import {View, StyleSheet, TouchableOpacity} from 'react-native';
import {Text, useTheme, ActivityIndicator} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';

interface ChatBubbleProps {
  type: 'user' | 'bot';
  content: string;
  status?: string;
  source?: string;
  result?: string;
  time: string;
  onDetailClick?: () => void;
}

export default function ChatBubble({
  type,
  content,
  status,
  source,
  result,
  time,
  onDetailClick,
}: ChatBubbleProps) {
  const theme = useTheme();
  const isUser = type === 'user';

  return (
    <View style={[styles.row, isUser ? styles.rowRight : styles.rowLeft]}>
      <View style={styles.bubbleWrap}>
        {/* Source label for user messages */}
        {isUser && source && source !== 'gui' && source !== 'mobile' && (
          <Text
            style={[
              styles.sourceLabel,
              {color: theme.colors.onSurfaceVariant},
              isUser && styles.sourceLabelRight,
            ]}>
            {sourceLabel(source)}
          </Text>
        )}

        {/* Bubble */}
        <View
          style={[
            styles.bubble,
            isUser
              ? [
                  styles.bubbleUser,
                  {backgroundColor: theme.colors.primary},
                ]
              : [
                  styles.bubbleBot,
                  {backgroundColor: theme.colors.surfaceVariant},
                ],
          ]}>
          <Text
            style={[
              styles.content,
              {
                color: isUser
                  ? theme.colors.onPrimary
                  : theme.colors.onSurface,
              },
            ]}>
            {content}
          </Text>

          {/* Bot bubble: status + result summary + detail link */}
          {!isUser && (
            <View style={styles.botExtra}>
              <StatusIndicator status={status} />

              {result && (
                <Text
                  style={[styles.resultSummary, {color: theme.colors.onSurfaceVariant}]}
                  numberOfLines={2}>
                  {getFirstLines(result)}
                </Text>
              )}

              {onDetailClick && (result || status === 'done' || status === 'failed') && (
                <TouchableOpacity
                  onPress={onDetailClick}
                  style={styles.detailButton}
                  hitSlop={{top: 8, bottom: 8, left: 8, right: 8}}>
                  <MaterialCommunityIcons
                    name="eye-outline"
                    size={14}
                    color={theme.colors.primary}
                  />
                  <Text style={[styles.detailText, {color: theme.colors.primary}]}>
                    자세히보기
                  </Text>
                </TouchableOpacity>
              )}
            </View>
          )}
        </View>

        {/* Timestamp */}
        <Text
          style={[
            styles.timestamp,
            {color: theme.colors.onSurfaceVariant},
            isUser && styles.timestampRight,
          ]}>
          {time}
        </Text>
      </View>
    </View>
  );
}

function StatusIndicator({status}: {status?: string}) {
  switch (status) {
    case 'processing':
      return (
        <View style={[styles.badge, {backgroundColor: '#f59e0b20'}]}>
          <ActivityIndicator size={12} color="#f59e0b" />
          <Text style={[styles.badgeText, {color: '#f59e0b'}]}>처리 중</Text>
        </View>
      );
    case 'done':
      return (
        <View style={[styles.badge, {backgroundColor: '#22c55e20'}]}>
          <MaterialCommunityIcons name="check" size={12} color="#22c55e" />
          <Text style={[styles.badgeText, {color: '#22c55e'}]}>완료</Text>
        </View>
      );
    case 'failed':
      return (
        <View style={[styles.badge, {backgroundColor: '#ef444420'}]}>
          <MaterialCommunityIcons name="alert-circle-outline" size={12} color="#ef4444" />
          <Text style={[styles.badgeText, {color: '#ef4444'}]}>실패</Text>
        </View>
      );
    case 'pending':
      return (
        <View style={[styles.badge, {backgroundColor: '#9ca3af20'}]}>
          <MaterialCommunityIcons name="clock-outline" size={12} color="#9ca3af" />
          <Text style={[styles.badgeText, {color: '#9ca3af'}]}>대기 중</Text>
        </View>
      );
    default:
      return null;
  }
}

function sourceLabel(source: string): string {
  switch (source) {
    case 'telegram':
      return 'Telegram';
    case 'cli':
      return 'CLI';
    case 'schedule':
      return 'Schedule';
    default:
      return source;
  }
}

function getFirstLines(text: string): string {
  const lines = text
    .replace(/^#+\s/gm, '')
    .replace(/\*\*/g, '')
    .split('\n')
    .filter(l => l.trim());
  return lines.slice(0, 2).join(' ').slice(0, 150);
}

const styles = StyleSheet.create({
  row: {
    marginBottom: 10,
  },
  rowRight: {
    alignItems: 'flex-end',
  },
  rowLeft: {
    alignItems: 'flex-start',
  },
  bubbleWrap: {
    maxWidth: '80%',
  },
  sourceLabel: {
    fontSize: 10,
    marginBottom: 2,
  },
  sourceLabelRight: {
    textAlign: 'right',
  },
  bubble: {
    borderRadius: 16,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  bubbleUser: {
    borderBottomRightRadius: 4,
  },
  bubbleBot: {
    borderBottomLeftRadius: 4,
  },
  content: {
    fontSize: 14,
    lineHeight: 20,
  },
  botExtra: {
    marginTop: 6,
    gap: 4,
  },
  badge: {
    flexDirection: 'row',
    alignItems: 'center',
    alignSelf: 'flex-start',
    gap: 4,
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
  },
  badgeText: {
    fontSize: 10,
    fontWeight: '600',
  },
  resultSummary: {
    fontSize: 12,
    marginTop: 2,
  },
  detailButton: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    marginTop: 2,
    minHeight: 28,
  },
  detailText: {
    fontSize: 12,
    fontWeight: '500',
  },
  timestamp: {
    fontSize: 10,
    marginTop: 2,
  },
  timestampRight: {
    textAlign: 'right',
  },
});
