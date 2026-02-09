import React, {useMemo, useCallback, useState} from 'react';
import {
  View,
  ScrollView,
  FlatList,
  StyleSheet,
  RefreshControl,
} from 'react-native';
import {Text, useTheme, ActivityIndicator, Switch} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {StackScreenProps} from '@react-navigation/stack';
import type {MoreStackParamList} from '../../navigation/MoreStackNavigator';
import type {Schedule, ScheduleRun} from '../../types';
import {
  useSchedules,
  useScheduleRuns,
  useToggleSchedule,
} from '../../hooks/useClaribot';
import EmptyState from '../../components/EmptyState';

type Props = StackScreenProps<MoreStackParamList, 'ScheduleDetail'>;

function findSchedule(data: any, id: number): Schedule | null {
  if (!data) return null;
  const d = data?.data;
  const items = Array.isArray(d) ? d : d?.items || [];
  return items.find((s: Schedule) => s.id === id) || null;
}

function parseRuns(data: any): ScheduleRun[] {
  if (!data) return [];
  const d = data?.data;
  if (Array.isArray(d)) return d;
  if (d?.items && Array.isArray(d.items)) return d.items;
  return [];
}

const runStatusConfig: Record<string, {icon: string; color: string}> = {
  running: {icon: 'loading', color: '#3b82f6'},
  done: {icon: 'check-circle', color: '#22c55e'},
  failed: {icon: 'close-circle', color: '#ef4444'},
};

export default function ScheduleDetailScreen({route}: Props) {
  const theme = useTheme();
  const {scheduleId} = route.params;
  const {data: schedulesData} = useSchedules(true);
  const {data: runsData, isLoading: runsLoading, refetch: refetchRuns} = useScheduleRuns(scheduleId);
  const toggleSchedule = useToggleSchedule();
  const [refreshing, setRefreshing] = useState(false);

  const schedule = useMemo(
    () => findSchedule(schedulesData, scheduleId),
    [schedulesData, scheduleId],
  );

  const runs = useMemo(() => parseRuns(runsData), [runsData]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await refetchRuns();
    setRefreshing(false);
  }, [refetchRuns]);

  const handleToggle = useCallback(() => {
    if (!schedule) return;
    toggleSchedule.mutate({id: schedule.id, enable: !schedule.enabled});
  }, [schedule, toggleSchedule]);

  if (!schedule) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  const typeIcon = schedule.type === 'bash' ? 'console' : 'robot-outline';

  const renderRun = ({item}: {item: ScheduleRun}) => {
    const cfg = runStatusConfig[item.status] || runStatusConfig.failed;
    const startDate = new Date(item.started_at);
    const endDate = item.completed_at ? new Date(item.completed_at) : null;

    return (
      <View
        style={[
          styles.runCard,
          {
            backgroundColor: theme.colors.surface,
            borderColor: theme.colors.outline,
          },
        ]}>
        <View style={styles.runHeader}>
          <MaterialCommunityIcons
            name={cfg.icon}
            size={18}
            color={cfg.color}
          />
          <Text
            variant="labelMedium"
            style={{color: cfg.color, textTransform: 'capitalize'}}>
            {item.status}
          </Text>
          <Text
            variant="labelSmall"
            style={{color: theme.colors.onSurfaceVariant, marginLeft: 'auto'}}>
            #{item.id}
          </Text>
        </View>
        <View style={styles.runTimes}>
          <View style={styles.timeRow}>
            <MaterialCommunityIcons
              name="play"
              size={12}
              color={theme.colors.onSurfaceVariant}
            />
            <Text
              variant="labelSmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              {startDate.toLocaleString()}
            </Text>
          </View>
          {endDate && (
            <View style={styles.timeRow}>
              <MaterialCommunityIcons
                name="stop"
                size={12}
                color={theme.colors.onSurfaceVariant}
              />
              <Text
                variant="labelSmall"
                style={{color: theme.colors.onSurfaceVariant}}>
                {endDate.toLocaleString()}
              </Text>
            </View>
          )}
        </View>
        {item.error ? (
          <Text
            variant="bodySmall"
            style={{color: '#ef4444'}}
            numberOfLines={3}>
            {item.error}
          </Text>
        ) : null}
      </View>
    );
  };

  return (
    <View style={[styles.container, {backgroundColor: theme.colors.background}]}>
      <FlatList
        data={runs}
        keyExtractor={item => String(item.id)}
        renderItem={renderRun}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListHeaderComponent={
          <View style={styles.headerSection}>
            <View
              style={[
                styles.infoCard,
                {
                  backgroundColor: theme.colors.surface,
                  borderColor: theme.colors.outline,
                },
              ]}>
              <View style={styles.infoRow}>
                <MaterialCommunityIcons
                  name={typeIcon}
                  size={20}
                  color={theme.colors.primary}
                />
                <Text
                  variant="titleMedium"
                  style={{color: theme.colors.onSurface, flex: 1}}>
                  Schedule #{schedule.id}
                </Text>
                <Switch value={schedule.enabled} onValueChange={handleToggle} />
              </View>

              <View
                style={[
                  styles.cronBox,
                  {backgroundColor: theme.colors.elevation.level2},
                ]}>
                <MaterialCommunityIcons
                  name="clock-outline"
                  size={14}
                  color={theme.colors.onSurfaceVariant}
                />
                <Text
                  variant="bodySmall"
                  style={{
                    color: theme.colors.onSurfaceVariant,
                    fontFamily: 'monospace',
                  }}>
                  {schedule.cron_expr}
                </Text>
              </View>

              <Text
                variant="bodyMedium"
                style={{color: theme.colors.onSurface}}>
                {schedule.message}
              </Text>

              <View style={styles.metaGrid}>
                <View style={styles.metaItem}>
                  <Text
                    variant="labelSmall"
                    style={{color: theme.colors.onSurfaceVariant}}>
                    Type
                  </Text>
                  <Text
                    variant="bodySmall"
                    style={{color: theme.colors.onSurface}}>
                    {schedule.type}
                  </Text>
                </View>
                {schedule.project_id && (
                  <View style={styles.metaItem}>
                    <Text
                      variant="labelSmall"
                      style={{color: theme.colors.onSurfaceVariant}}>
                      Project
                    </Text>
                    <Text
                      variant="bodySmall"
                      style={{color: theme.colors.onSurface}}>
                      {schedule.project_id}
                    </Text>
                  </View>
                )}
                {schedule.next_run && (
                  <View style={styles.metaItem}>
                    <Text
                      variant="labelSmall"
                      style={{color: theme.colors.onSurfaceVariant}}>
                      Next Run
                    </Text>
                    <Text
                      variant="bodySmall"
                      style={{color: theme.colors.onSurface}}>
                      {new Date(schedule.next_run).toLocaleString()}
                    </Text>
                  </View>
                )}
                {schedule.last_run && (
                  <View style={styles.metaItem}>
                    <Text
                      variant="labelSmall"
                      style={{color: theme.colors.onSurfaceVariant}}>
                      Last Run
                    </Text>
                    <Text
                      variant="bodySmall"
                      style={{color: theme.colors.onSurface}}>
                      {new Date(schedule.last_run).toLocaleString()}
                    </Text>
                  </View>
                )}
              </View>
            </View>

            <Text
              variant="titleSmall"
              style={{
                color: theme.colors.onSurface,
                marginTop: 8,
                marginBottom: 4,
              }}>
              Run History
            </Text>
          </View>
        }
        ListEmptyComponent={
          runsLoading ? (
            <ActivityIndicator style={{marginTop: 24}} />
          ) : (
            <EmptyState icon="history" message="No run history" />
          )
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
  listContent: {
    padding: 12,
  },
  headerSection: {
    gap: 8,
    marginBottom: 8,
  },
  infoCard: {
    borderRadius: 10,
    borderWidth: 1,
    padding: 14,
    gap: 10,
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  cronBox: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    paddingHorizontal: 10,
    paddingVertical: 6,
    borderRadius: 6,
  },
  metaGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    marginTop: 4,
  },
  metaItem: {
    gap: 2,
  },
  runCard: {
    borderRadius: 8,
    borderWidth: 1,
    padding: 12,
    marginBottom: 8,
    gap: 6,
  },
  runHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  runTimes: {
    gap: 2,
  },
  timeRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
});
