import React, {useState, useCallback} from 'react';
import {Alert, ScrollView, StyleSheet, TouchableOpacity, View} from 'react-native';
import {
  Text,
  useTheme,
  Divider,
  Switch,
  ActivityIndicator,
} from 'react-native-paper';
import {useQueryClient} from '@tanstack/react-query';
import MaterialCommunityIcons from 'react-native-vector-icons/MaterialCommunityIcons';
import {useAuth} from '../hooks/useAuth';
import {useHealth} from '../hooks/useClaribot';
import {mmkv} from '../storage/cache';

export default function SettingsScreen() {
  const theme = useTheme();
  const queryClient = useQueryClient();
  const {serverUrl, logout, disconnectServer} = useAuth();
  const {data: healthData} = useHealth();
  const [clearingCache, setClearingCache] = useState(false);

  const handleClearCache = useCallback(() => {
    Alert.alert('캐시 초기화', '모든 캐시 데이터를 삭제합니다. 계속하시겠습니까?', [
      {text: '취소', style: 'cancel'},
      {
        text: '초기화',
        style: 'destructive',
        onPress: async () => {
          setClearingCache(true);
          try {
            // Clear MMKV cache (preserve server_url)
            const keys = mmkv.getAllKeys();
            for (const key of keys) {
              if (key !== 'server_url') {
                mmkv.remove(key);
              }
            }
            // Invalidate all TanStack Query caches
            await queryClient.invalidateQueries();
            Alert.alert('완료', '캐시가 초기화되었습니다.');
          } finally {
            setClearingCache(false);
          }
        },
      },
    ]);
  }, [queryClient]);

  const handleLogout = useCallback(() => {
    Alert.alert('로그아웃', '로그아웃하시겠습니까?', [
      {text: '취소', style: 'cancel'},
      {
        text: '로그아웃',
        style: 'destructive',
        onPress: async () => {
          await logout();
        },
      },
    ]);
  }, [logout]);

  const handleChangeServer = useCallback(() => {
    Alert.alert('서버 변경', '현재 서버에서 로그아웃하고 새 서버에 연결합니다.', [
      {text: '취소', style: 'cancel'},
      {
        text: '변경',
        onPress: async () => {
          await disconnectServer();
        },
      },
    ]);
  }, [disconnectServer]);

  const cacheSize = useCallback(() => {
    const keys = mmkv.getAllKeys();
    return `${keys.length} items`;
  }, []);

  const sectionStyle = [styles.section, {backgroundColor: theme.colors.surface}];
  const labelStyle = {color: theme.colors.onSurface};
  const subStyle = {color: theme.colors.onSurfaceVariant};

  return (
    <ScrollView
      style={[styles.container, {backgroundColor: theme.colors.background}]}
      contentContainerStyle={styles.content}>
      {/* Server Connection */}
      <Text
        variant="labelLarge"
        style={[styles.sectionHeader, subStyle]}>
        서버 연결
      </Text>
      <View style={sectionStyle}>
        <View style={styles.row}>
          <MaterialCommunityIcons
            name="server"
            size={22}
            color={theme.colors.primary}
          />
          <View style={styles.rowText}>
            <Text variant="bodyMedium" style={labelStyle}>
              서버 URL
            </Text>
            <Text variant="bodySmall" style={subStyle}>
              {serverUrl || '연결되지 않음'}
            </Text>
          </View>
          <View
            style={[
              styles.statusBadge,
              {
                backgroundColor: healthData
                  ? '#dcfce7'
                  : '#fef2f2',
              },
            ]}>
            <Text
              variant="labelSmall"
              style={{
                color: healthData ? '#16a34a' : '#dc2626',
              }}>
              {healthData ? '연결됨' : '끊김'}
            </Text>
          </View>
        </View>
        <Divider style={{backgroundColor: theme.colors.outline}} />
        <SettingsButton
          icon="swap-horizontal"
          label="서버 변경"
          theme={theme}
          onPress={handleChangeServer}
        />
      </View>

      {/* Cache Management */}
      <Text
        variant="labelLarge"
        style={[styles.sectionHeader, subStyle]}>
        캐시 관리
      </Text>
      <View style={sectionStyle}>
        <View style={styles.row}>
          <MaterialCommunityIcons
            name="database-outline"
            size={22}
            color={theme.colors.primary}
          />
          <View style={styles.rowText}>
            <Text variant="bodyMedium" style={labelStyle}>
              캐시 크기
            </Text>
            <Text variant="bodySmall" style={subStyle}>
              {cacheSize()}
            </Text>
          </View>
        </View>
        <Divider style={{backgroundColor: theme.colors.outline}} />
        <SettingsButton
          icon="delete-outline"
          label="캐시 초기화"
          theme={theme}
          onPress={handleClearCache}
          loading={clearingCache}
          destructive
        />
      </View>

      {/* Notifications (Phase 2) */}
      <Text
        variant="labelLarge"
        style={[styles.sectionHeader, subStyle]}>
        알림 설정
      </Text>
      <View style={sectionStyle}>
        <View style={styles.row}>
          <MaterialCommunityIcons
            name="bell-check-outline"
            size={22}
            color={theme.colors.onSurfaceVariant}
          />
          <View style={styles.rowText}>
            <Text
              variant="bodyMedium"
              style={{color: theme.colors.onSurfaceVariant}}>
              Task 완료 알림
            </Text>
          </View>
          <Switch value={false} disabled />
        </View>
        <Divider style={{backgroundColor: theme.colors.outline}} />
        <View style={styles.row}>
          <MaterialCommunityIcons
            name="bell-ring-outline"
            size={22}
            color={theme.colors.onSurfaceVariant}
          />
          <View style={styles.rowText}>
            <Text
              variant="bodyMedium"
              style={{color: theme.colors.onSurfaceVariant}}>
              Cycle 완료 알림
            </Text>
          </View>
          <Switch value={false} disabled />
        </View>
        <Divider style={{backgroundColor: theme.colors.outline}} />
        <View style={[styles.row, {paddingVertical: 10}]}>
          <MaterialCommunityIcons
            name="information-outline"
            size={18}
            color={theme.colors.onSurfaceVariant}
          />
          <Text variant="bodySmall" style={subStyle}>
            푸시 알림은 추후 지원 예정입니다.
          </Text>
        </View>
      </View>

      {/* Account */}
      <Text
        variant="labelLarge"
        style={[styles.sectionHeader, subStyle]}>
        계정
      </Text>
      <View style={sectionStyle}>
        <SettingsButton
          icon="logout"
          label="로그아웃"
          theme={theme}
          onPress={handleLogout}
          destructive
        />
      </View>

      {/* App Info */}
      <Text
        variant="labelLarge"
        style={[styles.sectionHeader, subStyle]}>
        앱 정보
      </Text>
      <View style={sectionStyle}>
        <View style={styles.row}>
          <MaterialCommunityIcons
            name="cellphone"
            size={22}
            color={theme.colors.primary}
          />
          <View style={styles.rowText}>
            <Text variant="bodyMedium" style={labelStyle}>
              앱 버전
            </Text>
          </View>
          <Text variant="bodySmall" style={subStyle}>
            0.0.1
          </Text>
        </View>
        <Divider style={{backgroundColor: theme.colors.outline}} />
        <View style={styles.row}>
          <MaterialCommunityIcons
            name="server"
            size={22}
            color={theme.colors.primary}
          />
          <View style={styles.rowText}>
            <Text variant="bodyMedium" style={labelStyle}>
              서버 버전
            </Text>
          </View>
          <Text variant="bodySmall" style={subStyle}>
            {healthData?.version ?? '-'}
          </Text>
        </View>
      </View>

      <View style={styles.footer} />
    </ScrollView>
  );
}

function SettingsButton({
  icon,
  label,
  theme,
  onPress,
  loading,
  destructive,
}: {
  icon: string;
  label: string;
  theme: any;
  onPress: () => void;
  loading?: boolean;
  destructive?: boolean;
}) {
  const color = destructive ? '#dc2626' : theme.colors.primary;
  return (
    <TouchableOpacity
      activeOpacity={0.7}
      onPress={onPress}
      disabled={loading}
      style={styles.row}>
      {loading ? (
        <ActivityIndicator size={22} color={color} />
      ) : (
        <MaterialCommunityIcons name={icon} size={22} color={color} />
      )}
      <Text variant="bodyMedium" style={{flex: 1, color}}>
        {label}
      </Text>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    paddingHorizontal: 16,
    paddingTop: 8,
  },
  sectionHeader: {
    paddingHorizontal: 4,
    paddingTop: 20,
    paddingBottom: 8,
  },
  section: {
    borderRadius: 12,
    overflow: 'hidden',
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 14,
    paddingVertical: 14,
    gap: 12,
    minHeight: 48,
  },
  rowText: {
    flex: 1,
    gap: 2,
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 10,
  },
  footer: {
    height: 40,
  },
});
