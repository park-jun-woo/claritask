import React from 'react';
import {ScrollView, StyleSheet, TouchableOpacity, View} from 'react-native';
import {Text, useTheme, Divider} from 'react-native-paper';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';

interface MenuItem {
  icon: string;
  label: string;
  description: string;
  screen?: string;
}

const menuItems: MenuItem[] = [
  {
    icon: 'folder-multiple-outline',
    label: 'Projects',
    description: 'Manage and switch projects',
    screen: 'Projects',
  },
  {
    icon: 'file-document-outline',
    label: 'Specs',
    description: 'Requirement specifications',
    screen: 'Specs',
  },
  {
    icon: 'clock-outline',
    label: 'Schedules',
    description: 'Scheduled tasks and automation',
    screen: 'Schedules',
  },
  {
    icon: 'cog-outline',
    label: 'Settings',
    description: 'App configuration',
    screen: 'Settings',
  },
];

export default function MoreScreen({navigation}: any) {
  const theme = useTheme();

  return (
    <ScrollView
      style={[styles.container, {backgroundColor: theme.colors.background}]}
      contentContainerStyle={styles.content}>
      {menuItems.map((item, index) => (
        <React.Fragment key={item.label}>
          {index > 0 && (
            <Divider style={{backgroundColor: theme.colors.outline}} />
          )}
          <TouchableOpacity
            activeOpacity={0.7}
            onPress={() => item.screen && navigation.navigate(item.screen)}
            disabled={!item.screen}
            style={[styles.menuItem, !item.screen && {opacity: 0.5}]}>
            <MaterialCommunityIcons
              name={item.icon}
              size={24}
              color={theme.colors.primary}
            />
            <View style={styles.menuText}>
              <Text
                variant="bodyLarge"
                style={{color: theme.colors.onSurface}}>
                {item.label}
              </Text>
              <Text
                variant="bodySmall"
                style={{color: theme.colors.onSurfaceVariant}}>
                {item.description}
              </Text>
            </View>
            <MaterialCommunityIcons
              name="chevron-right"
              size={24}
              color={theme.colors.onSurfaceVariant}
            />
          </TouchableOpacity>
        </React.Fragment>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    paddingVertical: 8,
  },
  menuItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 14,
    gap: 14,
  },
  menuText: {
    flex: 1,
    gap: 2,
  },
});
