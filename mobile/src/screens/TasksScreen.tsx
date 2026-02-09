import React, {useState, useMemo, useCallback} from 'react';
import {
  View,
  FlatList,
  StyleSheet,
  RefreshControl,
  TouchableOpacity,
  Alert,
  TextInput,
} from 'react-native';
import {
  Text,
  useTheme,
  ActivityIndicator,
  FAB,
  IconButton,
  Portal,
  Modal,
  Button,
} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {StackScreenProps} from '@react-navigation/stack';
import type {TasksStackParamList} from '../navigation/TasksStackNavigator';
import type {Task} from '../types';
import {
  useTasks,
  useStatus,
  useTaskCycle,
  useTaskStop,
  useTaskPlan,
  useTaskRun,
  useAddTask,
} from '../hooks/useClaribot';
import {statusColors} from '../theme';
import StatusBar from '../components/StatusBar';
import TaskTreeItem, {buildTree} from '../components/TaskTreeItem';
import EmptyState from '../components/EmptyState';

type Props = StackScreenProps<TasksStackParamList, 'TasksList'>;

type ViewMode = 'tree' | 'list';

const STATUS_ICONS: Record<string, string> = {
  todo: 'checkbox-blank-circle-outline',
  planned: 'clipboard-text-outline',
  split: 'source-branch',
  done: 'check-circle',
  failed: 'alert-circle',
};

function parseItems(data: any): Task[] {
  if (!data) return [];
  if (Array.isArray(data)) return data;
  if (data.items && Array.isArray(data.items)) return data.items;
  return [];
}

export default function TasksScreen({navigation}: Props) {
  const theme = useTheme();
  const {data: tasksData, isLoading, refetch} = useTasks();
  const {data: statusData} = useStatus();
  const taskCycle = useTaskCycle();
  const taskStop = useTaskStop();
  const planAll = useTaskPlan();
  const runAll = useTaskRun();
  const addTask = useAddTask();

  const [viewMode, setViewMode] = useState<ViewMode>('tree');
  const [statusFilter, setStatusFilter] = useState<string | null>(null);
  const [expandedNodes, setExpandedNodes] = useState<Set<number>>(new Set());
  const [refreshing, setRefreshing] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [addSpec, setAddSpec] = useState('');
  const [addParentId, setAddParentId] = useState('');

  const currentProject =
    statusData?.project_id || 'GLOBAL';

  const cycleStatus = useMemo(() => {
    const cs = statusData?.cycle_status;
    if (cs?.project_id === currentProject) return cs;
    const multi = statusData?.cycle_statuses?.find(
      c => c.project_id === currentProject,
    );
    return multi || cs;
  }, [statusData, currentProject]);

  const isProjectRunning =
    cycleStatus?.status === 'running';

  const taskItems = useMemo(
    () => parseItems(tasksData?.data),
    [tasksData],
  );

  const taskStats = useMemo(() => {
    if (statusData?.task_stats) return statusData.task_stats;
    const stats = {total: 0, leaf: 0, todo: 0, planned: 0, split: 0, done: 0, failed: 0};
    for (const t of taskItems) {
      stats.total++;
      if (t.is_leaf) stats.leaf++;
      const s = t.status as keyof typeof stats;
      if (s in stats) (stats[s] as number)++;
    }
    return stats;
  }, [statusData, taskItems]);

  const filteredItems = useMemo(() => {
    if (!statusFilter) return taskItems;
    return taskItems.filter(t => t.status === statusFilter);
  }, [taskItems, statusFilter]);

  const treeData = useMemo(
    () => buildTree(statusFilter ? filteredItems : taskItems),
    [taskItems, filteredItems, statusFilter],
  );

  const sortedListItems = useMemo(
    () => [...filteredItems].sort((a, b) => b.id - a.id),
    [filteredItems],
  );

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  }, [refetch]);

  const toggleNode = useCallback((id: number) => {
    setExpandedNodes(prev => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  }, []);

  const handleSelect = useCallback(
    (task: Task) => {
      navigation.navigate('TaskDetail', {taskId: task.id});
    },
    [navigation],
  );

  const handlePlanAll = () => {
    Alert.alert('Plan All', '모든 Todo Task를 Plan 하시겠습니까?', [
      {text: '취소', style: 'cancel'},
      {text: '실행', onPress: () => planAll.mutate(undefined)},
    ]);
  };

  const handleRunAll = () => {
    Alert.alert('Run All', '모든 Planned Task를 실행하시겠습니까?', [
      {text: '취소', style: 'cancel'},
      {text: '실행', onPress: () => runAll.mutate(undefined)},
    ]);
  };

  const handleCycle = () => {
    Alert.alert('Cycle', 'Plan + Run 순회를 시작하시겠습니까?', [
      {text: '취소', style: 'cancel'},
      {
        text: '실행',
        onPress: () =>
          taskCycle.mutate(
            currentProject !== 'GLOBAL' ? currentProject : undefined,
          ),
      },
    ]);
  };

  const handleStop = () => {
    Alert.alert('Stop', '순회를 중단하시겠습니까?', [
      {text: '취소', style: 'cancel'},
      {text: '중단', style: 'destructive', onPress: () => taskStop.mutate()},
    ]);
  };

  const handleAdd = async () => {
    if (!addSpec.trim()) return;
    await addTask.mutateAsync({
      spec: addSpec.trim(),
      parentId: addParentId ? Number(addParentId) : undefined,
    });
    setAddSpec('');
    setAddParentId('');
    setShowAddModal(false);
  };

  const renderListItem = useCallback(
    ({item}: {item: Task}) => {
      const title =
        item.title || (item.spec || '').split('\n')[0] || '(untitled)';
      const statusColor =
        statusColors[item.status as keyof typeof statusColors] || '#9ca3af';
      const statusIcon =
        STATUS_ICONS[item.status] || 'checkbox-blank-circle-outline';

      return (
        <TouchableOpacity
          activeOpacity={0.7}
          onPress={() => handleSelect(item)}
          style={[
            styles.listCard,
            {
              backgroundColor: theme.colors.surface,
              borderColor: theme.colors.outline,
            },
          ]}>
          <View style={styles.listCardHeader}>
            <Text
              variant="labelSmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              #{item.id}
            </Text>
            <MaterialCommunityIcons
              name={statusIcon}
              size={16}
              color={statusColor}
            />
          </View>
          <Text
            variant="bodyMedium"
            numberOfLines={2}
            style={{color: theme.colors.onSurface, fontWeight: '500'}}>
            {title}
          </Text>
          <View style={styles.listCardFooter}>
            <Text
              variant="labelSmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              depth: {item.depth}
            </Text>
            <Text
              variant="labelSmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              parent: {item.parent_id ? `#${item.parent_id}` : '-'}
            </Text>
          </View>
        </TouchableOpacity>
      );
    },
    [theme, handleSelect],
  );

  const renderTreeContent = () => {
    if (treeData.length === 0) {
      if (statusFilter && taskItems.length > 0) {
        return (
          <EmptyState
            icon="filter-off-outline"
            message="필터에 맞는 Task가 없습니다"
          />
        );
      }
      return (
        <EmptyState
          icon="clipboard-text-outline"
          message="아직 Task가 없습니다"
        />
      );
    }

    return (
      <FlatList
        data={treeData}
        keyExtractor={node => String(node.task.id)}
        renderItem={({item: node}) => (
          <TaskTreeItem
            node={node}
            depth={0}
            expandedNodes={expandedNodes}
            selectedId={undefined}
            onToggle={toggleNode}
            onSelect={handleSelect}
          />
        )}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        contentContainerStyle={styles.listContent}
      />
    );
  };

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  return (
    <View
      style={[styles.container, {backgroundColor: theme.colors.background}]}>
      {/* Header / Toolbar */}
      <View
        style={[
          styles.toolbar,
          {
            backgroundColor: theme.colors.surface,
            borderBottomColor: theme.colors.outline,
          },
        ]}>
        <Text variant="titleLarge" style={{color: theme.colors.onSurface}}>
          Tasks
        </Text>
        <View style={styles.toolbarActions}>
          {/* View mode toggle */}
          <View
            style={[
              styles.toggleGroup,
              {borderColor: theme.colors.outline},
            ]}>
            <TouchableOpacity
              onPress={() => setViewMode('tree')}
              style={[
                styles.toggleBtn,
                viewMode === 'tree' && {
                  backgroundColor: theme.colors.primaryContainer,
                },
              ]}>
              <MaterialCommunityIcons
                name="file-tree"
                size={18}
                color={
                  viewMode === 'tree'
                    ? theme.colors.primary
                    : theme.colors.onSurfaceVariant
                }
              />
            </TouchableOpacity>
            <TouchableOpacity
              onPress={() => setViewMode('list')}
              style={[
                styles.toggleBtn,
                viewMode === 'list' && {
                  backgroundColor: theme.colors.primaryContainer,
                },
              ]}>
              <MaterialCommunityIcons
                name="format-list-bulleted"
                size={18}
                color={
                  viewMode === 'list'
                    ? theme.colors.primary
                    : theme.colors.onSurfaceVariant
                }
              />
            </TouchableOpacity>
          </View>

          {/* Action buttons */}
          {isProjectRunning ? (
            <IconButton
              icon="stop-circle-outline"
              size={20}
              onPress={handleStop}
              disabled={taskStop.isPending}
            />
          ) : (
            <>
              <IconButton
                icon="clipboard-text-outline"
                size={20}
                onPress={handlePlanAll}
                disabled={
                  planAll.isPending || currentProject === 'GLOBAL'
                }
              />
              <IconButton
                icon="play-circle-outline"
                size={20}
                onPress={handleRunAll}
                disabled={
                  runAll.isPending || currentProject === 'GLOBAL'
                }
              />
              <IconButton
                icon="refresh"
                size={20}
                onPress={handleCycle}
                disabled={
                  taskCycle.isPending || currentProject === 'GLOBAL'
                }
              />
            </>
          )}
        </View>
      </View>

      {/* Status Bar */}
      {taskItems.length > 0 && (
        <View
          style={[
            styles.statusBarContainer,
            {backgroundColor: theme.colors.surface},
          ]}>
          <StatusBar
            taskStats={taskStats}
            cycleStatus={cycleStatus}
            statusFilter={statusFilter}
            onFilterChange={setStatusFilter}
          />
        </View>
      )}

      {/* Content */}
      {viewMode === 'tree' ? (
        renderTreeContent()
      ) : (
        <FlatList
          data={sortedListItems}
          keyExtractor={item => String(item.id)}
          renderItem={renderListItem}
          contentContainerStyle={styles.listContent}
          refreshControl={
            <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
          }
          ListEmptyComponent={
            statusFilter && taskItems.length > 0 ? (
              <EmptyState
                icon="filter-off-outline"
                message="필터에 맞는 Task가 없습니다"
              />
            ) : (
              <EmptyState
                icon="clipboard-text-outline"
                message="아직 Task가 없습니다"
              />
            )
          }
        />
      )}

      {/* FAB - Add Task */}
      <FAB
        icon="plus"
        onPress={() => setShowAddModal(true)}
        style={[styles.fab, {backgroundColor: theme.colors.primary}]}
        color={theme.colors.onPrimary}
      />

      {/* Add Task Modal */}
      <Portal>
        <Modal
          visible={showAddModal}
          onDismiss={() => setShowAddModal(false)}
          contentContainerStyle={[
            styles.modal,
            {backgroundColor: theme.colors.surface},
          ]}>
          <Text
            variant="titleMedium"
            style={{color: theme.colors.onSurface, marginBottom: 12}}>
            새 Task 추가
          </Text>

          <TextInput
            placeholder="Task spec (첫 줄이 제목이 됩니다)"
            value={addSpec}
            onChangeText={setAddSpec}
            multiline
            numberOfLines={4}
            style={[
              styles.textInput,
              {
                borderColor: theme.colors.outline,
                color: theme.colors.onSurface,
                backgroundColor: theme.colors.surfaceVariant,
              },
            ]}
            placeholderTextColor={theme.colors.onSurfaceVariant}
          />

          <TextInput
            placeholder="Parent ID (선택)"
            value={addParentId}
            onChangeText={setAddParentId}
            keyboardType="numeric"
            style={[
              styles.textInputSingle,
              {
                borderColor: theme.colors.outline,
                color: theme.colors.onSurface,
                backgroundColor: theme.colors.surfaceVariant,
              },
            ]}
            placeholderTextColor={theme.colors.onSurfaceVariant}
          />

          <View style={styles.modalActions}>
            <Button
              mode="outlined"
              onPress={() => setShowAddModal(false)}>
              취소
            </Button>
            <Button
              mode="contained"
              onPress={handleAdd}
              disabled={!addSpec.trim() || addTask.isPending}
              loading={addTask.isPending}>
              추가
            </Button>
          </View>
        </Modal>
      </Portal>
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
  toolbar: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderBottomWidth: 1,
  },
  toolbarActions: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 0,
  },
  toggleGroup: {
    flexDirection: 'row',
    borderWidth: 1,
    borderRadius: 6,
    overflow: 'hidden',
  },
  toggleBtn: {
    paddingHorizontal: 8,
    paddingVertical: 6,
  },
  statusBarContainer: {
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
  listContent: {
    padding: 12,
    paddingBottom: 80,
  },
  listCard: {
    borderWidth: 1,
    borderRadius: 8,
    padding: 12,
    marginBottom: 8,
  },
  listCardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  listCardFooter: {
    flexDirection: 'row',
    gap: 12,
    marginTop: 6,
  },
  fab: {
    position: 'absolute',
    right: 16,
    bottom: 16,
  },
  modal: {
    margin: 20,
    padding: 20,
    borderRadius: 12,
  },
  textInput: {
    borderWidth: 1,
    borderRadius: 8,
    padding: 12,
    fontSize: 14,
    textAlignVertical: 'top',
    minHeight: 100,
    marginBottom: 10,
  },
  textInputSingle: {
    borderWidth: 1,
    borderRadius: 8,
    padding: 12,
    fontSize: 14,
    height: 44,
    marginBottom: 16,
  },
  modalActions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 8,
  },
});
