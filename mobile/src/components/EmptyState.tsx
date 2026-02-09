import React from 'react';
import {View, StyleSheet} from 'react-native';
import {Text, useTheme} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';

interface EmptyStateProps {
  icon?: string;
  message?: string;
}

export default function EmptyState({
  icon = 'text-box-remove-outline',
  message = '아직 내용이 없습니다',
}: EmptyStateProps) {
  const theme = useTheme();

  return (
    <View style={styles.container}>
      <MaterialCommunityIcons
        name={icon}
        size={48}
        color={theme.colors.onSurfaceVariant}
      />
      <Text
        variant="bodyMedium"
        style={[styles.text, {color: theme.colors.onSurfaceVariant}]}>
        {message}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 48,
    gap: 12,
  },
  text: {
    textAlign: 'center',
  },
});
