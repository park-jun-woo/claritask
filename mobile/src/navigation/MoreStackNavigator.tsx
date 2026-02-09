import React from 'react';
import {createStackNavigator} from '@react-navigation/stack';
import {useTheme} from 'react-native-paper';
import MoreScreen from '../screens/MoreScreen';
import ProjectsScreen from '../screens/projects/ProjectsScreen';
import ProjectEditScreen from '../screens/projects/ProjectEditScreen';
import SpecsScreen from '../screens/specs/SpecsScreen';
import SpecDetailScreen from '../screens/specs/SpecDetailScreen';
import SchedulesScreen from '../screens/schedules/SchedulesScreen';
import ScheduleDetailScreen from '../screens/schedules/ScheduleDetailScreen';
import SettingsScreen from '../screens/SettingsScreen';

export type MoreStackParamList = {
  MoreMenu: undefined;
  Projects: undefined;
  ProjectEdit: {projectId: string};
  Specs: undefined;
  SpecDetail: {specId: number};
  Schedules: undefined;
  ScheduleDetail: {scheduleId: number};
  Settings: undefined;
};

const Stack = createStackNavigator<MoreStackParamList>();

export default function MoreStackNavigator() {
  const theme = useTheme();

  return (
    <Stack.Navigator
      screenOptions={{
        headerStyle: {backgroundColor: theme.colors.surface},
        headerTintColor: theme.colors.onSurface,
      }}>
      <Stack.Screen
        name="MoreMenu"
        component={MoreScreen}
        options={{title: 'More'}}
      />
      <Stack.Screen
        name="Projects"
        component={ProjectsScreen}
        options={{title: 'Projects'}}
      />
      <Stack.Screen
        name="ProjectEdit"
        component={ProjectEditScreen}
        options={({route}) => ({
          title: `Edit: ${route.params.projectId}`,
        })}
      />
      <Stack.Screen
        name="Specs"
        component={SpecsScreen}
        options={{title: 'Specs'}}
      />
      <Stack.Screen
        name="SpecDetail"
        component={SpecDetailScreen}
        options={{title: 'Spec Detail'}}
      />
      <Stack.Screen
        name="Schedules"
        component={SchedulesScreen}
        options={{title: 'Schedules'}}
      />
      <Stack.Screen
        name="ScheduleDetail"
        component={ScheduleDetailScreen}
        options={{title: 'Schedule Detail'}}
      />
      <Stack.Screen
        name="Settings"
        component={SettingsScreen}
        options={{title: 'Settings'}}
      />
    </Stack.Navigator>
  );
}
