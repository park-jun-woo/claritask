import React from 'react';
import {View, TouchableOpacity, StyleSheet} from 'react-native';
import {Text, useTheme} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {Task} from '../types';
import {statusColors} from '../theme';

export interface TreeNode {
  task: Task;
  children: TreeNode[];
}

interface TaskTreeItemProps {
  node: TreeNode;
  depth: number;
  expandedNodes: Set<number>;
  selectedId?: number;
  onToggle: (id: number) => void;
  onSelect: (task: Task) => void;
}

const STATUS_ICONS: Record<string, string> = {
  todo: 'checkbox-blank-circle-outline',
  planned: 'clipboard-text-outline',
  split: 'source-branch',
  done: 'check-circle',
  failed: 'alert-circle',
};

export function buildTree(items: Task[]): TreeNode[] {
  const map = new Map<number, TreeNode>();
  const roots: TreeNode[] = [];

  items.forEach(t => {
    map.set(t.id, {task: t, children: []});
  });

  items.forEach(t => {
    const node = map.get(t.id)!;
    if (t.parent_id && map.has(t.parent_id)) {
      map.get(t.parent_id)!.children.push(node);
    } else {
      roots.push(node);
    }
  });

  const sortDesc = (a: TreeNode, b: TreeNode) => b.task.id - a.task.id;
  roots.sort(sortDesc);
  map.forEach(node => node.children.sort(sortDesc));

  return roots;
}

export default function TaskTreeItem({
  node,
  depth,
  expandedNodes,
  selectedId,
  onToggle,
  onSelect,
}: TaskTreeItemProps) {
  const theme = useTheme();
  const {task} = node;
  const hasChildren = node.children.length > 0;
  const isExpanded = expandedNodes.has(task.id);
  const isSelected = task.id === selectedId;
  const statusColor =
    statusColors[task.status as keyof typeof statusColors] || '#9ca3af';
  const statusIcon = STATUS_ICONS[task.status] || 'checkbox-blank-circle-outline';
  const title =
    task.title || (task.spec || '').split('\n')[0] || '(untitled)';

  return (
    <View>
      <TouchableOpacity
        activeOpacity={0.7}
        onPress={() => onSelect(task)}
        style={[
          styles.row,
          {
            paddingLeft: depth * 12 + 8,
            backgroundColor: isSelected
              ? theme.colors.primaryContainer + '60'
              : 'transparent',
          },
        ]}>
        {hasChildren ? (
          <TouchableOpacity
            hitSlop={{top: 12, bottom: 12, left: 12, right: 12}}
            onPress={() => onToggle(task.id)}
            style={styles.expandBtn}>
            <MaterialCommunityIcons
              name={isExpanded ? 'chevron-down' : 'chevron-right'}
              size={18}
              color={theme.colors.onSurfaceVariant}
            />
          </TouchableOpacity>
        ) : (
          <View style={styles.expandPlaceholder} />
        )}

        <MaterialCommunityIcons
          name={statusIcon}
          size={16}
          color={statusColor}
        />

        <Text
          variant="labelSmall"
          style={{color: theme.colors.onSurfaceVariant, marginRight: 2}}>
          #{task.id}
        </Text>

        <Text
          variant="bodySmall"
          numberOfLines={1}
          style={{flex: 1, color: theme.colors.onSurface}}>
          {title}
        </Text>
      </TouchableOpacity>

      {hasChildren && isExpanded && (
        <View>
          {node.children.map(child => (
            <TaskTreeItem
              key={child.task.id}
              node={child}
              depth={depth + 1}
              expandedNodes={expandedNodes}
              selectedId={selectedId}
              onToggle={onToggle}
              onSelect={onSelect}
            />
          ))}
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingVertical: 10,
    paddingRight: 8,
    minHeight: 44,
  },
  expandBtn: {
    width: 24,
    height: 24,
    justifyContent: 'center',
    alignItems: 'center',
  },
  expandPlaceholder: {
    width: 24,
  },
});
