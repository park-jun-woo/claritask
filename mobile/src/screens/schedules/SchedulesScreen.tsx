import React, {useState, useMemo, useCallback} from 'react';
import {
  View,
  FlatList,
  StyleSheet,
  RefreshControl,
  TouchableOpacity,
} from 'react-native';
import {Text, useTheme, ActivityIndicator, Switch} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {StackScreenProps} from '@react-navigation/stack';
import type {MoreStackParamList} from '../../navigation/MoreStackNavigator';
import type {Schedule} from '../../types';
import {useSchedules, useToggleSchedule} from '../../hooks/useClaribot';
import EmptyState from '../../components/EmptyState';

type Props = StackScreenProps<MoreStackParamList, 'Schedules'>;

function parseItems(data: any): Schedule[] {
  if (!data) return [];
  const d = data?.data;
  if (Array.isArray(d)) return d;
  if (d?.items && Array.isArray(d.items)) return d.items;
  if (Array.isArray(data)) return data;
  return [];
}

export default function SchedulesScreen({navigation}: Props) {
  const theme = useTheme();
  const {data: schedulesData, isLoading, refetch} = useSchedules(true);
  const toggleSchedule = useToggleSchedule();
  const [refreshing, setRefreshing] = useState(false);

  const schedules = useMemo(() => parseItems(schedulesData), [schedulesData]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  }, [refetch]);

  const handleToggle = useCallback(
    (schedule: Schedule) => {
      toggleSchedule.mutate({
        id: schedule.id,
        enable: !schedule.enabled,
      });
    },
    [toggleSchedule],
  );

  const renderSchedule = useCallback(
    ({item}: {item: Schedule}) => {
      const typeIcon = item.type === 'bash' ? 'console' : 'robot-outline';
      const typeLabel = item.type === 'bash' ? 'Bash' : 'Claude';

      return (
        <TouchableOpacity
          activeOpacity={0.7}
          onPress={() =>
            navigation.navigate('ScheduleDetail', {scheduleId: item.id})
          }
          style={[
            styles.card,
            {
              backgroundColor: theme.colors.surface,
              borderColor: item.enabled
                ? theme.colors.primary
                : theme.colors.outline,
              borderWidth: item.enabled ? 1.5 : 1,
              opacity: item.enabled ? 1 : 0.7,
            },
          ]}>
          <View style={styles.cardBody}>
            <View style={styles.cardTopRow}>
              <View style={styles.typeChip}>
                <MaterialCommunityIcons
                  name={typeIcon}
                  size={14}
                  color={theme.colors.onSurfaceVariant}
                />
                <Text
                  variant="labelSmall"
                  style={{color: theme.colors.onSurfaceVariant}}>
                  {typeLabel}
                </Text>
              </View>
              <View
                style={[
                  styles.cronBadge,
                  {backgroundColor: theme.colors.elevation.level2},
                ]}>
                <MaterialCommunityIcons
                  name="clock-outline"
                  size={12}
                  color={theme.colors.onSurfaceVariant}
                />
                <Text
                  variant="labelSmall"
                  style={{
                    color: theme.colors.onSurfaceVariant,
                    fontFamily: 'monospace',
                  }}
                  numberOfLines={1}>
                  {item.cron_expr}
                </Text>
              </View>
              <Switch
                value={item.enabled}
                onValueChange={() => handleToggle(item)}
                style={styles.toggle}
              />
            </View>
            <Text
              variant="bodyMedium"
              style={{color: theme.colors.onSurface}}
              numberOfLines={2}>
              {item.message}
            </Text>
            <View style={styles.cardFooter}>
              {item.project_id && (
                <View style={styles.projectChip}>
                  <MaterialCommunityIcons
                    name="folder-outline"
                    size={12}
                    color={theme.colors.onSurfaceVariant}
                  />
                  <Text
                    variant="labelSmall"
                    style={{color: theme.colors.onSurfaceVariant}}>
                    {item.project_id}
                  </Text>
                </View>
              )}
              {item.next_run && (
                <Text
                  variant="labelSmall"
                  style={{color: theme.colors.onSurfaceVariant}}>
                  Next: {new Date(item.next_run).toLocaleString()}
                </Text>
              )}
            </View>
          </View>
        </TouchableOpacity>
      );
    },
    [theme, navigation, handleToggle],
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
        data={schedules}
        keyExtractor={item => String(item.id)}
        renderItem={renderSchedule}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListEmptyComponent={
          <EmptyState
            icon="clock-outline"
            message="No schedules found"
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
    overflow: 'hidden',
    marginBottom: 10,
  },
  cardBody: {
    padding: 14,
    gap: 8,
  },
  cardTopRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  typeChip: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  cronBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 6,
    flex: 1,
  },
  toggle: {
    marginLeft: 'auto',
  },
  cardFooter: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 8,
  },
  projectChip: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
});
