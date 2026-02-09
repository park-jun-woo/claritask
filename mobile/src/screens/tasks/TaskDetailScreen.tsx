import React, {useState} from 'react';
import {View, ScrollView, StyleSheet, Alert} from 'react-native';
import {
  Text,
  useTheme,
  SegmentedButtons,
  Button,
  ActivityIndicator,
  Chip,
} from 'react-native-paper';
import type {StackScreenProps} from '@react-navigation/stack';
import type {TasksStackParamList} from '../../navigation/TasksStackNavigator';
import {useTask, useTaskPlan, useTaskRun} from '../../hooks/useClaribot';
import {statusColors} from '../../theme';
import MarkdownRenderer from '../../components/MarkdownRenderer';
import EmptyState from '../../components/EmptyState';

type Props = StackScreenProps<TasksStackParamList, 'TaskDetail'>;

type TabKey = 'spec' | 'plan' | 'report';

const STATUS_LABELS: Record<string, string> = {
  todo: 'Todo',
  planned: 'Planned',
  split: 'Split',
  done: 'Done',
  failed: 'Failed',
};

export default function TaskDetailScreen({route}: Props) {
  const {taskId} = route.params;
  const theme = useTheme();
  const {data, isLoading} = useTask(taskId);
  const planMutation = useTaskPlan();
  const runMutation = useTaskRun();

  const [activeTab, setActiveTab] = useState<TabKey>('spec');

  const task = data?.data;

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  if (!task) {
    return (
      <View style={styles.center}>
        <EmptyState icon="alert-circle-outline" message="Task를 찾을 수 없습니다" />
      </View>
    );
  }

  const statusColor = statusColors[task.status as keyof typeof statusColors] || '#9ca3af';

  const handlePlan = () => {
    Alert.alert(
      'Plan 실행',
      `Task #${task.id}의 Plan을 실행하시겠습니까?`,
      [
        {text: '취소', style: 'cancel'},
        {
          text: '실행',
          onPress: () => planMutation.mutate(task.id),
        },
      ],
    );
  };

  const handleRun = () => {
    Alert.alert(
      'Run 실행',
      `Task #${task.id}를 실행하시겠습니까?`,
      [
        {text: '취소', style: 'cancel'},
        {
          text: '실행',
          onPress: () => runMutation.mutate(task.id),
        },
      ],
    );
  };

  const renderTabContent = () => {
    const content = task[activeTab] || '';

    if (!content.trim()) {
      const icons: Record<TabKey, string> = {
        spec: 'file-document-outline',
        plan: 'clipboard-text-outline',
        report: 'chart-box-outline',
      };
      const messages: Record<TabKey, string> = {
        spec: 'Spec이 아직 작성되지 않았습니다',
        plan: 'Plan이 아직 생성되지 않았습니다',
        report: 'Report가 아직 생성되지 않았습니다',
      };
      return <EmptyState icon={icons[activeTab]} message={messages[activeTab]} />;
    }

    return <MarkdownRenderer content={content} />;
  };

  const showPlanButton = task.status === 'todo';
  const showRunButton = task.status === 'planned';
  const isMutating = planMutation.isPending || runMutation.isPending;

  const formattedDate = task.created_at
    ? new Date(task.created_at).toLocaleDateString('ko-KR')
    : '';

  return (
    <View style={[styles.container, {backgroundColor: theme.colors.background}]}>
      {/* Header info */}
      <View
        style={[
          styles.header,
          {
            backgroundColor: theme.colors.surface,
            borderBottomColor: theme.colors.outline,
          },
        ]}>
        <View style={styles.headerTop}>
          <Text variant="titleLarge" style={{color: theme.colors.onSurface}}>
            #{task.id} {task.title}
          </Text>
          <Chip
            compact
            style={{backgroundColor: statusColor + '20'}}
            textStyle={{color: statusColor, fontSize: 12, fontWeight: '600'}}>
            {STATUS_LABELS[task.status] || task.status}
          </Chip>
        </View>

        <View style={styles.metaRow}>
          {formattedDate ? (
            <Text
              variant="bodySmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              생성일: {formattedDate}
            </Text>
          ) : null}
          {task.parent_id != null && (
            <Text
              variant="bodySmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              부모: #{task.parent_id}
            </Text>
          )}
          {task.priority > 0 && (
            <Text
              variant="bodySmall"
              style={{color: theme.colors.onSurfaceVariant}}>
              우선순위: {task.priority}
            </Text>
          )}
        </View>

        {task.error ? (
          <View
            style={[
              styles.errorBox,
              {backgroundColor: theme.colors.error + '15'},
            ]}>
            <Text
              variant="bodySmall"
              style={{color: theme.colors.error}}>
              {task.error}
            </Text>
          </View>
        ) : null}

        {/* Action buttons */}
        {(showPlanButton || showRunButton) && (
          <View style={styles.actions}>
            {showPlanButton && (
              <Button
                mode="contained"
                icon="clipboard-text-outline"
                onPress={handlePlan}
                loading={planMutation.isPending}
                disabled={isMutating}>
                Plan
              </Button>
            )}
            {showRunButton && (
              <Button
                mode="contained"
                icon="play"
                onPress={handleRun}
                loading={runMutation.isPending}
                disabled={isMutating}>
                Run
              </Button>
            )}
          </View>
        )}
      </View>

      {/* Tab selector */}
      <View style={[styles.tabContainer, {backgroundColor: theme.colors.surface}]}>
        <SegmentedButtons
          value={activeTab}
          onValueChange={v => setActiveTab(v as TabKey)}
          buttons={[
            {value: 'spec', label: 'Spec'},
            {value: 'plan', label: 'Plan'},
            {value: 'report', label: 'Report'},
          ]}
          density="small"
        />
      </View>

      {/* Tab content */}
      <ScrollView
        style={styles.content}
        contentContainerStyle={styles.contentContainer}>
        {renderTabContent()}
      </ScrollView>
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
  header: {
    padding: 16,
    borderBottomWidth: 1,
    gap: 8,
  },
  headerTop: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    gap: 8,
  },
  metaRow: {
    flexDirection: 'row',
    gap: 16,
    flexWrap: 'wrap',
  },
  errorBox: {
    padding: 8,
    borderRadius: 6,
  },
  actions: {
    flexDirection: 'row',
    gap: 8,
    marginTop: 4,
  },
  tabContainer: {
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  content: {
    flex: 1,
  },
  contentContainer: {
    padding: 16,
    paddingBottom: 32,
  },
});
