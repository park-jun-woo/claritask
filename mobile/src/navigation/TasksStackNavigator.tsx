import React from 'react';
import {createStackNavigator} from '@react-navigation/stack';
import {useTheme} from 'react-native-paper';
import TasksScreen from '../screens/TasksScreen';
import TaskDetailScreen from '../screens/tasks/TaskDetailScreen';

export type TasksStackParamList = {
  TasksList: undefined;
  TaskDetail: {taskId: number};
};

const Stack = createStackNavigator<TasksStackParamList>();

export default function TasksStackNavigator() {
  const theme = useTheme();

  return (
    <Stack.Navigator
      screenOptions={{
        headerStyle: {backgroundColor: theme.colors.surface},
        headerTintColor: theme.colors.onSurface,
      }}>
      <Stack.Screen
        name="TasksList"
        component={TasksScreen}
        options={{headerShown: false}}
      />
      <Stack.Screen
        name="TaskDetail"
        component={TaskDetailScreen}
        options={({route}) => ({
          title: `Task #${route.params.taskId}`,
        })}
      />
    </Stack.Navigator>
  );
}
