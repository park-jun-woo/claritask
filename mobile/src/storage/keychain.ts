import * as Keychain from 'react-native-keychain';

const AUTH_SERVICE = 'com.claribot.auth';
const SERVER_SERVICE = 'com.claribot.server';

export async function saveToken(token: string): Promise<void> {
  await Keychain.setGenericPassword('token', token, {service: AUTH_SERVICE});
}

export async function getToken(): Promise<string | null> {
  const credentials = await Keychain.getGenericPassword({
    service: AUTH_SERVICE,
  });
  if (credentials) {
    return credentials.password;
  }
  return null;
}

export async function deleteToken(): Promise<void> {
  await Keychain.resetGenericPassword({service: AUTH_SERVICE});
}

export async function saveServerUrl(url: string): Promise<void> {
  await Keychain.setGenericPassword('server', url, {service: SERVER_SERVICE});
}

export async function getServerUrl(): Promise<string | null> {
  const credentials = await Keychain.getGenericPassword({
    service: SERVER_SERVICE,
  });
  if (credentials) {
    return credentials.password;
  }
  return null;
}

export async function deleteServerUrl(): Promise<void> {
  await Keychain.resetGenericPassword({service: SERVER_SERVICE});
}
