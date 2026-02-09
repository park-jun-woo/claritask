import React, {useState, useMemo, useCallback} from 'react';
import {
  View,
  FlatList,
  StyleSheet,
  RefreshControl,
  TouchableOpacity,
} from 'react-native';
import {
  Text,
  Searchbar,
  useTheme,
  ActivityIndicator,
  Chip,
} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {StackScreenProps} from '@react-navigation/stack';
import type {MoreStackParamList} from '../../navigation/MoreStackNavigator';
import type {Project, ProjectStats} from '../../types';
import {
  useProjects,
  useProjectStats,
  useStatus,
  useSwitchProject,
} from '../../hooks/useClaribot';
import {setSelectedProject} from '../../storage/cache';
import {statusColors} from '../../theme';
import EmptyState from '../../components/EmptyState';

type Props = StackScreenProps<MoreStackParamList, 'Projects'>;

function TaskStatsBadge({stats}: {stats: ProjectStats['stats']}) {
  const theme = useTheme();
  const items = [
    {key: 'todo', count: stats.todo, color: statusColors.todo},
    {key: 'planned', count: stats.planned, color: statusColors.planned},
    {key: 'split', count: stats.split, color: statusColors.split},
    {key: 'done', count: stats.done, color: statusColors.done},
    {key: 'failed', count: stats.failed, color: statusColors.failed},
  ].filter(i => i.count > 0);

  if (items.length === 0) {
    return (
      <Text
        variant="bodySmall"
        style={{color: theme.colors.onSurfaceVariant}}>
        No tasks
      </Text>
    );
  }

  return (
    <View style={styles.statsRow}>
      {items.map(item => (
        <View key={item.key} style={styles.statItem}>
          <View
            style={[styles.statDot, {backgroundColor: item.color}]}
          />
          <Text
            variant="labelSmall"
            style={{color: theme.colors.onSurfaceVariant}}>
            {item.count}
          </Text>
        </View>
      ))}
      <Text
        variant="labelSmall"
        style={{color: theme.colors.onSurfaceVariant, marginLeft: 4}}>
        ({stats.total})
      </Text>
    </View>
  );
}

export default function ProjectsScreen({navigation}: Props) {
  const theme = useTheme();
  const {data: projectsData, isLoading, refetch} = useProjects(true);
  const {data: statsData} = useProjectStats();
  const {data: statusData} = useStatus();
  const switchProject = useSwitchProject();

  const [search, setSearch] = useState('');
  const [refreshing, setRefreshing] = useState(false);

  const currentProjectId = statusData?.project_id;

  const projects: Project[] = useMemo(() => {
    const raw = projectsData?.data;
    if (!Array.isArray(raw)) return [];
    return raw as Project[];
  }, [projectsData]);

  const statsMap = useMemo(() => {
    const map: Record<string, ProjectStats['stats']> = {};
    const raw = statsData?.data;
    if (Array.isArray(raw)) {
      (raw as ProjectStats[]).forEach(ps => {
        map[ps.project_id] = ps.stats;
      });
    }
    return map;
  }, [statsData]);

  const filtered = useMemo(() => {
    if (!search.trim()) return projects;
    const q = search.toLowerCase();
    return projects.filter(
      p =>
        p.id.toLowerCase().includes(q) ||
        (p.description || '').toLowerCase().includes(q) ||
        (p.path || '').toLowerCase().includes(q),
    );
  }, [projects, search]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  }, [refetch]);

  const handleSelect = useCallback(
    (proj: Project) => {
      setSelectedProject(proj.id);
      switchProject.mutate(proj.id, {
        onSuccess: () => {
          // Navigate to Dashboard tab (parent is TabNavigator)
          const tabNav = navigation.getParent();
          if (tabNav) {
            tabNav.navigate('Dashboard');
          }
        },
      });
    },
    [switchProject, navigation],
  );

  const handleEdit = useCallback(
    (project: Project) => {
      navigation.navigate('ProjectEdit', {projectId: project.id});
    },
    [navigation],
  );

  const renderProject = useCallback(
    ({item}: {item: Project}) => {
      const isActive = item.id === currentProjectId;
      const stats = statsMap[item.id];

      return (
        <TouchableOpacity
          activeOpacity={0.7}
          onPress={() => handleSelect(item)}
          style={[
            styles.card,
            {
              backgroundColor: theme.colors.surface,
              borderColor: isActive
                ? theme.colors.primary
                : theme.colors.outline,
              borderWidth: isActive ? 2 : 1,
            },
          ]}>
          <View style={styles.cardHeader}>
            <View style={styles.cardTitleRow}>
              {isActive && (
                <MaterialCommunityIcons
                  name="check-circle"
                  size={18}
                  color={statusColors.done}
                />
              )}
              <Text
                variant="titleMedium"
                style={{color: theme.colors.onSurface, flex: 1}}
                numberOfLines={1}>
                {item.id}
              </Text>
              <TouchableOpacity
                hitSlop={{top: 12, bottom: 12, left: 12, right: 12}}
                onPress={() => handleEdit(item)}>
                <MaterialCommunityIcons
                  name="pencil-outline"
                  size={20}
                  color={theme.colors.onSurfaceVariant}
                />
              </TouchableOpacity>
            </View>
            {item.description ? (
              <Text
                variant="bodySmall"
                style={{color: theme.colors.onSurfaceVariant}}
                numberOfLines={2}>
                {item.description}
              </Text>
            ) : null}
          </View>

          <View style={[styles.cardFooter, {borderTopColor: theme.colors.outline}]}>
            <View style={styles.pathRow}>
              <MaterialCommunityIcons
                name="folder-outline"
                size={14}
                color={theme.colors.onSurfaceVariant}
              />
              <Text
                variant="labelSmall"
                style={{color: theme.colors.onSurfaceVariant, flex: 1}}
                numberOfLines={1}>
                {item.path}
              </Text>
            </View>
            {stats ? (
              <TaskStatsBadge stats={stats} />
            ) : (
              <Text
                variant="bodySmall"
                style={{color: theme.colors.onSurfaceVariant}}>
                -
              </Text>
            )}
          </View>
        </TouchableOpacity>
      );
    },
    [currentProjectId, statsMap, theme, handleSelect, handleEdit],
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
      <View style={[styles.searchContainer, {backgroundColor: theme.colors.surface}]}>
        <Searchbar
          placeholder="Search projects..."
          value={search}
          onChangeText={setSearch}
          style={{backgroundColor: theme.colors.surfaceVariant}}
          inputStyle={{minHeight: 0}}
        />
        {currentProjectId ? (
          <Chip
            compact
            icon="check-circle"
            style={{alignSelf: 'flex-start'}}
            onPress={() => {
              switchProject.mutate('none');
            }}>
            {currentProjectId} (tap to deselect)
          </Chip>
        ) : null}
      </View>

      <FlatList
        data={filtered}
        keyExtractor={item => item.id}
        renderItem={renderProject}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListEmptyComponent={
          <EmptyState
            icon="folder-open-outline"
            message={
              search ? 'No matching projects' : 'No projects found'
            }
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
  searchContainer: {
    padding: 12,
    gap: 8,
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
  cardHeader: {
    padding: 14,
    gap: 4,
  },
  cardTitleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  cardFooter: {
    paddingHorizontal: 14,
    paddingVertical: 10,
    borderTopWidth: 1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: 8,
  },
  pathRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    flex: 1,
  },
  statsRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  statItem: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 2,
  },
  statDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
});
