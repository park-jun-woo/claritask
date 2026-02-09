import React from 'react';
import {createStackNavigator} from '@react-navigation/stack';
import {useAuth} from '../hooks/useAuth';
import AuthNavigator from './AuthNavigator';
import TabNavigator from './TabNavigator';
import {ActivityIndicator, View, StyleSheet} from 'react-native';

export type RootParamList = {
  Auth: undefined;
  Main: undefined;
};

const Stack = createStackNavigator<RootParamList>();

export default function RootNavigator() {
  const {state, isLoading} = useAuth();

  if (isLoading || state === 'loading') {
    return (
      <View style={styles.loading}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  const isAuthenticated = state === 'authenticated';

  return (
    <Stack.Navigator screenOptions={{headerShown: false}}>
      {isAuthenticated ? (
        <Stack.Screen name="Main" component={TabNavigator} />
      ) : (
        <Stack.Screen name="Auth" component={AuthNavigator} />
      )}
    </Stack.Navigator>
  );
}

const styles = StyleSheet.create({
  loading: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
});
