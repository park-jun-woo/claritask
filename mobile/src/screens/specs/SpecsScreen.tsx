import React, {useState, useMemo, useCallback} from 'react';
import {
  View,
  FlatList,
  StyleSheet,
  RefreshControl,
  TouchableOpacity,
} from 'react-native';
import {Text, useTheme, ActivityIndicator, Badge} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {StackScreenProps} from '@react-navigation/stack';
import type {MoreStackParamList} from '../../navigation/MoreStackNavigator';
import type {Spec} from '../../types';
import {useSpecs} from '../../hooks/useClaribot';
import EmptyState from '../../components/EmptyState';

type Props = StackScreenProps<MoreStackParamList, 'Specs'>;

const statusConfig: Record<
  Spec['status'],
  {label: string; color: string; bg: string}
> = {
  draft: {label: 'Draft', color: '#fff', bg: '#6b7280'},
  review: {label: 'Review', color: '#fff', bg: '#3b82f6'},
  approved: {label: 'Approved', color: '#fff', bg: '#22c55e'},
  deprecated: {label: 'Deprecated', color: '#fff', bg: '#ef4444'},
};

function parseItems(data: any): Spec[] {
  if (!data) return [];
  const d = data?.data;
  if (Array.isArray(d)) return d;
  if (d?.items && Array.isArray(d.items)) return d.items;
  if (Array.isArray(data)) return data;
  return [];
}

export default function SpecsScreen({navigation}: Props) {
  const theme = useTheme();
  const {data: specsData, isLoading, refetch} = useSpecs(true);
  const [refreshing, setRefreshing] = useState(false);

  const specs = useMemo(() => parseItems(specsData), [specsData]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  }, [refetch]);

  const renderSpec = useCallback(
    ({item}: {item: Spec}) => {
      const cfg = statusConfig[item.status] || statusConfig.draft;

      return (
        <TouchableOpacity
          activeOpacity={0.7}
          onPress={() => navigation.navigate('SpecDetail', {specId: item.id})}
          style={[
            styles.card,
            {
              backgroundColor: theme.colors.surface,
              borderColor: theme.colors.outline,
            },
          ]}>
          <View style={styles.cardBody}>
            <View style={styles.cardTitleRow}>
              <Text
                variant="titleMedium"
                style={{color: theme.colors.onSurface, flex: 1}}
                numberOfLines={1}>
                #{item.id} {item.title}
              </Text>
              <View style={[styles.statusBadge, {backgroundColor: cfg.bg}]}>
                <Text style={[styles.statusText, {color: cfg.color}]}>
                  {cfg.label}
                </Text>
              </View>
            </View>
            {item.priority > 0 && (
              <View style={styles.metaRow}>
                <MaterialCommunityIcons
                  name="flag-outline"
                  size={14}
                  color={theme.colors.onSurfaceVariant}
                />
                <Text
                  variant="labelSmall"
                  style={{color: theme.colors.onSurfaceVariant}}>
                  Priority: {item.priority}
                </Text>
              </View>
            )}
          </View>
          <View style={styles.chevron}>
            <MaterialCommunityIcons
              name="chevron-right"
              size={20}
              color={theme.colors.onSurfaceVariant}
            />
          </View>
        </TouchableOpacity>
      );
    },
    [theme, navigation],
  );

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  return (
    <View style={[styles.container, {backgroundColor: theme.colors.background}]}>
      <FlatList
        data={specs}
        keyExtractor={item => String(item.id)}
        renderItem={renderSpec}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListEmptyComponent={
          <EmptyState
            icon="file-document-outline"
            message="No specs found"
          />
        }
      />
    </View>
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
  list: {
    padding: 12,
    gap: 10,
  },
  card: {
    borderRadius: 10,
    borderWidth: 1,
    overflow: 'hidden',
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 10,
  },
  cardBody: {
    flex: 1,
    padding: 14,
    gap: 6,
  },
  cardTitleRow: {
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
  },
  metaRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  chevron: {
    paddingRight: 12,
  },
});
