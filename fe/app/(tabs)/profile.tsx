import React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  Alert,
  ScrollView,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '@/context/auth';
import { TabBar } from '@/components/tab-bar';

export default function ProfileScreen() {
  const router = useRouter();
  const { user, logout } = useAuth();

  const initial = user?.username?.[0]?.toUpperCase() ?? '?';

  function handleLogout() {
    Alert.alert('退出登录', '确定要退出登录吗？', [
      { text: '取消', style: 'cancel' },
      {
        text: '退出',
        style: 'destructive',
        onPress: async () => {
          await logout();
          router.replace('/auth');
        },
      },
    ]);
  }

  return (
    <SafeAreaView style={styles.root}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>我的</Text>
      </View>

      {/* Avatar + Info */}
      <View style={styles.profileSection}>
        <View style={styles.avatar}>
          <Text style={styles.avatarText}>{initial}</Text>
        </View>
        <Text style={styles.username}>{user?.username ?? '—'}</Text>
        <Text style={styles.email}>{user?.email ?? '—'}</Text>
      </View>

      <View style={styles.divider} />

      {/* Info rows */}
      <View style={styles.infoSection}>
        <View style={styles.infoRow}>
          <View style={styles.infoRowLeft}>
            <Ionicons name="person-outline" size={16} color="#555555" />
            <Text style={styles.infoLabel}>用户名</Text>
          </View>
          <Text style={styles.infoValue}>{user?.username ?? '—'}</Text>
        </View>
        <View style={styles.separator} />
        <View style={styles.infoRow}>
          <View style={styles.infoRowLeft}>
            <Ionicons name="mail-outline" size={16} color="#555555" />
            <Text style={styles.infoLabel}>邮箱</Text>
          </View>
          <Text style={styles.infoValue}>{user?.email ?? '—'}</Text>
        </View>
        <View style={styles.separator} />
        <View style={styles.infoRow}>
          <View style={styles.infoRowLeft}>
            <Ionicons name="shield-outline" size={16} color="#555555" />
            <Text style={styles.infoLabel}>角色</Text>
          </View>
          <Text style={styles.infoValue}>{user?.role ?? '—'}</Text>
        </View>
      </View>

      {/* Action cards */}
      <TouchableOpacity
        style={styles.actionCard}
        onPress={() => router.push('/analysis-history' as any)}
        activeOpacity={0.8}
      >
        <View style={styles.actionLeft}>
          <Ionicons name="time-outline" size={16} color="#AAAAAA" />
          <Text style={styles.actionText}>AI 摄影分析记录</Text>
        </View>
        <Ionicons name="chevron-forward" size={16} color="#444444" />
      </TouchableOpacity>

      {/* Logout */}
      <TouchableOpacity style={styles.logoutCard} onPress={handleLogout} activeOpacity={0.8}>
        <Ionicons name="log-out-outline" size={16} color="#FF5555" />
        <Text style={styles.logoutText}>退出登录</Text>
      </TouchableOpacity>

      <TabBar />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0A0A0A' },
  header: {
    height: 56,
    paddingHorizontal: 24,
    justifyContent: 'center',
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  headerTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 18,
    color: '#FFFFFF',
  },
  profileSection: {
    alignItems: 'center',
    paddingTop: 40,
    paddingBottom: 36,
    gap: 10,
  },
  avatar: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: '#1E1E1E',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 4,
  },
  avatarText: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 28,
    color: '#FFFFFF',
  },
  username: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 22,
    color: '#FFFFFF',
  },
  email: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 13,
    color: '#555555',
  },
  divider: { height: 1, backgroundColor: '#2A2A2A', marginHorizontal: 24 },
  infoSection: {
    marginHorizontal: 24,
    marginTop: 24,
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 18,
    paddingVertical: 16,
  },
  infoRowLeft: { flexDirection: 'row', alignItems: 'center', gap: 10 },
  infoLabel: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 14,
    color: '#FFFFFF',
  },
  infoValue: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 12,
    color: '#555555',
  },
  separator: { height: 1, backgroundColor: '#2A2A2A', marginHorizontal: 18 },
  actionCard: {
    marginHorizontal: 24,
    marginTop: 16,
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 18,
    paddingVertical: 18,
  },
  actionLeft: { flexDirection: 'row', alignItems: 'center', gap: 10 },
  actionText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 14,
    color: '#FFFFFF',
  },
  logoutCard: {
    marginHorizontal: 24,
    marginTop: 16,
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
    paddingVertical: 18,
  },
  logoutText: {
    fontFamily: 'DMSans_500Medium',
    fontSize: 15,
    color: '#FF5555',
  },
});
