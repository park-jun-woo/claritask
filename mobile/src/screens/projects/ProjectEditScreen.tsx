import React, {useState, useEffect} from 'react';
import {View, ScrollView, StyleSheet, Alert} from 'react-native';
import {
  Text,
  TextInput,
  Button,
  useTheme,
  ActivityIndicator,
} from 'react-native-paper';
import type {StackScreenProps} from '@react-navigation/stack';
import type {MoreStackParamList} from '../../navigation/MoreStackNavigator';
import type {Project} from '../../types';
import {useProject, useUpdateProject} from '../../hooks/useClaribot';
import EmptyState from '../../components/EmptyState';

type Props = StackScreenProps<MoreStackParamList, 'ProjectEdit'>;

export default function ProjectEditScreen({route, navigation}: Props) {
  const {projectId} = route.params;
  const theme = useTheme();
  const {data, isLoading} = useProject(projectId);
  const updateProject = useUpdateProject();

  const project: Project | undefined = data?.data as Project | undefined;

  const [description, setDescription] = useState('');
  const [parallel, setParallel] = useState('');
  const [dirty, setDirty] = useState(false);

  useEffect(() => {
    if (project) {
      setDescription(project.description || '');
      // parallel is not in the Project type directly, but we handle it via API
    }
  }, [project]);

  const handleSave = () => {
    const updates: {description?: string; parallel?: number} = {};

    if (description !== (project?.description || '')) {
      updates.description = description;
    }

    const parallelNum = parseInt(parallel, 10);
    if (parallel.trim() && !isNaN(parallelNum) && parallelNum > 0) {
      updates.parallel = parallelNum;
    }

    if (Object.keys(updates).length === 0) {
      navigation.goBack();
      return;
    }

    updateProject.mutate(
      {id: projectId, data: updates},
      {
        onSuccess: () => {
          Alert.alert('Saved', 'Project updated successfully.', [
            {text: 'OK', onPress: () => navigation.goBack()},
          ]);
        },
        onError: err => {
          Alert.alert('Error', err.message || 'Failed to update project.');
        },
      },
    );
  };

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  if (!project) {
    return (
      <View style={styles.center}>
        <EmptyState
          icon="alert-circle-outline"
          message="Project not found"
        />
      </View>
    );
  }

  return (
    <ScrollView
      style={[styles.container, {backgroundColor: theme.colors.background}]}
      contentContainerStyle={styles.content}>
      {/* Project Info (read-only) */}
      <View style={[styles.section, {backgroundColor: theme.colors.surface, borderColor: theme.colors.outline}]}>
        <Text variant="titleMedium" style={{color: theme.colors.onSurface}}>
          {project.id}
        </Text>
        <Text
          variant="bodySmall"
          style={{color: theme.colors.onSurfaceVariant}}>
          {project.path}
        </Text>
      </View>

      {/* Description */}
      <View style={styles.fieldGroup}>
        <Text variant="labelLarge" style={{color: theme.colors.onSurface}}>
          Description
        </Text>
        <TextInput
          mode="outlined"
          value={description}
          onChangeText={v => {
            setDescription(v);
            setDirty(true);
          }}
          placeholder="Project description"
          multiline
          numberOfLines={3}
          style={styles.input}
        />
      </View>

      {/* Parallel */}
      <View style={styles.fieldGroup}>
        <Text variant="labelLarge" style={{color: theme.colors.onSurface}}>
          Claude Parallel Instances
        </Text>
        <Text
          variant="bodySmall"
          style={{color: theme.colors.onSurfaceVariant, marginBottom: 4}}>
          Number of Claude Code instances to run in parallel for this project.
        </Text>
        <TextInput
          mode="outlined"
          value={parallel}
          onChangeText={v => {
            setParallel(v.replace(/[^0-9]/g, ''));
            setDirty(true);
          }}
          placeholder="Default (from config)"
          keyboardType="numeric"
          style={styles.input}
        />
      </View>

      {/* Save Button */}
      <Button
        mode="contained"
        onPress={handleSave}
        loading={updateProject.isPending}
        disabled={!dirty || updateProject.isPending}
        style={styles.saveButton}
        icon="content-save">
        Save
      </Button>
    </ScrollView>
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
  content: {
    padding: 16,
    gap: 16,
    paddingBottom: 32,
  },
  section: {
    padding: 14,
    borderRadius: 10,
    borderWidth: 1,
    gap: 4,
  },
  fieldGroup: {
    gap: 6,
  },
  input: {
    fontSize: 14,
  },
  saveButton: {
    marginTop: 8,
  },
});
