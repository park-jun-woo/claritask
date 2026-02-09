import React, {useState, useMemo} from 'react';
import {View, StyleSheet, ScrollView} from 'react-native';
import {
  Text,
  useTheme,
  Menu,
  TouchableRipple,
  Divider,
} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import type {Project} from '../types';

interface ProjectSelectorProps {
  projects: Project[];
  selectedId: string | null;
  onSelect: (id: string | null) => void;
}

export default function ProjectSelector({
  projects,
  selectedId,
  onSelect,
}: ProjectSelectorProps) {
  const theme = useTheme();
  const [visible, setVisible] = useState(false);

  const sortedProjects = useMemo(() => {
    return [...projects].sort((a, b) => {
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
      return (b.last_accessed || b.created_at).localeCompare(
        a.last_accessed || a.created_at,
      );
    });
  }, [projects]);

  const selectedLabel = selectedId
    ? projects.find(p => p.id === selectedId)?.id || selectedId
    : '전체';

  return (
    <Menu
      visible={visible}
      onDismiss={() => setVisible(false)}
      anchor={
        <TouchableRipple
          onPress={() => setVisible(true)}
          style={[
            styles.anchor,
            {
              backgroundColor: theme.colors.surfaceVariant,
              borderColor: theme.colors.outline,
            },
          ]}>
          <View style={styles.anchorContent}>
            <MaterialCommunityIcons
              name="folder-outline"
              size={18}
              color={theme.colors.onSurface}
            />
            <Text
              variant="bodyMedium"
              numberOfLines={1}
              style={[styles.anchorText, {color: theme.colors.onSurface}]}>
              {selectedLabel}
            </Text>
            <MaterialCommunityIcons
              name="chevron-down"
              size={18}
              color={theme.colors.onSurfaceVariant}
            />
          </View>
        </TouchableRipple>
      }
      contentStyle={{maxHeight: 300}}>
      <ScrollView style={{maxHeight: 280}}>
        <Menu.Item
          title="전체"
          leadingIcon={selectedId === null ? 'check' : undefined}
          onPress={() => {
            onSelect(null);
            setVisible(false);
          }}
        />
        <Divider />
        {sortedProjects.map(project => (
          <Menu.Item
            key={project.id}
            title={project.id}
            leadingIcon={
              project.pinned
                ? 'pin'
                : selectedId === project.id
                  ? 'check'
                  : undefined
            }
            onPress={() => {
              onSelect(project.id);
              setVisible(false);
            }}
          />
        ))}
      </ScrollView>
    </Menu>
  );
}

const styles = StyleSheet.create({
  anchor: {
    borderRadius: 8,
    borderWidth: 1,
    paddingHorizontal: 12,
    paddingVertical: 8,
    minHeight: 44,
    justifyContent: 'center',
  },
  anchorContent: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  anchorText: {
    flex: 1,
    fontWeight: '500',
  },
});
