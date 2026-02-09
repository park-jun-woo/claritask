import React, {useMemo} from 'react';
import {View, TouchableOpacity, StyleSheet} from 'react-native';
import {Text, useTheme} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {TaskStats, CycleStatus} from '../types';
import {statusColors} from '../theme';

interface StatusBarProps {
  taskStats: TaskStats;
  cycleStatus?: CycleStatus;
  statusFilter: string | null;
  onFilterChange: (status: string | null) => void;
}

const STATUSES = ['todo', 'split', 'planned', 'done', 'failed'] as const;

function formatElapsed(sec: number): string {
  if (sec < 60) return `${sec}s`;
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  if (m < 60) return `${m}m ${s}s`;
  const h = Math.floor(m / 60);
  return `${h}h ${m % 60}m`;
}

export default function StatusBar({
  taskStats,
  cycleStatus,
  statusFilter,
  onFilterChange,
}: StatusBarProps) {
  const theme = useTheme();

  const total = taskStats.total || 1;

  const barSegments = useMemo(() => {
    return STATUSES.map(s => ({
      status: s,
      count: taskStats[s] || 0,
      ratio: ((taskStats[s] || 0) / total) * 100,
      color: statusColors[s],
    })).filter(seg => seg.count > 0);
  }, [taskStats, total]);

  const leafDone = taskStats.done || 0;
  const leafTotal = taskStats.leaf || 0;
  const progress = leafTotal > 0 ? Math.round((leafDone / leafTotal) * 100) : 0;

  const showCycle = cycleStatus && cycleStatus.status !== 'idle';
  const isRunning = cycleStatus?.status === 'running';

  return (
    <View style={styles.container}>
      {/* Cycle status */}
      {showCycle && (
        <View
          style={[
            styles.cycleRow,
            {
              backgroundColor: isRunning
                ? statusColors.done + '15'
                : '#eab30815',
              borderColor: isRunning
                ? statusColors.done + '40'
                : '#eab30840',
            },
          ]}>
          <MaterialCommunityIcons
            name={isRunning ? 'loading' : 'alert-circle-outline'}
            size={14}
            color={isRunning ? statusColors.done : '#eab308'}
          />
          <Text
            variant="labelSmall"
            style={{fontWeight: '600', color: theme.colors.onSurface}}>
            {isRunning ? 'Running' : 'Interrupted'}
          </Text>
          {cycleStatus?.phase && (
            <View
              style={[
                styles.phaseBadge,
                {borderColor: theme.colors.outline},
              ]}>
              <Text
                variant="labelSmall"
                style={{fontSize: 10, color: theme.colors.onSurfaceVariant}}>
                {cycleStatus.phase}
              </Text>
            </View>
          )}
          {cycleStatus?.current_task_id ? (
            <Text
              variant="labelSmall"
              style={{color: theme.colors.onSurface}}>
              Task #{cycleStatus.current_task_id}
            </Text>
          ) : null}
          {cycleStatus?.target_total != null &&
            cycleStatus.target_total > 0 && (
              <Text
                variant="labelSmall"
                style={{color: theme.colors.onSurfaceVariant}}>
                {cycleStatus.completed ?? 0}/{cycleStatus.target_total}
              </Text>
            )}
          {cycleStatus?.elapsed_sec != null && (
            <Text
              variant="labelSmall"
              style={{
                color: theme.colors.onSurfaceVariant,
                marginLeft: 'auto',
              }}>
              {formatElapsed(cycleStatus.elapsed_sec)}
            </Text>
          )}
        </View>
      )}

      {/* Progress bar */}
      <View
        style={[
          styles.progressBar,
          {backgroundColor: theme.colors.surfaceVariant},
        ]}>
        {barSegments.map(seg => (
          <TouchableOpacity
            key={seg.status}
            activeOpacity={0.7}
            onPress={() =>
              onFilterChange(statusFilter === seg.status ? null : seg.status)
            }
            style={[
              styles.barSegment,
              {
                backgroundColor: seg.color,
                width: `${Math.max(seg.ratio, 2)}%` as any,
                opacity: statusFilter && statusFilter !== seg.status ? 0.3 : 1,
              },
            ]}
          />
        ))}
      </View>

      {/* Status chips */}
      <View style={styles.chipRow}>
        {STATUSES.map(s => {
          const count = taskStats[s] || 0;
          const isActive = statusFilter === s;
          return (
            <TouchableOpacity
              key={s}
              activeOpacity={0.7}
              onPress={() => onFilterChange(isActive ? null : s)}
              style={[
                styles.chip,
                {
                  backgroundColor: isActive
                    ? statusColors[s] + '25'
                    : 'transparent',
                  borderColor: isActive
                    ? statusColors[s]
                    : theme.colors.outline,
                },
              ]}>
              <View
                style={[styles.chipDot, {backgroundColor: statusColors[s]}]}
              />
              <Text
                variant="labelSmall"
                style={{
                  color: isActive
                    ? statusColors[s]
                    : theme.colors.onSurfaceVariant,
                  fontSize: 11,
                }}>
                {s}
              </Text>
              <Text
                variant="labelSmall"
                style={{
                  color: theme.colors.onSurface,
                  fontWeight: '600',
                  fontSize: 11,
                }}>
                {count}
              </Text>
            </TouchableOpacity>
          );
        })}
      </View>

      {/* Leaf progress */}
      <View style={styles.leafRow}>
        <Text
          variant="labelSmall"
          style={{color: theme.colors.onSurfaceVariant}}>
          done/leaf
        </Text>
        <Text
          variant="labelSmall"
          style={{color: theme.colors.onSurface, fontWeight: '600'}}>
          {leafDone}/{leafTotal}
        </Text>
        <Text
          variant="labelSmall"
          style={{color: theme.colors.onSurfaceVariant}}>
          ({progress}%)
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    gap: 6,
    paddingVertical: 4,
  },
  cycleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    paddingHorizontal: 8,
    paddingVertical: 6,
    borderRadius: 6,
    borderWidth: 1,
    flexWrap: 'wrap',
  },
  phaseBadge: {
    borderWidth: 1,
    borderRadius: 4,
    paddingHorizontal: 4,
    paddingVertical: 1,
  },
  progressBar: {
    height: 6,
    borderRadius: 3,
    flexDirection: 'row',
    overflow: 'hidden',
  },
  barSegment: {
    height: '100%',
  },
  chipRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 4,
  },
  chip: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 3,
    paddingHorizontal: 6,
    paddingVertical: 3,
    borderRadius: 4,
    borderWidth: 1,
  },
  chipDot: {
    width: 7,
    height: 7,
    borderRadius: 4,
  },
  leafRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    justifyContent: 'flex-end',
  },
});
