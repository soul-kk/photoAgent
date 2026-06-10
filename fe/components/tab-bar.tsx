import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { useRouter, usePathname } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';

type Tab = {
  name: string;
  label: string;
  href: string;
  icon: keyof typeof Ionicons.glyphMap;
  iconActive: keyof typeof Ionicons.glyphMap;
};

const TABS: Tab[] = [
  { name: 'index', label: '首页', href: '/', icon: 'home-outline', iconActive: 'home' },
  { name: 'analysis', label: '分析', href: '/analysis', icon: 'scan-outline', iconActive: 'scan' },
  { name: 'advice', label: '拍摄', href: '/advice', icon: 'camera-outline', iconActive: 'camera' },
  { name: 'profile', label: '我的', href: '/profile', icon: 'person-outline', iconActive: 'person' },
];

export function TabBar() {
  const router = useRouter();
  const pathname = usePathname();

  return (
    <View style={styles.container}>
      <View style={styles.pill}>
        {TABS.map((tab) => {
          const isActive =
            tab.href === '/' ? pathname === '/' : pathname.startsWith(tab.href);
          return (
            <TouchableOpacity
              key={tab.name}
              style={[styles.tab, isActive && styles.tabActive]}
              onPress={() => router.push(tab.href as any)}
              activeOpacity={0.8}
            >
              <Ionicons
                name={isActive ? tab.iconActive : tab.icon}
                size={18}
                color={isActive ? '#0A0A0A' : '#555555'}
              />
              <Text style={[styles.label, isActive && styles.labelActive]}>
                {tab.label}
              </Text>
            </TouchableOpacity>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    height: 82,
    paddingHorizontal: 21,
    paddingTop: 12,
    paddingBottom: 21,
    backgroundColor: '#0A0A0A',
  },
  pill: {
    flex: 1,
    flexDirection: 'row',
    backgroundColor: '#141414',
    borderRadius: 36,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 4,
  },
  tab: {
    flex: 1,
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 4,
    borderRadius: 26,
  },
  tabActive: {
    backgroundColor: '#FFFFFF',
  },
  label: {
    fontFamily: 'DMMono_500Medium',
    fontSize: 10,
    letterSpacing: 1,
    color: '#555555',
  },
  labelActive: {
    color: '#0A0A0A',
  },
});
