import React from 'react';
import {createStackNavigator} from '@react-navigation/stack';
import ServerConnectScreen from '../screens/ServerConnectScreen';
import LoginScreen from '../screens/LoginScreen';

export type AuthParamList = {
  ServerConnect: undefined;
  Login: undefined;
};

const Stack = createStackNavigator<AuthParamList>();

export default function AuthNavigator() {
  return (
    <Stack.Navigator screenOptions={{headerShown: false}}>
      <Stack.Screen name="ServerConnect" component={ServerConnectScreen} />
      <Stack.Screen name="Login" component={LoginScreen} />
    </Stack.Navigator>
  );
}
