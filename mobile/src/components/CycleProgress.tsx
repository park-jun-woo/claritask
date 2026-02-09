import React from 'react';
import {View, StyleSheet} from 'react-native';
import {
  Text,
  useTheme,
  Button,
  ProgressBar,
  Card,
  Chip,
} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {CycleStatus} from '../types';

interface CycleProgressProps {
  cycleStatuses: CycleStatus[];
  onStop: () => void;
  isStopping: boolean;
}

function formatElapsed(sec?: number): string {
  if (!sec || sec <= 0) return '0s';
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  if (m === 0) return `${s}s`;
  return `${m}m ${s}s`;
}

function CycleCard({
  cycle,
  onStop,
  isStopping,
}: {
  cycle: CycleStatus;
  onStop: () => void;
  isStopping: boolean;
}) {
  const theme = useTheme();
  const isRunning = cycle.status === 'running';
  const total = cycle.target_total || 0;
  const completed = cycle.completed || 0;
  const progress = total > 0 ? completed / total : 0;
  const phaseLabel = cycle.phase === 'planning' ? 'Planning' : 'Running';

  return (
    <Card
      style={[styles.card, {backgroundColor: theme.colors.surface}]}
      mode="outlined">
      <Card.Content style={styles.cardContent}>
        <View style={styles.cardHeader}>
          <View style={styles.cardHeaderLeft}>
            <MaterialCommunityIcons
              name={isRunning ? 'sync' : 'pause-circle-outline'}
              size={20}
              color={isRunning ? '#3b82f6' : '#f59e0b'}
            />
            <Text
              variant="titleSmall"
              style={{color: theme.colors.onSurface}}>
              {cycle.project_id || 'Global'}
            </Text>
            <Chip compact style={styles.phaseChip} textStyle={styles.phaseText}>
              {phaseLabel}
            </Chip>
          </View>
          {isRunning && (
            <Button
              mode="outlined"
              compact
              icon="stop"
              onPress={onStop}
              loading={isStopping}
              disabled={isStopping}
              style={styles.stopButton}
              labelStyle={styles.stopLabel}>
              Stop
            </Button>
          )}
        </View>

        <ProgressBar
          progress={progress}
          color="#3b82f6"
          style={styles.progressBar}
        />

        <View style={styles.statsRow}>
          <Text
            variant="bodySmall"
            style={{color: theme.colors.onSurfaceVariant}}>
            {completed}/{total} tasks
          </Text>
          {cycle.active_workers != null && cycle.active_workers > 0 && (
            <Text
              variant="bodySmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              Workers: {cycle.active_workers}
            </Text>
          )}
          <Text
            variant="bodySmall"
            style={{color: theme.colors.onSurfaceVariant}}>
            {formatElapsed(cycle.elapsed_sec)}
          </Text>
        </View>
      </Card.Content>
    </Card>
  );
}

export default function CycleProgress({
  cycleStatuses,
  onStop,
  isStopping,
}: CycleProgressProps) {
  const activeCycles = cycleStatuses.filter(c => c.status !== 'idle');

  if (activeCycles.length === 0) return null;

  return (
    <View style={styles.container}>
      {activeCycles.map((cycle, idx) => (
        <CycleCard
          key={cycle.project_id || idx}
          cycle={cycle}
          onStop={onStop}
          isStopping={isStopping}
        />
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    gap: 8,
  },
  card: {
    borderRadius: 12,
  },
  cardContent: {
    gap: 8,
  },
  cardHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  cardHeaderLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    flex: 1,
  },
  phaseChip: {
    height: 24,
  },
  phaseText: {
    fontSize: 11,
  },
  progressBar: {
    height: 6,
    borderRadius: 3,
  },
  statsRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    flexWrap: 'wrap',
    gap: 8,
  },
  stopButton: {
    borderColor: '#ef4444',
    minWidth: 0,
  },
  stopLabel: {
    color: '#ef4444',
    fontSize: 12,
  },
});
