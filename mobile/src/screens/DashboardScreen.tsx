import React, {useState, useCallback, useMemo} from 'react';
import {
  View,
  ScrollView,
  StyleSheet,
  RefreshControl,
  TouchableOpacity,
} from 'react-native';
import {
  Text,
  useTheme,
  Card,
  ProgressBar,
  ActivityIndicator,
} from 'react-native-paper';
import {useNavigation} from '@react-navigation/native';
import type {BottomTabNavigationProp} from '@react-navigation/bottom-tabs';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {TabParamList} from '../navigation/TabNavigator';
import {
  useStatus,
  useProjects,
  useProjectStats,
  useMessages,
  useTaskStop,
} from '../hooks/useClaribot';
import {statusColors} from '../theme';
import type {Project, ProjectStats, Message} from '../types';
import ProjectSelector from '../components/ProjectSelector';
import CycleProgress from '../components/CycleProgress';
import EmptyState from '../components/EmptyState';

type Nav = BottomTabNavigationProp<TabParamList>;

function parseItems<T>(data: any): T[] {
  if (!data) return [];
  if (Array.isArray(data)) return data;
  if (data.items && Array.isArray(data.items)) return data.items;
  if (data.data) {
    if (Array.isArray(data.data)) return data.data;
    if (data.data.items && Array.isArray(data.data.items))
      return data.data.items;
  }
  return [];
}

const STAT_CARDS: {
  key: 'todo' | 'planned' | 'done' | 'failed';
  label: string;
  icon: string;
  color: string;
}[] = [
  {
    key: 'todo',
    label: 'Todo',
    icon: 'checkbox-blank-outline',
    color: statusColors.todo,
  },
  {
    key: 'planned',
    label: 'Planned',
    icon: 'clipboard-text-outline',
    color: statusColors.planned,
  },
  {
    key: 'done',
    label: 'Done',
    icon: 'check-circle-outline',
    color: statusColors.done,
  },
  {
    key: 'failed',
    label: 'Failed',
    icon: 'alert-circle-outline',
    color: statusColors.failed,
  },
];

export default function DashboardScreen() {
  const theme = useTheme();
  const navigation = useNavigation<Nav>();
  const [selectedProject, setSelectedProject] = useState<string | null>(null);

  const {data: statusData, refetch: refetchStatus} = useStatus();
  const {data: projectsData, refetch: refetchProjects} = useProjects(true);
  const {data: statsData, refetch: refetchStats} = useProjectStats();
  const {data: messagesData, refetch: refetchMessages} = useMessages(true);
  const stopMutation = useTaskStop();

  const [refreshing, setRefreshing] = useState(false);

  const projects = useMemo(() => parseItems<Project>(projectsData), [projectsData]);
  const allStats: ProjectStats[] = useMemo(
    () => parseItems(statsData),
    [statsData],
  );
  const messages: Message[] = useMemo(() => {
    const items = parseItems<Message>(messagesData);
    return items.slice(0, 5);
  }, [messagesData]);

  // Filter stats by selected project
  const filteredStats = useMemo(() => {
    if (!selectedProject) return allStats;
    return allStats.filter(s => s.project_id === selectedProject);
  }, [allStats, selectedProject]);

  // Aggregate task stats
  const taskStats = useMemo(() => {
    const agg = {todo: 0, planned: 0, done: 0, failed: 0, leaf: 0};
    filteredStats.forEach(ps => {
      agg.todo += ps.stats.todo || 0;
      agg.planned += ps.stats.planned || 0;
      agg.done += ps.stats.done || 0;
      agg.failed += ps.stats.failed || 0;
      agg.leaf += ps.stats.leaf || 0;
    });
    return agg;
  }, [filteredStats]);

  const cycleStatuses = useMemo(
    () => statusData?.cycle_statuses || [],
    [statusData],
  );

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await Promise.all([
      refetchStatus(),
      refetchProjects(),
      refetchStats(),
      refetchMessages(),
    ]);
    setRefreshing(false);
  }, [refetchStatus, refetchProjects, refetchStats, refetchMessages]);

  const handleStopCycle = () => {
    stopMutation.mutate();
  };

  const navigateToMessages = () => {
    navigation.navigate('Messages');
  };

  const navigateToTasks = () => {
    navigation.navigate('Tasks');
  };

  if (!statusData && !projectsData) {
    return (
      <View
        style={[styles.center, {backgroundColor: theme.colors.background}]}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  return (
    <ScrollView
      style={[styles.container, {backgroundColor: theme.colors.background}]}
      contentContainerStyle={styles.scrollContent}
      refreshControl={
        <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
      }>
      {/* 1. Project Selector */}
      <ProjectSelector
        projects={projects}
        selectedId={selectedProject}
        onSelect={setSelectedProject}
      />

      {/* 2. Cycle Status */}
      <CycleProgress
        cycleStatuses={cycleStatuses}
        onStop={handleStopCycle}
        isStopping={stopMutation.isPending}
      />

      {/* 3. Task Stats Cards */}
      <View style={styles.statsGrid}>
        {STAT_CARDS.map(({key, label, icon, color}) => (
          <TouchableOpacity
            key={key}
            style={styles.statCardWrapper}
            activeOpacity={0.7}
            onPress={navigateToTasks}>
            <Card
              style={[
                styles.statCard,
                {backgroundColor: theme.colors.surface},
              ]}
              mode="outlined">
              <Card.Content style={styles.statCardContent}>
                <MaterialCommunityIcons name={icon} size={24} color={color} />
                <Text
                  variant="headlineSmall"
                  style={[
                    styles.statNumber,
                    {color: theme.colors.onSurface},
                  ]}>
                  {taskStats[key]}
                </Text>
                <Text
                  variant="bodySmall"
                  style={{color: theme.colors.onSurfaceVariant}}>
                  {label}
                </Text>
              </Card.Content>
            </Card>
          </TouchableOpacity>
        ))}
      </View>

      {/* 4. Project Progress List */}
      {allStats.length > 0 && (
        <View style={styles.section}>
          <Text
            variant="titleMedium"
            style={[styles.sectionTitle, {color: theme.colors.onSurface}]}>
            프로젝트별 진행률
          </Text>
          {allStats.map(ps => {
            const leafTotal = ps.stats.leaf || 0;
            const done = ps.stats.done || 0;
            const progress = leafTotal > 0 ? done / leafTotal : 0;
            const percent = Math.round(progress * 100);

            return (
              <TouchableOpacity
                key={ps.project_id}
                activeOpacity={0.7}
                onPress={() => setSelectedProject(ps.project_id)}>
                <Card
                  style={[
                    styles.projectCard,
                    {backgroundColor: theme.colors.surface},
                  ]}
                  mode="outlined">
                  <Card.Content style={styles.projectCardContent}>
                    <View style={styles.projectHeader}>
                      <Text
                        variant="bodyMedium"
                        style={{
                          color: theme.colors.onSurface,
                          fontWeight: '600',
                          flex: 1,
                        }}
                        numberOfLines={1}>
                        {ps.project_id}
                      </Text>
                      <Text
                        variant="bodySmall"
                        style={{color: theme.colors.onSurfaceVariant}}>
                        {percent}%
                      </Text>
                    </View>
                    <ProgressBar
                      progress={progress}
                      color={statusColors.done}
                      style={styles.projectProgress}
                    />
                    <View style={styles.projectStats}>
                      <StatDot
                        color={statusColors.done}
                        count={done}
                        label="done"
                      />
                      <StatDot
                        color={statusColors.planned}
                        count={ps.stats.planned || 0}
                        label="planned"
                      />
                      <StatDot
                        color={statusColors.todo}
                        count={ps.stats.todo || 0}
                        label="todo"
                      />
                      <StatDot
                        color={statusColors.failed}
                        count={ps.stats.failed || 0}
                        label="failed"
                      />
                    </View>
                  </Card.Content>
                </Card>
              </TouchableOpacity>
            );
          })}
        </View>
      )}

      {/* 5. Recent Messages */}
      <View style={styles.section}>
        <View style={styles.sectionHeader}>
          <Text
            variant="titleMedium"
            style={[styles.sectionTitle, {color: theme.colors.onSurface}]}>
            최근 메시지
          </Text>
          <TouchableOpacity onPress={navigateToMessages}>
            <Text
              variant="bodySmall"
              style={{color: theme.colors.primary, fontWeight: '500'}}>
              더보기
            </Text>
          </TouchableOpacity>
        </View>
        {messages.length === 0 ? (
          <EmptyState
            icon="message-text-outline"
            message="최근 메시지가 없습니다"
          />
        ) : (
          messages.map(msg => (
            <TouchableOpacity
              key={msg.id}
              activeOpacity={0.7}
              onPress={navigateToMessages}>
              <Card
                style={[
                  styles.messageCard,
                  {backgroundColor: theme.colors.surface},
                ]}
                mode="outlined">
                <Card.Content style={styles.messageContent}>
                  <View style={styles.messageHeader}>
                    <MaterialCommunityIcons
                      name={
                        msg.source === 'telegram'
                          ? 'send'
                          : msg.source === 'schedule'
                            ? 'clock-outline'
                            : 'message-text-outline'
                      }
                      size={16}
                      color={theme.colors.onSurfaceVariant}
                    />
                    <Text
                      variant="bodySmall"
                      style={{color: theme.colors.onSurfaceVariant}}>
                      {msg.project_id || 'Global'}
                    </Text>
                    <Text
                      variant="bodySmall"
                      style={{
                        color: theme.colors.onSurfaceVariant,
                        marginLeft: 'auto',
                      }}>
                      {formatTime(msg.created_at)}
                    </Text>
                  </View>
                  <Text
                    variant="bodyMedium"
                    numberOfLines={2}
                    style={{color: theme.colors.onSurface}}>
                    {msg.content}
                  </Text>
                </Card.Content>
              </Card>
            </TouchableOpacity>
          ))
        )}
      </View>
    </ScrollView>
  );
}

function StatDot({
  color,
  count,
  label,
}: {
  color: string;
  count: number;
  label: string;
}) {
  const theme = useTheme();
  return (
    <View style={statDotStyles.row}>
      <View style={[statDotStyles.dot, {backgroundColor: color}]} />
      <Text
        variant="labelSmall"
        style={{color: theme.colors.onSurfaceVariant}}>
        {count} {label}
      </Text>
    </View>
  );
}

const statDotStyles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  dot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
});

function formatTime(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffMin = Math.floor(diffMs / 60000);

  if (diffMin < 1) return '방금';
  if (diffMin < 60) return `${diffMin}분 전`;

  const diffHour = Math.floor(diffMin / 60);
  if (diffHour < 24) return `${diffHour}시간 전`;

  return d.toLocaleDateString('ko-KR', {month: 'short', day: 'numeric'});
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  scrollContent: {
    padding: 16,
    gap: 16,
    paddingBottom: 32,
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  statsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  statCardWrapper: {
    width: '48%',
    flexGrow: 1,
  },
  statCard: {
    borderRadius: 12,
  },
  statCardContent: {
    alignItems: 'center',
    gap: 4,
    paddingVertical: 12,
  },
  statNumber: {
    fontWeight: '700',
  },
  section: {
    gap: 8,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  sectionTitle: {
    fontWeight: '600',
  },
  projectCard: {
    borderRadius: 10,
  },
  projectCardContent: {
    gap: 6,
  },
  projectHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 8,
  },
  projectProgress: {
    height: 6,
    borderRadius: 3,
  },
  projectStats: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
  },
  messageCard: {
    borderRadius: 10,
  },
  messageContent: {
    gap: 4,
  },
  messageHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
});
