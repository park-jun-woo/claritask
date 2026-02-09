import React from 'react';
import {View, ScrollView, StyleSheet} from 'react-native';
import {Text, useTheme, ActivityIndicator} from 'react-native-paper';
import type {StackScreenProps} from '@react-navigation/stack';
import type {MoreStackParamList} from '../../navigation/MoreStackNavigator';
import type {Spec} from '../../types';
import {useSpec} from '../../hooks/useClaribot';
import MarkdownRenderer from '../../components/MarkdownRenderer';

type Props = StackScreenProps<MoreStackParamList, 'SpecDetail'>;

const statusConfig: Record<
  Spec['status'],
  {label: string; bg: string}
> = {
  draft: {label: 'Draft', bg: '#6b7280'},
  review: {label: 'Review', bg: '#3b82f6'},
  approved: {label: 'Approved', bg: '#22c55e'},
  deprecated: {label: 'Deprecated', bg: '#ef4444'},
};

function parseSpec(data: any): Spec | null {
  if (!data) return null;
  const d = data?.data;
  if (d && typeof d === 'object' && 'id' in d) return d as Spec;
  return null;
}

export default function SpecDetailScreen({route}: Props) {
  const theme = useTheme();
  const {specId} = route.params;
  const {data: specData, isLoading} = useSpec(specId);

  const spec = parseSpec(specData);

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  if (!spec) {
    return (
      <View style={styles.center}>
        <Text style={{color: theme.colors.onSurfaceVariant}}>
          Spec not found
        </Text>
      </View>
    );
  }

  const cfg = statusConfig[spec.status] || statusConfig.draft;

  return (
    <ScrollView
      style={[styles.container, {backgroundColor: theme.colors.background}]}
      contentContainerStyle={styles.content}>
      <View style={[styles.header, {backgroundColor: theme.colors.surface, borderColor: theme.colors.outline}]}>
        <View style={styles.titleRow}>
          <Text
            variant="titleLarge"
            style={{color: theme.colors.onSurface, flex: 1}}>
            #{spec.id} {spec.title}
          </Text>
          <View style={[styles.statusBadge, {backgroundColor: cfg.bg}]}>
            <Text style={styles.statusText}>{cfg.label}</Text>
          </View>
        </View>
        <View style={styles.metaRow}>
          {spec.priority > 0 && (
            <Text
              variant="labelSmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              Priority: {spec.priority}
            </Text>
          )}
          <Text
            variant="labelSmall"
            style={{color: theme.colors.onSurfaceVariant}}>
            Updated: {new Date(spec.updated_at).toLocaleDateString()}
          </Text>
        </View>
      </View>

      <View style={[styles.contentCard, {backgroundColor: theme.colors.surface, borderColor: theme.colors.outline}]}>
        <MarkdownRenderer content={spec.content} />
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  content: {
    padding: 12,
    gap: 12,
  },
  header: {
    borderRadius: 10,
    borderWidth: 1,
    padding: 14,
    gap: 8,
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 10,
  },
  statusText: {
    fontSize: 11,
    fontWeight: '600',
    color: '#fff',
  },
  metaRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  contentCard: {
    borderRadius: 10,
    borderWidth: 1,
    padding: 14,
  },
});
