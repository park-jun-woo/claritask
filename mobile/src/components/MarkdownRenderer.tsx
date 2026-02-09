import React from 'react';
import {StyleSheet} from 'react-native';
import {useTheme} from 'react-native-paper';
import Markdown from 'react-native-markdown-display';
import type {MD3Theme} from 'react-native-paper';
import EmptyState from './EmptyState';

interface MarkdownRendererProps {
  content: string;
}

function buildStyles(theme: MD3Theme) {
  const codeBackground = theme.dark
    ? theme.colors.elevation.level2
    : theme.colors.elevation.level2;

  return StyleSheet.create({
    body: {
      color: theme.colors.onSurface,
      fontSize: 14,
      lineHeight: 22,
    },
    heading1: {
      fontSize: 22,
      fontWeight: '700',
      color: theme.colors.onSurface,
      marginTop: 16,
      marginBottom: 8,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.outline,
      paddingBottom: 6,
    },
    heading2: {
      fontSize: 19,
      fontWeight: '600',
      color: theme.colors.onSurface,
      marginTop: 14,
      marginBottom: 6,
    },
    heading3: {
      fontSize: 16,
      fontWeight: '600',
      color: theme.colors.onSurface,
      marginTop: 12,
      marginBottom: 4,
    },
    heading4: {
      fontSize: 15,
      fontWeight: '600',
      color: theme.colors.onSurface,
      marginTop: 10,
      marginBottom: 4,
    },
    paragraph: {
      marginTop: 4,
      marginBottom: 4,
    },
    link: {
      color: theme.dark ? '#60a5fa' : '#2563eb',
      textDecorationLine: 'underline',
    },
    blockquote: {
      backgroundColor: codeBackground,
      borderLeftWidth: 4,
      borderLeftColor: theme.colors.onSurfaceVariant,
      paddingLeft: 12,
      paddingVertical: 4,
      marginVertical: 8,
    },
    code_inline: {
      backgroundColor: codeBackground,
      borderRadius: 4,
      paddingHorizontal: 4,
      paddingVertical: 1,
      fontFamily: 'monospace',
      fontSize: 13,
      color: theme.colors.onSurface,
    },
    code_block: {
      backgroundColor: codeBackground,
      borderRadius: 6,
      padding: 12,
      fontFamily: 'monospace',
      fontSize: 13,
      color: theme.colors.onSurface,
      marginVertical: 8,
    },
    fence: {
      backgroundColor: codeBackground,
      borderRadius: 6,
      padding: 12,
      fontFamily: 'monospace',
      fontSize: 13,
      color: theme.colors.onSurface,
      marginVertical: 8,
    },
    table: {
      borderWidth: 1,
      borderColor: theme.colors.outline,
      borderRadius: 4,
      marginVertical: 8,
    },
    thead: {
      backgroundColor: codeBackground,
    },
    th: {
      padding: 8,
      fontWeight: '600',
      borderRightWidth: 1,
      borderBottomWidth: 1,
      borderColor: theme.colors.outline,
    },
    td: {
      padding: 8,
      borderRightWidth: 1,
      borderBottomWidth: 1,
      borderColor: theme.colors.outline,
    },
    tr: {
      borderBottomWidth: 1,
      borderColor: theme.colors.outline,
    },
    bullet_list: {
      marginVertical: 4,
    },
    ordered_list: {
      marginVertical: 4,
    },
    list_item: {
      marginVertical: 2,
    },
    hr: {
      backgroundColor: theme.colors.outline,
      height: 1,
      marginVertical: 12,
    },
    strong: {
      fontWeight: '700',
    },
    em: {
      fontStyle: 'italic',
    },
    s: {
      textDecorationLine: 'line-through',
    },
  });
}

export default function MarkdownRenderer({content}: MarkdownRendererProps) {
  const theme = useTheme();

  if (!content || content.trim() === '') {
    return <EmptyState />;
  }

  const markdownStyles = buildStyles(theme);

  return <Markdown style={markdownStyles}>{content}</Markdown>;
}
